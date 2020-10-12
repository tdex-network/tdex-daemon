package dbbadger

import (
	"context"
	"github.com/dgraph-io/badger"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"os"
)

var ctx context.Context
var marketRepository domain.MarketRepository
var unspentRepository domain.UnspentRepository
var dbManager *DbManager
var testDbDir = "testdb"

func before() {
	var err error

	dbManager, err = NewDbManager(testDbDir)
	if err != nil {
		panic(err)
	}
	marketRepository = NewMarketRepositoryImpl(dbManager)
	unspentRepository = NewUnspentRepositoryImpl(dbManager)
	tx := dbManager.Store.Badger().NewTransaction(true)
	err = insertMarkets(tx, dbManager)
	if err != nil {
		panic(err)
	}
	if err := tx.Commit(); err != nil {
		panic(err)
	}

	ctx = context.WithValue(
		context.Background(),
		"tx",
		dbManager.Store.Badger().NewTransaction(true))
}

func after() {
	tx := ctx.Value("tx").(*badger.Txn)
	tx.Discard()
	dbManager.Store.Close()

	err := os.RemoveAll(testDbDir)
	if err != nil {
		panic(err)
	}
}

func insertMarkets(tx *badger.Txn, db *DbManager) error {
	markets := []Market{
		{
			AccountIndex: 5,
			BaseAsset:    "ah5",
			QuoteAsset:   "qh5",
			Fee:          0,
			FeeAsset:     "",
			Tradable:     true,
			Strategy:     1,
			BasePrice:    nil,
			QuotePrice:   nil,
		},
		{
			AccountIndex: 6,
			BaseAsset:    "ah6",
			QuoteAsset:   "qh6",
			Fee:          0,
			FeeAsset:     "",
			Tradable:     true,
			Strategy:     1,
			BasePrice:    nil,
			QuotePrice:   nil,
		},
		{
			AccountIndex: 7,
			BaseAsset:    "ah7",
			QuoteAsset:   "qh7",
			Fee:          0,
			FeeAsset:     "",
			Tradable:     false,
			Strategy:     1,
			BasePrice:    nil,
			QuotePrice:   nil,
		},
		{
			AccountIndex: 8,
			BaseAsset:    "ah8",
			QuoteAsset:   "qh8",
			Fee:          0,
			FeeAsset:     "",
			Tradable:     false,
			Strategy:     1,
			BasePrice:    nil,
			QuotePrice:   nil,
		},
		{
			AccountIndex: 9,
			BaseAsset:    "ah9",
			QuoteAsset:   "qh9",
			Fee:          0,
			FeeAsset:     "",
			Tradable:     false,
			Strategy:     1,
			BasePrice:    nil,
			QuotePrice:   nil,
		},
	}
	for _, v := range markets {
		err := db.Store.TxInsert(tx, v.AccountIndex, v)
		if err != nil {
			return err
		}
	}

	return nil
}
