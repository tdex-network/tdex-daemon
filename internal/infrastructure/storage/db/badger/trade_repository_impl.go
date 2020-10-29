package dbbadger

import (
	"context"
	"errors"

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
	db *DbManager
}

func NewTradeRepositoryImpl(db *DbManager) domain.TradeRepository {
	return tradeRepositoryImpl{
		db: db,
	}
}

func (t tradeRepositoryImpl) GetOrCreateTrade(
	ctx context.Context,
	tradeID *uuid.UUID,
) (*domain.Trade, error) {
	tx := ctx.Value("tx").(*badger.Txn)
	return t.getOrCreateTrade(tx, tradeID)
}

func (t tradeRepositoryImpl) GetAllTrades(
	ctx context.Context,
) ([]*domain.Trade, error) {
	tx := ctx.Value("tx").(*badger.Txn)
	if tx == nil {
		return nil, errors.New("context's transaction is nil")
	}

	return t.getAllTrades(tx), nil
}

func (t tradeRepositoryImpl) GetAllTradesByMarket(
	ctx context.Context,
	marketQuoteAsset string,
) ([]*domain.Trade, error) {
	tx := ctx.Value("tx").(*badger.Txn)
	query := badgerhold.Where("MarketQuoteAsset").Eq(marketQuoteAsset)
	tr, err := t.findTrades(tx, query)
	if err != nil {
		return nil, err
	}
	trades := make([]*domain.Trade, 0, len(tr))
	for _, v := range tr {
		trades = append(trades, &v)
	}

	return trades, nil
}

func (t tradeRepositoryImpl) GetTradeBySwapAcceptID(
	ctx context.Context,
	swapAcceptID string,
) (*domain.Trade, error) {
	tx := ctx.Value("tx").(*badger.Txn)
	query := badgerhold.Where("SwapAccept.ID").Eq(swapAcceptID)

	trades, err := t.findTrades(tx, query)
	if err != nil {
		return nil, err
	}

	if len(trades) <= 0 {
		return nil, errors.New("trade not found")
	}

	trade := &trades[0]
	return trade, nil
}

func (t tradeRepositoryImpl) UpdateTrade(
	ctx context.Context,
	ID *uuid.UUID,
	updateFn func(t *domain.Trade) (*domain.Trade, error),
) error {
	tx := ctx.Value("tx").(*badger.Txn)

	currentTrade, err := t.getOrCreateTrade(tx, ID)
	if err != nil {
		return err
	}

	updatedTrade, err := updateFn(currentTrade)
	if err != nil {
		return err
	}

	return t.updateTrade(tx, updatedTrade.ID, *updatedTrade)
}

func (t tradeRepositoryImpl) getOrCreateTrade(
	tx *badger.Txn,
	ID *uuid.UUID,
) (*domain.Trade, error) {
	if ID != nil {
		return t.getTrade(tx, *ID)
	}

	trade := domain.NewTrade()
	if err := t.insertTrade(tx, *trade); err != nil {
		return nil, err
	}
	return trade, nil
}

func (t tradeRepositoryImpl) findTrades(
	tx *badger.Txn,
	query *badgerhold.Query,
) ([]domain.Trade, error) {
	var trades []domain.Trade
	err := t.db.Store.TxFind(
		tx,
		&trades,
		query,
	)

	return trades, err
}

func (t tradeRepositoryImpl) getTrade(
	tx *badger.Txn,
	ID uuid.UUID,
) (*domain.Trade, error) {
	var trade domain.Trade
	err := t.db.Store.TxGet(
		tx,
		ID,
		&trade,
	)
	if err != nil {
		return nil, err
	}

	return &trade, nil
}

func (t tradeRepositoryImpl) updateTrade(
	tx *badger.Txn,
	ID uuid.UUID,
	trade domain.Trade,
) error {
	return t.db.Store.TxUpdate(
		tx,
		ID,
		trade,
	)
}

func (t tradeRepositoryImpl) insertTrade(
	tx *badger.Txn,
	trade domain.Trade,
) error {
	if err := t.db.Store.TxInsert(
		tx,
		trade.ID,
		&trade,
	); err != nil {
		if err != badgerhold.ErrKeyExists {
			return err
		}
	}
	return nil
}

func (t tradeRepositoryImpl) getAllTrades(tx *badger.Txn) []*domain.Trade {
	trades := make([]*domain.Trade, 0)

	iter := badger.DefaultIteratorOptions
	iter.PrefetchValues = false
	it := tx.NewIterator(iter)
	defer it.Close()

	for it.Seek(tradeTablePrefixKey); it.ValidForPrefix(tradeTablePrefixKey); it.Next() {
		item := it.Item()
		data, _ := item.ValueCopy(nil)
		var trade domain.Trade
		err := JSONDecode(data, &trade)
		if err == nil {
			trades = append(trades, &trade)
		}
	}

	return trades
}
