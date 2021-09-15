package dbbadger

import (
	"context"

	"github.com/dgraph-io/badger/v2"
	"github.com/google/uuid"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/timshannon/badgerhold/v2"
)

const (
	TradeBadgerholdKeyPrefix = "bh_Trade"
)

//badgerhold internal implementation adds prefix to the key
var tradeTablePrefixKey = []byte(TradeBadgerholdKeyPrefix)

type tradeRepositoryImpl struct {
	store *badgerhold.Store
}

func NewTradeRepositoryImpl(store *badgerhold.Store) domain.TradeRepository {
	return tradeRepositoryImpl{store}
}

func (t tradeRepositoryImpl) GetOrCreateTrade(
	ctx context.Context,
	tradeID *uuid.UUID,
) (*domain.Trade, error) {
	return t.getOrCreateTrade(ctx, tradeID)
}

func (t tradeRepositoryImpl) GetAllTrades(
	ctx context.Context, page *domain.Page,
) ([]*domain.Trade, error) {
	query := &badgerhold.Query{}
	if page != nil {
		from := page.Number*page.Size - page.Size
		query.Skip(from).Limit(page.Size)
	}

	return t.findTrades(ctx, query)
}

func (t tradeRepositoryImpl) GetAllTradesByMarket(
	ctx context.Context, marketQuoteAsset string, page *domain.Page,
) ([]*domain.Trade, error) {
	query := badgerhold.Where("MarketQuoteAsset").Eq(marketQuoteAsset)
	if page != nil {
		from := page.Number*page.Size - page.Size
		query.Skip(from).Limit(page.Size)
	}

	return t.findTrades(ctx, query)
}

func (t tradeRepositoryImpl) GetCompletedTradesByMarket(
	ctx context.Context,
	marketQuoteAsset string,
	page *domain.Page,
) ([]*domain.Trade, error) {
	query := badgerhold.
		Where("MarketQuoteAsset").Eq(marketQuoteAsset).
		And("Status.Code").Ge(domain.Completed).
		And("Status.Failed").Eq(false)
	if page != nil {
		from := page.Number*page.Size - page.Size
		query.Skip(from).Limit(page.Size)
	}

	return t.findTrades(ctx, query)
}

func (t tradeRepositoryImpl) GetTradeBySwapAcceptID(
	ctx context.Context,
	swapAcceptID string,
) (*domain.Trade, error) {
	query := badgerhold.Where("SwapAccept.ID").Eq(swapAcceptID)

	trades, err := t.findTrades(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(trades) <= 0 {
		return nil, nil
	}

	return trades[0], nil
}

func (t tradeRepositoryImpl) GetTradeByTxID(
	ctx context.Context,
	txID string,
) (*domain.Trade, error) {
	query := badgerhold.Where("TxID").Eq(txID)

	trades, err := t.findTrades(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(trades) <= 0 {
		return nil, ErrTradeNotFound
	}

	return trades[0], nil
}

func (t tradeRepositoryImpl) UpdateTrade(
	ctx context.Context,
	ID *uuid.UUID,
	updateFn func(t *domain.Trade) (*domain.Trade, error),
) error {
	txIsNotGiven := ctx.Value("tx") == nil

	currentTrade, err := t.getTrade(ctx, *ID)
	if err != nil {
		return err
	}

	updatedTrade, err := updateFn(currentTrade)
	if err != nil {
		return err
	}

	for {
		err := t.updateTrade(ctx, updatedTrade.ID, *updatedTrade)
		if err != nil {
			if txIsNotGiven && isTransactionConflict(err) {
				continue
			}
		}
		return err
	}
}

func (t tradeRepositoryImpl) getOrCreateTrade(
	ctx context.Context,
	ID *uuid.UUID,
) (*domain.Trade, error) {
	if ID != nil {
		tr, err := t.getTrade(ctx, *ID)
		if err != nil {
			return nil, err
		}
		if tr != nil {
			return tr, nil
		}
	}

	trade := domain.NewTrade()
	if ID != nil {
		trade.ID = *ID
	}

	if err := t.insertTrade(ctx, *trade); err != nil {
		return nil, err
	}
	return trade, nil
}

func (t tradeRepositoryImpl) findTrades(
	ctx context.Context,
	query *badgerhold.Query,
) ([]*domain.Trade, error) {
	var tr []domain.Trade
	var err error
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = t.store.TxFind(tx, &tr, query)
	} else {
		err = t.store.Find(&tr, query)
	}
	if err != nil {
		return nil, err
	}

	trades := make([]*domain.Trade, 0, len(tr))
	for i := range tr {
		trade := tr[i]
		trades = append(trades, &trade)
	}

	return trades, nil
}

func (t tradeRepositoryImpl) getTrade(
	ctx context.Context,
	ID uuid.UUID,
) (*domain.Trade, error) {
	var trade domain.Trade
	var err error
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = t.store.TxGet(tx, ID, &trade)
	} else {
		err = t.store.Get(ID, &trade)
	}
	if err != nil {
		if err == badgerhold.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &trade, nil
}

func (t tradeRepositoryImpl) updateTrade(
	ctx context.Context,
	ID uuid.UUID,
	trade domain.Trade,
) error {
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		return t.store.TxUpdate(tx, ID, trade)
	}
	return t.store.Update(ID, trade)
}

func (t tradeRepositoryImpl) insertTrade(
	ctx context.Context,
	trade domain.Trade,
) error {
	var err error
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = t.store.TxInsert(tx, trade.ID, &trade)
	} else {
		err = t.store.Insert(trade.ID, &trade)
	}
	if err != nil {
		if err != badgerhold.ErrKeyExists {
			return err
		}
	}
	return nil
}
