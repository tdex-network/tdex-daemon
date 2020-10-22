package application

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tdex-network/tdex-daemon/pkg/trade"
	"github.com/vulpemventures/go-elements/network"
)

func TestGetPriceAndPreviewForMarket(t *testing.T) {
	t.Run("with market default strategy", func(t *testing.T) {
		defaultStrategy := true
		tt, err := mocksForPriceAndPreview(defaultStrategy)
		if err != nil {
			t.Fatal(err)
		}

		price, previewAmount, err := getPriceAndPreviewForMarket(
			tt.unspents,
			tt.market,
			TradeBuy,
			tt.lbtcAmount,
		)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, int(tt.expectedBuyAmount), int(previewAmount))
		assert.Equal(t, tt.expectedPrice.BasePrice.String(), price.BasePrice.String())
		assert.Equal(t, tt.expectedPrice.QuotePrice.String(), price.QuotePrice.String())

		_, previewAmount, err = getPriceAndPreviewForMarket(
			tt.unspents,
			tt.market,
			TradeSell,
			tt.lbtcAmount,
		)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, int(tt.expectedSellAmount), int(previewAmount))
	})

	t.Run("with market pluggable strategy", func(t *testing.T) {
		defaultStrategy := false
		tt, err := mocksForPriceAndPreview(defaultStrategy)
		if err != nil {
			t.Fatal(err)
		}

		price, previewAmount, err := getPriceAndPreviewForMarket(
			tt.unspents,
			tt.market,
			TradeBuy,
			tt.lbtcAmount,
		)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, int(tt.expectedBuyAmount), int(previewAmount))
		assert.Equal(t, tt.expectedPrice.BasePrice.String(), price.BasePrice.String())
		assert.Equal(t, tt.expectedPrice.QuotePrice.String(), price.QuotePrice.String())

		_, previewAmount, err = getPriceAndPreviewForMarket(
			tt.unspents,
			tt.market,
			TradeSell,
			tt.lbtcAmount,
		)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, int(tt.expectedSellAmount), int(previewAmount))
	})
}

func TestGetTradableMarkets(t *testing.T) {
	traderSvc, ctx, close := newTestTrader()
	defer close()

	markets, err := traderSvc.GetTradableMarkets(ctx)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(markets))
}

func TestGetMarketPrice(t *testing.T) {
	traderSvc, ctx, close := newTestTrader()
	defer close()

	market := Market{
		BaseAsset:  marketUnspents[0].AssetHash,
		QuoteAsset: marketUnspents[1].AssetHash,
	}
	sellPreview, err := traderSvc.GetMarketPrice(ctx, market, TradeSell, 30000000)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, sellPreview)

	buyPreview, err := traderSvc.GetMarketPrice(ctx, market, TradeBuy, 30000000)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, buyPreview)
}

func TestTradePropose(t *testing.T) {
	traderSvc, ctx, close := newTestTrader()
	defer close()

	markets, err := traderSvc.GetTradableMarkets(ctx)
	if err != nil {
		t.Error(err)
	}

	market := markets[0].Market
	preview, err := traderSvc.GetMarketPrice(ctx, market, TradeSell, 30000000)
	if err != nil {
		t.Error(err)
	}

	proposerWallet, err := trade.NewRandomWallet(&network.Regtest)
	if err != nil {
		t.Error(err)
	}

	t.Run("propose rejected", func(t *testing.T) {
		swapRequest, err := newSwapRequest(
			proposerWallet,
			market.BaseAsset, 30000000,
			market.QuoteAsset, preview.Amount,
		)
		if err != nil {
			t.Error(err)
		}

		// alter the swap request params to make it invalid
		swapRequest.AmountR = 50000000
		swapAccept, swapFail, _, err := traderSvc.TradePropose(ctx, market, TradeSell, swapRequest)
		if err != nil {
			t.Error(err)
		}
		assert.NotNil(t, swapFail)
		assert.Nil(t, swapAccept)
	})

	t.Run("propose accepted", func(t *testing.T) {
		swapRequest, err := newSwapRequest(
			proposerWallet,
			market.BaseAsset, 30000000,
			market.QuoteAsset, preview.Amount,
		)
		if err != nil {
			t.Error(err)
		}

		swapAccept, swapFail, _, err := traderSvc.TradePropose(ctx, market, TradeSell, swapRequest)
		if err != nil {
			t.Error(err)
		}
		if swapFail != nil {
			t.Error(errors.New(swapFail.GetFailureMessage()))
		}
		assert.NotNil(t, swapAccept)
		assert.Nil(t, swapFail)

		swapComplete, err := newSwapComplete(proposerWallet, swapAccept)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, true, isFinalizableTransaction(swapComplete.GetTransaction()))
	})
}

// func TestTradeFailedToComplete(t *testing.T) {
// 	traderSvc, ctx, close := newTestTrader()
// 	defer close()

// 	txID, swapFail, err := traderSvc.TradeComplete(ctx, mockWrongSwapComplete(), nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	assert.Equal(t, true, len(txID) <= 0)
// 	assert.NotNil(t, swapFail)
// }
