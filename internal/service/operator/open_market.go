package operatorservice

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/domain/market"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// OpenMarket makes the given market tradable
func (s *Service) OpenMarket(ctx context.Context, req *pb.OpenMarketRequest) (*pb.OpenMarketReply, error) {

	// Checks if base asset is correct
	if err := validateBaseAsset(req.GetMarket().GetBaseAsset()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	//Checks if market exist
	_, accountIndex, err := s.marketRepository.GetMarketByAsset(ctx, req.GetMarket().GetQuoteAsset())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.marketRepository.UpdateMarket(context.Background(), accountIndex, func(m *market.Market) (*market.Market, error) {

		if err := m.MakeTradable(); err != nil {
			return nil, err
		}

		return m, nil
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.OpenMarketReply{}, nil
}
