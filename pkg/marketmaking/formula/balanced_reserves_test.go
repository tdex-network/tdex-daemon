package formula

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking"
	"github.com/tdex-network/tdex-daemon/pkg/mathutil"
)

func TestBalancedReserves_SpotPrice(t *testing.T) {
	type args struct {
		opts *marketmaking.FormulaOpts
	}
	tests := []struct {
		name          string
		b             BalancedReserves
		args          args
		wantSpotPrice decimal.Decimal
	}{
		{
			"OutGivenIn",
			BalancedReserves{},
			args{
				opts: &marketmaking.FormulaOpts{
					BalanceIn:  2 * mathutil.BigOne,
					BalanceOut: 2 * 9760 * mathutil.BigOne,
				},
			},
			decimal.NewFromInt(9760),
		},
	}
	b := &BalancedReserves{}
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
		opts     *marketmaking.FormulaOpts
		amountIn uint64
	}
	tests := []struct {
		name          string
		args          args
		wantAmountOut uint64
	}{
		{
			"OutGivenIn with fee taken on the input",
			args{
				opts: &marketmaking.FormulaOpts{
					BalanceIn:           100000000,
					BalanceOut:          650000000000,
					Fee:                 25,
					ChargeFeeOnTheWayIn: true,
				},
				amountIn: 10000,
			},
			64831000,
		},
		{
			"OutGivenIn with the fee taken on the output",
			args{
				opts: &marketmaking.FormulaOpts{
					BalanceIn:           100000000,
					BalanceOut:          650000000000,
					Fee:                 25,
					ChargeFeeOnTheWayIn: false,
				},
				amountIn: 10000,
			},
			65156000,
		},
	}

	failingTests := []struct {
		name      string
		args      args
		wantError error
	}{
		{
			"OutGivenIn fails if provided amount is 0",
			args{
				opts: &marketmaking.FormulaOpts{
					BalanceIn:           100000000,
					BalanceOut:          650000000000,
					Fee:                 25,
					ChargeFeeOnTheWayIn: true,
				},
				amountIn: 0,
			},
			ErrAmountTooLow,
		},
	}

	b := BalancedReserves{}
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
		opts      *marketmaking.FormulaOpts
		amountOut uint64
	}
	tests := []struct {
		name         string
		b            BalancedReserves
		args         args
		wantAmountIn uint64
	}{
		{
			"InGivenOut with fee taken on the input",
			BalancedReserves{},
			args{
				opts: &marketmaking.FormulaOpts{
					BalanceIn:           650000000000,
					BalanceOut:          100000000,
					Fee:                 25,
					ChargeFeeOnTheWayIn: true,
				},
				amountOut: 10000,
			},
			64844388,
		},
		{
			"InGivenOut with fee taken on the output",
			BalancedReserves{},
			args{
				opts: &marketmaking.FormulaOpts{
					BalanceIn:           650000000000,
					BalanceOut:          100000000,
					Fee:                 25,
					ChargeFeeOnTheWayIn: false,
				},
				amountOut: 10000,
			},
			65169423,
		},
	}

	failingTests := []struct {
		name      string
		args      args
		wantError error
	}{
		{
			"InGivenOut fails if provided amount is 0",
			args{
				opts: &marketmaking.FormulaOpts{
					BalanceIn:           100000000,
					BalanceOut:          650000000000,
					Fee:                 25,
					ChargeFeeOnTheWayIn: true,
				},
				amountOut: 0,
			},
			ErrAmountTooLow,
		},
		{
			"InGivenOut fails if provided amount is equal or exceeds the balance",
			args{
				opts: &marketmaking.FormulaOpts{
					BalanceIn:           100000000,
					BalanceOut:          650000000000,
					Fee:                 25,
					ChargeFeeOnTheWayIn: true,
				},
				amountOut: 650000000000,
			},
			ErrAmountTooBig,
		},
	}

	b := BalancedReserves{}
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
