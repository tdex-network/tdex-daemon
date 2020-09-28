package grpchandler

import (
	"context"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"
)

type operatorHandler struct {
	pb.UnimplementedOperatorServer
}

func NewOperatorHandler() pb.OperatorServer {
	return &operatorHandler{}
}

func (o operatorHandler) DepositMarket(
	ctx context.Context,
	request *pb.DepositMarketRequest,
) (*pb.DepositMarketReply, error) {
	panic("implement me")
}

func (o operatorHandler) ListDepositMarket(
	ctx context.Context,
	request *pb.ListDepositMarketRequest,
) (*pb.ListDepositMarketReply, error) {
	panic("implement me")
}

func (o operatorHandler) DepositFeeAccount(
	ctx context.Context,
	request *pb.DepositFeeAccountRequest,
) (*pb.DepositFeeAccountReply, error) {
	panic("implement me")
}

func (o operatorHandler) BalanceFeeAccount(
	ctx context.Context,
	request *pb.BalanceFeeAccountRequest,
) (*pb.BalanceFeeAccountReply, error) {
	panic("implement me")
}

func (o operatorHandler) OpenMarket(
	ctx context.Context,
	request *pb.OpenMarketRequest,
) (*pb.OpenMarketReply, error) {
	panic("implement me")
}

func (o operatorHandler) CloseMarket(
	ctx context.Context,
	request *pb.CloseMarketRequest,
) (*pb.CloseMarketReply, error) {
	panic("implement me")
}

func (o operatorHandler) ListMarket(
	ctx context.Context,
	request *pb.ListMarketRequest,
) (*pb.ListMarketReply, error) {
	panic("implement me")
}

func (o operatorHandler) UpdateMarketFee(
	ctx context.Context,
	request *pb.UpdateMarketFeeRequest,
) (*pb.UpdateMarketFeeReply, error) {
	panic("implement me")
}

func (o operatorHandler) UpdateMarketPrice(
	ctx context.Context,
	request *pb.UpdateMarketPriceRequest,
) (*pb.UpdateMarketPriceReply, error) {
	panic("implement me")
}

func (o operatorHandler) UpdateMarketStrategy(
	ctx context.Context,
	request *pb.UpdateMarketStrategyRequest,
) (*pb.UpdateMarketStrategyReply, error) {
	panic("implement me")
}

func (o operatorHandler) WithdrawMarket(
	ctx context.Context,
	request *pb.WithdrawMarketRequest,
) (*pb.WithdrawMarketReply,
	error) {
	panic("implement me")
}

func (o operatorHandler) ListSwaps(
	ctx context.Context,
	request *pb.ListSwapsRequest,
) (*pb.ListSwapsReply, error) {
	panic("implement me")
}

func (o operatorHandler) ReportMarketFee(
	ctx context.Context,
	request *pb.ReportMarketFeeRequest,
) (*pb.ReportMarketFeeReply, error) {
	panic("implement me")
}
