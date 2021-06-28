package formula_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking/formula"
	"github.com/tdex-network/tdex-daemon/pkg/mathutil"
)

func TestBalancedReserves_SpotPrice(t *testing.T) {
	type args struct {
		opts interface{}
	}
	tests := []struct {
		name          string
		b             formula.BalancedReserves
		args          args
		wantSpotPrice decimal.Decimal
	}{
		{
			"OutGivenIn",
			formula.BalancedReserves{},
			args{
				opts: toInterface(formula.BalancedReservesOpts{
					BalanceIn:  2 * mathutil.BigOne,
					BalanceOut: 2 * 9760 * mathutil.BigOne,
				}),
			},
			decimal.NewFromInt(9760),
		},
	}
	b := &formula.BalancedReserves{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSpotPrice, err := b.SpotPrice(tt.args.opts)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.wantSpotPrice.BigInt().Int64(), gotSpotPrice.BigInt().Int64())
		})
	}
}

func TestBalancedReserves_OutGivenIn(t *testing.T) {
	type args struct {
		opts     interface{}
		amountIn uint64
	}
	tests := []struct {
		name          string
		args          args
		wantAmountOut uint64
	}{
		{
			"with fee taken on the input",
			args{
				opts: toInterface(formula.BalancedReservesOpts{
					BalanceIn:           100000000,
					BalanceOut:          650000000000,
					Fee:                 25,
					ChargeFeeOnTheWayIn: true,
				}),
				amountIn: 10000,
			},
			64831000,
		},
		{
			"with the fee taken on the output",
			args{
				opts: toInterface(formula.BalancedReservesOpts{
					BalanceIn:           100000000,
					BalanceOut:          650000000000,
					Fee:                 25,
					ChargeFeeOnTheWayIn: false,
				}),
				amountIn: 10000,
			},
			64831016,
		},
	}

	failingTests := []struct {
		name      string
		args      args
		wantError error
	}{
		{
			"provided amount is zero",
			args{
				opts: toInterface(formula.BalancedReservesOpts{
					BalanceIn:           100000000,
					BalanceOut:          650000000000,
					Fee:                 25,
					ChargeFeeOnTheWayIn: true,
				}),
				amountIn: 0,
			},
			formula.ErrAmountTooLow,
		},
		{
			"provided amount too low",
			args{
				opts: toInterface(formula.BalancedReservesOpts{
					BalanceIn:           650000000000,
					BalanceOut:          100000000,
					Fee:                 25,
					ChargeFeeOnTheWayIn: true,
				}),
				amountIn: 1,
			},
			formula.ErrAmountTooLow,
		},
		{
			"calculated amount too low",
			args{
				opts: toInterface(formula.BalancedReservesOpts{
					BalanceIn:           650000000000,
					BalanceOut:          100000000,
					Fee:                 25,
					ChargeFeeOnTheWayIn: true,
				}),
				amountIn: 3259,
			},
			formula.ErrAmountTooLow,
		},
		{
			"provided amount too big",
			args{
				opts: toInterface(formula.BalancedReservesOpts{
					BalanceIn:           650000000000,
					BalanceOut:          100000000,
					Fee:                 25,
					ChargeFeeOnTheWayIn: true,
				}),
				amountIn: 650000000000,
			},
			formula.ErrAmountTooBig,
		},
	}

	b := formula.BalancedReserves{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAmountOut, err := b.OutGivenIn(tt.args.opts, tt.args.amountIn)
			if err != nil {
				t.Error(err)
			}
			assert.Equal(t, int64(tt.wantAmountOut), int64(gotAmountOut))
		})
	}

	for _, tt := range failingTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := b.OutGivenIn(tt.args.opts, tt.args.amountIn)
			assert.Equal(t, tt.wantError, err)
		})
	}
}

func TestBalancedReserves_InGivenOut(t *testing.T) {
	type args struct {
		opts      interface{}
		amountOut uint64
	}
	tests := []struct {
		name         string
		b            formula.BalancedReserves
		args         args
		wantAmountIn uint64
	}{
		{
			"with fees taken on the input",
			formula.BalancedReserves{},
			args{
				opts: toInterface(formula.BalancedReservesOpts{
					BalanceIn:           650000000000,
					BalanceOut:          100000000,
					Fee:                 25,
					ChargeFeeOnTheWayIn: true,
				}),
				amountOut: 10000,
			},
			65169016,
		},
		{
			"with fees taken on the output",
			formula.BalancedReserves{},
			args{
				opts: toInterface(formula.BalancedReservesOpts{
					BalanceIn:           650000000000,
					BalanceOut:          100000000,
					Fee:                 25,
					ChargeFeeOnTheWayIn: false,
				}),
				amountOut: 10000,
			},
			65169000,
		},
	}

	failingTests := []struct {
		name      string
		args      args
		wantError error
	}{
		{
			"provided amount is zero",
			args{
				opts: toInterface(formula.BalancedReservesOpts{
					BalanceIn:           650000000000,
					BalanceOut:          100000000,
					Fee:                 25,
					ChargeFeeOnTheWayIn: true,
				}),
				amountOut: 0,
			},
			formula.ErrAmountTooLow,
		},
		{
			"provided amount too big",
			args{
				opts: toInterface(formula.BalancedReservesOpts{
					BalanceIn:           650000000000,
					BalanceOut:          100000000,
					Fee:                 5000,
					ChargeFeeOnTheWayIn: true,
				}),
				amountOut: 100000000,
			},
			formula.ErrAmountTooBig,
		},
		{
			"calculated amount too big",
			args{
				opts: toInterface(formula.BalancedReservesOpts{
					BalanceIn:           650000000000,
					BalanceOut:          100000000,
					Fee:                 5000,
					ChargeFeeOnTheWayIn: true,
				}),
				amountOut: 50000000,
			},
			formula.ErrAmountTooBig,
		},
	}

	b := formula.BalancedReserves{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAmountIn, err := b.InGivenOut(tt.args.opts, tt.args.amountOut)
			if err != nil {
				t.Error(err)
			}
			assert.Equal(t, int64(tt.wantAmountIn), int64(gotAmountIn))
		})
	}

	for _, tt := range failingTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := b.InGivenOut(tt.args.opts, tt.args.amountOut)
			assert.Equal(t, tt.wantError, err)
		})
	}
}

func toInterface(o formula.BalancedReservesOpts) interface{} {
	return o
}
