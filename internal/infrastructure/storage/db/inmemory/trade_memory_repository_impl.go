package inmemory

import (
	"context"
	"encoding/hex"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"

	"github.com/google/uuid"
)

// TradeRepositoryImpl represents an in memory storage
type TradeRepositoryImpl struct {
	db *DbManager
}

// NewTradeRepositoryImpl returns a new empty TradeRepositoryImpl
func NewTradeRepositoryImpl(db *DbManager) domain.TradeRepository {
	return &TradeRepositoryImpl{
		db: db,
	}
}

func (r TradeRepositoryImpl) GetTradeByTxID(
	ctx context.Context,
	txID string,
) (*domain.Trade, error) {
	r.db.tradeStore.locker.Lock()
	defer r.db.tradeStore.locker.Unlock()

	trades, err := r.getAllTrades()
	if err != nil {
		return nil, err
	}
	for _, t := range trades {
		if t.TxID == txID {
			return t, nil
		}
	}

	return nil, nil
}

// GetCompletedTradesByMarket returns the copmpleted trades for a given quote asset
func (r TradeRepositoryImpl) GetCompletedTradesByMarket(ctx context.Context, marketQuoteAsset string) ([]*domain.Trade, error) {
	r.db.tradeStore.locker.Lock()
	defer r.db.tradeStore.locker.Unlock()

	tradesByMarkets, err := r.getAllTradesByMarket(marketQuoteAsset)
	if err != nil {
		return nil, err
	}

	completedTrades := make([]*domain.Trade, 0)
	for _, trade := range tradesByMarkets {
		if trade.Status.Code == domain.CompletedStatus.Code {
			completedTrades = append(completedTrades, trade)
		}
	}

	return completedTrades, nil
}

// GetOrCreateTrade gets a trade with a given swapID that can be either a
// request, accept, or complete ID. They all identify the same Trade.
// If not found, a new entry is inserted
func (r TradeRepositoryImpl) GetOrCreateTrade(_ context.Context, tradeID *uuid.UUID) (*domain.Trade, error) {
	r.db.tradeStore.locker.Lock()
	defer r.db.tradeStore.locker.Unlock()

	return r.getOrCreateTrade(tradeID)
}

// GetAllTrades returns all the trades processed by the daemon
func (r TradeRepositoryImpl) GetAllTrades(_ context.Context) ([]*domain.Trade, error) {
	r.db.tradeStore.locker.Lock()
	defer r.db.tradeStore.locker.Unlock()

	return r.getAllTrades()
}

// GetAllTradesByMarket returns all the trades processed for the given market
func (r TradeRepositoryImpl) GetAllTradesByMarket(_ context.Context, marketQuoteAsset string) ([]*domain.Trade, error) {
	r.db.tradeStore.locker.Lock()
	defer r.db.tradeStore.locker.Unlock()

	return r.getAllTradesByMarket(marketQuoteAsset)
}

// GetAllTradesByTrader returns all the trades processed for the given trader
func (r TradeRepositoryImpl) GetAllTradesByTrader(_ context.Context, traderID string) ([]*domain.Trade, error) {
	r.db.tradeStore.locker.Lock()
	defer r.db.tradeStore.locker.Unlock()

	return r.getAllTradesByTrader(traderID)
}

// GetTradeBySwapAcceptID returns the trade idetified by the swap accept message's id
func (r TradeRepositoryImpl) GetTradeBySwapAcceptID(_ context.Context, swapAcceptID string) (*domain.Trade, error) {
	r.db.tradeStore.locker.Lock()
	defer r.db.tradeStore.locker.Unlock()

	return r.getTradeBySwapAcceptID(swapAcceptID)
}

// UpdateTrade updates data to a trade identified by any of its swap ids (request, accept, complete) passing an update function
func (r TradeRepositoryImpl) UpdateTrade(
	ctx context.Context,
	tradeID *uuid.UUID,
	updateFn func(t *domain.Trade) (*domain.Trade, error),
) error {
	r.db.tradeStore.locker.Lock()
	defer r.db.tradeStore.locker.Unlock()

	currentTrade, err := r.getOrCreateTrade(tradeID)
	if err != nil {
		return err
	}

	updatedTrade, err := updateFn(currentTrade)
	if err != nil {
		return err
	}

	if swapAccept := updatedTrade.SwapAcceptMessage(); swapAccept != nil {
		if _, ok := r.db.tradeStore.tradesBySwapAcceptID[swapAccept.GetId()]; !ok {
			r.db.tradeStore.tradesBySwapAcceptID[swapAccept.GetId()] = currentTrade.ID
		}
	}

	r.addTradeByMarket(updatedTrade.MarketQuoteAsset, currentTrade.ID)
	r.addTradeByTrader(hex.EncodeToString(updatedTrade.TraderPubkey), currentTrade.ID)

	r.db.tradeStore.trades[updatedTrade.ID] = *updatedTrade

	return nil
}

func (r TradeRepositoryImpl) getOrCreateTrade(tradeID *uuid.UUID) (*domain.Trade, error) {
	createTrade := func() (*domain.Trade, error) {
		t := domain.NewTrade()
		r.db.tradeStore.trades[t.ID] = *t
		return t, nil
	}

	if tradeID == nil {
		return createTrade()
	}

	currentTrade, ok := r.db.tradeStore.trades[*tradeID]
	if !ok {
		return nil, ErrTradesNotFound
	}

	return &currentTrade, nil
}

func (r TradeRepositoryImpl) getAllTrades() ([]*domain.Trade, error) {
	allTrades := make([]*domain.Trade, 0)
	for _, trade := range r.db.tradeStore.trades {
		allTrades = append(allTrades, &trade)
	}
	return allTrades, nil
}

func (r TradeRepositoryImpl) getAllTradesByMarket(marketQuoteAsset string) ([]*domain.Trade, error) {
	tradeIDs, ok := r.db.tradeStore.tradesByMarket[marketQuoteAsset]
	if !ok {
		return nil, nil
	}

	tradeList := tradesFromIDs(r.db.tradeStore.trades, tradeIDs)
	return tradeList, nil
}

func (r TradeRepositoryImpl) getAllTradesByTrader(traderID string) ([]*domain.Trade, error) {
	tradeIDs, ok := r.db.tradeStore.tradesByTrader[traderID]
	if !ok {
		return nil, ErrTradesNotFound
	}

	tradeList := tradesFromIDs(r.db.tradeStore.trades, tradeIDs)
	return tradeList, nil
}

func (r TradeRepositoryImpl) getTradeBySwapAcceptID(swapAcceptID string) (*domain.Trade, error) {
	tradeID, ok := r.db.tradeStore.tradesBySwapAcceptID[swapAcceptID]
	if !ok {
		return nil, ErrTradesNotFound
	}
	trade := r.db.tradeStore.trades[tradeID]
	return &trade, nil
}

func tradesFromIDs(trades map[uuid.UUID]domain.Trade, tradeIDs []uuid.UUID) []*domain.Trade {
	tradesByID := make([]*domain.Trade, 0, len(tradeIDs))
	for _, tradeID := range tradeIDs {
		trade := trades[tradeID]
		tradesByID = append(tradesByID, &trade)
	}
	return tradesByID
}

func (r TradeRepositoryImpl) addTradeByMarket(key string, val uuid.UUID) {
	trades, ok := r.db.tradeStore.tradesByMarket[key]
	if !ok {
		r.db.tradeStore.tradesByMarket[key] = []uuid.UUID{val}
		return
	}

	if !contain(trades, val) {
		r.db.tradeStore.tradesByMarket[key] = append(
			r.db.tradeStore.tradesByMarket[key],
			val,
		)
	}
}

func (r TradeRepositoryImpl) addTradeByTrader(key string, val uuid.UUID) {
	trades, ok := r.db.tradeStore.tradesByTrader[key]
	if !ok {
		r.db.tradeStore.tradesByTrader[key] = []uuid.UUID{val}
		return
	}

	if !contain(trades, val) {
		r.db.tradeStore.tradesByTrader[key] = append(
			r.db.tradeStore.tradesByTrader[key],
			val,
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
