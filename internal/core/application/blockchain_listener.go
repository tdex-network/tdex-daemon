package application

import (
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
)

const readOnlyTx = true

//BlockchainListener defines the needed method sto start and stop a blockchain listener
type BlockchainListener interface {
	StartObservation()
	StopObservation()

	StartObserveAddress(accountIndex int, addr string, blindKey []byte)
	StopObserveAddress(addr string)

	StartObserveTx(txid string)
	StopObserveTx(txid string)
}

type blockchainListener struct {
	unspentRepository  domain.UnspentRepository
	marketRepository   domain.MarketRepository
	vaultRepository    domain.VaultRepository
	tradeRepository    domain.TradeRepository
	crawlerSvc         crawler.Service
	explorerSvc        explorer.Service
	dbManager          ports.DbManager
	started            bool
	pendingObservables []crawler.Observable
	// Loggers
	feeDepositLogged    bool
	feeBalanceLowLogged bool
	marketBaseAsset     string
	feeBalanceThreshold uint64

	mutex *sync.RWMutex
}

// NewBlockchainListener returns a BlockchainListener with all the needed services
func NewBlockchainListener(
	unspentRepository domain.UnspentRepository,
	marketRepository domain.MarketRepository,
	vaultRepository domain.VaultRepository,
	tradeRepository domain.TradeRepository,
	crawlerSvc crawler.Service,
	explorerSvc explorer.Service,
	dbManager ports.DbManager,
	marketBaseAsset string,
	feeBalanceThreshold uint64,
) BlockchainListener {
	return newBlockchainListener(
		unspentRepository,
		marketRepository,
		vaultRepository,
		tradeRepository,
		crawlerSvc,
		explorerSvc,
		dbManager,
		marketBaseAsset,
		feeBalanceThreshold,
	)
}

func newBlockchainListener(
	unspentRepository domain.UnspentRepository,
	marketRepository domain.MarketRepository,
	vaultRepository domain.VaultRepository,
	tradeRepository domain.TradeRepository,
	crawlerSvc crawler.Service,
	explorerSvc explorer.Service,
	dbManager ports.DbManager,
	marketBaseAsset string,
	feeBalanceThreshold uint64,
) *blockchainListener {
	return &blockchainListener{
		unspentRepository:   unspentRepository,
		marketRepository:    marketRepository,
		vaultRepository:     vaultRepository,
		tradeRepository:     tradeRepository,
		crawlerSvc:          crawlerSvc,
		explorerSvc:         explorerSvc,
		dbManager:           dbManager,
		mutex:               &sync.RWMutex{},
		pendingObservables:  make([]crawler.Observable, 0),
		marketBaseAsset:     marketBaseAsset,
		feeBalanceThreshold: feeBalanceThreshold,
	}
}

func (b *blockchainListener) StartObservation() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if !b.started {
		log.Debug("start crawler")
		go b.crawlerSvc.Start()
		log.Debug("start listening on event channel")
		go b.listenToEventChannel()
		go b.startPendingObservables()
		b.started = true
	}
}

func (b *blockchainListener) StopObservation() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.started {
		log.Debug("stop crawler")
		b.crawlerSvc.Stop()
		b.started = false
	}
}

func (b *blockchainListener) StartObserveAddress(
	accountIndex int,
	addr string,
	blindKey []byte,
) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	observable := &crawler.AddressObservable{
		AccountIndex: accountIndex,
		Address:      addr,
		BlindingKey:  blindKey,
	}

	if !b.started {
		b.pendingObservables = append(b.pendingObservables, observable)
		return
	}
	b.crawlerSvc.AddObservable(observable)
}

func (b *blockchainListener) StartObserveTx(txid string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	observable := &crawler.TransactionObservable{TxID: txid}

	if !b.started {
		b.pendingObservables = append(b.pendingObservables, observable)
		return
	}
	b.crawlerSvc.AddObservable(observable)
}

func (b *blockchainListener) StopObserveAddress(addr string) {
	b.crawlerSvc.RemoveObservable(&crawler.AddressObservable{
		Address: addr,
	})
}

func (b *blockchainListener) StopObserveTx(txid string) {
	b.crawlerSvc.RemoveObservable(&crawler.TransactionObservable{TxID: txid})
}

func (b *blockchainListener) listenToEventChannel() {
	for {
		select {
		case event := <-b.crawlerSvc.GetEventChannel():
			switch event.Type() {
			default:
				// unnoticeable sleep to prevent high cpu usage
				// https://github.com/golang/go/issues/27707#issuecomment-698487427
				time.Sleep(time.Microsecond)
			case crawler.CloseSignal:
				log.Debug("CloseEvent detected")
				log.Debug("stop listening on event channel")
				return
			case crawler.TransactionConfirmed:
				e := event.(crawler.TransactionEvent)
				ctx := context.Background()
				if _, err := b.dbManager.RunTransaction(
					ctx,
					!readOnlyTx,
					func(ctx context.Context) (interface{}, error) {
						return nil, b.updateTrade(ctx, e)
					},
				); err != nil {
					log.Warnf(
						"trying to update trade completion status %s\n",
						err.Error(),
					)
					break
				}
				// stop watching for a tx after it's confirmed
				b.StopObserveTx(e.TxID)
			}
		}
	}
}

func (b *blockchainListener) startPendingObservables() {
	if len(b.pendingObservables) <= 0 {
		return
	}

	for _, observable := range b.pendingObservables {
		b.crawlerSvc.AddObservable(observable)
		time.Sleep(200 * time.Millisecond)
	}

	b.pendingObservables = nil
}

func (b *blockchainListener) updateTrade(
	ctx context.Context,
	event crawler.TransactionEvent,
) error {
	trade, err := b.tradeRepository.GetTradeByTxID(ctx, event.TxID)
	if err != nil {
		return err
	}

	if trade.Status.Code == domain.CompletedStatus.Code {
		return nil
	}

	err = b.tradeRepository.UpdateTrade(
		ctx,
		&trade.ID,
		func(t *domain.Trade) (*domain.Trade, error) {
			if _, err := t.Settle(uint64(event.BlockTime)); err != nil {
				return nil, err
			}

			return t, nil
		},
	)
	if err != nil {
		return err
	}

	log.Infof("trade %s completed", trade.ID)

	return nil
}

func (b *blockchainListener) handleUnspents(
	event crawler.Event,
) (context.Context, error) {
	e := event.(crawler.AddressEvent)
	unspents := unspentsFromEvent(e)
	ctx := context.Background()

	if _, err := b.dbManager.RunUnspentsTransaction(
		ctx,
		!readOnlyTx,
		func(ctx context.Context) (interface{}, error) {
			return nil, b.updateUnspentsForAddress(ctx, unspents, e.Address)
		},
	); err != nil {
		log.Warnf("trying to update unspents for address %s: %s\n", e.Address, err.Error())
		return nil, err
	}
	return ctx, nil
}

func (b *blockchainListener) checkFeeAccountBalance(ctx context.Context) error {
	info, err := b.vaultRepository.GetAllDerivedAddressesInfoForAccount(
		ctx,
		domain.FeeAccount,
	)
	if err != nil {
		return err
	}

	feeAccountBalance, err := b.unspentRepository.GetBalance(
		ctx,
		info.Addresses(),
		b.marketBaseAsset,
	)
	if err != nil {
		return err
	}

	if feeAccountBalance < b.feeBalanceThreshold {
		if !b.getSafeFeeBalanceLowLogged() {
			log.Warn(
				"fee account balance for account index too low. Trades for markets won't be " +
					"served properly. Fund the fee account as soon as possible",
			)
			b.updateSafeFeeBalanceLowLogged(true)
		}
		b.updateSafeFeeDepositLogged(false)
	} else {
		if !b.getSafeFeeDepositLogged() {
			log.Info("fee account funded. Trades can be served")
			b.updateSafeFeeDepositLogged(true)
		}
		b.updateSafeFeeBalanceLowLogged(true)
	}
	return nil
}

func (b *blockchainListener) updateSafeFeeBalanceLowLogged(logged bool) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.feeBalanceLowLogged = logged
}

func (b *blockchainListener) updateSafeFeeDepositLogged(logged bool) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.feeDepositLogged = logged
}

func (b *blockchainListener) getSafeFeeBalanceLowLogged() bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.feeBalanceLowLogged
}

func (b *blockchainListener) getSafeFeeDepositLogged() bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.feeDepositLogged
}

func (b *blockchainListener) checkMarketAccountFundings(ctx context.Context, accountIndex int) error {
	market, err := b.marketRepository.GetMarketByAccount(ctx, accountIndex)
	if err != nil {
		return err
	}

	// if market is not found it means it's never been opened, therefore
	// let's notify whether the market can be safely opened, base or quote
	// asset are missing, or if the market account owns too many assets.
	if market == nil || !market.IsFunded() {
		info, err := b.vaultRepository.GetAllDerivedAddressesInfoForAccount(ctx, accountIndex)
		if err != nil {
			return err
		}
		unspents, err := b.unspentRepository.GetUnspentsForAddresses(ctx, info.Addresses())
		if err != nil {
			return err
		}
		unspentsAssetType := map[string]bool{}
		for _, u := range unspents {
			unspentsAssetType[u.AssetHash] = true
		}

		switch len(unspentsAssetType) {
		case 0:
			log.Warnf("no funds detected for market %d", accountIndex)
		case 1:
			asset := "base"
			for k := range unspentsAssetType {
				if k == b.marketBaseAsset {
					asset = "quote"
				}
			}
			log.Warnf("%s asset is missing for market %d", asset, accountIndex)
		case 2:
			baseAsset := b.marketBaseAsset
			var asset string
			for k := range unspentsAssetType {
				if k != baseAsset {
					asset = k
				}
			}
			log.Infof("funding market with quote asset %s", asset)

			// Prepare unspents to become outpoint for the market to run validations
			outpoints := make([]domain.OutpointWithAsset, 0, len(unspents))
			for _, u := range unspents {
				outpoints = append(outpoints, domain.OutpointWithAsset{
					Txid:  u.TxID,
					Vout:  int(u.VOut),
					Asset: u.AssetHash,
				})
			}

			// Update the market trying to funding attaching the newly found quote asset.
			if err := b.marketRepository.UpdateMarket(
				ctx,
				accountIndex,
				func(m *domain.Market) (*domain.Market, error) {
					if err := m.FundMarket(outpoints, baseAsset); err != nil {
						return nil, err
					}

					return m, nil
				},
			); err != nil {
				log.Warn(err)
			}

		default:
			log.Warnf(
				"market with account %d funded with more than 2 different assets."+
					"It will be impossible to determine the correct quote asset "+
					"and market won't be opened. Funds must be moved away from "+
					"this account so that it owns only unspents of 2 type of assets",
				accountIndex,
			)
		}
	}
	return nil
}

func (b *blockchainListener) updateUnspentsForAddress(
	ctx context.Context,
	unspents []domain.Unspent,
	address string,
) error {
	existingUnspents, err := b.unspentRepository.GetAllUnspentsForAddresses(
		ctx,
		[]string{address},
	)
	if err != nil {
		return err
	}

	// check for unspents to add to the storage
	unspentsToAdd := make([]domain.Unspent, 0)
	for _, u := range unspents {
		if index := findUnspent(existingUnspents, u); index < 0 {
			unspentsToAdd = append(unspentsToAdd, u)
		}
	}

	//update spent
	unspentsToMarkAsSpent := make([]domain.UnspentKey, 0)
	unspentsToMarkAsConfirmed := make([]domain.UnspentKey, 0)
	for _, existingUnspent := range existingUnspents {
		if index := findUnspent(unspents, existingUnspent); index < 0 {
			unspentsToMarkAsSpent = append(unspentsToMarkAsSpent, existingUnspent.Key())
		} else {
			if existingUnspent.IsConfirmed() != unspents[index].IsConfirmed() {
				unspentsToMarkAsConfirmed = append(unspentsToMarkAsConfirmed, existingUnspent.Key())
			}
		}
	}

	if len(unspentsToAdd) > 0 {
		if err := b.unspentRepository.AddUnspents(ctx, unspentsToAdd); err != nil {
			return err
		}
	}
	if len(unspentsToMarkAsSpent) > 0 {
		if err := b.unspentRepository.SpendUnspents(ctx, unspentsToMarkAsSpent); err != nil {
			return err
		}
	}
	if len(unspentsToMarkAsConfirmed) > 0 {
		if err := b.unspentRepository.ConfirmUnspents(ctx, unspentsToMarkAsConfirmed); err != nil {
			return err
		}
	}
	return nil
}

func unspentsFromEvent(event crawler.AddressEvent) []domain.Unspent {
	unspents := make([]domain.Unspent, 0, len(event.Utxos))
	for _, utxo := range event.Utxos {
		u := domain.Unspent{
			TxID:            utxo.Hash(),
			VOut:            utxo.Index(),
			Value:           utxo.Value(),
			AssetHash:       utxo.Asset(),
			ValueCommitment: utxo.ValueCommitment(),
			AssetCommitment: utxo.AssetCommitment(),
			ValueBlinder:    utxo.ValueBlinder(),
			AssetBlinder:    utxo.AssetBlinder(),
			ScriptPubKey:    utxo.Script(),
			Nonce:           utxo.Nonce(),
			RangeProof:      utxo.RangeProof(),
			SurjectionProof: utxo.SurjectionProof(),
			Confirmed:       utxo.IsConfirmed(),
			Address:         event.Address,
		}
		unspents = append(unspents, u)
	}
	return unspents
}

func findUnspent(list []domain.Unspent, unspent domain.Unspent) int {
	for i, u := range list {
		if u.IsKeyEqual(unspent.Key()) {
			return i
		}
	}
	return -1
}
