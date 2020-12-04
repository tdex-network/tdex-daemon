package application

import (
	"errors"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tdex-network/tdex-daemon/pkg/trade"
	"github.com/vulpemventures/go-elements/network"
)

const marketWithPluggableStrategy = true

func TestGetTradableMarkets(t *testing.T) {
	traderSvc, ctx, close := newTestTrader(!marketWithPluggableStrategy)

	markets, err := traderSvc.GetTradableMarkets(ctx)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(markets))

	t.Cleanup(close)
}

func TestGetMarketPrice_BalancedStrategy(t *testing.T) {
	traderSvc, ctx, close := newTestTrader(!marketWithPluggableStrategy)

	market := Market{
		BaseAsset:  marketUnspents[0].AssetHash,
		QuoteAsset: marketUnspents[1].AssetHash,
	}

	// market reserves: 1 LBTC, 6500 USDT,
	// SELL trade:
	//   - if one wants to SELL 0.3 LBTC he gets back 1500 USDT, and with a
	// 		 market fee of the 0,25% he gets 1496,25 USDT.
	// 	 - if one wants to receive 1500 USDT, he needs to SELL 0.3 LBTC, and
	// 		 with a market fee of the 0,25% he has to SELL 0.30075 LBTC.
	// BUY trade:
	//   - if one wants to BUY 0.3 LBTC he has to send 2785,71428571 USDT, and with a
	// 		 market fee of the 0,25% he needs to send 2792,67857142 USDT.
	// 	 - if one wants to send 2785,71428571 USDT, he's BUYing 0.3 LBTC, and
	// 		 with a market fee of the 0,25% he BUYs 0,29977329 LBTC

	tests := []struct {
		tradeType                int
		amount                   uint64
		asset                    string
		previewAmountWithoutFees int
		expectedDiff             int
	}{
		{
			TradeSell,
			30000000,
			market.BaseAsset,
			150000000000,
			288626500,
		},
		{
			TradeSell,
			150000000000,
			market.QuoteAsset,
			30000000,
			97573,
		},
		{
			TradeBuy,
			30000000,
			market.BaseAsset,
			278571428571,
			995963429,
		},
		{
			TradeBuy,
			278571428571,
			market.QuoteAsset,
			30000000,
			52539,
		},
	}

	for _, tt := range tests {
		preview, err := traderSvc.GetMarketPrice(
			ctx, market, tt.tradeType, tt.amount, tt.asset,
		)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(
			t,
			tt.expectedDiff,
			int(math.Abs(float64(tt.previewAmountWithoutFees)-float64(preview.Amount))),
		)
	}

	t.Cleanup(close)
}
func TestGetMarketPrice_PluggableStrategy(t *testing.T) {
	traderSvc, ctx, close := newTestTrader(marketWithPluggableStrategy)

	market := Market{
		BaseAsset:  marketUnspents[0].AssetHash,
		QuoteAsset: marketUnspents[1].AssetHash,
	}

	// market prices: LBTC/USDT = 0.00015385, USDT/LBTC = 6500
	// SELL trade:
	//   - if one wants to SELL 0.3 LBTC he gets back 1950 USDT, and with a
	// 		 market fee of the 0.25% he gets 1462.5 USDT.
	// 	 - if one wants to receive 1950 USDT, he needs to SELL 0.3000075 LBTC, and
	// 		 with a market fee of the 0,25% he has to SELL 0.37500938 LBTC.
	// BUY trade:
	//   - if one wants to BUY 0.3 LBTC he has to send 1950 USDT, and with a
	// 		 market fee of the 0.25% he needs to send 2437.5 USDT.
	// 	 - if one wants to send 1950 USDT, he's BUYing 0.3000075 LBTC, and
	// 		 with a market fee of the 0.25% he BUYs 0.37500938 LBTC

	tests := []struct {
		tradeType                int
		amount                   uint64
		asset                    string
		previewAmountWithoutFees int
		expectedDiff             int
	}{
		{
			TradeSell,
			30000000,
			market.BaseAsset,
			195000000000,
			487500000,
		},
		{
			TradeSell,
			195000000000,
			market.QuoteAsset,
			30000750,
			75001,
		},
		{
			TradeBuy,
			30000000,
			market.BaseAsset,
			195000000000,
			487500000,
		},
		{
			TradeBuy,
			195000000000,
			market.QuoteAsset,
			30000750,
			75002,
		},
	}

	for _, tt := range tests {
		preview, err := traderSvc.GetMarketPrice(
			ctx, market, tt.tradeType, tt.amount, tt.asset,
		)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(
			t,
			tt.expectedDiff,
			int(math.Abs(float64(tt.previewAmountWithoutFees)-float64(preview.Amount))),
		)
	}

	t.Cleanup(close)
}

func TestTradePropose(t *testing.T) {
	traderSvc, ctx, close := newTestTrader(!marketWithPluggableStrategy)

	markets, err := traderSvc.GetTradableMarkets(ctx)
	if err != nil {
		t.Fatal(err)
	}

	market := markets[0].Market
	preview, err := traderSvc.GetMarketPrice(ctx, market, TradeSell, 30000000, market.BaseAsset)
	if err != nil {
		t.Fatal(err)
	}

	proposerWallet, err := trade.NewRandomWallet(&network.Regtest)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("propose rejected", func(t *testing.T) {
		swapRequest, err := newSwapRequest(proposerWallet, market.BaseAsset, 30000000, market.QuoteAsset, preview.Amount)
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
			market.BaseAsset, 30000000, //assetP, amountP
			market.QuoteAsset, preview.Amount, // assetR, amountR
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

	t.Cleanup(close)
}
