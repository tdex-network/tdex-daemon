package operatorservice

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/domain/market"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UpdateMarketPrice rpc updates the price for the given market
func (s *Service) UpdateMarketPrice(ctx context.Context, req *pb.UpdateMarketPriceRequest) (*pb.UpdateMarketPriceReply, error) {

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

	//Updates the base price and the quote price
	if err := s.marketRepository.UpdateMarket(context.Background(), accountIndex, func(m *market.Market) (*market.Market, error) {

		price := req.GetPrice()

		if err := m.ChangeBasePrice(price.GetBasePrice()); err != nil {
			return nil, err
		}

		if err := m.ChangeQuotePrice(price.GetQuotePrice()); err != nil {
			return nil, err
		}

		return m, nil
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateMarketPriceReply{}, nil
}
