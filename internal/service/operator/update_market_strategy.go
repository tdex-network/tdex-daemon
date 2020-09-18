package operatorservice

import (
	"context"
	"errors"

	"github.com/tdex-network/tdex-daemon/internal/domain/market"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UpdateMarketStrategy changes the current market making strategy, either using an automated
// market making formula or a pluggable price feed
func (s *Service) UpdateMarketStrategy(ctx context.Context, req *pb.UpdateMarketStrategyRequest) (*pb.UpdateMarketStrategyReply, error) {

	requestMkt := req.GetMarket()
	// Checks if base asset is correct
	if err := validateBaseAsset(requestMkt.GetBaseAsset()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	//Checks if market exist
	_, accountIndex, err := s.marketRepository.GetMarketByAsset(ctx, requestMkt.GetQuoteAsset())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	//For now we support only BALANCED or PLUGGABLE (ie. price feed)
	requestStrategy := req.GetStrategyType()
	//Updates the strategy
	if err := s.marketRepository.UpdateMarket(context.Background(), accountIndex, func(m *market.Market) (*market.Market, error) {

		switch requestStrategy {

		case pb.StrategyType_PLUGGABLE:
			if err := m.MakeStrategyPluggable(); err != nil {
				return nil, err
			}

		case pb.StrategyType_BALANCED:
			if err := m.MakeStrategyBalanced(); err != nil {
				return nil, err
			}

		default:
			return nil, errors.New("Strategy not supported")
		}

		return m, nil
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateMarketStrategyReply{}, nil
}
