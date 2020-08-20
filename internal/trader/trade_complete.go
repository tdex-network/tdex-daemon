package tradeservice

import (
	pb "github.com/tdex-network/tdex-protobuf/generated/go/trade"
)

// TradeComplete is the domain controller for the TradeComplete RPC
func (*Server) TradeComplete(req *pb.TradeCompleteRequest, stream pb.Trade_TradeCompleteServer) error {
	if err := stream.Send(&pb.TradeCompleteReply{}); err != nil {
		return err
	}
	return nil
}
