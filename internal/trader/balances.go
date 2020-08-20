package tradeservice

import (
	"context"

	pb "github.com/tdex-network/tdex-protobuf/generated/go/trade"
)

// Balances is the domain controller for the Balances RPC
func (s *Server) Balances(ctx context.Context, req *pb.BalancesRequest) (res *pb.BalancesReply, err error) {
	return &pb.BalancesReply{}, nil
}
