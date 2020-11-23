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
	return newWalletService(
		vaultRepository,
		unspentRepository,
		crawlerService,
		explorerService,
	)
}

func newWalletService(
	vaultRepository domain.VaultRepository,
	unspentRepository domain.UnspentRepository,
	crawlerService crawler.Service,
	explorerService explorer.Service,
) *walletService {
	w := &walletService{
		vaultRepository:   vaultRepository,
		unspentRepository: unspentRepository,
		crawlerService:    crawlerService,
		explorerService:   explorerService,
	}
	// to understand if the service has an already initialized wallet we check
	// if the inner vaultRepo is able to return a Vault without passing mnemonic
	// and passphrase. If it does, it means it's been retrieved from storage,
	// therefore we let the crawler to start watch all derived addresses and mark
	// the wallet as initialized
	if vault, err := w.vaultRepository.GetOrCreateVault(
		context.Background(), nil, "",
	); err == nil {
		addressesInfo := vault.AllDerivedAddressesInfo()
		for _, info := range addressesInfo {
			if info.AccountIndex != domain.WalletAccount {
				w.crawlerService.AddObservable(&crawler.AddressObservable{
					AccountIndex: info.AccountIndex,
					Address:      info.Address,
					BlindingKey:  info.BlindingKey,
				})
			}
		}
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

	derivedAddresses, prvBlindingKeys, err := w.vaultRepository.
		GetAllDerivedAddressesAndBlindingKeysForAccount(ctx, domain.WalletAccount)
	if err != nil {
		return nil, err
	}

	unspents, err := w.getUnspents(derivedAddresses, prvBlindingKeys)
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
		return nil, ErrWalletNotFunded
	}

	var rawTx []byte
	var addressToObserve *crawler.AddressObservable

	err = w.vaultRepository.UpdateVault(
		ctx,
		nil,
		"",
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
				_, script, _, err := v.DeriveNextInternalAddressForAccount(
					domain.WalletAccount,
				)
				if err != nil {
					return nil, err
				}
				derivationPath, _ := walletAccount.DerivationPathByScript[script]
				changePathsByAsset[asset] = derivationPath
			}
			feeAddress, script, feeBlindkey, err := v.DeriveNextInternalAddressForAccount(domain.FeeAccount)
			if err != nil {
				return nil, err
			}
			feeChangePathByAsset[config.GetNetwork().AssetID] = feeAccount.DerivationPathByScript[script]
			addressToObserve = &crawler.AddressObservable{
				AccountIndex: domain.FeeAccount,
				Address:      feeAddress,
				BlindingKey:  feeBlindkey,
			}

			txHex, _, err := sendToMany(sendToManyOpts{
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
			})
			if err != nil {
				return nil, err
			}

			if req.Push {
				if _, err := w.explorerService.BroadcastTransaction(txHex); err != nil {
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

	// of course, do not forget of starting watching new address of fee account
	w.crawlerService.AddObservable(addressToObserve)

	return rawTx, nil
}

func (w *walletService) getAllUnspentsForAccount(
	ctx context.Context,
	accountIndex int,
	useExplorer bool,
) ([]explorer.Utxo, error) {
	addresses, blindingKeys, err := w.vaultRepository.
		GetAllDerivedAddressesAndBlindingKeysForAccount(ctx, accountIndex)
	if err != nil {
		return nil, err
	}

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
}

func sendToMany(opts sendToManyOpts) (string, string, error) {
	w, err := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: opts.mnemonic,
	})
	if err != nil {
		return "", "", err
	}

	// default to MinMilliSatPerByte if needed
	milliSatPerByte := opts.milliSatPerByte
	if milliSatPerByte < domain.MinMilliSatPerByte {
		milliSatPerByte = domain.MinMilliSatPerByte
	}

	// create the transaction
	newPset, err := w.CreateTx()
	if err != nil {
		return "", "", err
	}

	// add inputs and outputs
	updateResult, err := w.UpdateTx(wallet.UpdateTxOpts{
		PsetBase64:         newPset,
		Unspents:           opts.unspents,
		Outputs:            opts.outputs,
		ChangePathsByAsset: opts.changePathsByAsset,
		MilliSatsPerBytes:  milliSatPerByte,
		Network:            config.GetNetwork(),
	})
	if err != nil {
		return "", "", err
	}

	// update the list of output blinding keys with those of the eventual changes
	outputsBlindingKeys := opts.outputsBlindingKeys
	for _, v := range updateResult.ChangeOutputsBlindingKeys {
		outputsBlindingKeys = append(outputsBlindingKeys, v)
	}

	// add inputs for paying network fees
	updateResult, err = w.UpdateTx(wallet.UpdateTxOpts{
		PsetBase64:         updateResult.PsetBase64,
		Unspents:           opts.feeUnspents,
		ChangePathsByAsset: opts.feeChangePathByAsset,
		MilliSatsPerBytes:  milliSatPerByte,
		Network:            config.GetNetwork(),
		WantChangeForFees:  true,
	})
	if err != nil {
		return "", "", err
	}

	// again, add changes' blinding keys to the list of those of the outputs
	for _, v := range updateResult.ChangeOutputsBlindingKeys {
		outputsBlindingKeys = append(outputsBlindingKeys, v)
	}

	// blind the transaction
	blindedPset, err := w.BlindTransaction(wallet.BlindTransactionOpts{
		PsetBase64:         updateResult.PsetBase64,
		OutputBlindingKeys: outputsBlindingKeys,
	})
	if err != nil {
		return "", "", err
	}

	// add the explicit fee amount
	blindedPlusFees, err := w.UpdateTx(wallet.UpdateTxOpts{
		PsetBase64: blindedPset,
		Outputs:    transactionutil.NewFeeOutput(updateResult.FeeAmount),
		Network:    config.GetNetwork(),
	})
	if err != nil {
		return "", "", err
	}

	// sign the inputs
	inputPathsByScript := mergeDerivationPaths(opts.inputPathsByScript, opts.feeInputPathsByScript)
	signedPset, err := w.SignTransaction(wallet.SignTransactionOpts{
		PsetBase64:        blindedPlusFees.PsetBase64,
		DerivationPathMap: inputPathsByScript,
	})
	if err != nil {
		return "", "", err
	}

	// finalize, extract and return the transaction
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

func getLatestDerivationIndexForAccount(w *wallet.Wallet, accountIndex int, explorerSvc explorer.Service) *accountLastDerivedIndex {
	lastDerivedIndex := &accountLastDerivedIndex{}
	for chainIndex := 0; chainIndex <= 1; chainIndex++ {
		firstUnfundedAddress := -1
		unfundedAddressesCounter := 0
		i := 0
		for unfundedAddressesCounter < 20 {
			ctAddress, _, _ := w.DeriveConfidentialAddress(wallet.DeriveConfidentialAddressOpts{
				DerivationPath: fmt.Sprintf("%d'/%d/%d", accountIndex, chainIndex, i),
				Network:        config.GetNetwork(),
			})

			if !isAddressFunded(ctAddress, explorerSvc) {
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

func isAddressFunded(addr string, explorerSvc explorer.Service) bool {
	txs, err := explorerSvc.GetTransactionsForAddress(addr)
	if err != nil {
		// should we retry?
		return false
	}
	return len(txs) > 0
}

func getBalancesByAsset(unspents []explorer.Utxo) map[string]domain.BalanceInfo {
	balances := map[string]domain.BalanceInfo{}
	for _, unspent := range unspents {
		if _, ok := balances[unspent.Asset()]; !ok {
			balances[unspent.Asset()] = domain.BalanceInfo{}
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
