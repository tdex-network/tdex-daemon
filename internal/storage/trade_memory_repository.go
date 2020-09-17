package storage

import (
	"context"
	"encoding/hex"
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/internal/domain/trade"
)

var (
	// ErrEmptyTradesByMarket ...
	ErrEmptyTradesByMarket = errors.New("no trades found for market")
)

// InMemoryTradeRepository represents an in memory storage
type InMemoryTradeRepository struct {
	trades         map[uuid.UUID]*trade.Trade
	tradesBySwapAcceptID map[string]uuid.UUID
	tradesByTrader map[string][]uuid.UUID
	tradesByMarket map[string][]uuid.UUID

	lock *sync.RWMutex
}

// NewInMemoryTradeRepository returns a new empty InMemoryTradeRepository
func NewInMemoryTradeRepository() *InMemoryTradeRepository {
	return &InMemoryTradeRepository{
		trades:         map[uuid.UUID]*trade.Trade{},
		tradesBySwapAcceptID: map[string]uuid.UUID{},
		tradesByTrader: map[string][]uuid.UUID{},
		tradesByMarket: map[string][]uuid.UUID{},
		lock:           &sync.RWMutex{},
	}
}

// GetOrCreateTrade gets a trade with a given swapID that can be either a
// request, accept, or complete ID. They all identify the same Trade.
// If not found, a new entry is inserted
func (r InMemoryTradeRepository) GetOrCreateTrade(_ context.Context, tradeID *uuid.UUID) (*trade.Trade, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.getOrCreateTrade(tradeID)
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
	tradeID *uuid.UUID,
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

	if swapAccept := updatedTrade.SwapAcceptMessage(); swapAccept != nil {
		if _, ok := r.tradesBySwapAcceptID[swapAccept.GetId()]; !ok {
			r.tradesBySwapAcceptID[swapAccept.GetId()] = currentTrade.ID()
		}
	}

	r.addTradeByMarket(updatedTrade.MarketQuoteAsset(), currentTrade.ID())
	r.addTradeByTrader(hex.EncodeToString(updatedTrade.TraderPubkey()), currentTrade.ID())
	return nil
}

func (r InMemoryTradeRepository) getOrCreateTrade(tradeID *uuid.UUID) (*trade.Trade, error) {
	if tradeID == nil {
		t := trade.NewTrade()
		r.trades[t.ID()] = t
		return t, nil
	}

	currentTrade, ok := r.trades[*tradeID]
	if !ok {
		t := trade.NewTrade()
		r.trades[t.ID()] = t
		return t, nil	
	}

	return currentTrade, nil
}

func (r InMemoryTradeRepository) getAllTrades() ([]*trade.Trade, error) {
	trades := make([]*trade.Trade, 0, len(r.trades))
	for _, trade := range r.trades {
		trades = append(trades, trade)
	}
	return trades, nil
}

func (r InMemoryTradeRepository) getAllTradesByMarket(marketQuoteAsset string) ([]*trade.Trade, error) {
	tradeIDs, ok := r.tradesByMarket[marketQuoteAsset]
	if !ok {
		return nil, ErrEmptyTradesByMarket
	}

	trades := r.tradesFromIDs(tradeIDs)
	return trades, nil
}

func (r InMemoryTradeRepository) getAllTradesByTrader(traderID string) ([]*trade.Trade, error) {
	tradeIDs, ok := r.tradesByTrader[traderID]
	if !ok {
		return nil, nil
	}

	trades := r.tradesFromIDs(tradeIDs)
	return trades, nil
}

func (r InMemoryTradeRepository) tradesFromIDs(tradeIDs []uuid.UUID) []*trade.Trade {
	trades := make([]*trade.Trade, 0, len(tradeIDs))
	for _, tradeID := range tradeIDs {
		trades = append(trades, r.trades[tradeID])
	}
	return trades
}

func (r InMemoryTradeRepository) addTradeByMarket(marketQuoteAsset string, tradeID uuid.UUID) {
	trades, ok := r.tradesByMarket[marketQuoteAsset]
	if !ok {
		r.tradesByMarket[marketQuoteAsset] = []uuid.UUID{tradeID}
		return
	}

	if !contain(trades, tradeID) {
		r.tradesByMarket[marketQuoteAsset] = append(
			r.tradesByMarket[marketQuoteAsset],
			tradeID,
		)
	}
}

func (r InMemoryTradeRepository) addTradeByTrader(traderID string, tradeID uuid.UUID) {
	trades, ok := r.tradesByTrader[traderID]
	if !ok {
		r.tradesByTrader[traderID] = []uuid.UUID{tradeID}
		return
	}

	if !contain(trades, tradeID) {
		r.tradesByTrader[traderID] = append(
			r.tradesByTrader[traderID],
			tradeID,
		)
	}
}

func contain(list []uuid.UUID, id uuid.UUID) bool {
	for _, l := range list {
		if id == l {
			return true
		}
	}
	return false
}
