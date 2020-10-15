package dbbadger

import (
	"github.com/magiconair/properties/assert"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"testing"
)

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
