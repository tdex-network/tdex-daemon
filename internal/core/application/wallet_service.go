package application

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/btcsuite/btcd/btcec"
	log "github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
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
		chRes chan *InitWalletReply,
		chErr chan error,
	)
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
	dbManager          ports.DbManager
	explorerService    explorer.Service
	blockchainListener BlockchainListener
	walletInitialized  bool
	walletIsSyncing    bool
	withElements       bool
	network            *network.Network
	marketFee          int64
	marketBaseAsset    string

	lock *sync.RWMutex
}

func NewWalletService(
	dbManager ports.DbManager,
	explorerService explorer.Service,
	blockchainListener BlockchainListener,
	withElements bool,
	net *network.Network,
	marketFee int64,
	marketBaseAsset string,
) (WalletService, error) {
	return newWalletService(
		dbManager,
		explorerService,
		blockchainListener,
		withElements,
		net,
		marketFee,
		marketBaseAsset,
	)
}

func newWalletService(
	dbManager ports.DbManager,
	explorerService explorer.Service,
	blockchainListener BlockchainListener,
	withElements bool,
	net *network.Network,
	marketFee int64,
	marketBaseAsset string,
) (*walletService, error) {
	w := &walletService{
		dbManager:          dbManager,
		explorerService:    explorerService,
		blockchainListener: blockchainListener,
		withElements:       withElements,
		network:            net,
		marketFee:          marketFee,
		marketBaseAsset:    marketBaseAsset,
		lock:               &sync.RWMutex{},
	}
	// to understand if the service has an already initialized wallet we check
	// if the inner vaultRepo is able to return a Vault without passing mnemonic
	// and passphrase. If it does, it means it's been retrieved from storage,
	// therefore we let the crawler to start watch all derived addresses and mark
	// the wallet as initialized
	if vault, err := w.dbManager.VaultRepository().GetOrCreateVault(
		context.Background(), nil, "", nil,
	); err == nil {
		info := vault.AllDerivedAddressesInfo()
		if err := fetchAndAddUnspents(
			w.explorerService,
			w.dbManager.UnspentRepository(),
			w.blockchainListener,
			info,
		); err != nil {
			return nil, err
		}

		w.setInitialized(true)
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

type InitWalletReply struct {
	AccountIndex int
	AddressIndex int
	Status       int
	Data         string
}

func (w *walletService) InitWallet(
	ctx context.Context,
	mnemonic []string,
	passphrase string,
	restore bool,
	chRes chan *InitWalletReply,
	chErr chan error,
) {
	if w.walletInitialized {
		chRes <- nil
		return
	}
	// this prevents strange behaviors by making consecutive calls to InitWallet
	// while it's still syncing
	if w.walletIsSyncing {
		chRes <- nil
		return
	}

	if restore && w.withElements {
		chErr <- fmt.Errorf(
			"Restoring a wallet through the Elements explorer is not availble at the " +
				"moment. Please restart the daemon using the Esplora block explorer.",
		)
		return
	}

	w.setSyncing(true)
	if restore {
		log.Debug("restoring wallet")
	} else {
		log.Debug("creating wallet")
	}

	vault, err := w.dbManager.VaultRepository().GetOrCreateVault(ctx, mnemonic, passphrase, w.network)
	if err != nil {
		chErr <- fmt.Errorf("unable to retrieve vault: %v", err)
		return
	}
	defer vault.Lock()

	if restore {
		data := "addresses discovery"
		if w.withElements {
			data += ". With elements this may take a while"
		}
		chRes <- &InitWalletReply{
			Status: Processing,
			Data:   data,
		}
	}

	allInfo := make(domain.AddressesInfo, 0)
	feeInfo, marketInfoByAccount, err := w.restoreVault(ctx, mnemonic, vault, restore)
	if err != nil {
		chErr <- fmt.Errorf("unable to restore vault: %v", err)
		return
	}

	if restore {
		chRes <- &InitWalletReply{
			Status: Done,
			Data:   "addresses discovery",
		}
	}

	allInfo = append(allInfo, feeInfo...)
	for _, marketInfo := range marketInfoByAccount {
		allInfo = append(allInfo, marketInfo...)
	}

	var unspents []domain.Unspent
	var markets []*domain.Market

	if restore {
		// restore unspents
		unspents, err = w.restoreUnspents(allInfo, chRes)
		if err != nil {
			chErr <- fmt.Errorf("unable to restore unspents: %v", err)
			return
		}

		// group unspents by address to facilitate market restoration
		unspentsByAddress := make(map[string][]domain.Unspent)
		for _, u := range unspents {
			unspentsByAddress[u.Address] = append(unspentsByAddress[u.Address], u)
		}

		// restore markets
		mLen := len(marketInfoByAccount)
		markets = make([]*domain.Market, mLen, mLen)
		i := 0
		for accountIndex, info := range marketInfoByAccount {
			market, err := w.restoreMarket(ctx, accountIndex, info, unspentsByAddress)
			if err != nil {
				chErr <- fmt.Errorf("unable to restore market: %v", err)
				return
			}
			markets[i] = market
			i++
		}
	}

	if err := w.persistRestoredState(ctx, vault, unspents, markets); err != nil {
		chErr <- fmt.Errorf("unable to persist restored state: %v", err)
		return
	}

	// notify the interface that no more replies will be sent and the channel
	// can be closed.
	chRes <- nil
	go startObserveUnconfirmedUnspents(w.blockchainListener, unspents)
	w.setInitialized(true)
	w.setSyncing(false)
	log.Debug("done")
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

	if err := w.dbManager.VaultRepository().UpdateVault(
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

	return w.dbManager.VaultRepository().UpdateVault(
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

	err = w.dbManager.VaultRepository().UpdateVault(
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

	info, err := w.dbManager.VaultRepository().GetAllDerivedAddressesInfoForAccount(ctx, domain.WalletAccount)
	if err != nil {
		return nil, err
	}

	addresses, keys := info.AddressesAndKeys()
	unspents, err := w.explorerService.GetUnspentsForAddresses(addresses, keys)
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

	var txHex string

	if err := w.dbManager.VaultRepository().UpdateVault(
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

			_txHex, err := sendToMany(sendToManyOpts{
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

	go extractUnspentsFromTxAndUpdateUtxoSet(
		w.dbManager.UnspentRepository(),
		w.dbManager.VaultRepository(),
		w.network,
		txHex,
		domain.FeeAccount,
	)

	rawTx, _ := hex.DecodeString(txHex)
	return rawTx, nil
}

func (w *walletService) isSyncing() bool {
	w.lock.RLock()
	defer w.lock.RUnlock()

	return w.walletIsSyncing
}

func (w *walletService) setSyncing(val bool) {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.walletIsSyncing = val
}

func (w *walletService) isInitialized() bool {
	w.lock.RLock()
	defer w.lock.RUnlock()

	return w.walletInitialized
}

func (w *walletService) setInitialized(val bool) {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.walletInitialized = val
}

func (w *walletService) restoreVault(
	ctx context.Context,
	mnemonic []string,
	vault *domain.Vault,
	restore bool,
) (domain.AddressesInfo, map[int]domain.AddressesInfo, error) {
	ww, _ := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: mnemonic,
	})

	var feeRestoreInfo, walletRestoreInfo *accountLastDerivedIndex
	var marketsRestoreInfo []*accountLastDerivedIndex

	if restore {
		feeRestoreInfo = w.restoreAccount(ww, domain.FeeAccount)
		walletRestoreInfo = w.restoreAccount(ww, domain.WalletAccount)
		marketsRestoreInfo = w.restoreMarketAccounts(ww)
	}

	// restore vault accounts
	feeInfo, err := initVaultAccount(vault, domain.FeeAccount, feeRestoreInfo)
	if err != nil {
		return nil, nil, err
	}

	if _, err := initVaultAccount(vault, domain.WalletAccount, walletRestoreInfo); err != nil {
		return nil, nil, err
	}

	marketInfoByAccount := make(map[int]domain.AddressesInfo)
	for i, m := range marketsRestoreInfo {
		accountIndex := domain.MarketAccountStart + i
		info, err := initVaultAccount(vault, accountIndex, m)
		if err != nil {
			return nil, nil, err
		}
		marketInfoByAccount[accountIndex] = info
	}

	return feeInfo, marketInfoByAccount, nil
}

func (w *walletService) restoreAccount(
	ww *wallet.Wallet,
	accountIndex int,
) *accountLastDerivedIndex {
	lastDerivedIndex := &accountLastDerivedIndex{}
	for chainIndex := 0; chainIndex <= 1; chainIndex++ {
		firstUnusedAddress := -1
		unusedAddressesCounter := 0
		i := 0
		for unusedAddressesCounter < 20 {
			ctAddress, script, _ := ww.DeriveConfidentialAddress(wallet.DeriveConfidentialAddressOpts{
				DerivationPath: fmt.Sprintf("%d'/%d/%d", accountIndex, chainIndex, i),
				Network:        w.network,
			})
			blindKey, _, _ := ww.DeriveBlindingKeyPair(wallet.DeriveBlindingKeyPairOpts{
				Script: script,
			})

			if !isAddressFunded(ctAddress, blindKey.Serialize(), w.explorerService) {
				if firstUnusedAddress < 0 {
					firstUnusedAddress = i
				}
				unusedAddressesCounter++
			} else {
				if firstUnusedAddress >= 0 {
					firstUnusedAddress = -1
					unusedAddressesCounter = 0
				}
			}
			i++
		}
		if chainIndex == 0 {
			lastDerivedIndex.external = firstUnusedAddress - 1
		} else {
			lastDerivedIndex.internal = firstUnusedAddress - 1
		}
	}

	if lastDerivedIndex.external < 0 && lastDerivedIndex.internal < 0 {
		log.Debugf("account %d empty", accountIndex)
		return nil
	}
	log.Debugf(
		"account %d total fetched addresses: %d",
		accountIndex, lastDerivedIndex.total(),
	)
	return lastDerivedIndex
}

func (w *walletService) restoreMarketAccounts(
	ww *wallet.Wallet,
) []*accountLastDerivedIndex {
	marketsLastIndex := make([]*accountLastDerivedIndex, 0)
	i := 0
	for {
		marketIndex := domain.MarketAccountStart + i
		lastDerivedIndex := w.restoreAccount(
			ww,
			marketIndex,
		)
		if lastDerivedIndex == nil {
			break
		}
		marketsLastIndex = append(marketsLastIndex, lastDerivedIndex)
		i++
	}
	return marketsLastIndex
}

func (w *walletService) restoreMarket(
	ctx context.Context,
	accountIndex int,
	info domain.AddressesInfo,
	unspentsByAddress map[string][]domain.Unspent,
) (*domain.Market, error) {
	market, err := w.dbManager.MarketRepository().GetOrCreateMarket(ctx, &domain.Market{
		AccountIndex: accountIndex,
		Fee:          w.marketFee,
	})
	if err != nil {
		return nil, err
	}

	if len(unspentsByAddress) > 0 {
		outpoints := make([]domain.OutpointWithAsset, 0)
		for _, i := range info {
			unspentsForAddress := unspentsByAddress[i.Address]
			for _, u := range unspentsForAddress {
				outpoints = append(outpoints, u.ToOutpointWithAsset())
			}
		}

		if err := market.FundMarket(outpoints, w.marketBaseAsset); err != nil {
			return nil, err
		}
	}

	return market, nil
}

func (w *walletService) persistRestoredState(
	ctx context.Context,
	vault *domain.Vault,
	unspents []domain.Unspent,
	markets []*domain.Market,
) error {
	// update changes to vault
	if err := w.dbManager.VaultRepository().UpdateVault(
		ctx,
		func(v *domain.Vault) (*domain.Vault, error) {
			return vault, nil
		},
	); err != nil {
		return fmt.Errorf("unable to persist changes to the vault repo: %s", err)
	}

	// update utxo set
	if err := w.dbManager.UnspentRepository().AddUnspents(ctx, unspents); err != nil {
		return fmt.Errorf("unable to persist changes to the unspent repo: %s", err)
	}

	// update changes to markets
	for _, m := range markets {
		if err := w.dbManager.MarketRepository().UpdateMarket(
			ctx,
			m.AccountIndex,
			func(_ *domain.Market) (*domain.Market, error) {
				return m, nil
			},
		); err != nil {
			return fmt.Errorf("unable to persist changes to the market repo: %s", err)
		}
	}

	return nil
}

func (w *walletService) getAllUnspentsForAccount(
	ctx context.Context,
	accountIndex int,
	useExplorer bool,
) ([]explorer.Utxo, error) {
	info, err := w.dbManager.VaultRepository().GetAllDerivedAddressesInfoForAccount(ctx, accountIndex)
	if err != nil {
		return nil, err
	}
	addresses, blindingKeys := info.AddressesAndKeys()

	if useExplorer {
		return w.explorerService.GetUnspentsForAddresses(addresses, blindingKeys)
	}

	unspents, err := w.dbManager.UnspentRepository().GetAvailableUnspentsForAddresses(ctx, addresses)
	if err != nil {
		return nil, err
	}

	utxos := make([]explorer.Utxo, 0, len(unspents))
	for _, u := range unspents {
		utxos = append(utxos, u.ToUtxo())
	}
	return utxos, nil
}

type unspentInfo struct {
	info     domain.AddressInfo
	unspents []domain.Unspent
}

func (w *walletService) restoreUnspents(
	info domain.AddressesInfo,
	chRes chan *InitWalletReply,
) ([]domain.Unspent, error) {
	chUnspentsInfo := make(chan unspentInfo)
	chErr := make(chan error, 1)
	unspents := make([]domain.Unspent, 0)
	cb := newCircuitBreaker()

	for _, in := range info {
		path := strings.Split(in.DerivationPath, "/")
		addressIndex, _ := strconv.Atoi(path[len(path)-1])
		chRes <- &InitWalletReply{
			AccountIndex: in.AccountIndex,
			AddressIndex: addressIndex,
			Status:       Processing,
			Data: fmt.Sprintf(
				"account %d index %d", in.AccountIndex, addressIndex,
			),
		}
		go w.restoreUnspentsForAddress(cb, in, chUnspentsInfo, chErr)

		select {
		case err := <-chErr:
			close(chErr)
			close(chUnspentsInfo)
			return nil, err
		case unspentsInfo := <-chUnspentsInfo:
			unspents = append(unspents, unspentsInfo.unspents...)

			info := unspentsInfo.info
			path := strings.Split(info.DerivationPath, "/")
			addressIndex, _ := strconv.Atoi(path[len(path)-1])
			chRes <- &InitWalletReply{
				AccountIndex: info.AccountIndex,
				AddressIndex: addressIndex,
				Status:       Done,
				Data: fmt.Sprintf(
					"account %d index %d",
					info.AccountIndex,
					addressIndex,
				),
			}
		}
	}

	return unspents, nil
}

func (w *walletService) restoreUnspentsForAddress(
	cb *gobreaker.CircuitBreaker,
	info domain.AddressInfo,
	chUnspentsInfo chan unspentInfo,
	chErr chan error,
) {
	addr := info.Address
	blindingKeys := [][]byte{info.BlindingKey}

	iUtxos, err := cb.Execute(func() (interface{}, error) {
		return w.explorerService.GetUnspents(addr, blindingKeys)
	})
	if err != nil {
		chErr <- err
		return
	}
	utxos := iUtxos.([]explorer.Utxo)

	unspents := make([]domain.Unspent, len(utxos), len(utxos))
	for i, u := range utxos {
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
	chUnspentsInfo <- unspentInfo{info, unspents}
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

func sendToMany(opts sendToManyOpts) (string, error) {
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

	// again, add changes' blinding keys to the list of those of the outputs
	for _, v := range feeUpdateResult.ChangeOutputsBlindingKeys {
		outputsBlindingKeys = append(outputsBlindingKeys, v)
	}

	// blind the transaction
	blindedPset, err := w.BlindTransactionWithKeys(wallet.BlindTransactionWithKeysOpts{
		PsetBase64:         feeUpdateResult.PsetBase64,
		OutputBlindingKeys: outputsBlindingKeys,
	})
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
	if err != nil {
		return "", err
	}

	return txHex, nil
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

func fetchUnspents(explorerSvc explorer.Service, info domain.AddressesInfo) ([]domain.Unspent, error) {
	if len(info) <= 0 {
		return nil, nil
	}
	cb := newCircuitBreaker()

	iUtxos, err := cb.Execute(func() (interface{}, error) {
		return explorerSvc.GetUnspentsForAddresses(info.AddressesAndKeys())
	})
	if err != nil {
		return nil, err
	}
	utxos := iUtxos.([]explorer.Utxo)

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

	return unspents, nil
}

func addUnspents(unspentRepo domain.UnspentRepository, unspents []domain.Unspent) error {
	return unspentRepo.AddUnspents(context.Background(), unspents)
}

func fetchAndAddUnspents(
	explorerSvc explorer.Service,
	unspentRepo domain.UnspentRepository,
	bcListener BlockchainListener,
	info domain.AddressesInfo,
) error {
	unspents, err := fetchUnspents(explorerSvc, info)
	if err != nil {
		return err
	}
	if unspents == nil {
		return nil
	}
	go startObserveUnconfirmedUnspents(bcListener, unspents)
	return addUnspents(unspentRepo, unspents)
}

func spendUnspents(
	unspentRepo domain.UnspentRepository,
	unspentKeys []domain.UnspentKey,
) (int, error) {
	return unspentRepo.SpendUnspents(context.Background(), unspentKeys)
}

func addUnspentsAsync(unspentRepo domain.UnspentRepository, unspents []domain.Unspent) {
	if err := addUnspents(unspentRepo, unspents); err != nil {
		log.Warnf(
			"unexpected error occured while adding unspents to the utxo set. "+
				"You must manually run ReloadUtxo RPC as soon as possible to restore "+
				"the utxo set of the internal wallet. Error: %v", err,
		)
	}
	log.Debugf("added %d unspents", len(unspents))
}

func spendUnspentsAsync(
	unspentRepo domain.UnspentRepository,
	unspentKeys []domain.UnspentKey,
) {
	count, err := spendUnspents(unspentRepo, unspentKeys)
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

func startObserveUnconfirmedUnspents(
	bcListener BlockchainListener,
	unspents []domain.Unspent) {
	count := 0
	for _, u := range unspents {
		if !u.IsConfirmed() {
			bcListener.StartObserveTx(u.TxID)
			count++
		}
	}
	if count > 0 {
		log.Debugf("num of unconfirmed unspents to watch: %d", count)
	}
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

func extractUnspentsFromTxAndUpdateUtxoSet(
	unspentRepo domain.UnspentRepository,
	vaultRepo domain.VaultRepository,
	net *network.Network,
	txHex string,
	accountIndex int,
) {
	unspentsToAdd, unspentsToSpend, err := extractUnspentsFromTx(
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
	addUnspentsAsync(unspentRepo, unspentsToAdd)
	spendUnspentsAsync(unspentRepo, unspentsToSpend)
}

func newCircuitBreaker() *gobreaker.CircuitBreaker {
	return gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name: "explorer",
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests > 20 && failureRatio >= 0.7
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			if to == gobreaker.StateOpen {
				log.Warn("explorer seems down, stop allowing requests")
			}
			if from == gobreaker.StateOpen && to == gobreaker.StateHalfOpen {
				log.Info("checking explorer status")
			}
			if from == gobreaker.StateHalfOpen && to == gobreaker.StateClosed {
				log.Info("explorer seems ok, restart allowing requests")
			}
		},
	})
}
