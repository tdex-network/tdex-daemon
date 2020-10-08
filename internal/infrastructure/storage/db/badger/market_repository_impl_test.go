package dbbadger

import (
	"context"
	"github.com/dgraph-io/badger"
	"github.com/magiconair/properties/assert"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"os"
	"path/filepath"
	"testing"
)

var ctx context.Context
var marketRepository domain.MarketRepository
var dbManager *DbManager

func before() {
	var err error
	dbDir := filepath.Join(config.GetString(config.DataDirPathKey), "testdb")
	dbManager, err = NewDbManager(dbDir)
	if err != nil {
		panic(err)
	}
	marketRepository = NewMarketRepositoryImpl(dbManager)
	tx := dbManager.Store.Badger().NewTransaction(true)
	err = insertMarkets(tx, dbManager)
	if err != nil {
		panic(err)
	}
	if err := tx.Commit(); err != nil {
		panic(err)
	}

	ctx = context.WithValue(
		ctx,
		"tx",
		dbManager.Store.Badger().NewTransaction(true))
}

func after() {
	tx := ctx.Value("tx").(*badger.Txn)
	tx.Discard()
	dbManager.Store.Close()

	dbDir := filepath.Join(config.GetString(config.DataDirPathKey), "testdb")
	err := os.RemoveAll(dbDir)
	if err != nil {
		panic(err)
	}
}

func TestGetCreateOrUpdate(t *testing.T) {
	before()
	defer after()

	market, err := marketRepository.GetOrCreateMarket(ctx, domain.MarketAccountStart)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, market.BaseAssetHash(), "ah5")

	market, err = marketRepository.GetOrCreateMarket(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, market.FeeAsset(), config.GetString(config.BaseAssetKey))
}

func TestGetAll(t *testing.T) {
	before()
	defer after()
	market, err := marketRepository.GetAllMarkets(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(market), 5)
}

func TestGetMarketByAsset(t *testing.T) {
	before()
	defer after()
	market, accountIndex, err := marketRepository.GetMarketByAsset(ctx, "qh7")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, market.BaseAssetHash(), "ah7")
	assert.Equal(t, accountIndex, 7)
}

func TestGetLatestMarket(t *testing.T) {
	before()
	defer after()
	market, accountIndex, err := marketRepository.GetLatestMarket(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, market.BaseAssetHash(), "ah9")
	assert.Equal(t, accountIndex, 9)
}

func TestTradableMarket(t *testing.T) {
	before()
	defer after()
	markets, err := marketRepository.GetTradableMarkets(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(markets), 2)
}

func TestUpdateMarket(t *testing.T) {
	before()
	defer after()
	err := marketRepository.UpdateMarket(
		ctx,
		5,
		func(m *domain.Market) (*domain.Market, error) {
			err := m.MakeNotTradable()
			if err != nil {
				return nil, err
			}
			return m, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	market, _, err := marketRepository.GetMarketByAsset(ctx, "qh5")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, market.IsTradable(), false)
}

func TestOpenMarket(t *testing.T) {
	before()
	defer after()
	market, _, err := marketRepository.GetMarketByAsset(ctx, "qh9")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, market.IsTradable(), false)

	err = marketRepository.OpenMarket(ctx, "qh9")
	if err != nil {
		t.Fatal(err)
	}

	market, _, err = marketRepository.GetMarketByAsset(ctx, "qh9")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, market.IsTradable(), true)
}

func TestCloseMarket(t *testing.T) {
	before()
	defer after()
	market, _, err := marketRepository.GetMarketByAsset(ctx, "qh6")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, market.IsTradable(), true)

	err = marketRepository.CloseMarket(ctx, "qh6")
	if err != nil {
		t.Fatal(err)
	}

	market, _, err = marketRepository.GetMarketByAsset(ctx, "qh6")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, market.IsTradable(), false)
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
