package application

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/transactionutil"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/transaction"
)

type WalletService interface {
	GenSeed(ctx context.Context) ([]string, error)
	InitWallet(
		ctx context.Context,
		mnemonic []string,
		passphrase string,
	) error
	UnlockWallet(
		ctx context.Context,
		passphrase string,
	) error
	ChangePassword(
		ctx context.Context,
		currentPassphrase string,
		newPassphrase string) error
	GenerateAddressAndBlindingKey(
		ctx context.Context,
	) (address string, blindingKey string, err error)
	GetWalletBalance(
		ctx context.Context,
	) (map[string]domain.BalanceInfo, error)
	SendToMany(
		ctx context.Context,
		req SendToManyRequest,
	) ([]byte, error)
}

type walletService struct {
	vaultRepository   domain.VaultRepository
	unspentRepository domain.UnspentRepository
	explorerService   explorer.Service
}

func NewWalletService(
	vaultRepository domain.VaultRepository,
	unspentRepository domain.UnspentRepository,
	explorerService explorer.Service,
) WalletService {
	return &walletService{
		vaultRepository:   vaultRepository,
		unspentRepository: unspentRepository,
		explorerService:   explorerService,
	}
}

func (w *walletService) GenSeed(ctx context.Context) ([]string, error) {
	mnemonic, err := wallet.NewMnemonic(wallet.NewMnemonicOpts{EntropySize: 256})
	if err != nil {
		return nil, err
	}
	return mnemonic, nil
}

func (w *walletService) InitWallet(
	ctx context.Context,
	mnemonic []string,
	passphrase string,
) error {
	//validate mnemonic
	return w.vaultRepository.UpdateVault(
		ctx,
		mnemonic,
		passphrase,
		func(v *domain.Vault) (*domain.Vault, error) {
			v.InitAccount(domain.FeeAccount)
			v.InitAccount(domain.WalletAccount)
			return v, nil
		},
	)
}

func (w *walletService) UnlockWallet(
	ctx context.Context,
	passphrase string,
) error {

	return w.vaultRepository.UpdateVault(
		ctx,
		nil,
		"",
		func(v *domain.Vault) (*domain.Vault, error) {
			if err := v.Unlock(passphrase); err != nil {
				return nil, err
			}
			return v, nil
		},
	)
}

func (w *walletService) ChangePassword(
	ctx context.Context,
	currentPassphrase string,
	newPassphrase string) error {

	return w.vaultRepository.UpdateVault(
		ctx,
		nil,
		"",
		func(v *domain.Vault) (*domain.Vault, error) {
			err := v.ChangePassphrase(currentPassphrase, newPassphrase)
			if err != nil {
				return nil, err
			}
			return v, nil
		},
	)
}

func (w *walletService) GenerateAddressAndBlindingKey(
	ctx context.Context,
) (address string, blindingKey string, err error) {

	err = w.vaultRepository.UpdateVault(
		ctx,
		nil,
		"",
		func(v *domain.Vault) (*domain.Vault, error) {
			adr, _, bk, err1 := v.DeriveNextExternalAddressForAccount(
				domain.WalletAccount,
			)
			if err1 != nil {
				err = err1
				return nil, err1
			}

			address = adr
			blindingKey = hex.EncodeToString(bk)

			return v, nil
		},
	)

	return
}

func (w *walletService) GetWalletBalance(
	ctx context.Context,
) (map[string]domain.BalanceInfo, error) {

	derivedAddresses, _, err := w.vaultRepository.
		GetAllDerivedAddressesAndBlindingKeysForAccount(ctx, domain.WalletAccount)
	if err != nil {
		return nil, err
	}

	unspents := w.unspentRepository.GetAvailableUnspentsForAddresses(
		ctx,
		derivedAddresses,
	)

	result := w.unspentRepository.GetBalanceInfoForAsset(unspents)

	return result, nil
}

type SendToManyRequest struct {
	Outputs         []TxOut
	MillisatPerByte int64
	Push            bool
}

type TxOut struct {
	Asset   string
	Value   int64
	Address string
}

func (w *walletService) SendToMany(
	ctx context.Context,
	req SendToManyRequest,
) ([]byte, error) {
	var rawTx []byte

	outputs, outputsBlindingKeys, err := parseRequestOutputs(req.Outputs)
	if err != nil {
		return nil, err
	}

	derivedAddresses, blindingKeys, err := w.vaultRepository.
		GetAllDerivedAddressesAndBlindingKeysForAccount(ctx, domain.WalletAccount)
	if err != nil {
		return nil, err
	}

	unspents, err := w.explorerService.GetUnspentsForAddresses(
		derivedAddresses,
		blindingKeys,
	)
	if err != nil {
		return nil, err
	}

	if len(unspents) <= 0 {
		return nil, errors.New("wallet not funded")
	}

	err = w.vaultRepository.UpdateVault(
		ctx,
		nil,
		"",
		func(v *domain.Vault) (*domain.Vault, error) {
			mnemonic, err := v.Mnemonic()
			if err != nil {
				return nil, err
			}
			account, err := v.AccountByIndex(domain.WalletAccount)
			if err != nil {
				return nil, err
			}

			changePathsByAsset := map[string]string{}
			for _, asset := range getAssetsOfOutputs(outputs) {
				_, script, _, err := v.DeriveNextInternalAddressForAccount(
					domain.WalletAccount,
				)
				if err != nil {
					return nil, err
				}
				derivationPath, _ := account.DerivationPathByScript(script)
				changePathsByAsset[asset] = derivationPath
			}

			txHex, _, err := sendToMany(
				mnemonic,
				account,
				unspents,
				outputs,
				outputsBlindingKeys,
				int(req.MillisatPerByte),
				changePathsByAsset,
			)
			if err != nil {
				return nil, err
			}

			if req.Push {
				if _, err := w.explorerService.BroadcastTransaction(
					txHex); err != nil {
					return nil, err
				}
			}

			tx, err := hex.DecodeString(txHex)
			if err != nil {
				return nil, err
			}
			rawTx = tx

			return v, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return rawTx, nil
}

func parseRequestOutputs(reqOutputs []TxOut) ([]*transaction.TxOutput,
	[][]byte, error) {
	outputs := make([]*transaction.TxOutput, 0, len(reqOutputs))
	blindingKeys := make([][]byte, 0, len(reqOutputs))

	for _, out := range reqOutputs {
		asset, err := bufferutil.AssetHashToBytes(out.Asset)
		if err != nil {
			return nil, nil, err
		}
		value, err := bufferutil.ValueToBytes(uint64(out.Value))
		if err != nil {
			return nil, nil, err
		}
		script, blindingKey, err := parseConfidentialAddress(out.Address)
		if err != nil {
			return nil, nil, err
		}

		output := transaction.NewTxOutput(asset, value, script)
		outputs = append(outputs, output)
		blindingKeys = append(blindingKeys, blindingKey)
	}
	return outputs, blindingKeys, nil
}

func parseConfidentialAddress(addr string) ([]byte, []byte, error) {
	script, err := address.ToOutputScript(addr, *config.GetNetwork())
	if err != nil {
		return nil, nil, err
	}
	blindingKey, err := extractBlindingKey(addr, script)
	if err != nil {
		return nil, nil, err
	}
	return script, blindingKey, nil
}

func extractBlindingKey(addr string, script []byte) ([]byte, error) {
	addrType, _ := address.DecodeType(addr, *config.GetNetwork())
	switch addrType {
	case address.ConfidentialP2Pkh, address.ConfidentialP2Sh:
		decoded, _ := address.FromBase58(addr)
		return decoded.Data[1:34], nil
	case address.ConfidentialP2Wpkh, address.ConfidentialP2Wsh:
		decoded, _ := address.FromBlech32(addr)
		return decoded.PublicKey, nil
	default:
		return nil, fmt.Errorf("failed to extract blinding key from address '%s'", addr)
	}
}

func getAssetsOfOutputs(outputs []*transaction.TxOutput) []string {
	assets := make([]string, 0)
	for _, out := range outputs {
		asset := bufferutil.AssetHashFromBytes(out.Asset)
		if !containsAsset(assets, asset) {
			assets = append(assets, asset)
		}
	}
	return assets
}

func containsAsset(assets []string, asset string) bool {
	for _, a := range assets {
		if a == asset {
			return true
		}
	}
	return false
}

func sendToMany(
	mnemonic []string,
	account *domain.Account,
	unspents []explorer.Utxo,
	outputs []*transaction.TxOutput,
	outputsBlindingKeys [][]byte,
	milliSatsPerBytes int,
	changePathsByAsset map[string]string,
) (string, string, error) {
	w, err := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: mnemonic,
	})
	if err != nil {
		return "", "", err
	}

	if milliSatsPerBytes < domain.MinMilliSatPerByte {
		milliSatsPerBytes = domain.MinMilliSatPerByte
	}

	newPset, err := w.CreateTx()
	if err != nil {
		return "", "", err
	}
	updateResult, err := w.UpdateTx(wallet.UpdateTxOpts{
		PsetBase64:         newPset,
		Unspents:           unspents,
		Outputs:            outputs,
		ChangePathsByAsset: changePathsByAsset,
		MilliSatsPerBytes:  milliSatsPerBytes,
		Network:            config.GetNetwork(),
	})
	if err != nil {
		return "", "", err
	}

	changeOutputsBlindingKeys := make([][]byte, 0, len(updateResult.ChangeOutputsBlindingKeys))
	for _, v := range updateResult.ChangeOutputsBlindingKeys {
		changeOutputsBlindingKeys = append(changeOutputsBlindingKeys, v)
	}
	outputsPlusChangesBlindingKeys := append(
		outputsBlindingKeys,
		changeOutputsBlindingKeys...,
	)
	blindedPset, err := w.BlindTransaction(wallet.BlindTransactionOpts{
		PsetBase64:         updateResult.PsetBase64,
		OutputBlindingKeys: outputsPlusChangesBlindingKeys,
	})
	if err != nil {
		return "", "", err
	}
	blindedPlusFees, err := w.UpdateTx(wallet.UpdateTxOpts{
		PsetBase64: blindedPset,
		Outputs:    transactionutil.NewFeeOutput(updateResult.FeeAmount),
		Network:    config.GetNetwork(),
	})
	if err != nil {
		return "", "", err
	}

	inputPathsByScript := getDerivationPathsForUnspents(account, unspents)
	signedPset, err := w.SignTransaction(wallet.SignTransactionOpts{
		PsetBase64:        blindedPlusFees.PsetBase64,
		DerivationPathMap: inputPathsByScript,
	})
	if err != nil {
		return "", "", err
	}

	return wallet.FinalizeAndExtractTransaction(
		wallet.FinalizeAndExtractTransactionOpts{
			PsetBase64: signedPset,
		},
	)
}

func getDerivationPathsForUnspents(
	account *domain.Account,
	unspents []explorer.Utxo,
) map[string]string {
	paths := map[string]string{}
	for _, unspent := range unspents {
		script := hex.EncodeToString(unspent.Script())
		derivationPath, _ := account.DerivationPathByScript(script)
		paths[script] = derivationPath
	}
	return paths
}
