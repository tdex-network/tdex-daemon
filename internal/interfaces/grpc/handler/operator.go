package grpchandler

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const readOnlyTx = true

type operatorHandler struct {
	pb.UnimplementedOperatorServer
	operatorSvc application.OperatorService
	dbManager   ports.DbManager
}

// NewOperatorHandler is a constructor function returning an protobuf OperatorServer.
func NewOperatorHandler(
	operatorSvc application.OperatorService,
	dbManager ports.DbManager,
) pb.OperatorServer {
	return newOperatorHandler(operatorSvc, dbManager)
}

func newOperatorHandler(
	operatorSvc application.OperatorService,
	dbManager ports.DbManager,
) *operatorHandler {
	return &operatorHandler{
		operatorSvc: operatorSvc,
		dbManager:   dbManager,
	}
}

func (o operatorHandler) DepositMarket(
	ctx context.Context,
	req *pb.DepositMarketRequest,
) (*pb.DepositMarketReply, error) {
	return o.depositMarket(ctx, req)
}

func (o operatorHandler) DepositFeeAccount(
	ctx context.Context,
	req *pb.DepositFeeAccountRequest,
) (*pb.DepositFeeAccountReply, error) {
	return o.depositFeeAccount(ctx, req)
}

func (o operatorHandler) OpenMarket(
	ctx context.Context,
	req *pb.OpenMarketRequest,
) (*pb.OpenMarketReply, error) {
	return o.openMarket(ctx, req)
}

func (o operatorHandler) CloseMarket(
	ctx context.Context,
	req *pb.CloseMarketRequest,
) (*pb.CloseMarketReply, error) {
	return o.closeMarket(ctx, req)
}

func (o operatorHandler) UpdateMarketFee(
	ctx context.Context,
	req *pb.UpdateMarketFeeRequest,
) (*pb.UpdateMarketFeeReply, error) {
	return o.updateMarketFee(ctx, req)
}

func (o operatorHandler) UpdateMarketPrice(
	ctx context.Context,
	req *pb.UpdateMarketPriceRequest,
) (*pb.UpdateMarketPriceReply, error) {
	return o.updateMarketPrice(ctx, req)
}

func (o operatorHandler) UpdateMarketStrategy(
	ctx context.Context,
	req *pb.UpdateMarketStrategyRequest,
) (*pb.UpdateMarketStrategyReply, error) {
	return o.updateMarketStrategy(ctx, req)
}

func (o operatorHandler) ListSwaps(
	ctx context.Context,
	req *pb.ListSwapsRequest,
) (*pb.ListSwapsReply, error) {
	return o.listSwaps(ctx, req)
}

func (o operatorHandler) WithdrawMarket(
	ctx context.Context,
	req *pb.WithdrawMarketRequest,
) (*pb.WithdrawMarketReply, error) {
	return o.withdrawMarket(ctx, req)
}

func (o operatorHandler) BalanceFeeAccount(
	ctx context.Context,
	req *pb.BalanceFeeAccountRequest,
) (*pb.BalanceFeeAccountReply, error) {
	return o.balanceFeeAccount(ctx, req)
}

func (o operatorHandler) ListDepositMarket(
	ctx context.Context,
	req *pb.ListDepositMarketRequest,
) (*pb.ListDepositMarketReply, error) {
	return o.listDepositMarket(ctx, req)
}

func (o operatorHandler) ListMarket(
	ctx context.Context,
	req *pb.ListMarketRequest,
) (*pb.ListMarketReply, error) {
	return o.listMarket(ctx, req)
}

func (o operatorHandler) ReportMarketFee(
	ctx context.Context,
	req *pb.ReportMarketFeeRequest,
) (*pb.ReportMarketFeeReply, error) {
	return o.reportMarketFee(ctx, req)
}

func (o operatorHandler) depositMarket(
	reqCtx context.Context,
	req *pb.DepositMarketRequest,
) (*pb.DepositMarketReply, error) {
	res, err := o.dbManager.RunTransaction(
		reqCtx,
		!readOnlyTx,
		func(ctx context.Context) (interface{}, error) {
			addresses, err := o.operatorSvc.DepositMarket(
				ctx,
				req.GetMarket().GetBaseAsset(),
				req.GetMarket().GetQuoteAsset(),
				int(req.GetNumOfAddresses()),
			)
			if err != nil {
				return nil, err
			}

			return &pb.DepositMarketReply{Addresses: addresses}, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return res.(*pb.DepositMarketReply), nil
}

func (o operatorHandler) depositFeeAccount(
	reqCtx context.Context,
	req *pb.DepositFeeAccountRequest,
) (*pb.DepositFeeAccountReply, error) {
	res, err := o.dbManager.RunTransaction(
		reqCtx,
		!readOnlyTx,
		func(ctx context.Context) (interface{}, error) {
			addr, blindingKey, err := o.operatorSvc.DepositFeeAccount(ctx)
			if err != nil {
				return nil, err
			}

			return &pb.DepositFeeAccountReply{
				Address:  addr,
				Blinding: blindingKey,
			}, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res.(*pb.DepositFeeAccountReply), nil
}

func (o operatorHandler) openMarket(
	reqCtx context.Context,
	req *pb.OpenMarketRequest,
) (*pb.OpenMarketReply, error) {
	market := req.GetMarket()
	if err := validateMarket(market); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	res, err := o.dbManager.RunTransaction(
		reqCtx,
		!readOnlyTx,
		func(ctx context.Context) (interface{}, error) {
			if err := o.operatorSvc.OpenMarket(
				ctx,
				market.GetBaseAsset(),
				market.GetQuoteAsset(),
			); err != nil {
				return nil, err
			}

			return &pb.OpenMarketReply{}, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res.(*pb.OpenMarketReply), nil
}

func (o operatorHandler) closeMarket(
	reqCtx context.Context,
	req *pb.CloseMarketRequest,
) (*pb.CloseMarketReply, error) {
	market := req.GetMarket()
	if err := validateMarket(market); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	res, err := o.dbManager.RunTransaction(
		reqCtx,
		!readOnlyTx,
		func(ctx context.Context) (interface{}, error) {
			if err := o.operatorSvc.CloseMarket(
				ctx,
				market.GetBaseAsset(),
				market.GetQuoteAsset(),
			); err != nil {
				return nil, err
			}

			return &pb.CloseMarketReply{}, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res.(*pb.CloseMarketReply), nil
}

func (o operatorHandler) updateMarketFee(
	reqCtx context.Context,
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

	res, err := o.dbManager.RunTransaction(
		reqCtx,
		!readOnlyTx,
		func(ctx context.Context) (interface{}, error) {
			result, err := o.operatorSvc.UpdateMarketFee(ctx, mwf)
			if err != nil {
				return nil, err
			}

			return &pb.UpdateMarketFeeReply{
				MarketWithFee: &pbtypes.MarketWithFee{
					Market: &pbtypes.Market{
						BaseAsset:  result.BaseAsset,
						QuoteAsset: result.QuoteAsset,
					},
					Fee: &pbtypes.Fee{
						Asset:      result.FeeAsset,
						BasisPoint: result.BasisPoint,
					},
				},
			}, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res.(*pb.UpdateMarketFeeReply), nil
}

func (o operatorHandler) updateMarketPrice(
	reqCtx context.Context,
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

	res, err := o.dbManager.RunPricesTransaction(
		reqCtx,
		!readOnlyTx,
		func(ctx context.Context) (interface{}, error) {
			if err := o.operatorSvc.UpdateMarketPrice(ctx, mwp); err != nil {
				return nil, err
			}

			return &pb.UpdateMarketPriceReply{}, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res.(*pb.UpdateMarketPriceReply), nil
}

func (o operatorHandler) updateMarketStrategy(
	reqCtx context.Context,
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

	res, err := o.dbManager.RunTransaction(
		reqCtx,
		!readOnlyTx,
		func(ctx context.Context) (interface{}, error) {
			if err := o.operatorSvc.UpdateMarketStrategy(ctx, ms); err != nil {
				return nil, err
			}

			return &pb.UpdateMarketStrategyReply{}, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res.(*pb.UpdateMarketStrategyReply), nil
}

func (o operatorHandler) listSwaps(
	ctx context.Context,
	req *pb.ListSwapsRequest,
) (*pb.ListSwapsReply, error) {
	res, err := o.dbManager.RunTransaction(
		ctx,
		readOnlyTx,
		func(ctx context.Context) (interface{}, error) {
			swapInfos, err := o.operatorSvc.ListSwaps(ctx)
			if err != nil {
				return nil, err
			}

			pbSwapInfos := make([]*pb.SwapInfo, len(swapInfos), len(swapInfos))

			for index, swapInfo := range swapInfos {
				pbSwapInfos[index] = &pb.SwapInfo{
					Status:  pb.SwapStatus(swapInfo.Status),
					AmountP: swapInfo.AmountP,
					AssetP:  swapInfo.AssetP,
					AmountR: swapInfo.AmountR,
					AssetR:  swapInfo.AssetR,
					MarketFee: &pbtypes.Fee{
						Asset:      swapInfo.MarketFee.FeeAsset,
						BasisPoint: swapInfo.MarketFee.BasisPoint,
					},
					RequestTimeUnix:  swapInfo.RequestTimeUnix,
					AcceptTimeUnix:   swapInfo.AcceptTimeUnix,
					CompleteTimeUnix: swapInfo.RequestTimeUnix,
					ExpiryTimeUnix:   swapInfo.ExpiryTimeUnix,
				}
			}

			return &pb.ListSwapsReply{Swaps: pbSwapInfos}, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res.(*pb.ListSwapsReply), nil
}

func (o operatorHandler) withdrawMarket(
	reqCtx context.Context,
	req *pb.WithdrawMarketRequest,
) (*pb.WithdrawMarketReply, error) {
	market := req.GetMarket()
	if err := validateMarket(market); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	wm := application.WithdrawMarketReq{
		Market: application.Market{
			BaseAsset:  req.GetMarket().GetBaseAsset(),
			QuoteAsset: req.GetMarket().GetQuoteAsset(),
		},
		BalanceToWithdraw: application.Balance{
			BaseAmount:  req.GetBalanceToWithdraw().GetBaseAmount(),
			QuoteAmount: req.GetBalanceToWithdraw().GetQuoteAmount(),
		},
		MillisatPerByte: req.GetMillisatPerByte(),
		Address:         req.GetAddress(),
		Push:            req.GetPush(),
	}

	res, err := o.dbManager.RunTransaction(
		reqCtx,
		!readOnlyTx,
		func(ctx context.Context) (interface{}, error) {
			rawTx, err := o.operatorSvc.WithdrawMarketFunds(ctx, wm)
			if err != nil {
				return nil, err
			}

			return &pb.WithdrawMarketReply{RawTx: rawTx}, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res.(*pb.WithdrawMarketReply), nil
}

func (o operatorHandler) balanceFeeAccount(
	ctx context.Context,
	req *pb.BalanceFeeAccountRequest,
) (*pb.BalanceFeeAccountReply, error) {
	res, err := o.dbManager.RunTransaction(
		ctx,
		readOnlyTx,
		func(ctx context.Context) (interface{}, error) {
			balance, err := o.operatorSvc.FeeAccountBalance(ctx)
			if err != nil {
				return nil, err
			}

			return &pb.BalanceFeeAccountReply{Balance: balance}, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res.(*pb.BalanceFeeAccountReply), nil
}

func (o operatorHandler) listDepositMarket(
	ctx context.Context,
	req *pb.ListDepositMarketRequest,
) (*pb.ListDepositMarketReply, error) {
	if err := validateMarket(req.GetMarket()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	res, err := o.dbManager.RunTransaction(
		ctx,
		readOnlyTx,
		func(ctx context.Context) (interface{}, error) {
			addresses, err := o.operatorSvc.ListMarketExternalAddresses(
				ctx,
				application.Market{
					BaseAsset:  req.GetMarket().GetBaseAsset(),
					QuoteAsset: req.GetMarket().GetQuoteAsset(),
				},
			)
			if err != nil {
				return nil, err
			}

			return &pb.ListDepositMarketReply{Address: addresses}, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res.(*pb.ListDepositMarketReply), nil
}

// ListMarket returns the result of the ListMarket method of the operator service.
func (o operatorHandler) listMarket(ctx context.Context, req *pb.ListMarketRequest) (*pb.ListMarketReply, error) {
	res, err := o.dbManager.RunTransaction(
		ctx,
		readOnlyTx,
		func(ctx context.Context) (interface{}, error) {
			marketInfos, err := o.operatorSvc.ListMarket(ctx)
			if err != nil {
				return nil, err
			}

			pbMarketInfos := make([]*pb.MarketInfo, len(marketInfos), len(marketInfos))

			for index, marketInfo := range marketInfos {
				pbMarketInfos[index] = &pb.MarketInfo{
					Fee: &pbtypes.Fee{
						BasisPoint: marketInfo.Fee.BasisPoint,
						Asset:      marketInfo.Fee.FeeAsset,
					},
					Market: &pbtypes.Market{
						BaseAsset:  marketInfo.Market.BaseAsset,
						QuoteAsset: marketInfo.Market.QuoteAsset,
					},
					Tradable:     marketInfo.Tradable,
					StrategyType: pb.StrategyType(marketInfo.StrategyType),
				}
			}

			return &pb.ListMarketReply{Markets: pbMarketInfos}, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res.(*pb.ListMarketReply), nil
}

func (o operatorHandler) reportMarketFee(
	ctx context.Context,
	req *pb.ReportMarketFeeRequest,
) (*pb.ReportMarketFeeReply, error) {
	if err := validateMarket(req.GetMarket()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	res, err := o.dbManager.RunTransaction(
		ctx,
		readOnlyTx,
		func(ctx context.Context) (interface{}, error) {
			report, err := o.operatorSvc.GetCollectedMarketFee(
				ctx,
				application.Market{
					BaseAsset:  req.GetMarket().GetBaseAsset(),
					QuoteAsset: req.GetMarket().GetQuoteAsset(),
				},
			)
			if err != nil {
				return nil, err
			}

			collectedFees := make([]*pbtypes.Fee, 0)
			for _, v := range report.CollectedFees {
				collectedFees = append(collectedFees, &pbtypes.Fee{
					Asset:      v.FeeAsset,
					BasisPoint: v.BasisPoint,
				})
			}

			return &pb.ReportMarketFeeReply{
				CollectedFees:              collectedFees,
				TotalCollectedFeesPerAsset: report.TotalCollectedFeesPerAsset,
			}, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res.(*pb.ReportMarketFeeReply), nil
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

	if market.GetBaseAsset() != config.GetString(config.BaseAssetKey) {
		return domain.ErrInvalidBaseAsset
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
