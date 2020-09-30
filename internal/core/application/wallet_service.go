package application

import (
	"context"
	"encoding/hex"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/transactionutil"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/transaction"
)

var (
	// ErrWalletNotFunded ...
	ErrWalletNotFunded = fmt.Errorf("wallet not funded")
	// ErrWalletIsSyncing ...
	ErrWalletIsSyncing = fmt.Errorf(
		"wallet is syncing data from blockchain. All functionalities are " +
			"disabled until this operation is completed",
	)
	// ErrWalletNotInitialized ...
	ErrWalletNotInitialized = fmt.Errorf("wallet not initialized")
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
	crawlerService    crawler.Service
	explorerService   explorer.Service
	walletInitialized bool
	walletIsSyncing   bool
}

func NewWalletService(
	vaultRepository domain.VaultRepository,
	unspentRepository domain.UnspentRepository,
	crawlerService crawler.Service,
	explorerService explorer.Service,
) WalletService {
	w := &walletService{
		vaultRepository:   vaultRepository,
		unspentRepository: unspentRepository,
		crawlerService:    crawlerService,
		explorerService:   explorerService,
	}
	// to understand if the service has an already initialized wallet we check
	// if the inner vaultRepo is able to return a Vault without passing mnemonic
	// and passphrase. If it does, it means it's been retrieved from storage,
	// hence we mark the wallet as initialized
	if _, err := w.vaultRepository.GetOrCreateVault(
		context.Background(), nil, "",
	); err == nil {
		w.walletInitialized = true
	}
	return w
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
	if w.walletInitialized {
		return nil
	}
	// this prevents strange behaviors by making consecutive calls to InitWallet
	// while it's still syncing
	if w.walletIsSyncing {
		return nil
	}

	w.walletIsSyncing = true

	err := w.vaultRepository.UpdateVault(
		ctx,
		mnemonic,
		passphrase,
		func(v *domain.Vault) (*domain.Vault, error) {
			log.Debug("start syncing wallet")
			ww, err := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
				SigningMnemonic: mnemonic,
			})
			if err != nil {
				return nil, err
			}

			feeLastDerivedIndex := getLatestDerivationIndexForAccount(ww, domain.FeeAccount, w.explorerService)
			walletLastDerivedIndex := getLatestDerivationIndexForAccount(ww, domain.WalletAccount, w.explorerService)
			marketsLastDerivedIndex := getLatestDerivationIndexForMarkets(ww, w.explorerService)

			// fmt.Println("feeLastDerivedIndex", feeLastDerivedIndex)
			// fmt.Println("walletLastDerivedIndex", walletLastDerivedIndex)
			// fmt.Println("marketsLastDerivedIndex", len(marketsLastDerivedIndex))
			if err := initVaultAccount(v, domain.FeeAccount, feeLastDerivedIndex, w.crawlerService); err != nil {
				return nil, err
			}
			// we dont't want to let the crawler watch for WalletAccount addresses
			if err := initVaultAccount(v, domain.WalletAccount, walletLastDerivedIndex, nil); err != nil {
				return nil, err
			}
			for i, m := range marketsLastDerivedIndex {
				if err := initVaultAccount(v, domain.MarketAccountStart+i, m, w.crawlerService); err != nil {
					return nil, err
				}
			}
			v.Lock()
			w.walletInitialized = true
			return v, nil
		},
	)

	w.walletIsSyncing = false
	log.Debug("ended syncing wallet")
	return err
}

func (w *walletService) UnlockWallet(
	ctx context.Context,
	passphrase string,
) error {
	if w.walletIsSyncing {
		return ErrWalletIsSyncing
	}
	if !w.walletInitialized {
		return ErrWalletNotInitialized
	}

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
	if w.walletIsSyncing {
		return ErrWalletIsSyncing
	}
	if !w.walletInitialized {
		return ErrWalletNotInitialized
	}

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
	if w.walletIsSyncing {
		return "", "", ErrWalletIsSyncing
	}
	if !w.walletInitialized {
		return "", "", ErrWalletNotInitialized
	}

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
	if w.walletIsSyncing {
		return nil, ErrWalletIsSyncing
	}
	if !w.walletInitialized {
		return nil, ErrWalletNotInitialized
	}

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
	if w.walletIsSyncing {
		return nil, ErrWalletIsSyncing
	}
	if !w.walletInitialized {
		return nil, ErrWalletNotInitialized
	}

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
		return nil, ErrWalletNotFunded
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

type accountLastDerivedIndex struct {
	external int
	internal int
}

func getLatestDerivationIndexForAccount(w *wallet.Wallet, accountIndex int, explorerSvc explorer.Service) *accountLastDerivedIndex {
	lastDerivedIndex := &accountLastDerivedIndex{}
	for chainIndex := 0; chainIndex <= 1; chainIndex++ {
		firstUnfundedAddress := -1
		unfundedAddressesCounter := 0
		i := 0
		for unfundedAddressesCounter < 20 {
			ctAddress, script, _ := w.DeriveConfidentialAddress(wallet.DeriveConfidentialAddressOpts{
				DerivationPath: fmt.Sprintf("%d'/%d/%d", accountIndex, chainIndex, i),
				Network:        config.GetNetwork(),
			})
			blindingKey, _, _ := w.DeriveBlindingKeyPair(wallet.DeriveBlindingKeyPairOpts{
				Script: script,
			})

			if !isAddressFunded(ctAddress, blindingKey.Serialize(), explorerSvc) {
				if firstUnfundedAddress < 0 {
					firstUnfundedAddress = i
				}
				unfundedAddressesCounter++
			} else {
				if firstUnfundedAddress >= 0 {
					firstUnfundedAddress = -1
					unfundedAddressesCounter = 0
				}
			}
			i++
		}
		if chainIndex == 0 {
			lastDerivedIndex.external = firstUnfundedAddress - 1
		} else {
			lastDerivedIndex.internal = firstUnfundedAddress - 1
		}
	}

	if lastDerivedIndex.external < 0 && lastDerivedIndex.internal < 0 {
		log.Debugf("account %d empty", accountIndex)
		return nil
	}
	log.Debugf("account %d last derived external address %d", accountIndex, lastDerivedIndex.external)
	return lastDerivedIndex
}

func getLatestDerivationIndexForMarkets(w *wallet.Wallet, explorerSvc explorer.Service) []*accountLastDerivedIndex {
	marketsLastIndex := make([]*accountLastDerivedIndex, 0)
	i := 0
	for {
		marketIndex := domain.MarketAccountStart + i
		lastDerivedIndex := getLatestDerivationIndexForAccount(w, marketIndex, explorerSvc)
		if lastDerivedIndex == nil {
			// fmt.Println("breaked loop at index", i)
			break
		}
		marketsLastIndex = append(marketsLastIndex, lastDerivedIndex)
		i++
	}
	return marketsLastIndex
}

func initVaultAccount(v *domain.Vault, accountIndex int, lastDerivedIndex *accountLastDerivedIndex, crawlerSvc crawler.Service) error {
	if lastDerivedIndex == nil {
		v.InitAccount(accountIndex)
		return nil
	}

	for i := 0; i <= lastDerivedIndex.external; i++ {
		addr, _, blindingKey, err := v.DeriveNextExternalAddressForAccount(accountIndex)
		if err != nil {
			return err
		}
		if crawlerSvc != nil {
			crawlerSvc.AddObservable(&crawler.AddressObservable{
				AccountIndex: accountIndex,
				Address:      addr,
				BlindingKey:  blindingKey,
			})
		}
	}
	for i := 0; i <= lastDerivedIndex.internal; i++ {
		addr, _, blindingKey, err := v.DeriveNextInternalAddressForAccount(accountIndex)
		if err != nil {
			return err
		}
		if crawlerSvc != nil {
			crawlerSvc.AddObservable(&crawler.AddressObservable{
				AccountIndex: accountIndex,
				Address:      addr,
				BlindingKey:  blindingKey,
			})
		}
	}
	return nil
}

func isAddressFunded(addr string, blindingKey []byte, explorerSvc explorer.Service) bool {
	unspents, err := explorerSvc.GetUnspentsForAddresses([]string{addr}, [][]byte{blindingKey})
	if err != nil {
		// should we retry?
		return false
	}
	return len(unspents) > 0
}
