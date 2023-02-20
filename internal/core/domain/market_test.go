package domain_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

const (
	baseAsset  = "0000000000000000000000000000000000000000000000000000000000000000"
	quoteAsset = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

func TestNewMarket(t *testing.T) {
	t.Parallel()

	fee := uint32(25)

	m, err := domain.NewMarket(baseAsset, quoteAsset, fee, 0, 0)
	require.NoError(t, err)
	require.NotNil(t, m)
	require.Equal(t, baseAsset, m.BaseAsset)
	require.Equal(t, quoteAsset, m.QuoteAsset)
	require.Equal(t, fee, m.PercentageFee)
	require.Zero(t, m.FixedFee.BaseFee)
	require.Zero(t, m.FixedFee.QuoteFee)
	require.False(t, m.IsStrategyPluggable())
}

func TestFailingNewMarket(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		baseAsset           string
		quoteAsset          string
		baseAssetPrecision  int
		quoteAssetPrecision int
		fee                 int64
		expectedError       error
	}{
		{
			name:          "invalid_base_asset",
			baseAsset:     "",
			quoteAsset:    quoteAsset,
			fee:           25,
			expectedError: domain.ErrMarketInvalidBaseAsset,
		},
		{
			name:          "invalid_quote_asset",
			baseAsset:     baseAsset,
			quoteAsset:    "invalidquoteasset",
			fee:           25,
			expectedError: domain.ErrMarketInvalidQuoteAsset,
		},
		{
			name:          "fee_too_low",
			baseAsset:     baseAsset,
			quoteAsset:    quoteAsset,
			fee:           -1,
			expectedError: domain.ErrMarketInvalidPercentageFee,
		},
		{
			name:          "fee_too_high",
			baseAsset:     baseAsset,
			quoteAsset:    quoteAsset,
			fee:           10000,
			expectedError: domain.ErrMarketInvalidPercentageFee,
		},
		{
			name:               "invalid_base_asset_precision",
			baseAsset:          baseAsset,
			quoteAsset:         quoteAsset,
			baseAssetPrecision: 10,
			fee:                25,
			expectedError:      domain.ErrMarketInvalidBaseAssetPrecision,
		},
		{
			name:                "invalid_quote_asset_precision",
			baseAsset:           baseAsset,
			quoteAsset:          quoteAsset,
			quoteAssetPrecision: -1,
			fee:                 25,
			expectedError:       domain.ErrMarketInvalidQuoteAssetPrecision,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewMarket(
				tt.baseAsset, tt.quoteAsset, uint32(tt.fee),
				uint(tt.baseAssetPrecision), uint(tt.quoteAssetPrecision),
			)
			require.EqualError(t, err, tt.expectedError.Error())
		})
	}
}

func TestMakeTradable(t *testing.T) {
	t.Parallel()

	m := newTestMarket()

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
			name:          "not_priced",
			market:        newTestMarketWithPluggableStrategy(),
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

	m.MakeNotTradable()
	require.False(t, m.IsTradable())
}

func TestMakeStrategyPluggable(t *testing.T) {
	t.Parallel()

	m := newTestMarket()
	require.True(t, m.IsStrategyBalanced())
	require.False(t, m.IsStrategyPluggable())

	err := m.MakeStrategyPluggable()
	require.NoError(t, err)
	require.True(t, m.IsStrategyPluggable())
	require.False(t, m.IsStrategyBalanced())
}

func TestFailingMakeStrategyPluggable(t *testing.T) {
	t.Parallel()

	m := newTestMarketTradable()

	err := m.MakeStrategyPluggable()
	require.EqualError(t, err, domain.ErrMarketIsOpen.Error())
}

func TestMakeStrategyBalanced(t *testing.T) {
	t.Parallel()

	m := newTestMarketWithPluggableStrategy()

	err := m.MakeStrategyBalanced()
	require.NoError(t, err)
	require.False(t, m.IsStrategyPluggable())
}

func TestFailingMakeStrategyBalanced(t *testing.T) {
	t.Parallel()

	m := newTestMarketTradable()

	err := m.MakeStrategyBalanced()
	require.EqualError(t, err, domain.ErrMarketIsOpen.Error())
}

func TestChangePercentageFee(t *testing.T) {
	t.Parallel()

	m := newTestMarket()
	newFee := uint32(50)

	err := m.ChangePercentageFee(newFee)
	require.NoError(t, err)
	require.Equal(t, newFee, m.PercentageFee)
}

func TestFailingChangePercentageFee(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		market        *domain.Market
		marketFee     int
		expectedError error
	}{
		{
			name:          "must_be_closed",
			market:        newTestMarketTradable(),
			marketFee:     50,
			expectedError: domain.ErrMarketIsOpen,
		},
		{
			name:          "fee_too_low",
			market:        newTestMarket(),
			marketFee:     -1,
			expectedError: domain.ErrMarketInvalidPercentageFee,
		},
		{
			name:          "fee_too_high",
			market:        newTestMarket(),
			marketFee:     10000,
			expectedError: domain.ErrMarketInvalidPercentageFee,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.market.ChangePercentageFee(uint32(tt.marketFee))
			require.EqualError(t, err, tt.expectedError.Error())
		})
	}
}

func TestChangeFixedFee(t *testing.T) {
	t.Parallel()

	m := newTestMarket()
	baseFee := int64(100)
	err := m.ChangeFixedFee(baseFee, -1)
	require.NoError(t, err)
	require.Equal(t, int(baseFee), int(m.FixedFee.BaseFee))
	require.Zero(t, m.FixedFee.QuoteFee)

	quoteFee := int64(200000)
	err = m.ChangeFixedFee(-1, quoteFee)
	require.NoError(t, err)
	require.Equal(t, int(baseFee), int(m.FixedFee.BaseFee))
	require.Equal(t, int(quoteFee), int(m.FixedFee.QuoteFee))
}

func TestFailingChangeFixedFee(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		market            *domain.Market
		baseFee, quoteFee int64
		expectedError     error
	}{
		{
			name:          "must_be_closed",
			market:        newTestMarketTradable(),
			expectedError: domain.ErrMarketIsOpen,
		},
		{
			name:          "invalid_fixed_base_fee",
			market:        newTestMarket(),
			baseFee:       -2,
			quoteFee:      1000,
			expectedError: domain.ErrMarketInvalidFixedFee,
		},
		{
			name:          "invalid_fixed_quote_fee",
			market:        newTestMarket(),
			baseFee:       100,
			quoteFee:      -2,
			expectedError: domain.ErrMarketInvalidFixedFee,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.market.ChangeFixedFee(tt.baseFee, tt.quoteFee)
			require.EqualError(t, err, tt.expectedError.Error())
		})
	}
}

func TestChangeMarketMarketPrice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		market     *domain.Market
		basePrice  decimal.Decimal
		quotePrice decimal.Decimal
	}{
		{
			name:       "change_prices_with_balanced_strategy",
			market:     newTestMarket(),
			basePrice:  decimal.NewFromFloat(0.00002),
			quotePrice: decimal.NewFromFloat(50000),
		},
		{
			name:       "change_prices_with_pluggable_strategy",
			market:     newTestMarketWithPluggableStrategy(),
			basePrice:  decimal.NewFromFloat(0.00002),
			quotePrice: decimal.NewFromFloat(50000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.market.ChangePrice(tt.basePrice, tt.quotePrice)
			require.NoError(t, err)
			require.Equal(t, tt.basePrice.String(), tt.market.Price.BasePrice)
			require.Equal(t, tt.quotePrice.String(), tt.market.Price.QuotePrice)
		})
	}
}

func TestFailingChangeBasePrice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		market        *domain.Market
		basePrice     decimal.Decimal
		quotePrice    decimal.Decimal
		expectedError error
	}{
		{
			name:          "negative_base_price",
			market:        newTestMarket(),
			basePrice:     decimal.NewFromInt(-1),
			quotePrice:    decimal.NewFromInt(1),
			expectedError: domain.ErrMarketInvalidBasePrice,
		},
		{
			name:          "zero_base_price",
			market:        newTestMarket(),
			basePrice:     decimal.Zero,
			quotePrice:    decimal.NewFromInt(1),
			expectedError: domain.ErrMarketInvalidBasePrice,
		},
		{
			name:          "negative_quote_price",
			market:        newTestMarket(),
			basePrice:     decimal.NewFromInt(1),
			quotePrice:    decimal.NewFromInt(-1),
			expectedError: domain.ErrMarketInvalidQuotePrice,
		},
		{
			name:          "zero_quote_price",
			market:        newTestMarket(),
			basePrice:     decimal.NewFromInt(1),
			quotePrice:    decimal.Zero,
			expectedError: domain.ErrMarketInvalidQuotePrice,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.market.ChangePrice(tt.basePrice, tt.quotePrice)
			require.EqualError(t, err, tt.expectedError.Error())
		})
	}
}

func TestPreview(t *testing.T) {
	t.Parallel()

	t.Run("market with balanced strategy", func(t *testing.T) {
		tests := []struct {
			baseAssetPrecision  uint
			quoteAssetPrecision uint
			percentageFee       uint32
			baseFixedFee        int64
			quoteFixedFee       int64
			baseBalance         uint64
			quoteBalance        uint64
			amount              uint64
			isBaseAsset         bool
			isBuy               bool
			expected            *domain.PreviewInfo
		}{
			{
				baseAssetPrecision:  8,
				quoteAssetPrecision: 2,
				percentageFee:       100,
				baseFixedFee:        650,
				quoteFixedFee:       20,
				baseBalance:         100000,
				quoteBalance:        4000,
				amount:              2000,
				isBaseAsset:         true,
				isBuy:               true,
				expected: &domain.PreviewInfo{
					Price: domain.MarketPrice{
						BasePrice:  decimal.NewFromFloat(0.000025).String(),
						QuotePrice: decimal.NewFromFloat(40000).String(),
					},
					Amount: 102,
					Asset:  quoteAsset,
				},
			},
			{
				baseAssetPrecision:  8,
				quoteAssetPrecision: 8,
				percentageFee:       100,
				baseFixedFee:        650,
				quoteFixedFee:       20000000,
				baseBalance:         100000,
				quoteBalance:        4000000000,
				amount:              100000000,
				isBaseAsset:         false,
				isBuy:               true,
				expected: &domain.PreviewInfo{
					Price: domain.MarketPrice{
						BasePrice:  decimal.NewFromFloat(0.000025).String(),
						QuotePrice: decimal.NewFromFloat(40000).String(),
					},
					Amount: 1765,
					Asset:  baseAsset,
				},
			},
			{
				baseAssetPrecision:  8,
				quoteAssetPrecision: 8,
				percentageFee:       100,
				baseFixedFee:        650,
				quoteFixedFee:       20000000,
				baseBalance:         100000,
				quoteBalance:        4000000000,
				amount:              2000,
				isBaseAsset:         true,
				isBuy:               false,
				expected: &domain.PreviewInfo{
					Price: domain.MarketPrice{
						BasePrice:  decimal.NewFromFloat(0.000025).String(),
						QuotePrice: decimal.NewFromFloat(40000).String(),
					},
					Amount: 57662280,
					Asset:  quoteAsset,
				},
			},
			{
				baseAssetPrecision:  8,
				quoteAssetPrecision: 2,
				percentageFee:       100,
				baseFixedFee:        650,
				quoteFixedFee:       20,
				baseBalance:         100000,
				quoteBalance:        4000,
				amount:              100,
				isBaseAsset:         false,
				isBuy:               false,
				expected: &domain.PreviewInfo{
					Price: domain.MarketPrice{
						BasePrice:  decimal.NewFromFloat(0.000025).String(),
						QuotePrice: decimal.NewFromFloat(40000).String(),
					},
					Amount: 3240,
					Asset:  baseAsset,
				},
			},
		}

		for _, tt := range tests {
			market := newTestMarketWithAssetsPrecision(
				tt.baseAssetPrecision, tt.quoteAssetPrecision,
			)
			market.ChangePercentageFee(tt.percentageFee)
			market.ChangeFixedFee(tt.baseFixedFee, tt.quoteFixedFee)
			market.MakeTradable()

			preview, err := market.Preview(tt.baseBalance, tt.quoteBalance, tt.amount, tt.isBaseAsset, tt.isBuy)
			require.NoError(t, err)
			require.NotNil(t, preview)
			require.Exactly(t, tt.expected.Price, preview.Price)
			require.Equal(t, int(tt.expected.Amount), int(preview.Amount))
			require.Equal(t, tt.expected.Asset, preview.Asset)
		}
	})

	t.Run("market with pluggable strategy", func(t *testing.T) {
		price := domain.MarketPrice{
			BasePrice:  decimal.NewFromFloat(0.000028571429).String(),
			QuotePrice: decimal.NewFromFloat(35000).String(),
		}

		tests := []struct {
			baseAssetPrecision  uint
			quoteAssetPrecision uint
			percentageFee       uint32
			baseFixedFee        int64
			quoteFixedFee       int64
			baseBalance         uint64
			quoteBalance        uint64
			amount              uint64
			isBaseAsset         bool
			isBuy               bool
			expected            *domain.PreviewInfo
		}{
			{
				baseAssetPrecision:  8,
				quoteAssetPrecision: 2,
				percentageFee:       100,
				baseFixedFee:        650,
				quoteFixedFee:       20,
				baseBalance:         100000,
				quoteBalance:        4000,
				amount:              2000,
				isBaseAsset:         true,
				isBuy:               true,
				expected: &domain.PreviewInfo{
					Price:  price,
					Amount: 90,
					Asset:  quoteAsset,
				},
			},
			{
				baseAssetPrecision:  8,
				quoteAssetPrecision: 8,
				percentageFee:       100,
				baseFixedFee:        650,
				quoteFixedFee:       20000000,
				baseBalance:         100000,
				quoteBalance:        4000000000,
				amount:              100000000,
				isBaseAsset:         false,
				isBuy:               true,
				expected: &domain.PreviewInfo{
					Price: domain.MarketPrice{
						BasePrice:  decimal.NewFromFloat(0.000028571429).String(),
						QuotePrice: decimal.NewFromFloat(35000).String(),
					},
					Amount: 2179,
					Asset:  baseAsset,
				},
			},
			{
				baseAssetPrecision:  8,
				quoteAssetPrecision: 8,
				percentageFee:       100,
				baseFixedFee:        650,
				quoteFixedFee:       20000000,
				baseBalance:         100000,
				quoteBalance:        4000000000,
				amount:              2000,
				isBaseAsset:         true,
				isBuy:               false,
				expected: &domain.PreviewInfo{
					Price: domain.MarketPrice{
						BasePrice:  decimal.NewFromFloat(0.000028571429).String(),
						QuotePrice: decimal.NewFromFloat(35000).String(),
					},
					Amount: 49300000,
					Asset:  quoteAsset,
				},
			},
			{
				baseAssetPrecision:  8,
				quoteAssetPrecision: 2,
				percentageFee:       100,
				baseFixedFee:        650,
				quoteFixedFee:       20,
				baseBalance:         100000,
				quoteBalance:        4000,
				amount:              100,
				isBaseAsset:         false,
				isBuy:               false,
				expected: &domain.PreviewInfo{
					Price: domain.MarketPrice{
						BasePrice:  decimal.NewFromFloat(0.000028571429).String(),
						QuotePrice: decimal.NewFromFloat(35000).String(),
					},
					Amount: 3536,
					Asset:  baseAsset,
				},
			},
		}

		for _, tt := range tests {
			market := newTestMarketWithAssetsPrecision(
				tt.baseAssetPrecision, tt.quoteAssetPrecision,
			)
			market.MakeStrategyPluggable()
			market.ChangePercentageFee(tt.percentageFee)
			market.ChangeFixedFee(tt.baseFixedFee, tt.quoteFixedFee)
			market.ChangePrice(price.GetBasePrice(), price.GetQuotePrice())
			market.MakeTradable()

			preview, err := market.Preview(tt.baseBalance, tt.quoteBalance, tt.amount, tt.isBaseAsset, tt.isBuy)
			require.NoError(t, err)
			require.NotNil(t, preview)
			require.Exactly(t, tt.expected.Price, preview.Price)
			require.Equal(t, tt.expected.Asset, preview.Asset)
			require.Equal(t, int(tt.expected.Amount), int(preview.Amount))
		}
	})
}

func TestFailingPreview(t *testing.T) {
	t.Parallel()

	t.Run("market with balanced strategy", func(t *testing.T) {
		market := newTestMarket()
		err := market.ChangePercentageFee(100)
		require.NoError(t, err)
		err = market.MakeTradable()
		require.NoError(t, err)

		tests := []struct {
			name         string
			baseBalance  uint64
			quoteBalance uint64
			amount       uint64
			isBaseAsset  bool
			isBuy        bool
			expectedErr  error
		}{
			{
				name:         "buy with base asset zero amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       0,
				isBaseAsset:  true,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with quote asset zero amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       0,
				isBaseAsset:  false,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with quote asset low amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       1,
				isBaseAsset:  false,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with base asset big amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       100000,
				isBaseAsset:  true,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooBig,
			},
			{
				name:         "sell with base asset zero amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       0,
				isBaseAsset:  true,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with quote asset zero amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       0,
				isBaseAsset:  false,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with base asset low amount",
				baseBalance:  4000000000,
				quoteBalance: 100000,
				amount:       1,
				isBaseAsset:  true,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with quote asset low amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       1,
				isBaseAsset:  false,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with quote asset big amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       4000000000,
				isBaseAsset:  false,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooBig,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				preview, err := market.Preview(
					tt.baseBalance, tt.quoteBalance, tt.amount, tt.isBaseAsset, tt.isBuy,
				)
				require.EqualError(t, err, tt.expectedErr.Error())
				require.Nil(t, preview)
			})
		}
	})

	t.Run("market with pluggable strategy", func(t *testing.T) {
		market := newTestMarketWithPluggableStrategy()
		market.MakeNotTradable()
		err := market.ChangePercentageFee(100)
		require.NoError(t, err)
		err = market.ChangePrice(
			decimal.NewFromFloat(0.000028571429), decimal.NewFromFloat(35000),
		)
		require.NoError(t, err)
		err = market.MakeTradable()
		require.NoError(t, err)

		tests := []struct {
			name         string
			baseBalance  uint64
			quoteBalance uint64
			amount       uint64
			isBaseAsset  bool
			isBuy        bool
			expectedErr  error
		}{
			{
				name:         "buy with base asset zero amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       0,
				isBaseAsset:  true,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with quote asset zero amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       0,
				isBaseAsset:  false,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with quote asset low amount",
				baseBalance:  100000,
				quoteBalance: 400000000,
				amount:       14999,
				isBaseAsset:  false,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with base asset big amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       100000,
				isBaseAsset:  true,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooBig,
			},
			{
				name:         "buy with quote asset big amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       3535384947,
				isBaseAsset:  false,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooBig,
			},
			{
				name:         "sell with base asset zero amount",
				baseBalance:  100000,
				quoteBalance: 400000000,
				amount:       0,
				isBaseAsset:  true,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with quote asset zero amount",
				baseBalance:  100000,
				quoteBalance: 400000000,
				amount:       0,
				isBaseAsset:  false,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with quote asset low amount",
				baseBalance:  100000,
				quoteBalance: 400000000,
				amount:       14999,
				isBaseAsset:  false,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with base asset big amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       115441,
				isBaseAsset:  true,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooBig,
			},
			{
				name:         "sell with quote asset big amount",
				baseBalance:  10000,
				quoteBalance: 40000000,
				amount:       40000000,
				isBaseAsset:  false,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooBig,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				preview, err := market.Preview(
					tt.baseBalance, tt.quoteBalance, tt.amount, tt.isBaseAsset, tt.isBuy,
				)
				require.EqualError(t, err, tt.expectedErr.Error())
				require.Nil(t, preview)
			})
		}
	})

	t.Run("market with balanced strategy and fixed fees", func(t *testing.T) {
		market := newTestMarket()
		err := market.ChangePercentageFee(100)
		require.NoError(t, err)
		err = market.ChangeFixedFee(650, 20000000)
		require.NoError(t, err)
		err = market.MakeTradable()
		require.NoError(t, err)

		tests := []struct {
			name         string
			baseBalance  uint64
			quoteBalance uint64
			amount       uint64
			isBaseAsset  bool
			isBuy        bool
			expectedErr  error
		}{
			{
				name:         "buy with base asset zero amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       0,
				isBaseAsset:  true,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with quote asset zero amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       0,
				isBaseAsset:  false,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with base asset low amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       649,
				isBaseAsset:  true,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with quote asset low amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       20034999,
				isBaseAsset:  false,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with base asset big amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       100000,
				isBaseAsset:  true,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooBig,
			},
			{
				name:         "sell with base asset zero amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       0,
				isBaseAsset:  true,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with quote asset zero amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       0,
				isBaseAsset:  false,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with base asset low amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       649,
				isBaseAsset:  true,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with quote asset low amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       19999999,
				isBaseAsset:  false,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with quote asset big amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       4000000000,
				isBaseAsset:  false,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooBig,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				preview, err := market.Preview(
					tt.baseBalance, tt.quoteBalance, tt.amount, tt.isBaseAsset, tt.isBuy,
				)
				require.EqualError(t, err, tt.expectedErr.Error())
				require.Nil(t, preview)
			})
		}
	})

	t.Run("market with pluggable strategy and fixed fees", func(t *testing.T) {
		t.Parallel()

		market := newTestMarketWithPluggableStrategy()
		market.MakeNotTradable()
		err := market.ChangePercentageFee(100)
		require.NoError(t, err)
		err = market.ChangeFixedFee(650, 20000000)
		require.NoError(t, err)
		err = market.ChangePrice(
			decimal.NewFromFloat(0.000028571429), decimal.NewFromFloat(35000),
		)
		require.NoError(t, err)
		err = market.MakeTradable()
		require.NoError(t, err)

		tests := []struct {
			name         string
			baseBalance  uint64
			quoteBalance uint64
			amount       uint64
			isBaseAsset  bool
			isBuy        bool
			expectedErr  error
		}{
			{
				name:         "buy with base asset zero amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       0,
				isBaseAsset:  true,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with quote asset zero amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       0,
				isBaseAsset:  false,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with base asset low amount",
				baseBalance:  100000,
				quoteBalance: 400000000,
				amount:       649,
				isBaseAsset:  false,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with quote asset low amount",
				baseBalance:  100000,
				quoteBalance: 400000000,
				amount:       20034999,
				isBaseAsset:  false,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "buy with base asset big amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       100000,
				isBaseAsset:  true,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooBig,
			},
			{
				name:         "buy with quote asset big amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       3558344947,
				isBaseAsset:  false,
				isBuy:        true,
				expectedErr:  domain.ErrMarketPreviewAmountTooBig,
			},
			{
				name:         "sell with base asset zero amount",
				baseBalance:  100000,
				quoteBalance: 400000000,
				amount:       0,
				isBaseAsset:  true,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with quote asset zero amount",
				baseBalance:  100000,
				quoteBalance: 400000000,
				amount:       0,
				isBaseAsset:  false,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with base asset low amount",
				baseBalance:  100000,
				quoteBalance: 400000000,
				amount:       649,
				isBaseAsset:  true,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with quote asset low amount",
				baseBalance:  100000,
				quoteBalance: 400000000,
				amount:       19999999,
				isBaseAsset:  false,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooLow,
			},
			{
				name:         "sell with base asset big amount",
				baseBalance:  100000,
				quoteBalance: 4000000000,
				amount:       116018,
				isBaseAsset:  true,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooBig,
			},
			{
				name:         "sell with quote asset big amount",
				baseBalance:  100000,
				quoteBalance: 400000000,
				amount:       400000000,
				isBaseAsset:  false,
				isBuy:        false,
				expectedErr:  domain.ErrMarketPreviewAmountTooBig,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				preview, err := market.Preview(
					tt.baseBalance, tt.quoteBalance, tt.amount, tt.isBaseAsset, tt.isBuy,
				)
				require.EqualError(t, err, tt.expectedErr.Error())
				require.Nil(t, preview)
			})
		}
	})
}

func newTestMarket() *domain.Market {
	m := newTestMarketWithAssetsPrecision(8, 8)
	return m
}

func newTestMarketWithAssetsPrecision(bp, qp uint) *domain.Market {
	m, _ := domain.NewMarket(baseAsset, quoteAsset, 25, bp, qp)
	return m
}

func newTestMarketTradable() *domain.Market {
	m := newTestMarketWithAssetsPrecision(8, 8)
	m.MakeTradable()
	return m
}

func newTestMarketWithPluggableStrategy() *domain.Market {
	m := newTestMarketWithAssetsPrecision(8, 8)
	m.MakeStrategyPluggable()
	return m
}
