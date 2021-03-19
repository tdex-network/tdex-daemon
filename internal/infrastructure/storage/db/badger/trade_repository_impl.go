package dbbadger

import (
	"context"
	"errors"

	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"

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
	return t.getOrCreateTrade(ctx, tradeID)
}

func (t tradeRepositoryImpl) GetAllTrades(
	ctx context.Context,
) ([]*domain.Trade, error) {
	return t.getAllTrades(ctx), nil
}

func (t tradeRepositoryImpl) GetAllTradesByMarket(
	ctx context.Context,
	marketQuoteAsset string,
) ([]*domain.Trade, error) {
	query := badgerhold.Where("MarketQuoteAsset").Eq(marketQuoteAsset)
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
		return nil, errors.New("trade not found")
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
		return nil, errors.New("trade not found")
	}

	return trades[0], nil
}

func (t tradeRepositoryImpl) UpdateTrade(
	ctx context.Context,
	ID *uuid.UUID,
	updateFn func(t *domain.Trade) (*domain.Trade, error),
) error {
	txIsNotGiven := ctx.Value("tx") == nil

	currentTrade, err := t.getOrCreateTrade(ctx, ID)
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

func (t tradeRepositoryImpl) GetCompletedTradesByMarket(
	ctx context.Context,
	marketQuoteAsset string,
) ([]*domain.Trade, error) {
	query := badgerhold.
		Where("MarketQuoteAsset").Eq(marketQuoteAsset).
		And("Status.Code").Eq(pb.SwapStatus_COMPLETE)
	return t.findTrades(ctx, query)
}

func (t tradeRepositoryImpl) getOrCreateTrade(
	ctx context.Context,
	ID *uuid.UUID,
) (*domain.Trade, error) {
	if ID != nil {
		return t.getTrade(ctx, *ID)
	}

	trade := domain.NewTrade()
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
		err = t.db.Store.TxFind(tx, &tr, query)
	} else {
		err = t.db.Store.Find(&tr, query)
	}

	trades := make([]*domain.Trade, len(tr), len(tr))
	for i := range tr {
		trade := tr[i]
		trades[i] = &trade
	}

	return trades, err
}

func (t tradeRepositoryImpl) getTrade(
	ctx context.Context,
	ID uuid.UUID,
) (*domain.Trade, error) {
	var trade domain.Trade
	var err error
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = t.db.Store.TxGet(tx, ID, &trade)
	} else {
		err = t.db.Store.Get(ID, &trade)
	}
	if err != nil {
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
		return t.db.Store.TxUpdate(tx, ID, trade)
	}
	return t.db.Store.Update(ID, trade)
}

func (t tradeRepositoryImpl) insertTrade(
	ctx context.Context,
	trade domain.Trade,
) error {
	var err error
	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		err = t.db.Store.TxInsert(tx, trade.ID, &trade)
	} else {
		err = t.db.Store.Insert(trade.ID, &trade)
	}
	if err != nil {
		if err != badgerhold.ErrKeyExists {
			return err
		}
	}
	return nil
}

func (t tradeRepositoryImpl) getAllTrades(ctx context.Context) []*domain.Trade {
	scan := func(tx *badger.Txn) []*domain.Trade {
		trades := make([]*domain.Trade, 0)

		iter := badger.DefaultIteratorOptions
		iter.PrefetchValues = false
		it := tx.NewIterator(iter)
		defer it.Close()

		for it.Seek(tradeTablePrefixKey); it.ValidForPrefix(tradeTablePrefixKey); it.Next() {
			item := it.Item()
			data, _ := item.ValueCopy(nil)
			var trade domain.Trade
			err := badgerhold.DefaultDecode(data, &trade)
			if err == nil {
				trades = append(trades, &trade)
			}
		}
		return trades
	}

	if ctx.Value("tx") != nil {
		tx := ctx.Value("tx").(*badger.Txn)
		return scan(tx)
	}

	var trades []*domain.Trade
	t.db.Store.Badger().View(func(tx *badger.Txn) error {
		trades = scan(tx)
		return nil
	})
	return trades
}
