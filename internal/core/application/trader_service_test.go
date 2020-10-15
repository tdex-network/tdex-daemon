package application

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/vulpemventures/go-elements/address"
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

type priceAndPreviewTestData struct {
	unspents           []domain.Unspent
	market             *domain.Market
	lbtcAmount         uint64
	expectedBuyAmount  uint64
	expectedSellAmount uint64
	expectedPrice      Price
}

func mocksForPriceAndPreview(withDefaultStrategy bool) (*priceAndPreviewTestData, error) {
	addr := "el1qqfmmhdayrxdqs60hecn6yzfzmpquwlhn5m39ytngr8gu63ar6zhqngyj0ak7n3jr8ypfz7s6v7nmnkdvmu8n5pev33ac5thm7"
	script, _ := address.ToOutputScript(addr, *config.GetNetwork())
	unspents := []domain.Unspent{
		{
			TxID:            "0000000000000000000000000000000000000000000000000000000000000000",
			VOut:            0,
			Value:           100000000,
			AssetHash:       config.GetNetwork().AssetID,
			ValueCommitment: "080000000000000000000000000000000000000000000000000000000000000000",
			AssetCommitment: "090000000000000000000000000000000000000000000000000000000000000000",
			ScriptPubKey:    script,
			Nonce:           make([]byte, 33),
			RangeProof:      make([]byte, 4174),
			SurjectionProof: make([]byte, 64),
			Address:         addr,
			Spent:           false,
			Locked:          false,
			LockedBy:        nil,
			Confirmed:       true,
		},
		{
			TxID:            "0000000000000000000000000000000000000000000000000000000000000000",
			VOut:            0,
			Value:           100000000,
			AssetHash:       config.GetNetwork().AssetID,
			ValueCommitment: "080000000000000000000000000000000000000000000000000000000000000000",
			AssetCommitment: "090000000000000000000000000000000000000000000000000000000000000000",
			ScriptPubKey:    script,
			Nonce:           make([]byte, 33),
			RangeProof:      make([]byte, 4174),
			SurjectionProof: make([]byte, 64),
			Address:         addr,
			Spent:           false,
			Locked:          false,
			LockedBy:        nil,
			Confirmed:       true,
		},
	}

	market, _ := domain.NewMarket(domain.MarketAccountStart)
	market.FundMarket([]domain.OutpointWithAsset{
		// LBTC
		domain.OutpointWithAsset{
			Asset: config.GetNetwork().AssetID,
			Txid:  "0000000000000000000000000000000000000000000000000000000000000000",
			Vout:  0,
		},
		// ASS
		domain.OutpointWithAsset{
			Asset: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Txid:  "0000000000000000000000000000000000000000000000000000000000000000",
			Vout:  1,
		},
	})

	bp, _ := decimal.NewFromString("0.00015385")
	qp, _ := decimal.NewFromString("6500")
	price := Price{
		BasePrice:  bp,
		QuotePrice: qp,
	}

	if withDefaultStrategy {
		market.MakeTradable()

		return &priceAndPreviewTestData{
			unspents:           unspents,
			market:             market,
			lbtcAmount:         10000, // 0.0001 LBTC
			expectedBuyAmount:  65169016,
			expectedSellAmount: 65155984,
			expectedPrice:      price,
		}, nil
	}

	market.MakeStrategyPluggable()
	market.ChangeBasePrice(bp)
	market.ChangeQuotePrice(qp)

	return &priceAndPreviewTestData{
		unspents:           unspents,
		market:             market,
		lbtcAmount:         10000, // 0.0001 LBTC
		expectedBuyAmount:  81250000,
		expectedSellAmount: 48750000,
		expectedPrice:      price,
	}, nil
}
