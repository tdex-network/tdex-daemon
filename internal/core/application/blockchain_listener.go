package application

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/vulpemventures/go-elements/network"
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

	PubSubService() SecurePubSub
}

type blockchainListener struct {
	crawlerSvc         crawler.Service
	explorerSvc        explorer.Service
	repoManager        ports.RepoManager
	pubsubSvc          SecurePubSub
	started            bool
	pendingObservables []crawler.Observable
	marketBaseAsset    string
	network            *network.Network

	mutex *sync.RWMutex
}

// NewBlockchainListener returns a BlockchainListener with all the needed services
func NewBlockchainListener(
	crawlerSvc crawler.Service,
	repoManager ports.RepoManager,
	pubsubSvc SecurePubSub,
	marketBaseAsset string,
	net *network.Network,
) BlockchainListener {
	return newBlockchainListener(
		crawlerSvc,
		repoManager,
		pubsubSvc,
		marketBaseAsset,
		net,
	)
}

func newBlockchainListener(
	crawlerSvc crawler.Service,
	repoManager ports.RepoManager,
	pubsubSvc SecurePubSub,
	marketBaseAsset string,
	net *network.Network,
) *blockchainListener {
	return &blockchainListener{
		crawlerSvc:         crawlerSvc,
		repoManager:        repoManager,
		pubsubSvc:          pubsubSvc,
		mutex:              &sync.RWMutex{},
		pendingObservables: make([]crawler.Observable, 0),
		marketBaseAsset:    marketBaseAsset,
		network:            net,
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

func (b *blockchainListener) PubSubService() SecurePubSub {
	return b.pubsubSvc
}

func (b *blockchainListener) listenToEventChannel() {
	for {
		event := <-b.crawlerSvc.GetEventChannel()

		switch event.Type() {
		default:
			// unnoticeable sleep to prevent high cpu usage
			// https://github.com/golang/go/issues/27707#issuecomment-698487427
			time.Sleep(time.Microsecond)
		case crawler.CloseSignal:
			log.Trace("CloseEvent detected")
			log.Trace("stop listening on event channel")
			return
		case crawler.TransactionConfirmed:
			e := event.(crawler.TransactionEvent)
			ctx := context.Background()

			trade, err := b.repoManager.TradeRepository().GetTradeByTxID(ctx, e.TxID)
			if err != nil {
				log.Warnf("unable to find trade with id %s: %v", e.TxID, err)
				break
			}

			if err := b.settleTrade(&trade.ID, e); err != nil {
				log.Warnf("trying to settle trade with id %s: %v", trade.ID, err)
				break
			}
			if err := b.confirmOrAddUnspents(e.TxHex, e.TxID, trade.MarketQuoteAsset); err != nil {
				log.Warnf("trying to confirm or add unspents: %v", err)
				break
			}

			// Publish message for topic TradeSettled to pubsub service.
			if b.pubsubSvc != nil {
				go func() {
					payload := map[string]interface{}{
						"txid": trade.TxID,
						"swap": map[string]interface{}{
							"amount_p": trade.SwapRequestMessage().GetAmountP(),
							"asset_p":  trade.SwapRequestMessage().GetAssetP(),
							"amount_r": trade.SwapRequestMessage().GetAmountR(),
							"asset_r":  trade.SwapRequestMessage().GetAssetR(),
						},
						"price": map[string]string{
							"base_price":  trade.MarketPrice.BasePrice.String(),
							"quote_price": trade.MarketPrice.QuotePrice.String(),
						},
						"market": map[string]string{
							"base_asset":  b.marketBaseAsset,
							"quote_asset": trade.MarketQuoteAsset,
						},
					}
					message, _ := json.Marshal(payload)
					topics := b.pubsubSvc.TopicsByCode()
					topic := topics[TradeSettled]

					if err := b.pubsubSvc.Publish(topic.Label(), string(message)); err != nil {
						log.WithError(err).Warnf(
							"an error occured while publishing message for topic %s",
							topic.Label(),
						)
					}
				}()
			}
			// stop watching for a tx after it's confirmed
			b.StopObserveTx(e.TxID)
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

func (b *blockchainListener) settleTrade(tradeID *uuid.UUID, event crawler.TransactionEvent) error {
	if err := b.repoManager.TradeRepository().UpdateTrade(
		context.Background(),
		tradeID,
		func(t *domain.Trade) (*domain.Trade, error) {
			mustAddTxHex := t.IsAccepted()
			if _, err := t.Settle(uint64(event.BlockTime)); err != nil {
				return nil, err
			}
			if mustAddTxHex {
				t.TxHex = event.TxHex
			}

			return t, nil
		},
	); err != nil {
		return err
	}

	log.Infof("trade with id %s settled", tradeID)
	return nil
}

func (b *blockchainListener) confirmOrAddUnspents(
	txHex string,
	txID string,
	mktAsset string,
) error {
	ctx := context.Background()
	_, accountIndex, err := b.repoManager.MarketRepository().GetMarketByAsset(ctx, mktAsset)
	if err != nil {
		return err
	}

	unspentsToAdd, unspentsToSpend, err := extractUnspentsFromTx(
		b.repoManager.VaultRepository(),
		b.network,
		txHex,
		accountIndex,
	)
	if err != nil {
		return err
	}

	uLen := len(unspentsToAdd)
	unspentAddresses := make([]string, uLen, uLen)
	for i, u := range unspentsToAdd {
		unspentAddresses[i] = u.Address
	}

	u, err := b.repoManager.UnspentRepository().GetAllUnspentsForAddresses(ctx, unspentAddresses)
	if err != nil {
		return err
	}
	if len(u) > 0 {
		unspentKeys := make([]domain.UnspentKey, uLen, uLen)
		for i, u := range unspentsToAdd {
			unspentKeys[i] = u.Key()
		}
		count, err := b.repoManager.UnspentRepository().ConfirmUnspents(ctx, unspentKeys)
		if err != nil {
			return err
		}
		log.Debugf("confirmed %d unspents", count)
		return nil
	}

	go func() {
		// these unspents must be inserted already confirmed.
		for i := range unspentsToAdd {
			unspentsToAdd[i].Confirmed = true
		}
		addUnspentsAsync(b.repoManager.UnspentRepository(), unspentsToAdd)
		spendUnspentsAsync(b.repoManager.UnspentRepository(), unspentsToSpend)
	}()

	return nil
}
