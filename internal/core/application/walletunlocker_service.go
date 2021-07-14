package application

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
	"github.com/vulpemventures/go-elements/network"
)

var (
	// ErrWalletNotInitialized is returned when attempting to unlock or change
	// the password of a not initialized wallet.
	ErrWalletNotInitialized = fmt.Errorf("wallet not initialized")
)

type WalletUnlockerService interface {
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
		newPassphrase string,
	) error
	IsReady(ctx context.Context) bool
	PassphraseChan() chan PassphraseMsg
	ReadyChan() chan bool
}

type walletUnlockerService struct {
	repoManager        ports.RepoManager
	explorerService    explorer.Service
	blockchainListener BlockchainListener
	walletInitialized  bool
	walletIsSyncing    bool
	walletRestored     bool
	network            *network.Network
	marketFee          int64
	marketBaseAsset    string

	lock      *sync.RWMutex
	pwChan    chan PassphraseMsg
	readyChan chan bool
}

func NewWalletUnlockerService(
	repoManager ports.RepoManager,
	explorerService explorer.Service,
	blockchainListener BlockchainListener,
	net *network.Network,
	marketFee int64,
	marketBaseAsset string,
) WalletUnlockerService {
	return newWalletUnlockerService(
		repoManager,
		explorerService,
		blockchainListener,
		net,
		marketFee,
		marketBaseAsset,
	)
}

func newWalletUnlockerService(
	repoManager ports.RepoManager,
	explorerService explorer.Service,
	blockchainListener BlockchainListener,
	net *network.Network,
	marketFee int64,
	marketBaseAsset string,
) *walletUnlockerService {
	w := &walletUnlockerService{
		repoManager:        repoManager,
		explorerService:    explorerService,
		blockchainListener: blockchainListener,
		network:            net,
		marketFee:          marketFee,
		marketBaseAsset:    marketBaseAsset,
		lock:               &sync.RWMutex{},
		pwChan:             make(chan PassphraseMsg, 1),
		readyChan:          make(chan bool, 1),
	}
	// to understand if the service has an already initialized wallet we check
	// if the inner vaultRepo is able to return a Vault without passing mnemonic
	// and passphrase. If it does, it means it's been retrieved from storage,
	// therefore we let the crawler to start watch all derived addresses and mark
	// the wallet as initialized
	if vault, err := w.repoManager.VaultRepository().GetOrCreateVault(
		context.Background(), nil, "", nil,
	); err == nil {
		go func() {
			log.Info("Restoring internal wallet's utxo set. This could take a while...")

			info := vault.AllDerivedAddressesInfo()
			if err := fetchAndAddUnspents(
				w.explorerService,
				w.repoManager.UnspentRepository(),
				w.blockchainListener,
				info,
			); err != nil {
				log.Infof("Failed for reason: %s", err)
				w.setRestored(false)
				return
			}
			log.Info("Done")
			w.setRestored(true)
		}()
		w.setInitialized(true)

	}
	return w
}

func (w *walletUnlockerService) GenSeed(ctx context.Context) ([]string, error) {
	mnemonic, err := wallet.NewMnemonic(wallet.NewMnemonicOpts{EntropySize: 256})
	if err != nil {
		return nil, err
	}
	return mnemonic, nil
}

func (w *walletUnlockerService) IsReady(_ context.Context) bool {
	return w.isInitialized() && w.isRestored()
}

type InitWalletReply struct {
	AccountIndex int
	AddressIndex int
	Status       int
	Data         string
}

func (w *walletUnlockerService) InitWallet(
	ctx context.Context,
	mnemonic []string,
	passphrase string,
	restore bool,
	chRes chan *InitWalletReply,
	chErr chan error,
) {
	defer func() {
		close(chErr)
		close(chRes)
	}()
	if w.isInitialized() {
		return
	}
	// this prevents strange behaviors by making consecutive calls to InitWallet
	// while it's still syncing
	if w.isSyncing() {
		return
	}

	if restore {
		w.setSyncing(true)
		log.Debug("restoring wallet")
		start := time.Now()
		defer func() {
			elapsed := time.Since(start)
			log.Debugf("Restoration took: %.2fs", elapsed.Seconds())
		}()
	} else {
		log.Debug("creating wallet")
	}

	vault, err := w.repoManager.VaultRepository().GetOrCreateVault(ctx, mnemonic, passphrase, w.network)
	if err != nil {
		chErr <- fmt.Errorf("unable to retrieve vault: %v", err)
		return
	}
	defer vault.Lock()

	if restore {
		data := "addresses discovery"

		chRes <- &InitWalletReply{
			Status: Processing,
			Data:   data,
		}

		allInfo := make(domain.AddressesInfo, 0)
		feeInfo, marketInfoByAccount, err := w.restoreVault(ctx, mnemonic, vault)
		if err != nil {
			chErr <- fmt.Errorf("unable to restore vault: %v", err)
			return
		}
		chRes <- &InitWalletReply{
			Status: Done,
			Data:   data,
		}

		allInfo = append(allInfo, feeInfo...)
		for _, marketInfo := range marketInfoByAccount {
			allInfo = append(allInfo, marketInfo...)
		}

		// restore unspents
		unspents, err := w.restoreUnspents(allInfo, chRes)
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
		markets := make([]*domain.Market, mLen, mLen)
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

		if err := w.persistRestoredState(ctx, vault, unspents, markets); err != nil {
			chErr <- fmt.Errorf("unable to persist restored state: %v", err)
			return
		}

		go startObserveUnconfirmedUnspents(w.blockchainListener, unspents)
	}

	if w.blockchainListener.PubSubService() != nil {
		go func() {
			if err := w.blockchainListener.PubSubService().Store().Init(
				passphrase,
			); err != nil {
				log.WithError(err).Warn(
					"an error occured while initializing pubsub service. " +
						"Pubsub not available for the current session.",
				)
			}
		}()
	}

	w.pwChan <- PassphraseMsg{
		Method:     InitWallet,
		CurrentPwd: passphrase,
	}
	w.setInitialized(true)
	w.setRestored(true)
	if w.isSyncing() {
		w.setSyncing(false)
	}
	log.Debug("done")
}

func (w *walletUnlockerService) UnlockWallet(
	ctx context.Context,
	passphrase string,
) error {
	if !w.isInitialized() {
		return ErrWalletNotInitialized
	}

	vault, err := w.repoManager.VaultRepository().GetOrCreateVault(ctx, nil, "", nil)
	if err != nil {
		return err
	}

	if !vault.IsLocked() {
		return nil
	}

	if err := vault.Unlock(passphrase); err != nil {
		return err
	}

	if err := w.repoManager.VaultRepository().UpdateVault(
		ctx,
		func(_ *domain.Vault) (*domain.Vault, error) {
			return vault, nil
		},
	); err != nil {
		return err
	}

	if w.blockchainListener.PubSubService() != nil {
		go func() {
			// For backward compatibility, check if the pubsub store has been
			// initialized by wallet.Init, otherwise it is initialized before being
			// unlocked here.
			if w.blockchainListener.PubSubService().Store().IsLocked() {
				if err := w.blockchainListener.PubSubService().Store().Init(
					passphrase,
				); err != nil {
					log.WithError(err).Warn(
						"an error occured while initializing pubsub service. " +
							"Pubsub not available for the current session.",
					)
					return
				}
			}
			if err := w.blockchainListener.PubSubService().Store().Unlock(
				passphrase,
			); err != nil {
				log.WithError(err).Warn(
					"an error occured while unlocking pubsub internal store",
				)
			}
		}()
	}

	w.pwChan <- PassphraseMsg{
		Method:     UnlockWallet,
		CurrentPwd: passphrase,
	}
	w.blockchainListener.StartObservation()
	return nil
}

func (w *walletUnlockerService) ChangePassword(
	ctx context.Context,
	currentPassphrase string,
	newPassphrase string,
) error {
	if !w.isInitialized() {
		return ErrWalletNotInitialized
	}

	if err := w.repoManager.VaultRepository().UpdateVault(
		ctx,
		func(v *domain.Vault) (*domain.Vault, error) {
			err := v.ChangePassphrase(currentPassphrase, newPassphrase)
			if err != nil {
				return nil, err
			}
			return v, nil
		},
	); err != nil {
		return err
	}

	if w.blockchainListener.PubSubService() != nil {
		go func() {
			if err := w.blockchainListener.PubSubService().Store().ChangePassword(
				currentPassphrase, newPassphrase,
			); err != nil {
				log.WithError(err).Warn(
					"an error occured while updating pubsub service password",
				)
			}
		}()
	}

	w.pwChan <- PassphraseMsg{
		Method:     ChangePassphrase,
		CurrentPwd: currentPassphrase,
		NewPwd:     newPassphrase,
	}

	return nil
}

func (w *walletUnlockerService) PassphraseChan() chan PassphraseMsg {
	return w.pwChan
}

func (w *walletUnlockerService) ReadyChan() chan bool {
	return w.readyChan
}

func (w *walletUnlockerService) isSyncing() bool {
	w.lock.RLock()
	defer w.lock.RUnlock()

	return w.walletIsSyncing
}

func (w *walletUnlockerService) setSyncing(val bool) {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.walletIsSyncing = val
}

func (w *walletUnlockerService) isInitialized() bool {
	w.lock.RLock()
	defer w.lock.RUnlock()

	return w.walletInitialized
}

func (w *walletUnlockerService) setInitialized(val bool) {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.walletInitialized = val
}

func (w *walletUnlockerService) isRestored() bool {
	w.lock.RLock()
	defer w.lock.RUnlock()

	return w.walletRestored
}

func (w *walletUnlockerService) setRestored(val bool) {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.walletRestored = val

	w.readyChan <- val
}

func (w *walletUnlockerService) restoreVault(
	ctx context.Context, mnemonic []string, vault *domain.Vault,
) (domain.AddressesInfo, map[int]domain.AddressesInfo, error) {
	ww, _ := wallet.NewWalletFromMnemonic(wallet.NewWalletFromMnemonicOpts{
		SigningMnemonic: mnemonic,
	})

	var feeRestoreInfo, walletRestoreInfo *accountLastDerivedIndex
	var marketsRestoreInfo []*accountLastDerivedIndex

	feeRestoreInfo = w.restoreAccount(ww, domain.FeeAccount)
	walletRestoreInfo = w.restoreAccount(ww, domain.WalletAccount)
	marketsRestoreInfo = w.restoreMarketAccounts(ww)

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

func (w *walletUnlockerService) restoreAccount(
	ww *wallet.Wallet, accountIndex int,
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

func (w *walletUnlockerService) restoreMarketAccounts(
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

func (w *walletUnlockerService) restoreMarket(
	ctx context.Context,
	accountIndex int,
	info domain.AddressesInfo,
	unspentsByAddress map[string][]domain.Unspent,
) (*domain.Market, error) {
	market, err := w.repoManager.MarketRepository().GetOrCreateMarket(ctx, &domain.Market{
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

func (w *walletUnlockerService) persistRestoredState(
	ctx context.Context,
	vault *domain.Vault,
	unspents []domain.Unspent,
	markets []*domain.Market,
) error {
	if _, err := w.repoManager.RunTransaction(
		ctx,
		false,
		func(ctx context.Context) (interface{}, error) {
			// update changes to vault
			if err := w.repoManager.VaultRepository().UpdateVault(
				ctx,
				func(v *domain.Vault) (*domain.Vault, error) {
					return vault, nil
				},
			); err != nil {
				return nil, fmt.Errorf("unable to persist changes to the vault repo: %s", err)
			}

			// update changes to markets
			for _, m := range markets {
				if err := w.repoManager.MarketRepository().UpdateMarket(
					ctx,
					m.AccountIndex,
					func(_ *domain.Market) (*domain.Market, error) {
						return m, nil
					},
				); err != nil {
					return nil, fmt.Errorf("unable to persist changes to the market repo: %s", err)
				}
			}

			// update utxo set
			if err := w.repoManager.UnspentRepository().AddUnspents(ctx, unspents); err != nil {
				return nil, fmt.Errorf("unable to persist changes to the unspent repo: %s", err)
			}

			return nil, nil
		}); err != nil {
		return err
	}

	return nil
}

type unspentInfo struct {
	info     domain.AddressInfo
	unspents []domain.Unspent
	err      error
}

func (w *walletUnlockerService) restoreUnspents(
	info domain.AddressesInfo, chRes chan *InitWalletReply,
) ([]domain.Unspent, error) {
	chUnspentsInfo := make(chan unspentInfo)
	unspents := make([]domain.Unspent, 0)
	cb := newCircuitBreaker()
	wg := &sync.WaitGroup{}
	wg.Add(len(info))

	go func() {
		wg.Wait()
		close(chUnspentsInfo)
	}()

	for i := range info {
		in := info[i]
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
		go w.restoreUnspentsForAddress(cb, in, chUnspentsInfo, wg)
		time.Sleep(1 * time.Millisecond)
	}

	for r := range chUnspentsInfo {
		if r.err != nil {
			return nil, r.err
		}

		unspents = append(unspents, r.unspents...)

		path := strings.Split(r.info.DerivationPath, "/")
		addressIndex, _ := strconv.Atoi(path[len(path)-1])
		chRes <- &InitWalletReply{
			AccountIndex: r.info.AccountIndex,
			AddressIndex: addressIndex,
			Status:       Done,
			Data: fmt.Sprintf(
				"account %d index %d",
				r.info.AccountIndex,
				addressIndex,
			),
		}
	}

	return unspents, nil
}

func (w *walletUnlockerService) restoreUnspentsForAddress(
	cb *gobreaker.CircuitBreaker,
	info domain.AddressInfo,
	chUnspent chan unspentInfo,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	addr := info.Address
	blindingKeys := [][]byte{info.BlindingKey}

	iUtxos, err := cb.Execute(func() (interface{}, error) {
		return w.explorerService.GetUnspents(addr, blindingKeys)
	})
	if err != nil {
		chUnspent <- unspentInfo{err: err}
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
			RangeProof:      make([]byte, 1),
			SurjectionProof: make([]byte, 1),
			Confirmed:       u.IsConfirmed(),
			Address:         addr,
		}
	}
	chUnspent <- unspentInfo{info: info, unspents: unspents}
}

type accountLastDerivedIndex struct {
	external int
	internal int
}

func (a *accountLastDerivedIndex) total() int {
	return (a.external + 1) + (a.internal + 1)
}

func initVaultAccount(
	v *domain.Vault, accountIndex int, lastDerivedIndex *accountLastDerivedIndex,
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

func isAddressFunded(
	addr string, blindKey []byte, explorerSvc explorer.Service,
) bool {
	txs, err := explorerSvc.GetTransactionsForAddress(addr, blindKey)
	if err != nil {
		// should we retry?
		return false
	}
	return len(txs) > 0
}

func fetchUnspents(
	explorerSvc explorer.Service, info domain.AddressesInfo,
) ([]domain.Unspent, error) {
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
	unspents := make([]domain.Unspent, 0, unspentsLen)
	infoByScript := groupAddressesInfoByScript(info)

	for _, u := range utxos {
		addr := infoByScript[hex.EncodeToString(u.Script())].Address
		unspents = append(unspents, domain.Unspent{
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
			RangeProof:      make([]byte, 1),
			SurjectionProof: make([]byte, 1),
			Confirmed:       u.IsConfirmed(),
			Address:         addr,
		})
	}

	return unspents, nil
}

func addUnspents(
	unspentRepo domain.UnspentRepository, unspents []domain.Unspent,
) error {
	return unspentRepo.AddUnspents(context.Background(), unspents)
}

func fetchAndAddUnspents(
	explorerSvc explorer.Service,
	unspentRepo domain.UnspentRepository,
	bcListener BlockchainListener,
	info domain.AddressesInfo,
) error {
	var unspents []domain.Unspent
	var err error

	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		log.Debugf("Restored uspents: %d", len(unspents))
		log.Debugf("Restore took: %.2fs", elapsed.Seconds())
	}()

	unspents, err = fetchUnspents(explorerSvc, info)
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
	unspentRepo domain.UnspentRepository, unspentKeys []domain.UnspentKey,
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
	unspentRepo domain.UnspentRepository, unspentKeys []domain.UnspentKey,
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
	bcListener BlockchainListener, unspents []domain.Unspent,
) {
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

	return TransactionManager.ExtractUnspents(txHex, infoByScript, network)
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
