package formula_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking/formula"
)

func TestSpotPrice(t *testing.T) {
	tests := []struct {
		name          string
		opts          formula.BalancedReservesOpts
		wantSpotPrice decimal.Decimal
	}{
		{
			"SpotPrice",
			formula.BalancedReservesOpts{
				BalanceIn:  decimal.NewFromInt(2),
				BalanceOut: decimal.NewFromInt(2 * 9760),
			},
			decimal.NewFromInt(9760),
		},
	}

	b := &formula.BalancedReserves{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spotPrice, err := b.SpotPrice(tt.opts)
			require.NoError(t, err)
			require.Equal(t, tt.wantSpotPrice.String(), spotPrice.String())
		})
	}
}

func TestOutGivenIn(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name          string
			opts          formula.BalancedReservesOpts
			amountIn      decimal.Decimal
			wantAmountOut decimal.Decimal
		}{
			{
				"with fee taken on the input",
				formula.BalancedReservesOpts{
					BalanceIn:  decimal.NewFromInt(1),
					BalanceOut: decimal.NewFromInt(6500),
				},
				decimal.NewFromFloat(0.0001),
				decimal.NewFromFloat(0.64993501),
			},
			{
				"with the fee taken on the output",
				formula.BalancedReservesOpts{
					BalanceIn:  decimal.NewFromInt(1),
					BalanceOut: decimal.NewFromInt(6500),
				},
				decimal.NewFromFloat(0.0001),
				decimal.NewFromFloat(0.64993501),
			},
		}

		b := formula.BalancedReserves{}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				amountOut, err := b.OutGivenIn(tt.opts, tt.amountIn)
				require.NoError(t, err)
				require.Equal(t, tt.wantAmountOut.String(), amountOut.String())
			})
		}
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name      string
			opts      formula.BalancedReservesOpts
			amountIn  decimal.Decimal
			wantError error
		}{
			{
				"provided amount is zero",
				formula.BalancedReservesOpts{
					BalanceIn:  decimal.NewFromInt(1),
					BalanceOut: decimal.NewFromInt(6500),
				},
				decimal.Zero,
				formula.ErrAmountTooLow,
			},
			{
				"provided amount too low",
				formula.BalancedReservesOpts{
					BalanceIn:  decimal.NewFromInt(6500),
					BalanceOut: decimal.NewFromInt(1),
				},
				decimal.NewFromFloat(0.00000001),
				formula.ErrAmountTooLow,
			},
		}

		b := formula.BalancedReserves{}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				amountOut, err := b.OutGivenIn(tt.opts, tt.amountIn)
				require.EqualError(t, err, tt.wantError.Error())
				require.Zero(t, amountOut)
			})
		}
	})
}

func TestInGivenOut(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			opts         formula.BalancedReservesOpts
			amountOut    decimal.Decimal
			wantAmountIn decimal.Decimal
		}{
			{
				formula.BalancedReservesOpts{
					BalanceIn:  decimal.NewFromInt(6500),
					BalanceOut: decimal.NewFromInt(1),
				},
				decimal.NewFromFloat(0.0001),
				decimal.NewFromFloat(0.65006501),
			},
		}

		b := formula.BalancedReserves{}
		for _, tt := range tests {
			amountIn, err := b.InGivenOut(tt.opts, tt.amountOut)
			require.NoError(t, err)
			require.Equal(t, tt.wantAmountIn.String(), amountIn.String())
		}
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name      string
			opts      formula.BalancedReservesOpts
			amountOut decimal.Decimal
			wantError error
		}{
			{
				"provided amount is zero",
				formula.BalancedReservesOpts{
					BalanceIn:  decimal.NewFromFloat(6500),
					BalanceOut: decimal.NewFromFloat(1),
				},
				decimal.Zero,
				formula.ErrAmountTooLow,
			},
			{
				"provided amount too big",
				formula.BalancedReservesOpts{
					BalanceIn:  decimal.NewFromFloat(6500),
					BalanceOut: decimal.NewFromFloat(1),
				},
				decimal.NewFromFloat(1),
				formula.ErrAmountTooBig,
			},
			{
				"provided amount too low",
				formula.BalancedReservesOpts{
					BalanceIn:  decimal.NewFromFloat(1),
					BalanceOut: decimal.NewFromFloat(6500),
				},
				decimal.NewFromFloat(0.00001),
				formula.ErrAmountTooLow,
			},
		}

		b := formula.BalancedReserves{}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				amountIn, err := b.InGivenOut(tt.opts, tt.amountOut)
				require.EqualError(t, err, tt.wantError.Error())
				require.Zero(t, amountIn)
			})
		}
	})
}
