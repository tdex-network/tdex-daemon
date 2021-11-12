package application

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/transactionutil"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/transaction"
)

var (
	// ErrWalletNotFunded ...
	ErrWalletNotFunded = fmt.Errorf("wallet not funded")
)

type WalletService interface {
	GenerateAddressAndBlindingKey(
		ctx context.Context,
	) (address string, blindingKey string, err error)
	GetWalletBalance(
		ctx context.Context,
	) (map[string]BalanceInfo, error)
	SendToMany(
		ctx context.Context,
		req SendToManyRequest,
	) ([]byte, []byte, error)
}

type walletService struct {
	repoManager        ports.RepoManager
	explorerService    explorer.Service
	blockchainListener BlockchainListener
	network            *network.Network
	marketFee          int64

	lock   *sync.RWMutex
	pwChan chan PassphraseMsg
}

func NewWalletService(
	repoManager ports.RepoManager,
	explorerService explorer.Service,
	blockchainListener BlockchainListener,
	net *network.Network,
	marketFee int64,
) WalletService {
	return newWalletService(
		repoManager,
		explorerService,
		blockchainListener,
		net,
		marketFee,
	)
}

func newWalletService(
	repoManager ports.RepoManager,
	explorerService explorer.Service,
	blockchainListener BlockchainListener,
	net *network.Network,
	marketFee int64,
) *walletService {
	return &walletService{
		repoManager:        repoManager,
		explorerService:    explorerService,
		blockchainListener: blockchainListener,
		network:            net,
		marketFee:          marketFee,
		lock:               &sync.RWMutex{},
		pwChan:             make(chan PassphraseMsg, 1),
	}
}

func (w *walletService) GenerateAddressAndBlindingKey(
	ctx context.Context,
) (address string, blindingKey string, err error) {
	err = w.repoManager.VaultRepository().UpdateVault(
		ctx,
		func(v *domain.Vault) (*domain.Vault, error) {
			info, err := v.DeriveNextExternalAddressForAccount(
				domain.WalletAccount,
			)
			if err != nil {
				return nil, err
			}

			address = info.Address
			blindingKey = hex.EncodeToString(info.BlindingKey)

			return v, nil
		},
	)

	return
}

func (w *walletService) GetWalletBalance(
	ctx context.Context,
) (map[string]BalanceInfo, error) {
	return getAccountBalanceFromExplorer(
		w.repoManager, w.explorerService, ctx, domain.WalletAccount,
	)
}

type SendToManyRequest struct {
	Outputs         []TxOut
	MillisatPerByte int64
	Push            bool
}

func (w *walletService) SendToMany(
	ctx context.Context,
	req SendToManyRequest,
) ([]byte, []byte, error) {
	outputs, outputsBlindingKeys, err := parseRequestOutputs(req.Outputs)
	if err != nil {
		return nil, nil, err
	}

	vault, err := w.repoManager.VaultRepository().GetOrCreateVault(ctx, nil, "", nil)
	if err != nil {
		return nil, nil, err
	}

	walletUnspents, err := w.getAllUnspentsForAccount(ctx, domain.WalletAccount, true)
	if err != nil {
		return nil, nil, err
	}

	if len(walletUnspents) <= 0 {
		return nil, nil, ErrWalletNotFunded
	}

	feeUnspents, err := w.getAllUnspentsForAccount(ctx, domain.FeeAccount, false)
	if err != nil {
		return nil, nil, err
	}

	if len(feeUnspents) <= 0 {
		return nil, nil, ErrFeeAccountNotFunded
	}

	mnemonic, err := vault.GetMnemonicSafe()
	if err != nil {
		return nil, nil, err
	}
	walletAccount, err := vault.AccountByIndex(domain.WalletAccount)
	if err != nil {
		return nil, nil, err
	}
	feeAccount, err := vault.AccountByIndex(domain.FeeAccount)
	if err != nil {
		return nil, nil, err
	}

	changePathsByAsset := map[string]string{}
	feeChangePathByAsset := map[string]string{}
	for _, asset := range getAssetsOfOutputs(outputs) {
		info, err := vault.DeriveNextInternalAddressForAccount(domain.WalletAccount)
		if err != nil {
			return nil, nil, err
		}
		changePathsByAsset[asset] = info.DerivationPath
	}
	feeInfo, err := vault.DeriveNextInternalAddressForAccount(domain.FeeAccount)
	if err != nil {
		return nil, nil, err
	}
	feeChangePathByAsset[w.network.AssetID] = feeInfo.DerivationPath

	txHex, err := sendToManyWithFeeTopup(sendToManyWithFeeTopupOpts{
		mnemonic:              mnemonic,
		unspents:              walletUnspents,
		feeUnspents:           feeUnspents,
		outputs:               outputs,
		outputsBlindingKeys:   outputsBlindingKeys,
		changePathsByAsset:    changePathsByAsset,
		feeChangePathByAsset:  feeChangePathByAsset,
		inputPathsByScript:    walletAccount.DerivationPathByScript,
		feeInputPathsByScript: feeAccount.DerivationPathByScript,
		milliSatPerByte:       int(req.MillisatPerByte),
		network:               w.network,
	})
	if err != nil {
		return nil, nil, err
	}

	txid, _ := transactionutil.GetTxIdFromHex(txHex)
	if req.Push {
		txid, err = w.explorerService.BroadcastTransaction(txHex)
		if err != nil {
			return nil, nil, err
		}
		log.Debugf("wallet account tx broadcasted with id: %s", txid)
	}

	if err := w.repoManager.VaultRepository().UpdateVault(
		ctx, func(_ *domain.Vault) (*domain.Vault, error) {
			return vault, nil
		},
	); err != nil {
		return nil, nil, err
	}

	go extractUnspentsFromTxAndUpdateUtxoSet(
		w.repoManager.UnspentRepository(),
		w.repoManager.VaultRepository(),
		w.network,
		txHex,
		domain.FeeAccount,
	)

	go w.blockchainListener.StartObserveTx(txid, "")

	rawTx, _ := hex.DecodeString(txHex)
	rawTxid, _ := hex.DecodeString(txid)
	return rawTx, rawTxid, nil
}

func (w *walletService) getAllUnspentsForAccount(
	ctx context.Context,
	accountIndex int,
	useExplorer bool,
) ([]explorer.Utxo, error) {
	info, err := w.repoManager.VaultRepository().GetAllDerivedAddressesInfoForAccount(ctx, accountIndex)
	if err != nil {
		return nil, err
	}
	addresses, blindingKeys := info.AddressesAndKeys()

	if useExplorer {
		return w.explorerService.GetUnspentsForAddresses(addresses, blindingKeys)
	}

	unspents, err := w.repoManager.UnspentRepository().GetAvailableUnspentsForAddresses(ctx, addresses)
	if err != nil {
		return nil, err
	}

	utxos := make([]explorer.Utxo, 0, len(unspents))
	for _, u := range unspents {
		utxos = append(utxos, u.ToUtxo())
	}
	return utxos, nil
}

func parseRequestOutputs(
	reqOutputs []TxOut,
) ([]*transaction.TxOutput, [][]byte, error) {
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
	script, err := address.ToOutputScript(addr)
	if err != nil {
		return nil, nil, err
	}
	ctAddr, err := address.FromConfidential(addr)
	if err != nil {
		return nil, nil, err
	}
	return script, ctAddr.BlindingKey, nil
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

type sendToManyWithFeeTopupOpts struct {
	mnemonic              []string
	unspents              []explorer.Utxo
	feeUnspents           []explorer.Utxo
	outputs               []*transaction.TxOutput
	outputsBlindingKeys   [][]byte
	changePathsByAsset    map[string]string
	feeChangePathByAsset  map[string]string
	inputPathsByScript    map[string]string
	feeInputPathsByScript map[string]string
	milliSatPerByte       int
	network               *network.Network
}

func sendToManyWithFeeTopup(opts sendToManyWithFeeTopupOpts) (string, error) {
	w, err := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: opts.mnemonic,
	})
	if err != nil {
		return "", err
	}

	// default to MinMilliSatPerByte if needed
	milliSatPerByte := opts.milliSatPerByte
	if milliSatPerByte < domain.MinMilliSatPerByte {
		milliSatPerByte = domain.MinMilliSatPerByte
	}

	// create the transaction
	newPset, err := w.CreateTx()
	if err != nil {
		return "", err
	}
	network := opts.network

	// add inputs and outputs
	updateResult, err := w.UpdateTx(wallet.UpdateTxOpts{
		PsetBase64:         newPset,
		Unspents:           opts.unspents,
		Outputs:            opts.outputs,
		ChangePathsByAsset: opts.changePathsByAsset,
		MilliSatsPerBytes:  milliSatPerByte,
		Network:            network,
	})
	if err != nil {
		return "", err
	}

	inputBlindingData := make(map[int]wallet.BlindingData)
	index := 0
	for _, v := range updateResult.SelectedUnspents {
		inputBlindingData[index] = wallet.BlindingData{
			Asset:         v.Asset(),
			Amount:        v.Value(),
			AssetBlinder:  v.AssetBlinder(),
			AmountBlinder: v.ValueBlinder(),
		}
		index++
	}

	// update the list of output blinding keys with those of the eventual changes
	outputsBlindingKeys := opts.outputsBlindingKeys
	for _, v := range updateResult.ChangeOutputsBlindingKeys {
		outputsBlindingKeys = append(outputsBlindingKeys, v)
	}

	// add inputs for paying network fees
	feeUpdateResult, err := w.UpdateTx(wallet.UpdateTxOpts{
		PsetBase64:         updateResult.PsetBase64,
		Unspents:           opts.feeUnspents,
		ChangePathsByAsset: opts.feeChangePathByAsset,
		MilliSatsPerBytes:  milliSatPerByte,
		Network:            network,
		WantChangeForFees:  true,
	})
	if err != nil {
		return "", err
	}

	for _, v := range feeUpdateResult.SelectedUnspents {
		inputBlindingData[index] = wallet.BlindingData{
			Asset:         v.Asset(),
			Amount:        v.Value(),
			AssetBlinder:  v.AssetBlinder(),
			AmountBlinder: v.ValueBlinder(),
		}
		index++
	}

	// again, add changes' blinding keys to the list of those of the outputs
	for _, v := range feeUpdateResult.ChangeOutputsBlindingKeys {
		outputsBlindingKeys = append(outputsBlindingKeys, v)
	}

	// blind the transaction
	blindedPset, err := w.BlindTransactionWithData(
		wallet.BlindTransactionWithDataOpts{
			PsetBase64:         feeUpdateResult.PsetBase64,
			InputBlindingData:  inputBlindingData,
			OutputBlindingKeys: outputsBlindingKeys,
		},
	)
	if err != nil {
		return "", err
	}

	// add the explicit fee amount
	blindedPlusFees, err := w.UpdateTx(wallet.UpdateTxOpts{
		PsetBase64: blindedPset,
		Outputs:    transactionutil.NewFeeOutput(feeUpdateResult.FeeAmount, network),
		Network:    network,
	})
	if err != nil {
		return "", err
	}

	// sign the inputs
	inputPathsByScript := mergeDerivationPaths(opts.inputPathsByScript, opts.feeInputPathsByScript)
	signedPset, err := w.SignTransaction(wallet.SignTransactionOpts{
		PsetBase64:        blindedPlusFees.PsetBase64,
		DerivationPathMap: inputPathsByScript,
	})
	if err != nil {
		return "", err
	}

	// finalize, extract and return the transaction
	txHex, _, err := wallet.FinalizeAndExtractTransaction(
		wallet.FinalizeAndExtractTransactionOpts{
			PsetBase64: signedPset,
		},
	)

	return txHex, err
}
