package grpchandler

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const readOnlyTx = true

type operatorHandler struct {
	pb.UnimplementedOperatorServer
	operatorSvc application.OperatorService
}

// NewOperatorHandler is a constructor function returning an protobuf OperatorServer.
func NewOperatorHandler(
	operatorSvc application.OperatorService,
) pb.OperatorServer {
	return newOperatorHandler(operatorSvc)
}

func newOperatorHandler(
	operatorSvc application.OperatorService,
) *operatorHandler {
	return &operatorHandler{
		operatorSvc: operatorSvc,
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

func (o operatorHandler) ClaimMarketDeposit(
	ctx context.Context,
	req *pb.ClaimMarketDepositRequest,
) (*pb.ClaimMarketDepositReply, error) {
	return o.claimMarketDeposit(ctx, req)
}

func (o operatorHandler) ClaimFeeDeposit(
	ctx context.Context,
	req *pb.ClaimFeeDepositRequest,
) (*pb.ClaimFeeDepositReply, error) {
	return o.claimFeeDeposit(ctx, req)
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

func (o operatorHandler) ReloadUtxos(
	ctx context.Context,
	rew *pb.ReloadUtxosRequest,
) (*pb.ReloadUtxosReply, error) {
	if err := o.operatorSvc.ReloadUtxos(ctx); err != nil {
		return nil, err
	}
	return &pb.ReloadUtxosReply{}, nil
}

func (o operatorHandler) ListUtxos(
	ctx context.Context,
	req *pb.ListUtxosRequest,
) (*pb.ListUtxosReply, error) {
	return o.listUtxos(ctx, req)
}

func (o operatorHandler) DropMarket(
	ctx context.Context,
	req *pb.DropMarketRequest,
) (*pb.DropMarketReply, error) {
	return o.dropMarket(ctx, req)
}

func (o operatorHandler) dropMarket(
	ctx context.Context,
	req *pb.DropMarketRequest,
) (*pb.DropMarketReply, error) {
	if err := o.operatorSvc.DropMarket(
		ctx,
		int(req.GetAccountIndex()),
	); err != nil {
		return nil, err
	}

	return &pb.DropMarketReply{}, nil
}

func (o operatorHandler) listUtxos(
	ctx context.Context,
	req *pb.ListUtxosRequest,
) (*pb.ListUtxosReply, error) {
	utxoInfoPerAccount, err := o.operatorSvc.ListUtxos(
		ctx,
	)
	if err != nil {
		return nil, err
	}

	infoPerAccount := make(map[uint64]*pb.UtxoInfoList)

	for k, v := range utxoInfoPerAccount {
		uLen := len(v.Unspents)
		unspents := make([]*pb.UtxoInfo, uLen, uLen)
		for i, u := range v.Unspents {
			unspents[i] = &pb.UtxoInfo{
				Outpoint: &pb.TxOutpoint{
					Hash:  u.Outpoint.Hash,
					Index: int32(u.Outpoint.Index),
				},
				Value: u.Value,
				Asset: u.Asset,
			}
		}

		sLen := len(v.Spents)
		spents := make([]*pb.UtxoInfo, sLen, sLen)
		for i, u := range v.Spents {
			spents[i] = &pb.UtxoInfo{
				Outpoint: &pb.TxOutpoint{
					Hash:  u.Outpoint.Hash,
					Index: int32(u.Outpoint.Index),
				},
				Value: u.Value,
				Asset: u.Asset,
			}
		}

		lLen := len(v.Locks)
		locks := make([]*pb.UtxoInfo, lLen, lLen)
		for i, u := range v.Locks {
			locks[i] = &pb.UtxoInfo{
				Outpoint: &pb.TxOutpoint{
					Hash:  u.Outpoint.Hash,
					Index: int32(u.Outpoint.Index),
				},
				Value: u.Value,
				Asset: u.Asset,
			}
		}

		infoPerAccount[k] = &pb.UtxoInfoList{
			Unspents: unspents,
			Spents:   spents,
			Locks:    locks,
		}
	}

	return &pb.ListUtxosReply{
		InfoPerAccount: infoPerAccount,
	}, nil
}

func (o operatorHandler) depositMarket(
	ctx context.Context,
	req *pb.DepositMarketRequest,
) (*pb.DepositMarketReply, error) {
	addressesAndKeys, err := o.operatorSvc.DepositMarket(
		ctx,
		req.GetMarket().GetBaseAsset(),
		req.GetMarket().GetQuoteAsset(),
		int(req.GetNumOfAddresses()),
	)
	if err != nil {
		return nil, err
	}
	aLen := len(addressesAndKeys)
	addresses := make([]string, aLen, aLen)
	for i, a := range addressesAndKeys {
		addresses[i] = a.Address
	}

	return &pb.DepositMarketReply{Addresses: addresses}, nil
}

func (o operatorHandler) depositFeeAccount(
	ctx context.Context,
	req *pb.DepositFeeAccountRequest,
) (*pb.DepositFeeAccountReply, error) {
	deposits, err := o.operatorSvc.DepositFeeAccount(
		ctx,
		int(req.GetNumOfAddresses()),
	)
	if err != nil {
		return nil, err
	}

	addressesAndKeys := make([]*pbtypes.AddressWithBlindingKey, 0, len(deposits))
	for _, d := range deposits {
		addressesAndKeys = append(addressesAndKeys, &pbtypes.AddressWithBlindingKey{
			Address:  d.Address,
			Blinding: d.BlindingKey,
		})
	}

	return &pb.DepositFeeAccountReply{
		AddressWithBlindingKey: addressesAndKeys,
	}, nil
}

func (o operatorHandler) openMarket(
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
		return nil, err
	}

	return &pb.OpenMarketReply{}, nil
}

func (o operatorHandler) closeMarket(
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
		return nil, err
	}

	return &pb.CloseMarketReply{}, nil
}

func (o operatorHandler) updateMarketFee(
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
			BasisPoint: marketWithFee.GetFee().GetBasisPoint(),
		},
	}

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
				BasisPoint: result.BasisPoint,
			},
		},
	}, nil
}

func (o operatorHandler) updateMarketPrice(
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
		return nil, err
	}

	return &pb.UpdateMarketPriceReply{}, nil
}

func (o operatorHandler) updateMarketStrategy(
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
		return nil, err
	}

	return &pb.UpdateMarketStrategyReply{}, nil
}

func (o operatorHandler) listSwaps(
	ctx context.Context,
	req *pb.ListSwapsRequest,
) (*pb.ListSwapsReply, error) {
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
				BasisPoint: swapInfo.MarketFee.BasisPoint,
			},
			RequestTimeUnix:  swapInfo.RequestTimeUnix,
			AcceptTimeUnix:   swapInfo.AcceptTimeUnix,
			CompleteTimeUnix: swapInfo.CompleteTimeUnix,
			ExpiryTimeUnix:   swapInfo.ExpiryTimeUnix,
		}
	}

	return &pb.ListSwapsReply{Swaps: pbSwapInfos}, nil
}

func (o operatorHandler) withdrawMarket(
	ctx context.Context,
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
		Push:            true,
	}

	rawTx, err := o.operatorSvc.WithdrawMarketFunds(ctx, wm)
	if err != nil {
		return nil, err
	}

	return &pb.WithdrawMarketReply{RawTx: rawTx}, nil
}

func (o operatorHandler) balanceFeeAccount(
	ctx context.Context,
	req *pb.BalanceFeeAccountRequest,
) (*pb.BalanceFeeAccountReply, error) {
	balance, err := o.operatorSvc.FeeAccountBalance(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.BalanceFeeAccountReply{Balance: balance}, nil
}

func (o operatorHandler) claimMarketDeposit(
	ctx context.Context,
	req *pb.ClaimMarketDepositRequest,
) (*pb.ClaimMarketDepositReply, error) {
	outputs := make([]application.TxOutpoint, 0, len(req.GetOutpoints()))
	for _, v := range req.GetOutpoints() {
		outputs = append(outputs, application.TxOutpoint{
			Hash:  v.Hash,
			Index: int(v.Index),
		})
	}

	if err := o.operatorSvc.ClaimMarketDeposit(
		ctx,
		application.Market{
			BaseAsset:  req.GetMarket().GetBaseAsset(),
			QuoteAsset: req.GetMarket().GetQuoteAsset(),
		},
		outputs,
	); err != nil {
		return nil, err
	}

	return &pb.ClaimMarketDepositReply{}, nil
}

func (o operatorHandler) claimFeeDeposit(
	ctx context.Context,
	req *pb.ClaimFeeDepositRequest,
) (*pb.ClaimFeeDepositReply, error) {
	outputs := make([]application.TxOutpoint, 0, len(req.GetOutpoints()))
	for _, v := range req.GetOutpoints() {
		outputs = append(outputs, application.TxOutpoint{
			Hash:  v.Hash,
			Index: int(v.Index),
		})
	}

	if err := o.operatorSvc.ClaimFeeDeposit(
		ctx,
		outputs,
	); err != nil {
		return nil, err
	}

	return &pb.ClaimFeeDepositReply{}, nil
}

func (o operatorHandler) listDepositMarket(
	ctx context.Context,
	req *pb.ListDepositMarketRequest,
) (*pb.ListDepositMarketReply, error) {
	if err := validateMarket(req.GetMarket()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

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
}

// ListMarket returns the result of the ListMarket method of the operator service.
func (o operatorHandler) listMarket(ctx context.Context, req *pb.ListMarketRequest) (*pb.ListMarketReply, error) {
	marketInfos, err := o.operatorSvc.ListMarket(ctx)
	if err != nil {
		return nil, err
	}

	pbMarketInfos := make([]*pb.MarketInfo, 0, len(marketInfos))

	for _, marketInfo := range marketInfos {
		basePrice, _ := marketInfo.Price.BasePrice.BigFloat().Float32()
		quotePrice, _ := marketInfo.Price.QuotePrice.BigFloat().Float32()

		pbMarketInfos = append(pbMarketInfos, &pb.MarketInfo{
			Market: &pbtypes.Market{
				BaseAsset:  marketInfo.Market.BaseAsset,
				QuoteAsset: marketInfo.Market.QuoteAsset,
			},
			Fee: &pbtypes.Fee{
				BasisPoint: marketInfo.Fee.BasisPoint,
			},
			Tradable:     marketInfo.Tradable,
			StrategyType: pb.StrategyType(marketInfo.StrategyType),
			AccountIndex: marketInfo.AccountIndex,
			Price: &pbtypes.Price{
				BasePrice:  basePrice,
				QuotePrice: quotePrice,
			},
		})
	}

	return &pb.ListMarketReply{Markets: pbMarketInfos}, nil
}

func (o operatorHandler) reportMarketFee(
	ctx context.Context,
	req *pb.ReportMarketFeeRequest,
) (*pb.ReportMarketFeeReply, error) {
	if err := validateMarket(req.GetMarket()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

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

	collectedFees := make([]*pb.FeeInfo, 0)
	for _, fee := range report.CollectedFees {
		marketPrice, _ := fee.MarketPrice.BigFloat().Float32()
		collectedFees = append(collectedFees, &pb.FeeInfo{
			TradeId:     fee.TradeID,
			BasisPoint:  fee.BasisPoint,
			Asset:       fee.Asset,
			Amount:      fee.Amount,
			MarketPrice: marketPrice,
		})
	}

	return &pb.ReportMarketFeeReply{
		CollectedFees:              collectedFees,
		TotalCollectedFeesPerAsset: report.TotalCollectedFeesPerAsset,
	}, nil
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
		return domain.ErrMarketInvalidBaseAsset
	}
	return nil
}

func validateFee(fee *pbtypes.Fee) error {
	if fee == nil {
		return errors.New("fee is null")
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
