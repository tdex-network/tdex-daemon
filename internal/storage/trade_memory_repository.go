package storage

import (
	"context"
	"encoding/hex"
	"errors"
	"sync"

	"github.com/tdex-network/tdex-daemon/internal/domain/trade"
)

var (
	// ErrEmptyTradesByMarket ...
	ErrEmptyTradesByMarket = errors.New("no trades found for market")
)

// InMemoryTradeRepository represents an in memory storage
type InMemoryTradeRepository struct {
	trades         map[string]*trade.Trade
	tradesByTrader map[string]ids
	tradesByMarket map[string]ids

	lock *sync.RWMutex
}

// NewInMemoryTradeRepository returns a new empty InMemoryTradeRepository
func NewInMemoryTradeRepository() *InMemoryTradeRepository {
	return &InMemoryTradeRepository{
		trades:         map[string]*trade.Trade{},
		tradesByTrader: map[string]ids{},
		tradesByMarket: map[string]ids{},
		lock:           &sync.RWMutex{},
	}
}

// GetOrCreateTrade gets a trade with a given swapID that can be either a
// request, accept, or complete ID. They all identify the same Trade.
// If not found, a new entry is inserted
func (r InMemoryTradeRepository) GetOrCreateTrade(_ context.Context, swapID string) (*trade.Trade, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.getOrCreateTrade(swapID)
}

// GetAllTrades returns all the trades processed by the daemon
func (r InMemoryTradeRepository) GetAllTrades(_ context.Context) ([]*trade.Trade, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.getAllTrades()
}

// GetAllTradesByMarket returns all the trades processed for the given market
func (r InMemoryTradeRepository) GetAllTradesByMarket(_ context.Context, marketQuoteAsset string) ([]*trade.Trade, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.getAllTradesByMarket(marketQuoteAsset)
}

// GetAllTradesByTrader returns all the trades processed for the given trader
func (r InMemoryTradeRepository) GetAllTradesByTrader(_ context.Context, traderID string) ([]*trade.Trade, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.getAllTradesByTrader(traderID)
}

// UpdateTrade updates data to a trade identified by any of its swap ids (request, accept, complete) passing an update function
func (r InMemoryTradeRepository) UpdateTrade(
	_ context.Context,
	tradeID string,
	updateFn func(t *trade.Trade) (*trade.Trade, error),
) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	currentTrade, err := r.getOrCreateTrade(tradeID)
	if err != nil {
		return err
	}

	updatedTrade, err := updateFn(currentTrade)
	if err != nil {
		return err
	}

	tradeIDs := updatedTrade.ID()
	for _, tradeID := range tradeIDs {
		r.trades[tradeID] = updatedTrade
	}

	r.addTradeByMarket(updatedTrade.MarketQuoteAsset(), tradeIDs[0])
	r.addTradeByTrader(hex.EncodeToString(updatedTrade.TraderID()), tradeIDs[0])
	return nil
}

func (r InMemoryTradeRepository) getOrCreateTrade(swapID string) (*trade.Trade, error) {
	currentTrade, ok := r.trades[swapID]
	if !ok {
		t := trade.NewTrade()
		r.trades[swapID] = t
		return t, nil
	}

	return currentTrade, nil
}

func (r InMemoryTradeRepository) getAllTrades() ([]*trade.Trade, error) {
	contain := func(trades []*trade.Trade, tradeID string) bool {
		for _, trade := range trades {
			if ids(trade.ID()).contain(tradeID) {
				return true
			}
		}
		return false
	}

	trades := make([]*trade.Trade, 0)
	for tradeID, trade := range r.trades {
		if !contain(trades, tradeID) {
			trades = append(trades, trade)
		}
	}
	return trades, nil
}

func (r InMemoryTradeRepository) getAllTradesByMarket(marketQuoteAsset string) ([]*trade.Trade, error) {
	swapIDs, ok := r.tradesByMarket[marketQuoteAsset]
	if !ok {
		return nil, ErrEmptyTradesByMarket
	}

	trades := r.tradesFromSwapIDs(swapIDs)
	return trades, nil
}

func (r InMemoryTradeRepository) getAllTradesByTrader(traderID string) ([]*trade.Trade, error) {
	swapIDs, ok := r.tradesByTrader[traderID]
	if !ok {
		return nil, nil
	}

	trades := r.tradesFromSwapIDs(swapIDs)
	return trades, nil
}

func (r InMemoryTradeRepository) tradesFromSwapIDs(swapIDs []string) []*trade.Trade {
	trades := make([]*trade.Trade, 0, len(swapIDs))
	for _, swapID := range swapIDs {
		trades = append(trades, r.trades[swapID])
	}
	return trades
}

func (r InMemoryTradeRepository) addTradeByMarket(marketQuoteAsset string, tradeID string) {
	trades, ok := r.tradesByMarket[marketQuoteAsset]
	if !ok {
		r.tradesByMarket[marketQuoteAsset] = ids{tradeID}
		return
	}

	if !trades.contain(tradeID) {
		r.tradesByMarket[marketQuoteAsset] = append(
			r.tradesByMarket[marketQuoteAsset],
			tradeID,
		)
	}
}

func (r InMemoryTradeRepository) addTradeByTrader(traderID string, tradeID string) {
	trades, ok := r.tradesByTrader[traderID]
	if !ok {
		r.tradesByTrader[traderID] = ids{tradeID}
		return
	}

	if !trades.contain(tradeID) {
		r.tradesByTrader[traderID] = append(
			r.tradesByTrader[traderID],
			tradeID,
		)
	}
}

type ids []string

func (i ids) contain(str string) bool {
	for _, s := range i {
		if s == str {
			return true
		}
	}
	return false
}
