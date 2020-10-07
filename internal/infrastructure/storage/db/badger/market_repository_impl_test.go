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
	dbDir := filepath.Join(config.GetString(config.TdexDir), "testdb")
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

	dbDir := filepath.Join(config.GetString(config.TdexDir), "testdb")
	err := os.RemoveAll(dbDir)
	if err != nil {
		panic(err)
	}
}

func TestGetCreateOrUpdate(t *testing.T) {
	before()
	defer after()

	market, err := marketRepository.GetOrCreateMarket(ctx, 0)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, market.BaseAssetHash(), "ah0")

	market, err = marketRepository.GetOrCreateMarket(ctx, 5)
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
	market, accountIndex, err := marketRepository.GetMarketByAsset(ctx, "qh3")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, market.BaseAssetHash(), "ah3")
	assert.Equal(t, accountIndex, 3)
}

func TestGetLatestMarket(t *testing.T) {
	before()
	defer after()
	market, accountIndex, err := marketRepository.GetLatestMarket(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, market.BaseAssetHash(), "ah4")
	assert.Equal(t, accountIndex, 4)
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
		0,
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

	market, _, err := marketRepository.GetMarketByAsset(ctx, "qh0")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, market.IsTradable(), false)
}

func TestOpenMarket(t *testing.T) {
	before()
	defer after()
	market, _, err := marketRepository.GetMarketByAsset(ctx, "qh4")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, market.IsTradable(), false)

	err = marketRepository.OpenMarket(ctx, "qh4")
	if err != nil {
		t.Fatal(err)
	}

	market, _, err = marketRepository.GetMarketByAsset(ctx, "qh4")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, market.IsTradable(), true)
}

func TestCloseMarket(t *testing.T) {
	before()
	defer after()
	market, _, err := marketRepository.GetMarketByAsset(ctx, "qh1")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, market.IsTradable(), true)

	err = marketRepository.CloseMarket(ctx, "qh1")
	if err != nil {
		t.Fatal(err)
	}

	market, _, err = marketRepository.GetMarketByAsset(ctx, "qh1")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, market.IsTradable(), false)
}

func insertMarkets(tx *badger.Txn, db *DbManager) error {
	markets := []Market{
		{
			AccountIndex: 0,
			BaseAsset:    "ah0",
			QuoteAsset:   "qh0",
			Fee:          0,
			FeeAsset:     "",
			Tradable:     true,
			Strategy:     1,
			BasePrice:    nil,
			QuotePrice:   nil,
		},
		{
			AccountIndex: 1,
			BaseAsset:    "ah1",
			QuoteAsset:   "qh1",
			Fee:          0,
			FeeAsset:     "",
			Tradable:     true,
			Strategy:     1,
			BasePrice:    nil,
			QuotePrice:   nil,
		},
		{
			AccountIndex: 2,
			BaseAsset:    "ah2",
			QuoteAsset:   "qh2",
			Fee:          0,
			FeeAsset:     "",
			Tradable:     false,
			Strategy:     1,
			BasePrice:    nil,
			QuotePrice:   nil,
		},
		{
			AccountIndex: 3,
			BaseAsset:    "ah3",
			QuoteAsset:   "qh3",
			Fee:          0,
			FeeAsset:     "",
			Tradable:     false,
			Strategy:     1,
			BasePrice:    nil,
			QuotePrice:   nil,
		},
		{
			AccountIndex: 4,
			BaseAsset:    "ah4",
			QuoteAsset:   "qh4",
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
