package operatorservice

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/domain/market"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UpdateMarketFee changes the Liquidity Provider fee for the given market.
// MUST be expressed as basis point.
// Eg. To change the fee on each swap from 0.25% to 1% you need to pass down 100
// The Market MUST be closed before doing this change.
func (s *Service) UpdateMarketFee(ctx context.Context, req *pb.UpdateMarketFeeRequest) (res *pb.UpdateMarketFeeReply, err error) {

	//Checks if market exist
	_, accountIndex, err := s.marketRepository.GetMarketByAsset(ctx, req.GetMarketWithFee().GetMarket().GetQuoteAsset())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	//Updates the fee and the fee asset
	if err := s.marketRepository.UpdateMarket(context.Background(), accountIndex, func(m *market.Market) (*market.Market, error) {

		if err := m.ChangeFee(req.GetMarketWithFee().GetFee().GetBasisPoint()); err != nil {
			return nil, err
		}

		if err := m.ChangeFeeAsset(req.GetMarketWithFee().GetFee().GetAsset()); err != nil {
			return nil, err
		}

		return m, nil
	}); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Ignore errors. If we reached this point it must exists.
	mkt, _ := s.marketRepository.GetOrCreateMarket(context.Background(), accountIndex)

	return &pb.UpdateMarketFeeReply{
		MarketWithFee: &pbtypes.MarketWithFee{
			Market: &pbtypes.Market{
				BaseAsset:  mkt.BaseAssetHash(),
				QuoteAsset: mkt.QuoteAssetHash(),
			},
			Fee: &pbtypes.Fee{
				Asset:      mkt.FeeAsset(),
				BasisPoint: mkt.Fee(),
			},
		},
	}, nil
}
