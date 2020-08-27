package tradeservice

import (
	pb "github.com/tdex-network/tdex-protobuf/generated/go/trade"
)

// TradePropose is the domain controller for the TradePropose RPC
func (s *Service) TradePropose(req *pb.TradeProposeRequest, stream pb.Trade_TradeProposeServer) error {
	if err := stream.Send(&pb.TradeProposeReply{}); err != nil {
		return err
	}
	return nil
}
