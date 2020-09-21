package formula

import (
	"testing"

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
		wantSpotPrice uint64
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
			9760,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BalancedReserves{}
			if gotSpotPrice := b.SpotPrice(tt.args.opts); gotSpotPrice != tt.wantSpotPrice {
				t.Errorf("BalancedReserves.SpotPrice() = %v, want %v", gotSpotPrice, tt.wantSpotPrice)
			}
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
		b             BalancedReserves
		args          args
		wantAmountOut uint64
	}{
		{
			"OutGivenIn with fee taken on the input",
			BalancedReserves{},
			args{
				opts: &marketmaking.FormulaOpts{
					BalanceIn:           100000000,
					BalanceOut:          650000000000,
					Fee:                 25,
					ChargeFeeOnTheWayIn: true,
				},
				amountIn: 10000,
			},
			64831017,
		},
		{
			"OutGivenIn with the fee taken on the output",
			BalancedReserves{},
			args{
				opts: &marketmaking.FormulaOpts{
					BalanceIn:           100000000,
					BalanceOut:          650000000000,
					Fee:                 25,
					ChargeFeeOnTheWayIn: false,
				},
				amountIn: 10000,
			},
			65155984,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := BalancedReserves{}
			if gotAmountOut := b.OutGivenIn(tt.args.opts, tt.args.amountIn); gotAmountOut != tt.wantAmountOut {
				t.Errorf("BalancedReserves.OutGivenIn() = %v, want %v", gotAmountOut, tt.wantAmountOut)
			}
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
			65169016,
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
			64843983,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := BalancedReserves{}
			if gotAmountIn := b.InGivenOut(tt.args.opts, tt.args.amountOut); gotAmountIn != tt.wantAmountIn {
				t.Errorf("BalancedReserves.InGivenOut() = %v, want %v", gotAmountIn, tt.wantAmountIn)
			}
		})
	}
}
