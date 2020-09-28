package grpchandler

import (
	"context"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/trade"
)

type traderHandler struct {
	pb.UnimplementedTradeServer
}

func NewTraderHandler() pb.TradeServer {
	return &traderHandler{}
}

func (t traderHandler) Markets(
	ctx context.Context,
	request *pb.MarketsRequest,
) (*pb.MarketsReply, error) {
	panic("implement me")
}

func (t traderHandler) Balances(
	ctx context.Context,
	request *pb.BalancesRequest,
) (*pb.BalancesReply, error) {
	panic("implement me")
}

func (t traderHandler) MarketPrice(
	ctx context.Context,
	request *pb.MarketPriceRequest,
) (*pb.MarketPriceReply, error) {
	panic("implement me")
}

func (t traderHandler) TradePropose(
	request *pb.TradeProposeRequest,
	server pb.Trade_TradeProposeServer,
) error {
	panic("implement me")
}

func (t traderHandler) TradeComplete(
	request *pb.TradeCompleteRequest,
	server pb.Trade_TradeCompleteServer,
) error {
	panic("implement me")
}
