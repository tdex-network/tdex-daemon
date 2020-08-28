package tradeservice

import (
	"context"

	pb "github.com/tdex-network/tdex-protobuf/generated/go/trade"
)

// MarketPrice is the domain controller for the MarketPrice RPC
func (s *Service) MarketPrice(ctx context.Context, req *pb.MarketPriceRequest) (res *pb.MarketPriceReply, err error) {
	return &pb.MarketPriceReply{}, nil
}
