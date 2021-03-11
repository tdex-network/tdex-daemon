package domain_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

func TestFundMarket(t *testing.T) {
	t.Parallel()

	market := newTestMarket()
	baseAsset := "0000000000000000000000000000000000000000000000000000000000000000"
	quoteAsset := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	outpoints := []domain.OutpointWithAsset{
		{
			Asset: baseAsset,
			Txid:  "0000000000000000000000000000000000000000000000000000000000000000",
			Vout:  0,
		},
		{
			Asset: quoteAsset,
			Txid:  "0000000000000000000000000000000000000000000000000000000000000000",
			Vout:  1,
		},
	}

	err := market.FundMarket(outpoints, baseAsset)
	require.NoError(t, err)
	require.Equal(t, baseAsset, market.BaseAsset)
	require.Equal(t, quoteAsset, market.QuoteAsset)
	require.True(t, market.IsFunded())
}

func TestFailingFundMarket(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		market        *domain.Market
		baseAsset     string
		outpoints     []domain.OutpointWithAsset
		expectedError error
	}{
		{
			name:      "missing_quote_asset",
			market:    newTestMarket(),
			baseAsset: "0000000000000000000000000000000000000000000000000000000000000000",
			outpoints: []domain.OutpointWithAsset{
				{
					Asset: "0000000000000000000000000000000000000000000000000000000000000000",
					Txid:  "0000000000000000000000000000000000000000000000000000000000000000",
					Vout:  0,
				},
			},
			expectedError: domain.ErrMarketMissingQuoteAsset,
		},
		{
			name:      "missing_base_asset",
			market:    newTestMarket(),
			baseAsset: "0000000000000000000000000000000000000000000000000000000000000000",
			outpoints: []domain.OutpointWithAsset{
				{
					Asset: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					Txid:  "0000000000000000000000000000000000000000000000000000000000000000",
					Vout:  1,
				},
			},
			expectedError: domain.ErrMarketMissingBaseAsset,
		},
		{
			name:      "to_many_assets",
			market:    newTestMarket(),
			baseAsset: "0000000000000000000000000000000000000000000000000000000000000000",
			outpoints: []domain.OutpointWithAsset{
				{
					Asset: "0000000000000000000000000000000000000000000000000000000000000000",
					Txid:  "0000000000000000000000000000000000000000000000000000000000000000",
					Vout:  0,
				},
				{
					Asset: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					Txid:  "0000000000000000000000000000000000000000000000000000000000000000",
					Vout:  1,
				},
				{
					Asset: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
					Txid:  "0000000000000000000000000000000000000000000000000000000000000000",
					Vout:  2,
				},
			},
			expectedError: domain.ErrMarketTooManyAssets,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.market.FundMarket(tt.outpoints, tt.baseAsset)
			require.EqualError(t, err, tt.expectedError.Error())
		})
	}
}

func TestMakeTradable(t *testing.T) {
	t.Parallel()

	m := newTestMarketFunded()

	err := m.MakeTradable()
	require.NoError(t, err)
	require.True(t, m.IsTradable())
}

func TestFailingMakeTradable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		market        *domain.Market
		expectedError error
	}{
		{
			name:          "not_funded",
			market:        newTestMarket(),
			expectedError: domain.ErrMarketNotFunded,
		},
		{
			name:          "not_priced",
			market:        newTestMarketFundedWithPluggableStrategy(),
			expectedError: domain.ErrMarketNotPriced,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.market.MakeTradable()
			require.EqualError(t, err, tt.expectedError.Error())
		})
	}
}

func TestMakeNotTradable(t *testing.T) {
	t.Parallel()

	m := newTestMarketTradable()

	err := m.MakeNotTradable()
	require.NoError(t, err)
	require.False(t, m.IsTradable())
}

func TestFailingMakeNotTradable(t *testing.T) {
	t.Parallel()

	m := newTestMarket()

	err := m.MakeNotTradable()
	require.EqualError(t, err, domain.ErrMarketNotFunded.Error())
}

func TestMakeStrategyPluggable(t *testing.T) {
	t.Parallel()

	m := newTestMarketFunded()

	err := m.MakeStrategyPluggable()
	require.NoError(t, err)
	require.True(t, m.IsStrategyPluggable())
	require.False(t, m.IsStrategyPluggableInitialized())
}

func TestFailingMakeStrategyPluggable(t *testing.T) {
	t.Parallel()

	m := newTestMarketTradable()

	err := m.MakeStrategyPluggable()
	require.EqualError(t, err, domain.ErrMarketMustBeClosed.Error())
}

func TestMakeStrategyBalanced(t *testing.T) {
	t.Parallel()

	m := newTestMarketFundedWithPluggableStrategy()

	err := m.MakeStrategyBalanced()
	require.NoError(t, err)
	require.False(t, m.IsStrategyPluggable())
}

func TestFailingMakeStrategyBalanced(t *testing.T) {
	t.Parallel()

	m := newTestMarketTradable()

	err := m.MakeStrategyBalanced()
	require.EqualError(t, err, domain.ErrMarketMustBeClosed.Error())
}

func TestChangeFee(t *testing.T) {
	t.Parallel()

	m := newTestMarketFunded()
	newFee := int64(50)

	err := m.ChangeFee(newFee)
	require.NoError(t, err)
	require.Equal(t, newFee, m.Fee)
}

func TestFailingChangeFee(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		market        *domain.Market
		marketFee     int64
		expectedError error
	}{
		{
			name:          "not_funded",
			market:        newTestMarket(),
			marketFee:     50,
			expectedError: domain.ErrMarketNotFunded,
		},
		{
			name:          "must_be_closed",
			market:        newTestMarketTradable(),
			marketFee:     50,
			expectedError: domain.ErrMarketMustBeClosed,
		},
		{
			name:          "fee_too_low",
			market:        newTestMarketFunded(),
			marketFee:     0,
			expectedError: domain.ErrMarketFeeTooLow,
		},
		{
			name:          "fee_too_high",
			market:        newTestMarketFunded(),
			marketFee:     10000,
			expectedError: domain.ErrMarketFeeTooHigh,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.market.ChangeFee(tt.marketFee)
			require.EqualError(t, err, tt.expectedError.Error())
		})
	}
}

func TestChangeMarketPrices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		market     *domain.Market
		basePrice  decimal.Decimal
		quotePrice decimal.Decimal
	}{
		{
			name:       "change_prices_with_balanced_strategy",
			market:     newTestMarketFunded(),
			basePrice:  decimal.NewFromFloat(0.00002),
			quotePrice: decimal.NewFromFloat(50000),
		},
		{
			name:       "change_prices_with_pluggable_strategy",
			market:     newTestMarketFundedWithPluggableStrategy(),
			basePrice:  decimal.NewFromFloat(0.00002),
			quotePrice: decimal.NewFromFloat(50000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.market.ChangeBasePrice(tt.basePrice)
			require.NoError(t, err)
			require.Equal(t, tt.basePrice, tt.market.BaseAssetPrice())

			err = tt.market.ChangeQuotePrice(tt.quotePrice)
			require.NoError(t, err)
			require.Equal(t, tt.quotePrice, tt.market.QuoteAssetPrice())
			require.True(t, tt.market.IsStrategyPluggableInitialized())
		})
	}
}

func TestFailingChangeBasePrice(t *testing.T) {
	t.Parallel()

	m := newTestMarket()

	err := m.ChangeBasePrice(decimal.NewFromFloat(0.0002))
	require.EqualError(t, err, domain.ErrMarketNotFunded.Error())
}

func TestFailingChangeQuotePrice(t *testing.T) {
	t.Parallel()

	m := newTestMarket()

	err := m.ChangeQuotePrice(decimal.NewFromFloat(50000))
	require.EqualError(t, err, domain.ErrMarketNotFunded.Error())
}

func newTestMarket() *domain.Market {
	m, _ := domain.NewMarket(0, 25)
	return m
}

func newTestMarketFunded() *domain.Market {
	baseAsset := "0000000000000000000000000000000000000000000000000000000000000000"
	quoteAsset := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	outpoints := []domain.OutpointWithAsset{
		{
			Asset: baseAsset,
			Txid:  "0000000000000000000000000000000000000000000000000000000000000000",
			Vout:  0,
		},
		{
			Asset: quoteAsset,
			Txid:  "0000000000000000000000000000000000000000000000000000000000000000",
			Vout:  1,
		},
	}

	m := newTestMarket()
	m.FundMarket(outpoints, baseAsset)
	return m
}

func newTestMarketTradable() *domain.Market {
	m := newTestMarketFunded()
	m.MakeTradable()
	return m
}

func newTestMarketFundedWithPluggableStrategy() *domain.Market {
	m := newTestMarketFunded()
	m.MakeStrategyPluggable()
	return m
}
