package tradeservice

import (
	"context"

	pb "github.com/tdex-network/tdex-protobuf/generated/go/trade"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MarketPrice is the domain controller for the MarketPrice RPC
func (s *Service) MarketPrice(ctx context.Context, req *pb.MarketPriceRequest) (*pb.MarketPriceReply, error) {

	requestMkt := req.GetMarket()
	// Checks if base asset is correct
	if err := validateBaseAsset(requestMkt.GetBaseAsset()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	//Checks if market exist
	mkt, _, err := s.marketRepository.GetMarketByAsset(ctx, requestMkt.GetQuoteAsset())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !mkt.IsTradable() {
		return nil, status.Error(codes.FailedPrecondition, "Market is closed")
	}

	return &pb.MarketPriceReply{
		Prices: []*pbtypes.PriceWithFee{
			{
				Price: &pbtypes.Price{
					BasePrice:  mkt.BaseAssetPrice(),
					QuotePrice: mkt.QuoteAssetPrice(),
				},
				Fee: &pbtypes.Fee{
					Asset:      mkt.FeeAsset(),
					BasisPoint: mkt.Fee(),
				},
			},
		},
	}, nil
}
