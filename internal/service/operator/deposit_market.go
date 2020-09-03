package operatorservice

import (
	"context"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"
)

func (s *Service) DepositMarket(
	ctx context.Context,
	depositMarketReq *pb.DepositMarketRequest,
) (*pb.DepositMarketReply, error) {
	return nil, nil
}
