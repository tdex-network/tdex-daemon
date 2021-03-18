package application

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/transactionutil"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
	"github.com/vulpemventures/go-elements/address"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/payment"
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
		restore bool,
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
	) (map[string]BalanceInfo, error)
	SendToMany(
		ctx context.Context,
		req SendToManyRequest,
	) ([]byte, error)
}

type walletService struct {
	vaultRepository    domain.VaultRepository
	unspentRepository  domain.UnspentRepository
	explorerService    explorer.Service
	blockchainListener BlockchainListener
	walletInitialized  bool
	walletIsSyncing    bool
	withElements       bool
	network            *network.Network
}

func NewWalletService(
	vaultRepository domain.VaultRepository,
	unspentRepository domain.UnspentRepository,
	explorerService explorer.Service,
	blockchainListener BlockchainListener,
	withElements bool,
	net *network.Network,
) (WalletService, error) {
	return newWalletService(
		vaultRepository,
		unspentRepository,
		explorerService,
		blockchainListener,
		withElements,
		net,
	)
}

func newWalletService(
	vaultRepository domain.VaultRepository,
	unspentRepository domain.UnspentRepository,
	explorerService explorer.Service,
	blockchainListener BlockchainListener,
	withElements bool,
	net *network.Network,
) (*walletService, error) {
	w := &walletService{
		vaultRepository:    vaultRepository,
		unspentRepository:  unspentRepository,
		explorerService:    explorerService,
		blockchainListener: blockchainListener,
		withElements:       withElements,
		network:            net,
	}
	// to understand if the service has an already initialized wallet we check
	// if the inner vaultRepo is able to return a Vault without passing mnemonic
	// and passphrase. If it does, it means it's been retrieved from storage,
	// therefore we let the crawler to start watch all derived addresses and mark
	// the wallet as initialized
	if vault, err := w.vaultRepository.GetOrCreateVault(
		context.Background(), nil, "", nil,
	); err == nil {
		info := vault.AllDerivedAddressesInfo()
		if err := fetchUnspents(
			w.explorerService,
			w.unspentRepository,
			info,
		); err != nil {
			return nil, err
		}
		w.walletInitialized = true
	}
	return w, nil
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
	restore bool,
) error {
	if w.walletInitialized {
		return nil
	}
	// this prevents strange behaviors by making consecutive calls to InitWallet
	// while it's still syncing
	if w.walletIsSyncing {
		return nil
	}

	if restore && w.withElements {
		return fmt.Errorf(
			"Restoring a wallet through the Elements explorer is not availble at the " +
				"moment. Please restart the daemon using the Esplora block explorer.",
		)
	}

	w.walletIsSyncing = true
	info := make(domain.AddressesInfo, 0)

	vault, err := w.vaultRepository.GetOrCreateVault(ctx, mnemonic, passphrase, w.network)
	if err != nil {
		return err
	}
	defer vault.Lock()

	if err := w.vaultRepository.UpdateVault(
		ctx,
		func(v *domain.Vault) (*domain.Vault, error) {
			vault = v

			log.Debug("start syncing wallet")
			ww, err := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
				SigningMnemonic: mnemonic,
			})
			if err != nil {
				return nil, err
			}

			feeLastDerivedIndex := getLatestDerivationIndexForAccount(ww, domain.FeeAccount, w.explorerService, restore, w.network)
			walletLastDerivedIndex := getLatestDerivationIndexForAccount(ww, domain.WalletAccount, w.explorerService, restore, w.network)
			marketsLastDerivedIndex := getLatestDerivationIndexForMarkets(ww, w.explorerService, restore, w.network)

			feeAddresses, err := initVaultAccount(v, domain.FeeAccount, feeLastDerivedIndex)
			if err != nil {
				return nil, err
			}
			// we dont't want to let the crawler watch for WalletAccount addresses
			if _, err := initVaultAccount(v, domain.WalletAccount, walletLastDerivedIndex); err != nil {
				return nil, err
			}
			marketAddresses := make(domain.AddressesInfo, 0)
			for i, m := range marketsLastDerivedIndex {
				addresses, err := initVaultAccount(v, domain.MarketAccountStart+i, m)
				if err != nil {
					return nil, err
				}
				marketAddresses = append(marketAddresses, addresses...)
			}

			info = append(info, feeAddresses...)
			info = append(info, marketAddresses...)
			w.walletInitialized = true
			return v, nil
		},
	); err != nil {
		log.Debug("ended syncing wallet with error")
		w.walletIsSyncing = false
		return err
	}

	w.walletIsSyncing = false
	log.Debug("ended syncing wallet")

	go func() {
		if err := fetchUnspents(
			w.explorerService,
			w.unspentRepository,
			info,
		); err != nil {
			log.Warnf(
				"unexpected error occured while restoring the utxo set. You must "+
					"manually run ReloadUtxo RPC as soon as possible to restore the utxo "+
					"set of the internal wallet. Error: %v", err,
			)
		}
	}()
	return nil
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

	if err := w.vaultRepository.UpdateVault(
		ctx,
		func(v *domain.Vault) (*domain.Vault, error) {
			if err := v.Unlock(passphrase); err != nil {
				return nil, err
			}
			return v, nil
		},
	); err != nil {
		return err
	}

	w.blockchainListener.StartObservation()
	return nil
}

func (w *walletService) ChangePassword(
	ctx context.Context,
	currentPassphrase string,
	newPassphrase string,
) error {
	if w.walletIsSyncing {
		return ErrWalletIsSyncing
	}
	if !w.walletInitialized {
		return ErrWalletNotInitialized
	}

	return w.vaultRepository.UpdateVault(
		ctx,
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
	if w.walletIsSyncing {
		return nil, ErrWalletIsSyncing
	}
	if !w.walletInitialized {
		return nil, ErrWalletNotInitialized
	}

	info, err := w.vaultRepository.GetAllDerivedAddressesInfoForAccount(ctx, domain.WalletAccount)
	if err != nil {
		return nil, err
	}

	unspents, err := w.getUnspents(info.AddressesAndKeys())
	if err != nil {
		return nil, err
	}

	return getBalancesByAsset(unspents), nil
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

	outputs, outputsBlindingKeys, err := parseRequestOutputs(req.Outputs)
	if err != nil {
		return nil, err
	}

	walletUnspents, err := w.getAllUnspentsForAccount(ctx, domain.WalletAccount, true)
	if err != nil {
		return nil, err
	}

	if len(walletUnspents) <= 0 {
		return nil, ErrWalletNotFunded
	}

	feeUnspents, err := w.getAllUnspentsForAccount(ctx, domain.FeeAccount, false)
	if err != nil {
		return nil, err
	}

	if len(feeUnspents) <= 0 {
		return nil, ErrFeeAccountNotFunded
	}

	var selectedUnspentKeys []domain.UnspentKey
	var txHex string

	if err := w.vaultRepository.UpdateVault(
		ctx,
		func(v *domain.Vault) (*domain.Vault, error) {
			mnemonic, err := v.GetMnemonicSafe()
			if err != nil {
				return nil, err
			}
			walletAccount, err := v.AccountByIndex(domain.WalletAccount)
			if err != nil {
				return nil, err
			}
			feeAccount, err := v.AccountByIndex(domain.FeeAccount)
			if err != nil {
				return nil, err
			}

			changePathsByAsset := map[string]string{}
			feeChangePathByAsset := map[string]string{}
			for _, asset := range getAssetsOfOutputs(outputs) {
				info, err := v.DeriveNextInternalAddressForAccount(domain.WalletAccount)
				if err != nil {
					return nil, err
				}
				changePathsByAsset[asset] = info.DerivationPath
			}
			feeInfo, err := v.DeriveNextInternalAddressForAccount(domain.FeeAccount)
			if err != nil {
				return nil, err
			}
			feeChangePathByAsset[w.network.AssetID] = feeInfo.DerivationPath

			_txHex, usedUnspentKeys, err := sendToMany(sendToManyOpts{
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
				return nil, err
			}

			txHex = _txHex
			selectedUnspentKeys = usedUnspentKeys
			if !req.Push {
				return v, nil
			}

			txid, err := w.explorerService.BroadcastTransaction(txHex)
			if err != nil {
				return nil, err
			}
			log.Debugf("wallet account tx broadcasted with id: %s", txid)

			return v, nil
		},
	); err != nil {
		return nil, err
	}

	go spendUnspents(w.unspentRepository, selectedUnspentKeys)
	go extractAndAddUnspentsFromTx(
		w.unspentRepository,
		w.vaultRepository,
		w.network,
		txHex,
		domain.FeeAccount,
	)

	rawTx, _ := hex.DecodeString(txHex)
	return rawTx, nil
}

func (w *walletService) getAllUnspentsForAccount(
	ctx context.Context,
	accountIndex int,
	useExplorer bool,
) ([]explorer.Utxo, error) {
	info, err := w.vaultRepository.GetAllDerivedAddressesInfoForAccount(ctx, accountIndex)
	if err != nil {
		return nil, err
	}
	addresses, blindingKeys := info.AddressesAndKeys()

	if useExplorer {
		return w.explorerService.GetUnspentsForAddresses(addresses, blindingKeys)
	}

	unspents, err := w.unspentRepository.GetAvailableUnspentsForAddresses(ctx, addresses)
	if err != nil {
		return nil, err
	}

	utxos := make([]explorer.Utxo, 0, len(unspents))
	for _, u := range unspents {
		utxos = append(utxos, u.ToUtxo())
	}
	return utxos, nil
}

func (w *walletService) getUnspents(addresses []string, blindingKeys [][]byte) ([]explorer.Utxo, error) {
	chUnspents := make(chan []explorer.Utxo)
	chErr := make(chan error, 1)
	unspents := make([]explorer.Utxo, 0)

	for _, addr := range addresses {
		go w.getUnspentsForAddress(addr, blindingKeys, chUnspents, chErr)

		select {
		case err := <-chErr:
			close(chErr)
			close(chUnspents)
			return nil, err
		case unspentsForAddress := <-chUnspents:
			unspents = append(unspents, unspentsForAddress...)
		}
	}

	return unspents, nil
}

func (w *walletService) getUnspentsForAddress(addr string, blindingKeys [][]byte, chUnspents chan []explorer.Utxo, chErr chan error) {
	unspents, err := w.explorerService.GetUnspents(addr, blindingKeys)
	if err != nil {
		chErr <- err
		return
	}
	chUnspents <- unspents
}

func parseRequestOutputs(reqOutputs []TxOut) (
	[]*transaction.TxOutput,
	[][]byte,
	error,
) {
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

type sendToManyOpts struct {
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

func sendToMany(opts sendToManyOpts) (string, []domain.UnspentKey, error) {
	w, err := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: opts.mnemonic,
	})
	if err != nil {
		return "", nil, err
	}

	// default to MinMilliSatPerByte if needed
	milliSatPerByte := opts.milliSatPerByte
	if milliSatPerByte < domain.MinMilliSatPerByte {
		milliSatPerByte = domain.MinMilliSatPerByte
	}

	// create the transaction
	newPset, err := w.CreateTx()
	if err != nil {
		return "", nil, err
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
		return "", nil, err
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
		return "", nil, err
	}

	// again, add changes' blinding keys to the list of those of the outputs
	for _, v := range feeUpdateResult.ChangeOutputsBlindingKeys {
		outputsBlindingKeys = append(outputsBlindingKeys, v)
	}

	// blind the transaction
	blindedPset, err := w.BlindTransaction(wallet.BlindTransactionOpts{
		PsetBase64:         feeUpdateResult.PsetBase64,
		OutputBlindingKeys: outputsBlindingKeys,
	})
	if err != nil {
		return "", nil, err
	}

	// add the explicit fee amount
	blindedPlusFees, err := w.UpdateTx(wallet.UpdateTxOpts{
		PsetBase64: blindedPset,
		Outputs:    transactionutil.NewFeeOutput(feeUpdateResult.FeeAmount, network),
		Network:    network,
	})
	if err != nil {
		return "", nil, err
	}

	// sign the inputs
	inputPathsByScript := mergeDerivationPaths(opts.inputPathsByScript, opts.feeInputPathsByScript)
	signedPset, err := w.SignTransaction(wallet.SignTransactionOpts{
		PsetBase64:        blindedPlusFees.PsetBase64,
		DerivationPathMap: inputPathsByScript,
	})
	if err != nil {
		return "", nil, err
	}

	// finalize, extract and return the transaction
	txHex, _, err := wallet.FinalizeAndExtractTransaction(
		wallet.FinalizeAndExtractTransactionOpts{
			PsetBase64: signedPset,
		},
	)
	if err != nil {
		return "", nil, err
	}
	keysLen := len(updateResult.SelectedUnspents) + len(feeUpdateResult.SelectedUnspents)
	selectedUnspentKeys := make([]domain.UnspentKey, keysLen, keysLen)

	for i, u := range updateResult.SelectedUnspents {
		selectedUnspentKeys[i] = domain.UnspentKey{
			TxID: u.Hash(),
			VOut: u.Index(),
		}
	}
	for i, u := range feeUpdateResult.SelectedUnspents {
		i += len(updateResult.SelectedUnspents)
		selectedUnspentKeys[i] = domain.UnspentKey{
			TxID: u.Hash(),
			VOut: u.Index(),
		}
	}

	return txHex, selectedUnspentKeys, nil
}

func getDerivationPathsForUnspents(
	account *domain.Account,
	unspents []explorer.Utxo,
) map[string]string {
	paths := map[string]string{}
	for _, unspent := range unspents {
		script := hex.EncodeToString(unspent.Script())
		if derivationPath, ok := account.DerivationPathByScript[script]; ok {
			paths[script] = derivationPath
		}
	}
	return paths
}

type accountLastDerivedIndex struct {
	external int
	internal int
}

func (a *accountLastDerivedIndex) total() int {
	return (a.external + 1) + (a.internal + 1)
}

func getLatestDerivationIndexForAccount(
	w *wallet.Wallet,
	accountIndex int,
	explorerSvc explorer.Service,
	restore bool,
	net *network.Network,
) *accountLastDerivedIndex {
	if !restore {
		log.Debugf("skip restore - account %d empty", accountIndex)
		return nil
	}

	lastDerivedIndex := &accountLastDerivedIndex{}
	for chainIndex := 0; chainIndex <= 1; chainIndex++ {
		firstUnfundedAddress := -1
		unfundedAddressesCounter := 0
		i := 0
		for unfundedAddressesCounter < 20 {
			ctAddress, script, _ := w.DeriveConfidentialAddress(wallet.DeriveConfidentialAddressOpts{
				DerivationPath: fmt.Sprintf("%d'/%d/%d", accountIndex, chainIndex, i),
				Network:        net,
			})
			blindKey, _, _ := w.DeriveBlindingKeyPair(wallet.DeriveBlindingKeyPairOpts{
				Script: script,
			})

			if !isAddressFunded(ctAddress, blindKey.Serialize(), explorerSvc) {
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

func getLatestDerivationIndexForMarkets(
	w *wallet.Wallet,
	explorerSvc explorer.Service,
	restore bool,
	net *network.Network,
) []*accountLastDerivedIndex {
	marketsLastIndex := make([]*accountLastDerivedIndex, 0)
	i := 0
	for {
		marketIndex := domain.MarketAccountStart + i
		lastDerivedIndex := getLatestDerivationIndexForAccount(
			w,
			marketIndex,
			explorerSvc,
			restore,
			net,
		)
		if lastDerivedIndex == nil {
			break
		}
		marketsLastIndex = append(marketsLastIndex, lastDerivedIndex)
		i++
	}
	return marketsLastIndex
}

func initVaultAccount(
	v *domain.Vault,
	accountIndex int,
	lastDerivedIndex *accountLastDerivedIndex,
) (domain.AddressesInfo, error) {
	if lastDerivedIndex == nil {
		v.InitAccount(accountIndex)
		return nil, nil
	}

	addresses := make(domain.AddressesInfo, 0, lastDerivedIndex.total())
	for i := 0; i <= lastDerivedIndex.external; i++ {
		info, err := v.DeriveNextExternalAddressForAccount(accountIndex)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, *info)
	}
	for i := 0; i <= lastDerivedIndex.internal; i++ {
		info, err := v.DeriveNextInternalAddressForAccount(accountIndex)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, *info)
	}
	return addresses, nil
}

func isAddressFunded(addr string, blindKey []byte, explorerSvc explorer.Service) bool {
	txs, err := explorerSvc.GetTransactionsForAddress(addr, blindKey)
	if err != nil {
		// should we retry?
		return false
	}
	return len(txs) > 0
}

func getBalancesByAsset(unspents []explorer.Utxo) map[string]BalanceInfo {
	balances := map[string]BalanceInfo{}
	for _, unspent := range unspents {
		if _, ok := balances[unspent.Asset()]; !ok {
			balances[unspent.Asset()] = BalanceInfo{}
		}

		balance := balances[unspent.Asset()]
		balance.TotalBalance += unspent.Value()
		if unspent.IsConfirmed() {
			balance.ConfirmedBalance += unspent.Value()
		} else {
			balance.UnconfirmedBalance += unspent.Value()
		}
		balances[unspent.Asset()] = balance
	}
	return balances
}

func fetchUnspents(
	explorerSvc explorer.Service,
	unspentRepo domain.UnspentRepository,
	info domain.AddressesInfo,
) error {
	utxos, err := explorerSvc.GetUnspentsForAddresses(info.AddressesAndKeys())
	if err != nil {
		return err
	}

	unspentsLen := len(utxos)
	unspents := make([]domain.Unspent, unspentsLen, unspentsLen)
	infoByScript := groupAddressesInfoByScript(info)

	for i, u := range utxos {
		addr := infoByScript[hex.EncodeToString(u.Script())].Address
		unspents[i] = domain.Unspent{
			TxID:            u.Hash(),
			VOut:            u.Index(),
			Value:           u.Value(),
			AssetHash:       u.Asset(),
			ValueCommitment: u.ValueCommitment(),
			AssetCommitment: u.AssetCommitment(),
			ValueBlinder:    u.ValueBlinder(),
			AssetBlinder:    u.AssetBlinder(),
			ScriptPubKey:    u.Script(),
			Nonce:           u.Nonce(),
			RangeProof:      u.RangeProof(),
			SurjectionProof: u.SurjectionProof(),
			Confirmed:       u.IsConfirmed(),
			Address:         addr,
		}
	}

	return unspentRepo.AddUnspents(context.Background(), unspents)
}

func addUnspents(unspentRepo domain.UnspentRepository, unspents []domain.Unspent) {
	if err := unspentRepo.AddUnspents(context.Background(), unspents); err != nil {
		log.Warnf(
			"unexpected error occured while adding unspents to the utxo set. "+
				"You must manually run ReloadUtxo RPC as soon as possible to restore "+
				"the utxo set of the internal wallet. Error: %v", err,
		)
	}
	log.Debugf("added %d unspents", len(unspents))
}

func spendUnspents(
	unspentRepo domain.UnspentRepository,
	unspentKeys []domain.UnspentKey,
) {
	count, err := unspentRepo.SpendUnspents(context.Background(), unspentKeys)
	if err != nil {
		log.Warnf(
			"unexpected error occured while updating the utxo set trying to mark "+
				"some unspents as spent. You must manually run ReloadUtxo RPC as soon "+
				"as possible to restore the utxo set of the internal wallet. "+
				"Error: %v", err,
		)
	}
	log.Debugf("spent %d unspents", count)
}

func extractUnspentsFromTx(
	vaultRepo domain.VaultRepository,
	network *network.Network,
	txHex string,
	accountIndex int,
) ([]domain.Unspent, []domain.UnspentKey, error) {
	vault, err := vaultRepo.GetOrCreateVault(context.Background(), nil, "", nil)
	if err != nil {
		return nil, nil, err
	}

	info, err := vault.AllDerivedAddressesInfoForAccount(accountIndex)
	if err != nil {
		return nil, nil, err
	}
	infoByScript := groupAddressesInfoByScript(info)

	if accountIndex != domain.FeeAccount {
		info, _ = vault.AllDerivedAddressesInfoForAccount(domain.FeeAccount)
		for script, in := range groupAddressesInfoByScript(info) {
			infoByScript[script] = in
		}
	}

	tx, err := transaction.NewTxFromHex(txHex)
	if err != nil {
		return nil, nil, err
	}

	unspentsToAdd := make([]domain.Unspent, 0)
	unspentsToSpend := make([]domain.UnspentKey, 0)

	for _, in := range tx.Inputs {
		// our unspents are native-segiwt only
		if len(in.Witness) > 0 {
			pubkey, _ := btcec.ParsePubKey(in.Witness[1], btcec.S256())
			p := payment.FromPublicKey(pubkey, network, nil)

			script := hex.EncodeToString(p.WitnessScript)
			if _, ok := infoByScript[script]; ok {
				unspentsToSpend = append(unspentsToSpend, domain.UnspentKey{
					TxID: bufferutil.TxIDFromBytes(in.Hash),
					VOut: in.Index,
				})
			}
		}
	}

	for i, out := range tx.Outputs {
		script := hex.EncodeToString(out.Script)
		if info, ok := infoByScript[script]; ok {
			unconfidential, ok := transactionutil.UnblindOutput(out, info.BlindingKey)
			if !ok {
				return nil, nil, errors.New("unable to unblind output")
			}
			unspentsToAdd = append(unspentsToAdd, domain.Unspent{
				TxID:            tx.TxHash().String(),
				VOut:            uint32(i),
				Value:           unconfidential.Value,
				AssetHash:       unconfidential.AssetHash,
				ValueCommitment: bufferutil.CommitmentFromBytes(out.Value),
				AssetCommitment: bufferutil.CommitmentFromBytes(out.Asset),
				ValueBlinder:    unconfidential.ValueBlinder,
				AssetBlinder:    unconfidential.AssetBlinder,
				ScriptPubKey:    out.Script,
				Nonce:           out.Nonce,
				RangeProof:      out.RangeProof,
				SurjectionProof: out.SurjectionProof,
				Address:         info.Address,
				Confirmed:       false,
			})
		}
	}
	return unspentsToAdd, unspentsToSpend, nil
}

func extractAndAddUnspentsFromTx(
	unspentRepo domain.UnspentRepository,
	vaultRepo domain.VaultRepository,
	net *network.Network,
	txHex string,
	accountIndex int,
) {
	unspentsToAdd, _, err := extractUnspentsFromTx(
		vaultRepo,
		net,
		txHex,
		accountIndex,
	)
	if err != nil {
		log.Warnf(
			"unable to extract unspents from wallet account tx. You must run "+
				"ReloadUtxo as soon as possible to restore the utxo set of the "+
				"wallet. Error: %v", err,
		)
		return
	}
	addUnspents(unspentRepo, unspentsToAdd)
}
