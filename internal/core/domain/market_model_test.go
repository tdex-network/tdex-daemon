package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

const (
	baseAsset  = "0000000000000000000000000000000000000000000000000000000000000000"
	quoteAsset = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

func TestNewMarket(t *testing.T) {
	t.Parallel()

	accountIndex := 0
	fee := int64(25)

	m, err := domain.NewMarket(accountIndex, baseAsset, quoteAsset, fee)
	require.NoError(t, err)
	require.NotNil(t, m)
	require.Equal(t, accountIndex, m.AccountIndex)
	require.Equal(t, baseAsset, m.BaseAsset)
	require.Equal(t, quoteAsset, m.QuoteAsset)
	require.Equal(t, fee, m.Fee)
	require.Zero(t, m.FixedFee.BaseFee)
	require.Zero(t, m.FixedFee.QuoteFee)
	require.False(t, m.IsStrategyPluggable())
}

func TestFailingNewMarket(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		accountIndex  int
		baseAsset     string
		quoteAsset    string
		fee           int64
		expectedError error
	}{
		{"invalid_account", -1, baseAsset, quoteAsset, 25, domain.ErrInvalidAccount},
		{"invalid_base_asset", 0, "", quoteAsset, 25, domain.ErrMarketInvalidBaseAsset},
		{"invalid_quote_asset", 0, baseAsset, "invalidquoteasset", 25, domain.ErrMarketInvalidQuoteAsset},
		{"fee_too_low", 0, baseAsset, quoteAsset, -1, domain.ErrMarketFeeTooLow},
		{"fee_too_high", 0, baseAsset, quoteAsset, 10000, domain.ErrMarketFeeTooHigh},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewMarket(tt.accountIndex, tt.baseAsset, tt.quoteAsset, tt.fee)
			require.EqualError(t, err, tt.expectedError.Error())
		})
	}
}
