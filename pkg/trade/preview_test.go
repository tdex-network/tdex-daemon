package trade

import (
	"testing"

	"github.com/stretchr/testify/assert"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/types"
)

func TestCalcProposeAmount(t *testing.T) {
	tests := []struct {
		priceWithFee    *pb.PriceWithFee
		amountToReceive uint64
		assetToSend     string
		expectedAmount  uint64
	}{
		{
			priceWithFee: &pb.PriceWithFee{
				Price: &pb.Price{
					BasePrice:  0.0002, // LBTC/USDT
					QuotePrice: 5000,   // USDT/LBTC
				},
				Fee: &pb.Fee{
					Asset:      "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225", // LBTC
					BasisPoint: 25,                                                                 // 0.25
				},
			},
			amountToReceive: uint64(10000000), // 0.1 LBTC
			assetToSend:     "358ec5d1fff7ff4c176a01ab4938b8e25fde6ef431cfadcc0bfe04770b113e68",
			expectedAmount:  uint64(62500000000),
			// 0.1 LBTC * 0.25 = 0.025 LBTC (fee) -> 0.1 + 0.025 = 0.125 LBTC -> 0.125 LBTC * 5000 USDT/LBTC = 625 USDT
		},
		{
			priceWithFee: &pb.PriceWithFee{
				Price: &pb.Price{
					BasePrice:  0.0002,
					QuotePrice: 5000,
				},
				Fee: &pb.Fee{
					Asset:      "358ec5d1fff7ff4c176a01ab4938b8e25fde6ef431cfadcc0bfe04770b113e68", // USDT
					BasisPoint: 25,
				},
			},
			amountToReceive: uint64(10000000), // 0.1 LBTC
			assetToSend:     "358ec5d1fff7ff4c176a01ab4938b8e25fde6ef431cfadcc0bfe04770b113e68",
			expectedAmount:  uint64(62500000000),
			// 0.1 LBTC * 5000 USDT/LBTC = 500 USDT -> 500 * 0.25 = 125 USDT (fee) -> 500 + 125 = 625 USDT
		},
	}

	for _, tt := range tests {
		amountToSend := calcProposeAmount(tt.priceWithFee, tt.amountToReceive, tt.assetToSend)
		assert.Equal(t, int(tt.expectedAmount), int(amountToSend))
	}
}

func TestCalcExpectedAmount(t *testing.T) {
	tests := []struct {
		priceWithFee   *pb.PriceWithFee
		amountToSend   uint64
		assetToReceive string
		expectedAmount uint64
	}{
		{
			priceWithFee: &pb.PriceWithFee{
				Price: &pb.Price{
					BasePrice:  0.0002,
					QuotePrice: 5000,
				},
				Fee: &pb.Fee{
					Asset:      "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225",
					BasisPoint: 25, // 0.25
				},
			},
			amountToSend:   uint64(45000000000), // 450 USDT
			assetToReceive: "358ec5d1fff7ff4c176a01ab4938b8e25fde6ef431cfadcc0bfe04770b113e68",
			expectedAmount: uint64(6750000),
			// 450 USDT * 0.0002 LBTC/USDT = 0.09 LBTC -> 0.09 * 0.25 = 0.0225 LBTC (fee) -> 0.09 - 0.0225 = 0.0675 LBTC
		},
		{
			priceWithFee: &pb.PriceWithFee{
				Price: &pb.Price{
					BasePrice:  0.0002,
					QuotePrice: 5000,
				},
				Fee: &pb.Fee{
					Asset:      "358ec5d1fff7ff4c176a01ab4938b8e25fde6ef431cfadcc0bfe04770b113e68",
					BasisPoint: 25, // 0.25
				},
			},
			amountToSend:   uint64(45000000000), // 450 USDT
			assetToReceive: "358ec5d1fff7ff4c176a01ab4938b8e25fde6ef431cfadcc0bfe04770b113e68",
			expectedAmount: uint64(6750000),
			// 450 USDT * 0.25 = 112.5 USDT (fee) -> 150 - 112.5 = 337.5 USDT -> 337.5 USDT * 0.0002 LBTC/USDT= 0.0675 LBTC
		},
	}

	for _, tt := range tests {
		amountToSend := calcExpectedAmount(tt.priceWithFee, tt.amountToSend, tt.assetToReceive)
		assert.Equal(t, int(tt.expectedAmount), int(amountToSend))
	}
}
