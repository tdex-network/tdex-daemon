package operatorservice

import (
	"context"

	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// OpenMarket makes the given market tradable
func (s *Service) OpenMarket(ctx context.Context, req *pb.OpenMarketRequest) (*pb.OpenMarketReply, error) {

	// Requested market
	baseAsset := req.GetMarket().GetBaseAsset()
	quoteAsset := req.GetMarket().GetQuoteAsset()

	// Checks if base asset is correct
	if err := validateBaseAsset(baseAsset); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.marketRepository.OpenMarket(context.Background(), quoteAsset); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.OpenMarketReply{}, nil
}
