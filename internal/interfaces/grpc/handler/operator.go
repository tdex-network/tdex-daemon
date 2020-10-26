package grpchandler

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"
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
	address, err := o.operatorSvc.DepositMarket(ctx, req.GetMarket().GetQuoteAsset())
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
	market := req.GetMarket()
	if err := validateMarket(market); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := o.operatorSvc.OpenMarket(
		ctx,
		market.GetBaseAsset(),
		market.GetQuoteAsset(),
	); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.OpenMarketReply{}, nil
}

func (o operatorHandler) CloseMarket(
	ctx context.Context,
	req *pb.CloseMarketRequest,
) (*pb.CloseMarketReply, error) {
	market := req.GetMarket()
	if err := validateMarket(market); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := o.operatorSvc.CloseMarket(
		ctx,
		market.GetBaseAsset(),
		market.GetQuoteAsset(),
	); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CloseMarketReply{}, nil
}

func (o operatorHandler) UpdateMarketFee(
	ctx context.Context,
	req *pb.UpdateMarketFeeRequest,
) (*pb.UpdateMarketFeeReply, error) {
	marketWithFee := req.GetMarketWithFee()
	if err := validateMarketWithFee(marketWithFee); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	mwf := application.MarketWithFee{
		Market: application.Market{
			BaseAsset:  marketWithFee.GetMarket().GetBaseAsset(),
			QuoteAsset: marketWithFee.GetMarket().GetQuoteAsset(),
		},
		Fee: application.Fee{
			FeeAsset:   marketWithFee.GetFee().GetAsset(),
			BasisPoint: marketWithFee.GetFee().GetBasisPoint(),
		},
	}
	res, err := o.operatorSvc.UpdateMarketFee(ctx, mwf)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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
	market := req.GetMarket()
	if err := validateMarket(market); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	price := req.GetPrice()
	if err := validatePrice(price); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	mwp := application.MarketWithPrice{
		Market: application.Market{
			BaseAsset:  market.GetBaseAsset(),
			QuoteAsset: market.GetQuoteAsset(),
		},
		Price: application.Price{
			BasePrice:  decimal.NewFromFloat32(price.GetBasePrice()),
			QuotePrice: decimal.NewFromFloat32(price.GetQuotePrice()),
		},
	}
	if err := o.operatorSvc.UpdateMarketPrice(ctx, mwp); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateMarketPriceReply{}, nil
}

func (o operatorHandler) UpdateMarketStrategy(
	ctx context.Context,
	req *pb.UpdateMarketStrategyRequest,
) (*pb.UpdateMarketStrategyReply, error) {
	market := req.GetMarket()
	if err := validateMarket(market); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	strategyType := req.GetStrategyType()
	if err := validateStrategyType(strategyType); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ms := application.MarketStrategy{
		Market: application.Market{
			BaseAsset:  market.GetBaseAsset(),
			QuoteAsset: market.GetQuoteAsset(),
		},
		Strategy: domain.StrategyType(strategyType),
	}
	if err := o.operatorSvc.UpdateMarketStrategy(ctx, ms); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateMarketStrategyReply{}, nil
}

func (o operatorHandler) ListSwaps(
	ctx context.Context,
	req *pb.ListSwapsRequest,
) (*pb.ListSwapsReply, error) {
	swaps, err := o.operatorSvc.ListSwaps(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return swaps, nil
}

func validateMarketWithFee(marketWithFee *pbtypes.MarketWithFee) error {
	if marketWithFee == nil {
		return errors.New("market with fee is null")
	}

	if err := validateMarket(marketWithFee.GetMarket()); err != nil {
		return err
	}

	if err := validateFee(marketWithFee.GetFee()); err != nil {
		return err
	}

	return nil
}

func validateMarket(market *pbtypes.Market) error {
	if market == nil {
		return errors.New("market is null")
	}
	if len(market.GetBaseAsset()) <= 0 || len(market.GetQuoteAsset()) <= 0 {
		return errors.New("base asset or quote asset are null")
	}
	return nil
}

func validateFee(fee *pbtypes.Fee) error {
	if fee == nil {
		return errors.New("fee is null")
	}
	if len(fee.GetAsset()) <= 0 {
		return errors.New("fee asset is null")
	}
	if fee.GetBasisPoint() <= 0 {
		return errors.New("fee basis point is too low")
	}
	return nil
}

func validatePrice(price *pbtypes.Price) error {
	if price == nil {
		return errors.New("price is null")
	}
	if price.GetBasePrice() <= 0 || price.GetQuotePrice() <= 0 {
		return errors.New("base or quote price are too low")
	}
	return nil
}

func validateStrategyType(sType pb.StrategyType) error {
	if domain.StrategyType(sType) < domain.StrategyTypePluggable ||
		domain.StrategyType(sType) > domain.StrategyTypeUnbalanced {
		return errors.New("strategy type is unknown")
	}
	return nil
}
