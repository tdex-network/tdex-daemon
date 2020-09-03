package formula

import (
	"testing"

	"github.com/tdex-network/tdex-daemon/pkg/marketmaking"
)

func TestBalancedReserves_SpotPrice(t *testing.T) {
	type fields struct {
		MakingFormula marketmaking.MakingFormula
	}
	type args struct {
		opts marketmaking.SpotPriceOpts
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantSpotPrice int64
	}{
		{
			"Spot Price",
			fields{},
			args{
				opts: marketmaking.SpotPriceOpts{
					BalanceIn:           100000000,
					WeightIn:            50,
					BalanceOut:          1000000000000,
					WeightOut:           50,
					Fee:                 1000,
					ChargeFeeOnTheWayIn: true,
				},
			},
			100000000 + 10000000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BalancedReserves{}
			if gotSpotPrice := b.SpotPrice(&tt.args.opts); gotSpotPrice != tt.wantSpotPrice {
				t.Errorf("BalancedReserves.SpotPrice() = %v, want %v", gotSpotPrice, tt.wantSpotPrice)
			}
		})
	}
}
