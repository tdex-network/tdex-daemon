package inmemory

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"

	"github.com/google/uuid"
)

type tradeRepositoryImpl struct {
	store *tradeInmemoryStore
}

// NewTradeRepositoryImpl returns a new inmemory TradeRepository implementation.
func NewTradeRepositoryImpl(store *tradeInmemoryStore) domain.TradeRepository {
	return &tradeRepositoryImpl{store}
}

func (r tradeRepositoryImpl) GetOrCreateTrade(_ context.Context, tradeID *uuid.UUID) (*domain.Trade, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	return r.getOrCreateTrade(tradeID)
}

func (r tradeRepositoryImpl) GetAllTrades(_ context.Context) ([]*domain.Trade, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	return r.getAllTrades()
}

func (r tradeRepositoryImpl) GetAllTradesByMarket(_ context.Context, marketQuoteAsset string) ([]*domain.Trade, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	return r.getAllTradesByMarket(marketQuoteAsset)
}

func (r tradeRepositoryImpl) GetCompletedTradesByMarket(ctx context.Context, marketQuoteAsset string) ([]*domain.Trade, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	tradesByMarkets, err := r.getAllTradesByMarket(marketQuoteAsset)
	if err != nil {
		return nil, err
	}

	completedTrades := make([]*domain.Trade, 0)
	for _, trade := range tradesByMarkets {
		if trade.Status.Code >= domain.CompletedStatus.Code && !trade.Status.Failed {
			completedTrades = append(completedTrades, trade)
		}
	}

	return completedTrades, nil
}

func (r tradeRepositoryImpl) GetTradeByTxID(
	ctx context.Context,
	txID string,
) (*domain.Trade, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

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

func (r tradeRepositoryImpl) GetTradeBySwapAcceptID(_ context.Context, swapAcceptID string) (*domain.Trade, error) {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	return r.getTradeBySwapAcceptID(swapAcceptID)
}

func (r tradeRepositoryImpl) UpdateTrade(
	ctx context.Context,
	tradeID *uuid.UUID,
	updateFn func(t *domain.Trade) (*domain.Trade, error),
) error {
	r.store.locker.Lock()
	defer r.store.locker.Unlock()

	currentTrade, err := r.getOrCreateTrade(tradeID)
	if err != nil {
		return err
	}

	updatedTrade, err := updateFn(currentTrade)
	if err != nil {
		return err
	}

	if swapAccept := updatedTrade.SwapAcceptMessage(); swapAccept != nil {
		if _, ok := r.store.tradesBySwapAcceptID[swapAccept.GetId()]; !ok {
			r.store.tradesBySwapAcceptID[swapAccept.GetId()] = currentTrade.ID
		}
	}

	r.addTradeByMarket(updatedTrade.MarketQuoteAsset, currentTrade.ID)

	r.store.trades[updatedTrade.ID] = *updatedTrade

	return nil
}

func (r tradeRepositoryImpl) getOrCreateTrade(tradeID *uuid.UUID) (*domain.Trade, error) {
	if tradeID != nil {
		tr, ok := r.store.trades[*tradeID]
		if ok {
			return &tr, nil
		}
	}

	trade := domain.NewTrade()
	if tradeID != nil {
		trade.ID = *tradeID
	}
	r.store.trades[trade.ID] = *trade
	return trade, nil
}

func (r tradeRepositoryImpl) getAllTrades() ([]*domain.Trade, error) {
	allTrades := make([]*domain.Trade, 0)
	for _, trade := range r.store.trades {
		allTrades = append(allTrades, &trade)
	}
	return allTrades, nil
}

func (r tradeRepositoryImpl) getAllTradesByMarket(marketQuoteAsset string) ([]*domain.Trade, error) {
	tradeIDs, ok := r.store.tradesByMarket[marketQuoteAsset]
	if !ok {
		return nil, nil
	}

	tradeList := tradesFromIDs(r.store.trades, tradeIDs)
	return tradeList, nil
}

func (r tradeRepositoryImpl) getTradeBySwapAcceptID(swapAcceptID string) (*domain.Trade, error) {
	tradeID, ok := r.store.tradesBySwapAcceptID[swapAcceptID]
	if !ok {
		return nil, nil
	}
	trade := r.store.trades[tradeID]
	return &trade, nil
}

func (r tradeRepositoryImpl) addTradeByMarket(key string, val uuid.UUID) {
	trades, ok := r.store.tradesByMarket[key]
	if !ok {
		r.store.tradesByMarket[key] = []uuid.UUID{val}
		return
	}

	if !contain(trades, val) {
		r.store.tradesByMarket[key] = append(
			r.store.tradesByMarket[key],
			val,
		)
	}
}

func tradesFromIDs(trades map[uuid.UUID]domain.Trade, tradeIDs []uuid.UUID) []*domain.Trade {
	tradesByID := make([]*domain.Trade, 0, len(tradeIDs))
	for _, tradeID := range tradeIDs {
		trade := trades[tradeID]
		tradesByID = append(tradesByID, &trade)
	}
	return tradesByID
}

func contain(list []uuid.UUID, id uuid.UUID) bool {
	for _, l := range list {
		if id == l {
			return true
		}
	}
	return false
}
