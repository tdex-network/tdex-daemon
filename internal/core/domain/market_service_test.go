package domain_test

import (
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

func TestVerifyMarketFunds(t *testing.T) {
	t.Parallel()

	market := newTestMarket()
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

	err := market.VerifyMarketFunds(outpoints)
	require.NoError(t, err)
}

func TestFailingVerifyMarketFunds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		market        *domain.Market
		outpoints     []domain.OutpointWithAsset
		expectedError error
	}{
		{
			name:   "missing_quote_asset",
			market: newTestMarket(),
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
			name:   "missing_base_asset",
			market: newTestMarket(),
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
			name:   "to_many_assets",
			market: newTestMarket(),
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
			err := tt.market.VerifyMarketFunds(tt.outpoints)
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

	err := m.MakeNotTradable()
	require.NoError(t, err)
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
	require.False(t, m.IsStrategyPluggableInitialized())
	require.False(t, m.IsStrategyBalanced())
}

func TestFailingMakeStrategyPluggable(t *testing.T) {
	t.Parallel()

	m := newTestMarketTradable()

	err := m.MakeStrategyPluggable()
	require.EqualError(t, err, domain.ErrMarketMustBeClosed.Error())
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
	require.EqualError(t, err, domain.ErrMarketMustBeClosed.Error())
}

func TestChangeFeeBasisPoint(t *testing.T) {
	t.Parallel()

	m := newTestMarket()
	newFee := int64(50)

	err := m.ChangeFeeBasisPoint(newFee)
	require.NoError(t, err)
	require.Equal(t, newFee, m.Fee)
}

func TestFailingChangeFeeBasisPoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		market        *domain.Market
		marketFee     int64
		expectedError error
	}{
		{
			name:          "must_be_closed",
			market:        newTestMarketTradable(),
			marketFee:     50,
			expectedError: domain.ErrMarketMustBeClosed,
		},
		{
			name:          "fee_too_low",
			market:        newTestMarket(),
			marketFee:     -1,
			expectedError: domain.ErrMarketFeeTooLow,
		},
		{
			name:          "fee_too_high",
			market:        newTestMarket(),
			marketFee:     10000,
			expectedError: domain.ErrMarketFeeTooHigh,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.market.ChangeFeeBasisPoint(tt.marketFee)
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
	require.Equal(t, baseFee, m.FixedFee.BaseFee)
	require.Zero(t, m.FixedFee.QuoteFee)

	quoteFee := int64(200000)
	err = m.ChangeFixedFee(-1, quoteFee)
	require.NoError(t, err)
	require.Equal(t, baseFee, m.FixedFee.BaseFee)
	require.Equal(t, quoteFee, m.FixedFee.QuoteFee)
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
			expectedError: domain.ErrMarketMustBeClosed,
		},
		{
			name:          "invalid_fixed_base_fee",
			market:        newTestMarket(),
			baseFee:       -2,
			quoteFee:      1000,
			expectedError: domain.ErrInvalidFixedFee,
		},
		{
			name:          "invalid_fixed_quote_fee",
			market:        newTestMarket(),
			baseFee:       100,
			quoteFee:      -2,
			expectedError: domain.ErrInvalidFixedFee,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.market.ChangeFixedFee(tt.baseFee, tt.quoteFee)
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

	err := m.ChangeBasePrice(decimal.NewFromFloat(-0.0002))
	require.EqualError(t, err, domain.ErrMarketInvalidBasePrice.Error())
}

func TestFailingChangeQuotePrice(t *testing.T) {
	t.Parallel()

	m := newTestMarket()

	err := m.ChangeQuotePrice(decimal.NewFromFloat(-50000))
	require.EqualError(t, err, domain.ErrMarketInvalidQuotePrice.Error())
}

func TestPreview(t *testing.T) {
	t.Parallel()

	t.Run("market with balanced strategy", func(t *testing.T) {
		tests := []struct {
			baseAssetPrecision  uint
			quoteAssetPrecision uint
			percentageFee       int64
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
					Price: domain.Prices{
						BasePrice:  decimal.NewFromFloat(0.000025),
						QuotePrice: decimal.NewFromFloat(40000),
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
					Price: domain.Prices{
						BasePrice:  decimal.NewFromFloat(0.000025),
						QuotePrice: decimal.NewFromFloat(40000),
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
					Price: domain.Prices{
						BasePrice:  decimal.NewFromFloat(0.000025),
						QuotePrice: decimal.NewFromFloat(40000),
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
					Price: domain.Prices{
						BasePrice:  decimal.NewFromFloat(0.000025),
						QuotePrice: decimal.NewFromFloat(40000),
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
			market.ChangeFeeBasisPoint(tt.percentageFee)
			market.ChangeFixedFee(tt.baseFixedFee, tt.quoteFixedFee)
			market.MakeTradable()

			preview, err := market.Preview(tt.baseBalance, tt.quoteBalance, tt.amount, tt.isBaseAsset, tt.isBuy)
			require.NoError(t, err)
			require.NotNil(t, preview)
			require.Equal(t, tt.expected.Price.BasePrice.String(), preview.Price.BasePrice.String())
			require.Equal(t, tt.expected.Price.QuotePrice.String(), preview.Price.QuotePrice.String())
			require.Equal(t, int(tt.expected.Amount), int(preview.Amount))
			require.Equal(t, tt.expected.Asset, preview.Asset)
		}
	})

	t.Run("market with pluggable strategy", func(t *testing.T) {
		price := domain.Prices{
			BasePrice:  decimal.NewFromFloat(0.000028571429),
			QuotePrice: decimal.NewFromFloat(35000),
		}

		tests := []struct {
			baseAssetPrecision  uint
			quoteAssetPrecision uint
			percentageFee       int64
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
					Price: domain.Prices{
						BasePrice:  decimal.NewFromFloat(0.000028571429),
						QuotePrice: decimal.NewFromFloat(35000),
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
					Price: domain.Prices{
						BasePrice:  decimal.NewFromFloat(0.000028571429),
						QuotePrice: decimal.NewFromFloat(35000),
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
					Price: domain.Prices{
						BasePrice:  decimal.NewFromFloat(0.000028571429),
						QuotePrice: decimal.NewFromFloat(35000),
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
			market.ChangeFeeBasisPoint(tt.percentageFee)
			market.ChangeFixedFee(tt.baseFixedFee, tt.quoteFixedFee)
			market.ChangeBasePrice(price.BasePrice)
			market.ChangeQuotePrice(price.QuotePrice)
			market.MakeTradable()

			preview, err := market.Preview(tt.baseBalance, tt.quoteBalance, tt.amount, tt.isBaseAsset, tt.isBuy)
			require.NoError(t, err)
			require.NotNil(t, preview)
			require.Equal(t, tt.expected.Price.BasePrice.String(), preview.Price.BasePrice.String())
			require.Equal(t, tt.expected.Price.QuotePrice.String(), preview.Price.QuotePrice.String())
			require.Equal(t, tt.expected.Asset, preview.Asset)
			require.Equal(t, int(tt.expected.Amount), int(preview.Amount))
		}
	})
}

func TestFailingPreview(t *testing.T) {
	t.Parallel()

	t.Run("market with balanced strategy", func(t *testing.T) {
		market := newTestMarket()
		market.ChangeFeeBasisPoint(100)
		market.MakeTradable()

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
				if preview != nil {
					fmt.Println(preview.Amount)
				}
				require.EqualError(t, err, tt.expectedErr.Error())
				require.Nil(t, preview)
			})
		}
	})

	t.Run("market with pluggable strategy", func(t *testing.T) {
		market := newTestMarketWithPluggableStrategy()
		market.MakeNotTradable()
		market.ChangeFeeBasisPoint(100)
		market.ChangeBasePrice(decimal.NewFromFloat(0.000028571429))
		market.ChangeQuotePrice(decimal.NewFromFloat(35000))
		market.MakeTradable()

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
				if preview != nil {
					fmt.Println(preview.Amount)
				}
				require.EqualError(t, err, tt.expectedErr.Error())
				require.Nil(t, preview)
			})
		}
	})

	t.Run("market with balanced strategy and fixed fees", func(t *testing.T) {
		market := newTestMarket()
		market.ChangeFeeBasisPoint(100)
		market.ChangeFixedFee(650, 20000000)
		market.MakeTradable()

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
		market.ChangeFeeBasisPoint(100)
		market.ChangeFixedFee(650, 20000000)
		market.ChangeBasePrice(decimal.NewFromFloat(0.000028571429))
		market.ChangeQuotePrice(decimal.NewFromFloat(35000))
		market.MakeTradable()

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
	m, _ := domain.NewMarket(0, baseAsset, quoteAsset, 25, bp, qp)
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
