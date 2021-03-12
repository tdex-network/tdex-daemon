package inmemory

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"

	"github.com/google/uuid"
)

type tradeRepositoryImpl struct {
	db *DbManager
}

// NewTradeRepositoryImpl returns a new inmemory TradeRepository implementation.
func NewTradeRepositoryImpl(db *DbManager) domain.TradeRepository {
	return &tradeRepositoryImpl{
		db: db,
	}
}

func (r tradeRepositoryImpl) GetOrCreateTrade(_ context.Context, tradeID *uuid.UUID) (*domain.Trade, error) {
	r.db.tradeStore.locker.Lock()
	defer r.db.tradeStore.locker.Unlock()

	return r.getOrCreateTrade(tradeID)
}

func (r tradeRepositoryImpl) GetAllTrades(_ context.Context) ([]*domain.Trade, error) {
	r.db.tradeStore.locker.Lock()
	defer r.db.tradeStore.locker.Unlock()

	return r.getAllTrades()
}

func (r tradeRepositoryImpl) GetAllTradesForMarket(_ context.Context, marketQuoteAsset string) ([]*domain.Trade, error) {
	r.db.tradeStore.locker.Lock()
	defer r.db.tradeStore.locker.Unlock()

	return r.getAllTradesForMarket(marketQuoteAsset)
}

func (r tradeRepositoryImpl) GetCompletedTradesForMarket(ctx context.Context, marketQuoteAsset string) ([]*domain.Trade, error) {
	r.db.tradeStore.locker.Lock()
	defer r.db.tradeStore.locker.Unlock()

	tradesByMarkets, err := r.getAllTradesForMarket(marketQuoteAsset)
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

func (r tradeRepositoryImpl) GetTradeWithTxID(
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

func (r tradeRepositoryImpl) GetTradeWithSwapAcceptID(_ context.Context, swapAcceptID string) (*domain.Trade, error) {
	r.db.tradeStore.locker.Lock()
	defer r.db.tradeStore.locker.Unlock()

	return r.getTradeWithSwapAcceptID(swapAcceptID)
}

func (r tradeRepositoryImpl) UpdateTrade(
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

	r.addTradeWithMarket(updatedTrade.MarketQuoteAsset, currentTrade.ID)

	r.db.tradeStore.trades[updatedTrade.ID] = *updatedTrade

	return nil
}

func (r tradeRepositoryImpl) getOrCreateTrade(tradeID *uuid.UUID) (*domain.Trade, error) {
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
		return nil, ErrTradeNotFound
	}

	return &currentTrade, nil
}

func (r tradeRepositoryImpl) getAllTrades() ([]*domain.Trade, error) {
	allTrades := make([]*domain.Trade, 0)
	for _, trade := range r.db.tradeStore.trades {
		allTrades = append(allTrades, &trade)
	}
	return allTrades, nil
}

func (r tradeRepositoryImpl) getAllTradesForMarket(marketQuoteAsset string) ([]*domain.Trade, error) {
	tradeIDs, ok := r.db.tradeStore.tradesByMarket[marketQuoteAsset]
	if !ok {
		return nil, nil
	}

	tradeList := tradesFromIDs(r.db.tradeStore.trades, tradeIDs)
	return tradeList, nil
}

func (r tradeRepositoryImpl) getTradeWithSwapAcceptID(swapAcceptID string) (*domain.Trade, error) {
	tradeID, ok := r.db.tradeStore.tradesBySwapAcceptID[swapAcceptID]
	if !ok {
		return nil, nil
	}
	trade := r.db.tradeStore.trades[tradeID]
	return &trade, nil
}

func (r tradeRepositoryImpl) addTradeWithMarket(key string, val uuid.UUID) {
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
