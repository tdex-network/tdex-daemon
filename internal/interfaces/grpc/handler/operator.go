package grpchandler

import (
	"context"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type operatorHandler struct {
	pb.UnimplementedOperatorServer
	operatorSvc application.OperatorService
}

func NewOperatorHandler(operatorSvc application.OperatorService) pb.OperatorServer {
	return &operatorHandler{
		operatorSvc: operatorSvc,
	}
}

func (o operatorHandler) DepositMarket(
	ctx context.Context,
	req *pb.DepositMarketRequest,
) (*pb.DepositMarketReply, error) {
	if req.Market.QuoteAsset == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"quote asset must be populated",
		)
	}
	address, err := o.operatorSvc.DepositMarket(ctx, req.Market.QuoteAsset)
	if err != nil {
		return nil, status.Error(
			codes.Internal,
			err.Error(),
		)
	}
	return &pb.DepositMarketReply{
		Address: address,
	}, nil
}

func (o operatorHandler) DepositFeeAccount(
	ctx context.Context,
	req *pb.DepositFeeAccountRequest,
) (*pb.DepositFeeAccountReply, error) {

	address, blindingKey, err := o.operatorSvc.DepositFeeAccount(ctx)
	if err != nil {
		return nil, status.Error(
			codes.Internal,
			err.Error(),
		)
	}

	return &pb.DepositFeeAccountReply{
		Address:  address,
		Blinding: blindingKey,
	}, nil
}

func (o operatorHandler) OpenMarket(
	ctx context.Context,
	req *pb.OpenMarketRequest,
) (*pb.OpenMarketReply, error) {
	if req.Market.BaseAsset == "" || req.Market.QuoteAsset == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"base asset and quote asset must be populated",
		)
	}

	err := o.operatorSvc.OpenMarket(
		ctx, req.Market.BaseAsset,
		req.Market.QuoteAsset,
	)
	if err != nil {
		return nil, status.Error(
			codes.Internal,
			err.Error(),
		)
	}

	return &pb.OpenMarketReply{}, nil
}

func (o operatorHandler) CloseMarket(
	ctx context.Context,
	req *pb.CloseMarketRequest,
) (*pb.CloseMarketReply, error) {
	if req.Market.BaseAsset == "" || req.Market.QuoteAsset == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"base asset and quote asset must be populated",
		)
	}

	err := o.operatorSvc.CloseMarket(
		ctx, req.Market.BaseAsset,
		req.Market.QuoteAsset,
	)
	if err != nil {
		return nil, status.Error(
			codes.Internal,
			err.Error(),
		)
	}

	return &pb.CloseMarketReply{}, nil
}

func (o operatorHandler) UpdateMarketFee(
	ctx context.Context,
	req *pb.UpdateMarketFeeRequest,
) (*pb.UpdateMarketFeeReply, error) {
	mwf := application.MarketWithFee{
		BaseAsset:  req.MarketWithFee.Market.BaseAsset,
		QuoteAsset: req.MarketWithFee.Market.QuoteAsset,
		Fee: application.Fee{
			FeeAsset:   req.MarketWithFee.Fee.Asset,
			BasisPoint: req.MarketWithFee.Fee.BasisPoint,
		},
	}
	res, err := o.operatorSvc.UpdateMarketFee(ctx, mwf)
	if err != nil {
		return nil, status.Error(
			codes.Internal,
			err.Error(),
		)
	}

	return &pb.UpdateMarketFeeReply{
		MarketWithFee: &pbtypes.MarketWithFee{
			Market: &pbtypes.Market{
				BaseAsset:  res.BaseAsset,
				QuoteAsset: res.QuoteAsset,
			},
			Fee: &pbtypes.Fee{
				Asset:      res.FeeAsset,
				BasisPoint: res.BasisPoint,
			},
		},
	}, nil
}

func (o operatorHandler) UpdateMarketPrice(
	ctx context.Context,
	req *pb.UpdateMarketPriceRequest,
) (*pb.UpdateMarketPriceReply, error) {
	mwp := application.MarketWithPriceReq{
		BaseAsset:  req.Market.BaseAsset,
		QuoteAsset: req.Market.QuoteAsset,
		PriceReq: application.PriceReq{
			BasePrice:  req.Price.BasePrice,
			QuotePrice: req.Price.QuotePrice,
		},
	}
	err := o.operatorSvc.UpdateMarketPrice(ctx, mwp)
	if err != nil {
		return nil, status.Error(
			codes.Internal,
			err.Error(),
		)
	}

	return &pb.UpdateMarketPriceReply{}, nil
}

func (o operatorHandler) UpdateMarketStrategy(
	ctx context.Context,
	req *pb.UpdateMarketStrategyRequest,
) (*pb.UpdateMarketStrategyReply, error) {
	ms := application.MarketStrategyReq{
		BaseAsset:  req.Market.BaseAsset,
		QuoteAsset: req.Market.QuoteAsset,
		Strategy:   domain.StrategyType(req.StrategyType),
	}
	err := o.operatorSvc.UpdateMarketStrategy(ctx, ms)
	if err != nil {
		return nil, status.Error(
			codes.Internal,
			err.Error(),
		)
	}

	return &pb.UpdateMarketStrategyReply{}, nil
}

func (o operatorHandler) ListSwaps(
	ctx context.Context,
	req *pb.ListSwapsRequest,
) (*pb.ListSwapsReply, error) {
	swaps, err := o.operatorSvc.ListSwaps(ctx)
	if err != nil {
		return nil, status.Error(
			codes.Internal,
			err.Error(),
		)
	}

	return swaps, nil
}
