package dbbadger

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

func TestGetCreateOrUpdate(t *testing.T) {
	before()
	defer after()

	market, err := marketRepository.GetOrCreateMarket(ctx, domain.MarketAccountStart)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "ah5", market.BaseAsset)

	market, err = marketRepository.GetOrCreateMarket(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "", market.BaseAsset)
}

func TestGetAll(t *testing.T) {
	before()
	defer after()
	market, err := marketRepository.GetAllMarkets(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 5, len(market))
}

func TestGetMarketByAsset(t *testing.T) {
	before()
	defer after()
	market, accountIndex, err := marketRepository.GetMarketByAsset(ctx, "qh7")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "ah7", market.BaseAsset)
	assert.Equal(t, 7, accountIndex)
}

func TestGetLatestMarket(t *testing.T) {
	before()
	defer after()
	market, accountIndex, err := marketRepository.GetLatestMarket(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "ah9", market.BaseAsset)
	assert.Equal(t, 9, accountIndex)
}

func TestTradableMarket(t *testing.T) {
	before()
	defer after()
	markets, err := marketRepository.GetTradableMarkets(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(markets))
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

	assert.Equal(t, false, market.IsTradable())
}

func TestOpenMarket(t *testing.T) {
	before()
	defer after()
	market, _, err := marketRepository.GetMarketByAsset(ctx, "qh9")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, false, market.IsTradable())

	err = marketRepository.OpenMarket(ctx, "qh9")
	if err != nil {
		t.Fatal(err)
	}

	market, _, err = marketRepository.GetMarketByAsset(ctx, "qh9")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, true, market.IsTradable())
}

func TestCloseMarket(t *testing.T) {
	before()
	defer after()
	market, _, err := marketRepository.GetMarketByAsset(ctx, "qh6")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, true, market.IsTradable())

	err = marketRepository.CloseMarket(ctx, "qh6")
	if err != nil {
		t.Fatal(err)
	}

	market, _, err = marketRepository.GetMarketByAsset(ctx, "qh6")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, false, market.IsTradable())
}

func TestUpdateMarketPrice(t *testing.T) {
	const basePriceString = "13000.5"
	const quotePriceString = "0.02"

	before()
	defer after()
	bp, _ := decimal.NewFromString(basePriceString)
	qp, _ := decimal.NewFromString(quotePriceString)
	err := marketRepository.UpdatePrices(ctx, 5, domain.Prices{BasePrice: bp, QuotePrice: qp})
	if err != nil {
		t.Fatal(err)
	}

	market, _, err := marketRepository.GetMarketByAsset(ctx, "qh5")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, basePriceString, market.Price.BasePrice.String())
	assert.Equal(t, quotePriceString, market.Price.QuotePrice.String())

}
