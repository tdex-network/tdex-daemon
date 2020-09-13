package formula

import (
	"testing"

	"github.com/tdex-network/tdex-daemon/pkg/marketmaking"
)

func TestBalancedReserves_SpotPrice(t *testing.T) {
	type args struct {
		opts *marketmaking.FormulaOpts
	}
	tests := []struct {
		name          string
		b             BalancedReserves
		args          args
		wantSpotPrice int64
	}{
		{
			"Spot Price",
			BalancedReserves{},
			args{
				opts: &marketmaking.FormulaOpts{
					BalanceIn:  8000 * BigOne,
					BalanceOut: BigOne,
					Fee:        250000,
				},
			},
			802005016000,
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
		amountIn int64
	}
	tests := []struct {
		name          string
		b             BalancedReserves
		args          args
		wantAmountOut int64
	}{
		{
			"OutGivenIn",
			BalancedReserves{},
			args{
				opts: &marketmaking.FormulaOpts{
					BalanceIn:           8000 * BigOne,
					BalanceOut:          BigOne,
					Fee:                 1,
					ChargeFeeOnTheWayIn: true,
				},
				amountIn: 8000 * BigOne,
			},
			9999999999999987,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := BalancedReserves{}
			if gotAmountOut := b.OutGivenIn(tt.args.opts, tt.args.amountIn); gotAmountOut != tt.wantAmountOut {
				println("got amount ", gotAmountOut/BigOne)
				t.Errorf("BalancedReserves.OutGivenIn() = %v, want %v", gotAmountOut, tt.wantAmountOut)
			}
		})
	}
}
