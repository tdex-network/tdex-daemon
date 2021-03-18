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
	"github.com/vulpemventures/go-elements/transaction"
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
				var tradeTxHex string

				if _, err := b.dbManager.RunTransaction(
					ctx,
					!readOnlyTx,
					func(ctx context.Context) (interface{}, error) {
						trade, err := b.tradeRepository.GetTradeByTxID(ctx, e.TxID)
						if err != nil {
							return nil, err
						}

						if err := b.tradeRepository.UpdateTrade(
							ctx,
							&trade.ID,
							func(t *domain.Trade) (*domain.Trade, error) {
								if _, err := t.Settle(uint64(e.BlockTime)); err != nil {
									return nil, err
								}

								return t, nil
							},
						); err != nil {
							return nil, err
						}

						tradeTxHex = trade.TxHex
						log.Infof("trade with id %s settled", trade.ID)

						return nil, nil
					},
				); err != nil {
					log.Warnf("trying to settle trade  %v", err.Error())
					break
				}
				if _, err := b.dbManager.RunUnspentsTransaction(
					ctx,
					!readOnlyTx,
					func(ctx context.Context) (interface{}, error) {
						tx, err := transaction.NewTxFromHex(tradeTxHex)
						if err != nil {
							return nil, err
						}
						keysLen := len(tx.Outputs)
						unspentKeys := make([]domain.UnspentKey, keysLen, keysLen)
						for i := 0; i < keysLen; i++ {
							unspentKeys[i] = domain.UnspentKey{
								TxID: e.TxID,
								VOut: uint32(i),
							}
						}

						return nil, b.unspentRepository.ConfirmUnspents(ctx, unspentKeys)
					},
				); err != nil {
					log.Warnf("trying to confirm unspents: %v", err)
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
