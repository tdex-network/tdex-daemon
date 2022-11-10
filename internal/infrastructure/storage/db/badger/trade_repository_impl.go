package dbbadger

import (
	"context"
	"fmt"

	"github.com/dgraph-io/badger/v2"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/timshannon/badgerhold/v2"
)

type tradeRepositoryImpl struct {
	store *badgerhold.Store
}

func NewTradeRepositoryImpl(store *badgerhold.Store) domain.TradeRepository {
	return tradeRepositoryImpl{store}
}

func (t tradeRepositoryImpl) AddTrade(
	ctx context.Context, trade *domain.Trade,
) error {
	return t.insertTrade(ctx, *trade)
}

func (t tradeRepositoryImpl) GetTradeById(
	ctx context.Context, tradeId string,
) (*domain.Trade, error) {
	return t.getTrade(ctx, tradeId)
}

func (t tradeRepositoryImpl) GetAllTrades(
	ctx context.Context, page domain.Page,
) ([]domain.Trade, error) {
	query := &badgerhold.Query{}
	if page != nil {
		offset := int(page.GetNumber()*page.GetSize() - page.GetSize())
		query.Skip(offset).Limit(int(page.GetSize()))
	}
	return t.findTrades(ctx, query)
}

func (t tradeRepositoryImpl) GetAllTradesByMarket(
	ctx context.Context, marketName string, page domain.Page,
) ([]domain.Trade, error) {
	query := badgerhold.Where("MarketName").Eq(marketName)
	if page != nil {
		offset := int(page.GetNumber()*page.GetSize() - page.GetSize())
		query.Skip(offset).Limit(int(page.GetSize()))
	}
	return t.findTrades(ctx, query)
}

func (t tradeRepositoryImpl) GetCompletedTradesByMarket(
	ctx context.Context, marketName string, page domain.Page,
) ([]domain.Trade, error) {
	query := badgerhold.
		Where("MarketName").Eq(marketName).
		And("Status.Code").Ge(domain.TradeStatusCodeCompleted).
		And("Status.Failed").Eq(false)
	if page != nil {
		offset := int(page.GetNumber()*page.GetSize() - page.GetSize())
		query.Skip(offset).Limit(int(page.GetSize()))
	}
	return t.findTrades(ctx, query)
}

func (t tradeRepositoryImpl) GetTradeBySwapAcceptId(
	ctx context.Context, swapAcceptId string,
) (*domain.Trade, error) {
	query := badgerhold.Where("SwapAccept.Id").Eq(swapAcceptId)

	trades, err := t.findTrades(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(trades) <= 0 {
		return nil, fmt.Errorf(
			"trade with swap accept id %s not found", swapAcceptId,
		)
	}

	return &trades[0], nil
}

func (t tradeRepositoryImpl) GetTradeByTxId(
	ctx context.Context, txId string,
) (*domain.Trade, error) {
	query := badgerhold.Where("TxId").Eq(txId)

	trades, err := t.findTrades(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(trades) <= 0 {
		return nil, nil
	}

	return &trades[0], nil
}

func (t tradeRepositoryImpl) UpdateTrade(
	ctx context.Context,
	id string, updateFn func(t *domain.Trade) (*domain.Trade, error),
) error {
	txIsNotGiven := ctx.Value("tx") == nil

	currentTrade, err := t.getTrade(ctx, id)
	if err != nil {
		return err
	}

	updatedTrade, err := updateFn(currentTrade)
	if err != nil {
		return err
	}

	for {
		err := t.updateTrade(ctx, updatedTrade.Id, *updatedTrade)
		if err != nil {
			if txIsNotGiven && isTransactionConflict(err) {
				continue
			}
		}
		return err
	}
}

func (t tradeRepositoryImpl) findTrades(
	ctx context.Context, query *badgerhold.Query,
) ([]domain.Trade, error) {
	var trades []domain.Trade
	var err error

	query.SortBy("SwapRequest.Timestamp").Reverse()
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = t.store.TxFind(tx, &trades, query)
	} else {
		err = t.store.Find(&trades, query)
	}
	if err != nil {
		return nil, err
	}

	return trades, nil
}

func (t tradeRepositoryImpl) getTrade(
	ctx context.Context, id string,
) (*domain.Trade, error) {
	var trade domain.Trade
	var err error
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = t.store.TxGet(tx, id, &trade)
	} else {
		err = t.store.Get(id, &trade)
	}
	if err != nil {
		if err == badgerhold.ErrNotFound {
			return nil, fmt.Errorf("trade with id %s not found", id)
		}
		return nil, err
	}

	return &trade, nil
}

func (t tradeRepositoryImpl) updateTrade(
	ctx context.Context, id string, trade domain.Trade,
) error {
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		return t.store.TxUpdate(tx, id, trade)
	}
	return t.store.Update(id, trade)
}

func (t tradeRepositoryImpl) insertTrade(
	ctx context.Context, trade domain.Trade,
) error {
	var err error
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = t.store.TxInsert(tx, trade.Id, &trade)
	} else {
		err = t.store.Insert(trade.Id, &trade)
	}
	if err != nil {
		if err == badgerhold.ErrKeyExists {
			return fmt.Errorf("trade with id %s already exists", trade.Id)
		}
		return err
	}
	return nil
}
