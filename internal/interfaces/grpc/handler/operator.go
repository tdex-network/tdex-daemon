package grpchandler

import (
	"context"
	"encoding/hex"
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/operator"
	pbw "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/wallet"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"
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

func (o operatorHandler) GetInfo(
	ctx context.Context, req *pb.GetInfoRequest,
) (*pb.GetInfoReply, error) {
	return o.getInfo(ctx, req)
}

func (o operatorHandler) GetFeeAddress(
	ctx context.Context, req *pb.GetFeeAddressRequest,
) (*pb.GetFeeAddressReply, error) {
	return o.getFeeAddress(ctx, req)
}

func (o operatorHandler) ListFeeAddresses(
	ctx context.Context, req *pb.ListFeeAddressesRequest,
) (*pb.ListFeeAddressesReply, error) {
	return o.listFeeAddresses(ctx, req)
}

func (o operatorHandler) GetFeeBalance(
	ctx context.Context, req *pb.GetFeeBalanceRequest,
) (*pb.GetFeeBalanceReply, error) {
	return o.getFeeBalance(ctx, req)
}

func (o operatorHandler) ClaimFeeDeposits(
	ctx context.Context, req *pb.ClaimFeeDepositsRequest,
) (*pb.ClaimFeeDepositsReply, error) {
	return o.claimFeeDeposits(ctx, req)
}

func (o operatorHandler) WithdrawFee(
	ctx context.Context, req *pb.WithdrawFeeRequest,
) (*pb.WithdrawFeeReply, error) {
	return o.withdrawFee(ctx, req)
}

func (o operatorHandler) NewMarket(
	ctx context.Context, req *pb.NewMarketRequest,
) (*pb.NewMarketReply, error) {
	return o.newMarket(ctx, req)
}

func (o operatorHandler) GetMarketAddress(
	ctx context.Context, req *pb.GetMarketAddressRequest,
) (*pb.GetMarketAddressReply, error) {
	return o.getMarketAddress(ctx, req)
}

func (o operatorHandler) ListMarketAddresses(
	ctx context.Context, req *pb.ListMarketAddressesRequest,
) (*pb.ListMarketAddressesReply, error) {
	return o.listMarketAddresses(ctx, req)
}

func (o operatorHandler) GetMarketBalance(
	ctx context.Context, req *pb.GetMarketBalanceRequest,
) (*pb.GetMarketBalanceReply, error) {
	return o.getMarketBalance(ctx, req)
}

func (o operatorHandler) ClaimMarketDeposits(
	ctx context.Context, req *pb.ClaimMarketDepositsRequest,
) (*pb.ClaimMarketDepositsReply, error) {
	return o.claimMarketDeposits(ctx, req)
}

func (o operatorHandler) OpenMarket(
	ctx context.Context, req *pb.OpenMarketRequest,
) (*pb.OpenMarketReply, error) {
	return o.openMarket(ctx, req)
}

func (o operatorHandler) CloseMarket(
	ctx context.Context, req *pb.CloseMarketRequest,
) (*pb.CloseMarketReply, error) {
	return o.closeMarket(ctx, req)
}

func (o operatorHandler) DropMarket(
	ctx context.Context, req *pb.DropMarketRequest,
) (*pb.DropMarketReply, error) {
	return o.dropMarket(ctx, req)
}

func (o operatorHandler) GetMarketCollectedSwapFees(
	ctx context.Context, req *pb.GetMarketCollectedSwapFeesRequest,
) (*pb.GetMarketCollectedSwapFeesReply, error) {
	return o.getMarketCollectedSwapFees(ctx, req)
}

func (o operatorHandler) WithdrawMarket(
	ctx context.Context, req *pb.WithdrawMarketRequest,
) (*pb.WithdrawMarketReply, error) {
	return o.withdrawMarket(ctx, req)
}

func (o operatorHandler) UpdateMarketPercentageFee(
	ctx context.Context, req *pb.UpdateMarketPercentageFeeRequest,
) (*pb.UpdateMarketFeeReply, error) {
	return o.updateMarketPercentageFee(ctx, req)
}

func (o operatorHandler) UpdateMarketFixedFee(
	ctx context.Context, req *pb.UpdateMarketFixedFeeRequest,
) (*pb.UpdateMarketFeeReply, error) {
	return o.updateMarketFixedFee(ctx, req)
}

func (o operatorHandler) UpdateMarketPrice(
	ctx context.Context, req *pb.UpdateMarketPriceRequest,
) (*pb.UpdateMarketPriceReply, error) {
	return o.updateMarketPrice(ctx, req)
}

func (o operatorHandler) UpdateMarketStrategy(
	ctx context.Context, req *pb.UpdateMarketStrategyRequest,
) (*pb.UpdateMarketStrategyReply, error) {
	return o.updateMarketStrategy(ctx, req)
}

func (o operatorHandler) GetFeeFragmenterAddress(
	ctx context.Context, req *pb.GetFeeFragmenterAddressRequest,
) (*pb.GetFeeFragmenterAddressReply, error) {
	return o.getFeeFragmenterAddress(ctx, req)
}

func (o operatorHandler) ListFeeFragmenterAddresses(
	ctx context.Context, req *pb.ListFeeFragmenterAddressesRequest,
) (*pb.ListFeeFragmenterAddressesReply, error) {
	return o.listFeeFragmenterAddresses(ctx, req)
}

func (o operatorHandler) GetFeeFragmenterBalance(
	ctx context.Context,
	req *pb.GetFeeFragmenterBalanceRequest,
) (*pb.GetFeeFragmenterBalanceReply, error) {
	return o.getFeeFragmenterBalance(ctx, req)
}

func (o operatorHandler) FeeFragmenterSplitFunds(
	req *pb.FeeFragmenterSplitFundsRequest, stream pb.Operator_FeeFragmenterSplitFundsServer,
) error {
	return o.feeFragmenterSplitFunds(req, stream)
}

func (o operatorHandler) WithdrawFeeFragmenter(
	ctx context.Context, req *pb.WithdrawFeeFragmenterRequest,
) (*pb.WithdrawFeeFragmenterReply, error) {
	return o.withdrawFeeFragmenter(ctx, req)
}

func (o operatorHandler) GetMarketFragmenterAddress(
	ctx context.Context, req *pb.GetMarketFragmenterAddressRequest,
) (*pb.GetMarketFragmenterAddressReply, error) {
	return o.getMarketFragmenterAddress(ctx, req)
}

func (o operatorHandler) ListMarketFragmenterAddresses(
	ctx context.Context, req *pb.ListMarketFragmenterAddressesRequest,
) (*pb.ListMarketFragmenterAddressesReply, error) {
	return o.listMarketFragmenterAddresses(ctx, req)
}

func (o operatorHandler) GetMarketFragmenterBalance(
	ctx context.Context,
	req *pb.GetMarketFragmenterBalanceRequest,
) (*pb.GetMarketFragmenterBalanceReply, error) {
	return o.getMarketFragmenterBalance(ctx, req)
}

func (o operatorHandler) MarketFragmenterSplitFunds(
	req *pb.MarketFragmenterSplitFundsRequest, stream pb.Operator_MarketFragmenterSplitFundsServer,
) error {
	return o.marketFragmenterSplitFunds(req, stream)
}

func (o operatorHandler) WithdrawMarketFragmenter(
	ctx context.Context, req *pb.WithdrawMarketFragmenterRequest,
) (*pb.WithdrawMarketFragmenterReply, error) {
	return o.withdrawMarketFragmenter(ctx, req)
}

func (o operatorHandler) ListMarkets(
	ctx context.Context, req *pb.ListMarketsRequest,
) (*pb.ListMarketsReply, error) {
	return o.listMarkets(ctx, req)
}

func (o operatorHandler) ListTrades(
	ctx context.Context, req *pb.ListTradesRequest,
) (*pb.ListTradesReply, error) {
	return o.listTrades(ctx, req)
}

func (o operatorHandler) ReloadUtxos(
	ctx context.Context, rew *pb.ReloadUtxosRequest,
) (*pb.ReloadUtxosReply, error) {
	if err := o.operatorSvc.ReloadUtxos(ctx); err != nil {
		return nil, err
	}
	return &pb.ReloadUtxosReply{}, nil
}

func (o operatorHandler) ListUtxos(
	ctx context.Context, req *pb.ListUtxosRequest,
) (*pb.ListUtxosReply, error) {
	return o.listUtxos(ctx, req)
}

func (o operatorHandler) AddWebhook(
	ctx context.Context, req *pb.AddWebhookRequest,
) (*pb.AddWebhookReply, error) {
	return o.addWebhook(ctx, req)
}

func (o operatorHandler) RemoveWebhook(
	ctx context.Context, req *pb.RemoveWebhookRequest,
) (*pb.RemoveWebhookReply, error) {
	return o.removeWebhook(ctx, req)
}
func (o operatorHandler) ListWebhooks(
	ctx context.Context, req *pb.ListWebhooksRequest,
) (*pb.ListWebhooksReply, error) {
	return o.listWebhooks(ctx, req)
}

func (o operatorHandler) ListDeposits(
	ctx context.Context, req *pb.ListDepositsRequest,
) (*pb.ListDepositsReply, error) {
	return o.listDeposits(ctx, req)
}

func (o operatorHandler) ListWithdrawals(
	ctx context.Context, req *pb.ListWithdrawalsRequest,
) (*pb.ListWithdrawalsReply, error) {
	return o.listWithdrawals(ctx, req)
}

func (o operatorHandler) getInfo(
	ctx context.Context, _ *pb.GetInfoRequest,
) (*pb.GetInfoReply, error) {
	info, err := o.operatorSvc.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	accountInfo := make([]*pb.AccountInfo, 0, len(info.Accounts))
	for _, a := range info.Accounts {
		accountInfo = append(accountInfo, &pb.AccountInfo{
			AccountIndex:        a.Index,
			DerivationPath:      a.DerivationPath,
			Xpub:                a.Xpub,
			LastExternalDerived: a.LastExternalDerived,
			LastInternalDerived: a.LastInternalDerived,
		})
	}
	return &pb.GetInfoReply{
		RootPath:          info.RootPath,
		MasterBlindingKey: info.MasterBlindingKey,
		AccountInfo:       accountInfo,
	}, nil
}

func (o operatorHandler) getFeeAddress(
	ctx context.Context, req *pb.GetFeeAddressRequest,
) (*pb.GetFeeAddressReply, error) {
	info, err := o.operatorSvc.GetFeeAddress(
		ctx, int(req.GetNumOfAddresses()),
	)
	if err != nil {
		return nil, err
	}

	addressesAndKeys := make([]*pbtypes.AddressWithBlindingKey, 0, len(info))
	for _, d := range info {
		addressesAndKeys = append(addressesAndKeys, &pbtypes.AddressWithBlindingKey{
			Address:  d.Address,
			Blinding: d.BlindingKey,
		})
	}

	return &pb.GetFeeAddressReply{
		AddressWithBlindingKey: addressesAndKeys,
	}, nil
}

func (o operatorHandler) listFeeAddresses(
	ctx context.Context, _ *pb.ListFeeAddressesRequest,
) (*pb.ListFeeAddressesReply, error) {
	info, err := o.operatorSvc.ListFeeExternalAddresses(ctx)
	if err != nil {
		return nil, err
	}

	addressesAndKeys := make([]*pbtypes.AddressWithBlindingKey, 0, len(info))
	for _, i := range info {
		addressesAndKeys = append(addressesAndKeys, &pbtypes.AddressWithBlindingKey{
			Address:  i.Address,
			Blinding: i.BlindingKey,
		})
	}

	return &pb.ListFeeAddressesReply{
		AddressWithBlindingKey: addressesAndKeys,
	}, nil
}

func (o operatorHandler) getFeeBalance(
	ctx context.Context, req *pb.GetFeeBalanceRequest,
) (*pb.GetFeeBalanceReply, error) {
	unlockedBalance, totalBalance, err := o.operatorSvc.GetFeeBalance(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.GetFeeBalanceReply{
		AvailableBalance: uint64(unlockedBalance),
		TotalBalance:     uint64(totalBalance),
	}, nil
}

func (o operatorHandler) claimFeeDeposits(
	ctx context.Context, req *pb.ClaimFeeDepositsRequest,
) (*pb.ClaimFeeDepositsReply, error) {
	outpoints := parseOutpoints(req.GetOutpoints())

	if err := o.operatorSvc.ClaimFeeDeposits(ctx, outpoints); err != nil {
		return nil, err
	}

	return &pb.ClaimFeeDepositsReply{}, nil
}

func (o operatorHandler) withdrawFee(
	ctx context.Context, req *pb.WithdrawFeeRequest,
) (*pb.WithdrawFeeReply, error) {
	args := application.WithdrawFeeReq{
		Amount:          req.GetAmount(),
		Address:         req.GetAddress(),
		Asset:           req.GetAsset(),
		MillisatPerByte: req.GetMillisatsPerByte(),
		Push:            true,
	}
	if err := args.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	_, txid, err := o.operatorSvc.WithdrawFeeFunds(ctx, args)
	if err != nil {
		return nil, err
	}

	return &pb.WithdrawFeeReply{Txid: hex.EncodeToString(txid)}, nil
}

func (o operatorHandler) newMarket(
	ctx context.Context, req *pb.NewMarketRequest,
) (*pb.NewMarketReply, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := o.operatorSvc.NewMarket(ctx, market); err != nil {
		return nil, err
	}

	return &pb.NewMarketReply{}, nil
}

func (o operatorHandler) getMarketAddress(
	ctx context.Context, req *pb.GetMarketAddressRequest,
) (*pb.GetMarketAddressReply, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	info, err := o.operatorSvc.GetMarketAddress(
		ctx, market, int(req.GetNumOfAddresses()),
	)
	if err != nil {
		return nil, err
	}

	addressesAndKeys := make([]*pbtypes.AddressWithBlindingKey, 0, len(info))
	for _, d := range info {
		addressesAndKeys = append(addressesAndKeys, &pbtypes.AddressWithBlindingKey{
			Address:  d.Address,
			Blinding: d.BlindingKey,
		})
	}

	return &pb.GetMarketAddressReply{
		AddressWithBlindingKey: addressesAndKeys,
	}, nil
}

func (o operatorHandler) listMarketAddresses(
	ctx context.Context, req *pb.ListMarketAddressesRequest,
) (*pb.ListMarketAddressesReply, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	info, err := o.operatorSvc.ListMarketExternalAddresses(ctx, market)
	if err != nil {
		return nil, err
	}

	addressesAndKeys := make([]*pbtypes.AddressWithBlindingKey, 0, len(info))
	for _, i := range info {
		addressesAndKeys = append(addressesAndKeys, &pbtypes.AddressWithBlindingKey{
			Address:  i.Address,
			Blinding: i.BlindingKey,
		})
	}

	return &pb.ListMarketAddressesReply{
		AddressWithBlindingKey: addressesAndKeys,
	}, nil
}

func (o operatorHandler) getMarketBalance(
	ctx context.Context, req *pb.GetMarketBalanceRequest,
) (*pb.GetMarketBalanceReply, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	unlockedBalance, totalBalance, err := o.operatorSvc.GetMarketBalance(
		ctx, market,
	)
	if err != nil {
		return nil, err
	}

	return &pb.GetMarketBalanceReply{
		AvailableBalance: &pbtypes.Balance{
			BaseAmount:  unlockedBalance.BaseAmount,
			QuoteAmount: unlockedBalance.QuoteAmount,
		},
		TotalBalance: &pbtypes.Balance{
			BaseAmount:  totalBalance.BaseAmount,
			QuoteAmount: totalBalance.QuoteAmount,
		},
	}, nil
}

func (o operatorHandler) claimMarketDeposits(
	ctx context.Context,
	req *pb.ClaimMarketDepositsRequest,
) (*pb.ClaimMarketDepositsReply, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	outpoints := parseOutpoints(req.GetOutpoints())

	if err := o.operatorSvc.ClaimMarketDeposits(
		ctx, market, outpoints,
	); err != nil {
		return nil, err
	}

	return &pb.ClaimMarketDepositsReply{}, nil
}

func (o operatorHandler) openMarket(
	ctx context.Context, req *pb.OpenMarketRequest,
) (*pb.OpenMarketReply, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := o.operatorSvc.OpenMarket(ctx, market); err != nil {
		return nil, err
	}

	return &pb.OpenMarketReply{}, nil
}

func (o operatorHandler) closeMarket(
	ctx context.Context, req *pb.CloseMarketRequest,
) (*pb.CloseMarketReply, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := o.operatorSvc.CloseMarket(ctx, market); err != nil {
		return nil, err
	}

	return &pb.CloseMarketReply{}, nil
}

func (o operatorHandler) dropMarket(
	ctx context.Context, req *pb.DropMarketRequest,
) (*pb.DropMarketReply, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := o.operatorSvc.DropMarket(ctx, market); err != nil {
		return nil, err
	}

	return &pb.DropMarketReply{}, nil
}

func (o operatorHandler) getMarketCollectedSwapFees(
	ctx context.Context, req *pb.GetMarketCollectedSwapFeesRequest,
) (*pb.GetMarketCollectedSwapFeesReply, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	page := parsePage(req.GetPage())

	report, err := o.operatorSvc.GetMarketCollectedFee(ctx, market, page)
	if err != nil {
		return nil, err
	}

	collectedFees := make([]*pb.FeeInfo, 0)
	for _, fee := range report.CollectedFees {
		marketPrice, _ := fee.MarketPrice.BigFloat().Float32()
		collectedFees = append(collectedFees, &pb.FeeInfo{
			TradeId:             fee.TradeID,
			BasisPoint:          fee.BasisPoint,
			Asset:               fee.Asset,
			PercentageFeeAmount: fee.PercentageFeeAmount,
			FixedFeeAmount:      fee.FixedFeeAmount,
			MarketPrice:         marketPrice,
		})
	}

	return &pb.GetMarketCollectedSwapFeesReply{
		CollectedFees:              collectedFees,
		TotalCollectedFeesPerAsset: report.TotalCollectedFeesPerAsset,
	}, nil
}

func (o operatorHandler) withdrawMarket(
	ctx context.Context, req *pb.WithdrawMarketRequest,
) (*pb.WithdrawMarketReply, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	balanceToWithdraw := parseBalance(req.GetBalanceToWithdraw())

	args := application.WithdrawMarketReq{
		Market:            market,
		BalanceToWithdraw: balanceToWithdraw,
		MillisatPerByte:   req.GetMillisatsPerByte(),
		Address:           req.GetAddress(),
		Push:              true,
	}
	if err := args.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	_, txid, err := o.operatorSvc.WithdrawMarketFunds(ctx, args)
	if err != nil {
		return nil, err
	}

	return &pb.WithdrawMarketReply{Txid: hex.EncodeToString(txid)}, nil
}

func (o operatorHandler) updateMarketPercentageFee(
	ctx context.Context, req *pb.UpdateMarketPercentageFeeRequest,
) (*pb.UpdateMarketFeeReply, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	mwf := application.MarketWithFee{
		Market: market,
		Fee: application.Fee{
			BasisPoint: req.GetBasisPoint(),
		},
	}
	result, err := o.operatorSvc.UpdateMarketPercentageFee(ctx, mwf)
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
				BasisPoint: result.Fee.BasisPoint,
				Fixed: &pbtypes.Fixed{
					BaseFee:  result.FixedBaseFee,
					QuoteFee: result.FixedQuoteFee,
				},
			},
		},
	}, nil
}

func (o operatorHandler) updateMarketFixedFee(
	ctx context.Context, req *pb.UpdateMarketFixedFeeRequest,
) (*pb.UpdateMarketFeeReply, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	fee := parseFixedFee(req.GetFixed())

	mwf := application.MarketWithFee{market, fee}
	result, err := o.operatorSvc.UpdateMarketFixedFee(ctx, mwf)
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
				Fixed: &pbtypes.Fixed{
					BaseFee:  result.FixedBaseFee,
					QuoteFee: result.FixedQuoteFee,
				},
			},
		},
	}, nil
}

func (o operatorHandler) updateMarketPrice(
	ctx context.Context, req *pb.UpdateMarketPriceRequest,
) (*pb.UpdateMarketPriceReply, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	price, err := parsePrice(req.GetPrice())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	mwp := application.MarketWithPrice{market, price}
	if err := o.operatorSvc.UpdateMarketPrice(ctx, mwp); err != nil {
		return nil, err
	}

	return &pb.UpdateMarketPriceReply{}, nil
}

func (o operatorHandler) updateMarketStrategy(
	ctx context.Context, req *pb.UpdateMarketStrategyRequest,
) (*pb.UpdateMarketStrategyReply, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	strategyType, err := parseStrategy(req.GetStrategyType())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ms := application.MarketStrategy{market, strategyType}
	if err := o.operatorSvc.UpdateMarketStrategy(ctx, ms); err != nil {
		return nil, err
	}

	return &pb.UpdateMarketStrategyReply{}, nil
}

func (o operatorHandler) getFeeFragmenterAddress(
	ctx context.Context, req *pb.GetFeeFragmenterAddressRequest,
) (*pb.GetFeeFragmenterAddressReply, error) {
	info, err := o.operatorSvc.GetFeeFragmenterAddress(
		ctx, int(req.GetNumOfAddresses()),
	)
	if err != nil {
		return nil, err
	}

	addressesAndKeys := make([]*pbtypes.AddressWithBlindingKey, 0, len(info))
	for _, d := range info {
		addressesAndKeys = append(addressesAndKeys, &pbtypes.AddressWithBlindingKey{
			Address:  d.Address,
			Blinding: d.BlindingKey,
		})
	}

	return &pb.GetFeeFragmenterAddressReply{
		AddressWithBlindingKey: addressesAndKeys,
	}, nil
}

func (o operatorHandler) listFeeFragmenterAddresses(
	ctx context.Context, req *pb.ListFeeFragmenterAddressesRequest,
) (*pb.ListFeeFragmenterAddressesReply, error) {
	info, err := o.operatorSvc.ListFeeFragmenterExternalAddresses(ctx)
	if err != nil {
		return nil, err
	}

	addressesAndKeys := make([]*pbtypes.AddressWithBlindingKey, 0, len(info))
	for _, d := range info {
		addressesAndKeys = append(addressesAndKeys, &pbtypes.AddressWithBlindingKey{
			Address:  d.Address,
			Blinding: d.BlindingKey,
		})
	}

	return &pb.ListFeeFragmenterAddressesReply{
		AddressWithBlindingKey: addressesAndKeys,
	}, nil
}

func (o operatorHandler) getFeeFragmenterBalance(
	ctx context.Context, req *pb.GetFeeFragmenterBalanceRequest,
) (*pb.GetFeeFragmenterBalanceReply, error) {
	info, err := o.operatorSvc.GetFeeFragmenterBalance(ctx)
	if err != nil {
		return nil, err
	}

	balance := make(map[string]*pbw.BalanceInfo)
	for a, b := range info {
		balance[a] = &pbw.BalanceInfo{
			ConfirmedBalance:   b.ConfirmedBalance,
			UnconfirmedBalance: b.UnconfirmedBalance,
			TotalBalance:       b.TotalBalance,
		}
	}

	return &pb.GetFeeFragmenterBalanceReply{
		Balance: balance,
	}, nil
}

func (o operatorHandler) feeFragmenterSplitFunds(
	req *pb.FeeFragmenterSplitFundsRequest,
	stream pb.Operator_FeeFragmenterSplitFundsServer,
) error {
	chReplies := make(chan application.FragmenterSplitFundsReply)
	go o.operatorSvc.FeeFragmenterSplitFunds(
		stream.Context(), req.GetMaxFragments(), req.GetMillisatsPerByte(),
		chReplies,
	)

	for reply := range chReplies {
		if reply.Err != nil {
			return reply.Err
		}

		if err := stream.Send(&pb.FragmenterSplitFundsReply{
			Message: reply.Msg,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (o operatorHandler) withdrawFeeFragmenter(
	ctx context.Context, req *pb.WithdrawFeeFragmenterRequest,
) (*pb.WithdrawFeeFragmenterReply, error) {
	txid, err := o.operatorSvc.WithdrawFeeFragmenterFunds(
		ctx, req.GetAddress(), req.GetMillisatsPerByte(),
	)
	if err != nil {
		return nil, err
	}

	return &pb.WithdrawFeeFragmenterReply{
		Txid: txid,
	}, nil
}

func (o operatorHandler) getMarketFragmenterAddress(
	ctx context.Context, req *pb.GetMarketFragmenterAddressRequest,
) (*pb.GetMarketFragmenterAddressReply, error) {
	info, err := o.operatorSvc.GetMarketFragmenterAddress(
		ctx, int(req.GetNumOfAddresses()),
	)
	if err != nil {
		return nil, err
	}

	addressesAndKeys := make([]*pbtypes.AddressWithBlindingKey, 0, len(info))
	for _, d := range info {
		addressesAndKeys = append(addressesAndKeys, &pbtypes.AddressWithBlindingKey{
			Address:  d.Address,
			Blinding: d.BlindingKey,
		})
	}

	return &pb.GetMarketFragmenterAddressReply{
		AddressWithBlindingKey: addressesAndKeys,
	}, nil
}

func (o operatorHandler) listMarketFragmenterAddresses(
	ctx context.Context, req *pb.ListMarketFragmenterAddressesRequest,
) (*pb.ListMarketFragmenterAddressesReply, error) {
	info, err := o.operatorSvc.ListMarketFragmenterExternalAddresses(ctx)
	if err != nil {
		return nil, err
	}

	addressesAndKeys := make([]*pbtypes.AddressWithBlindingKey, 0, len(info))
	for _, d := range info {
		addressesAndKeys = append(addressesAndKeys, &pbtypes.AddressWithBlindingKey{
			Address:  d.Address,
			Blinding: d.BlindingKey,
		})
	}

	return &pb.ListMarketFragmenterAddressesReply{
		AddressWithBlindingKey: addressesAndKeys,
	}, nil
}

func (o operatorHandler) getMarketFragmenterBalance(
	ctx context.Context, req *pb.GetMarketFragmenterBalanceRequest,
) (*pb.GetMarketFragmenterBalanceReply, error) {
	info, err := o.operatorSvc.GetMarketFragmenterBalance(ctx)
	if err != nil {
		return nil, err
	}

	balance := make(map[string]*pbw.BalanceInfo)
	for a, b := range info {
		balance[a] = &pbw.BalanceInfo{
			ConfirmedBalance:   b.ConfirmedBalance,
			UnconfirmedBalance: b.UnconfirmedBalance,
			TotalBalance:       b.TotalBalance,
		}
	}

	return &pb.GetMarketFragmenterBalanceReply{
		Balance: balance,
	}, nil
}

func (o operatorHandler) marketFragmenterSplitFunds(
	req *pb.MarketFragmenterSplitFundsRequest,
	stream pb.Operator_MarketFragmenterSplitFundsServer,
) error {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	chReplies := make(chan application.FragmenterSplitFundsReply)
	go o.operatorSvc.MarketFragmenterSplitFunds(
		stream.Context(), market, req.GetMillisatsPerByte(), chReplies,
	)

	for reply := range chReplies {
		if reply.Err != nil {
			return reply.Err
		}

		if err := stream.Send(&pb.FragmenterSplitFundsReply{
			Message: reply.Msg,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (o operatorHandler) withdrawMarketFragmenter(
	ctx context.Context, req *pb.WithdrawMarketFragmenterRequest,
) (*pb.WithdrawMarketFragmenterReply, error) {
	txid, err := o.operatorSvc.WithdrawMarketFragmenterFunds(
		ctx, req.GetAddress(), req.GetMillisatsPerByte(),
	)
	if err != nil {
		return nil, err
	}

	return &pb.WithdrawMarketFragmenterReply{
		Txid: txid,
	}, nil
}

func (o operatorHandler) listTrades(
	ctx context.Context, req *pb.ListTradesRequest,
) (*pb.ListTradesReply, error) {
	page := parsePage(req.GetPage())

	var tradeInfo []application.TradeInfo
	var err error
	if mkt := req.GetMarket(); mkt == nil {
		tradeInfo, err = o.operatorSvc.ListTrades(ctx, page)
	} else {
		market, err := parseMarket(req.GetMarket())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		tradeInfo, err = o.operatorSvc.ListTradesForMarket(ctx, market, page)
	}
	if err != nil {
		return nil, err
	}

	pbTradeInfo := make([]*pb.TradeInfo, 0, len(tradeInfo))
	for _, info := range tradeInfo {
		basePrice, _ := info.Price.BasePrice.Float64()
		quotePrice, _ := info.Price.QuotePrice.Float64()

		pbInfo := &pb.TradeInfo{
			TradeId: info.ID,
			Status: &pb.TradeStatusInfo{
				Status: pb.TradeStatus(info.Status.Code),
				Failed: info.Status.Failed,
			},
			MarketWithFee: &pbtypes.MarketWithFee{
				Market: &pbtypes.Market{
					BaseAsset:  info.MarketWithFee.BaseAsset,
					QuoteAsset: info.MarketWithFee.QuoteAsset,
				},
				Fee: &pbtypes.Fee{
					BasisPoint: info.MarketWithFee.BasisPoint,
					Fixed: &pbtypes.Fixed{
						BaseFee:  info.MarketWithFee.FixedBaseFee,
						QuoteFee: info.MarketWithFee.FixedQuoteFee,
					},
				},
			},
			Price: &pb.TradePrice{
				BasePrice:  basePrice,
				QuotePrice: quotePrice,
			},
			TxUrl:            info.TxURL,
			RequestTimeUnix:  info.RequestTimeUnix,
			AcceptTimeUnix:   info.AcceptTimeUnix,
			CompleteTimeUnix: info.CompleteTimeUnix,
			SettleTimeUnix:   info.SettleTimeUnix,
			ExpiryTimeUnix:   info.ExpiryTimeUnix,
		}

		swapInfoEmpty := info.SwapInfo == application.SwapInfo{}
		if !swapInfoEmpty {
			pbInfo.SwapInfo = &pb.SwapInfo{
				AssetP:  info.SwapInfo.AssetP,
				AmountP: info.SwapInfo.AmountP,
				AssetR:  info.SwapInfo.AssetR,
				AmountR: info.SwapInfo.AmountR,
			}
		}

		failInfoEmpty := info.SwapFailInfo == application.SwapFailInfo{}
		if !failInfoEmpty {
			pbInfo.FailInfo = &pb.SwapFailInfo{
				FailureCode:    uint32(info.SwapFailInfo.Code),
				FailureMessage: info.SwapFailInfo.Message,
			}
		}

		if tt := info.RequestTimeUnix; tt > 0 {
			pbInfo.RequestTimeUtc = time.Unix(int64(tt), 0).UTC().String()
		}
		if tt := info.AcceptTimeUnix; tt > 0 {
			pbInfo.AcceptTimeUtc = time.Unix(int64(tt), 0).UTC().String()
		}
		if tt := info.CompleteTimeUnix; tt > 0 {
			pbInfo.CompleteTimeUtc = time.Unix(int64(tt), 0).UTC().String()
		}
		if tt := info.SettleTimeUnix; tt > 0 {
			pbInfo.SettleTimeUtc = time.Unix(int64(tt), 0).UTC().String()
		}
		if tt := info.ExpiryTimeUnix; tt > 0 {
			pbInfo.ExpiryTimeUtc = time.Unix(int64(tt), 0).UTC().String()
		}

		pbTradeInfo = append(pbTradeInfo, pbInfo)
	}

	return &pb.ListTradesReply{Trades: pbTradeInfo}, nil
}

func (o operatorHandler) listMarkets(
	ctx context.Context, req *pb.ListMarketsRequest,
) (*pb.ListMarketsReply, error) {
	marketInfos, err := o.operatorSvc.ListMarkets(ctx)
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
				Fixed: &pbtypes.Fixed{
					BaseFee:  marketInfo.Fee.FixedBaseFee,
					QuoteFee: marketInfo.Fee.FixedQuoteFee,
				},
			},
			Tradable:     marketInfo.Tradable,
			StrategyType: pb.StrategyType(marketInfo.StrategyType),
			AccountIndex: marketInfo.AccountIndex,
			Price: &pbtypes.Price{
				BasePrice:  basePrice,
				QuotePrice: quotePrice,
			},
			Balance: &pbtypes.Balance{
				BaseAmount:  marketInfo.Balance.BaseAmount,
				QuoteAmount: marketInfo.Balance.QuoteAmount,
			},
		})
	}

	return &pb.ListMarketsReply{Markets: pbMarketInfos}, nil
}

func (o operatorHandler) listUtxos(
	ctx context.Context, req *pb.ListUtxosRequest,
) (*pb.ListUtxosReply, error) {
	var page *application.Page
	if pg := req.GetPage(); pg != nil {
		page = &application.Page{
			Number: int(pg.PageNumber),
			Size:   int(pg.PageSize),
		}
	}
	accountIndex := int(req.GetAccountIndex())

	utxoInfo, err := o.operatorSvc.ListUtxos(ctx, accountIndex, page)
	if err != nil {
		return nil, err
	}

	unspents := toUtxoInfoList(utxoInfo.Unspents)
	spents := toUtxoInfoList(utxoInfo.Spents)
	locks := toUtxoInfoList(utxoInfo.Locks)

	return &pb.ListUtxosReply{
		Unspents: unspents,
		Spents:   spents,
		Locks:    locks,
	}, nil
}

func (o operatorHandler) addWebhook(
	ctx context.Context, req *pb.AddWebhookRequest,
) (*pb.AddWebhookReply, error) {
	hook := application.Webhook{
		ActionType: int(req.GetAction()),
		Endpoint:   req.GetEndpoint(),
		Secret:     req.GetSecret(),
	}
	hookID, err := o.operatorSvc.AddWebhook(ctx, hook)
	if err != nil {
		return nil, err
	}
	return &pb.AddWebhookReply{Id: hookID}, nil
}

func (o operatorHandler) removeWebhook(
	ctx context.Context, req *pb.RemoveWebhookRequest,
) (*pb.RemoveWebhookReply, error) {
	if err := o.operatorSvc.RemoveWebhook(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &pb.RemoveWebhookReply{}, nil
}

func (o operatorHandler) listWebhooks(
	ctx context.Context, req *pb.ListWebhooksRequest,
) (*pb.ListWebhooksReply, error) {
	hooks, err := o.operatorSvc.ListWebhooks(ctx, int(req.GetAction()))
	if err != nil {
		return nil, err
	}
	hooksInfo := make([]*pb.WebhookInfo, 0, len(hooks))
	for _, h := range hooks {
		hooksInfo = append(hooksInfo, &pb.WebhookInfo{
			Id:        h.Id,
			Endpoint:  h.Endpoint,
			IsSecured: h.IsSecured,
		})
	}
	return &pb.ListWebhooksReply{
		WebhookInfo: hooksInfo,
	}, nil
}

func (o operatorHandler) listDeposits(
	ctx context.Context, req *pb.ListDepositsRequest,
) (*pb.ListDepositsReply, error) {
	page := parsePage(req.GetPage())
	deposits, err := o.operatorSvc.ListDeposits(
		ctx, int(req.GetAccountIndex()), page,
	)
	if err != nil {
		return nil, err
	}

	depositsProto := make([]*pb.UtxoInfo, 0, len(deposits))
	for _, v := range deposits {
		depositsProto = append(depositsProto, &pb.UtxoInfo{
			Outpoint: &pb.TxOutpoint{
				Hash:  v.TxID,
				Index: int32(v.VOut),
			},
			Value: v.Value,
			Asset: v.Asset,
		})
	}

	return &pb.ListDepositsReply{
		AccountIndex: req.GetAccountIndex(),
		Deposits:     depositsProto,
	}, err
}

func (o operatorHandler) listWithdrawals(
	ctx context.Context, req *pb.ListWithdrawalsRequest,
) (*pb.ListWithdrawalsReply, error) {
	page := parsePage(req.GetPage())

	withdrawals, err := o.operatorSvc.ListWithdrawals(
		ctx, int(req.GetAccountIndex()), page,
	)

	withdrawalsProto := make([]*pb.Withdrawal, 0, len(withdrawals))
	for _, v := range withdrawals {
		withdrawalsProto = append(withdrawalsProto, &pb.Withdrawal{
			TxId: v.TxID,
			Balance: &pbtypes.Balance{
				BaseAmount:  v.BaseAmount,
				QuoteAmount: v.QuoteAmount,
			},
			Address: v.Address,
		})
	}

	return &pb.ListWithdrawalsReply{
		AccountIndex: req.GetAccountIndex(),
		Withdrawals:  withdrawalsProto,
	}, err
}

func parseMarket(mkt *pbtypes.Market) (market application.Market, err error) {
	var baseAsset, quoteAsset string
	if mkt != nil {
		baseAsset, quoteAsset = mkt.GetBaseAsset(), mkt.GetQuoteAsset()
	}
	m := application.Market{baseAsset, quoteAsset}
	if err = m.Validate(); err != nil {
		return
	}

	market = m
	return
}

func parseOutpoints(outs []*pb.TxOutpoint) []application.TxOutpoint {
	outpoints := make([]application.TxOutpoint, 0, len(outs))
	for _, v := range outs {
		outpoints = append(outpoints, application.TxOutpoint{
			Hash:  v.Hash,
			Index: int(v.Index),
		})
	}
	return outpoints
}

func parsePage(p *pb.Page) *application.Page {
	if p == nil {
		return nil
	}
	return &application.Page{
		Number: int(p.PageNumber),
		Size:   int(p.PageSize),
	}
}

func parseBalance(bal *pbtypes.Balance) application.Balance {
	var baseAmount, quoteAmount uint64
	if bal != nil {
		baseAmount = bal.GetBaseAmount()
		quoteAmount = bal.GetQuoteAmount()
	}
	return application.Balance{baseAmount, quoteAmount}
}

func parseFixedFee(fee *pbtypes.Fixed) application.Fee {
	var baseFee, quoteFee int64
	if fee != nil {
		baseFee = fee.GetBaseFee()
		quoteFee = fee.GetQuoteFee()
	}
	return application.Fee{
		FixedBaseFee:  baseFee,
		FixedQuoteFee: quoteFee,
	}
}

func parsePrice(p *pbtypes.Price) (price application.Price, err error) {
	var basePrice, quotePrice = decimal.NewFromInt(0), decimal.NewFromInt(0)
	if p != nil {
		basePrice = decimal.NewFromFloat32(p.GetBasePrice())
		quotePrice = decimal.NewFromFloat32(p.GetQuotePrice())
	}
	pp := application.Price{
		BasePrice:  basePrice,
		QuotePrice: quotePrice,
	}
	if err = pp.Validate(); err != nil {
		return
	}
	price = pp
	return
}

func parseStrategy(sType pb.StrategyType) (domain.StrategyType, error) {
	strategyType := domain.StrategyType(sType)
	if strategyType < domain.StrategyTypePluggable ||
		strategyType > domain.StrategyTypeUnbalanced {
		return -1, errors.New("strategy type is unknown")
	}
	return strategyType, nil
}

func toUtxoInfoList(list []application.UtxoInfo) []*pb.UtxoInfo {
	res := make([]*pb.UtxoInfo, 0, len(list))
	for _, u := range list {
		res = append(res, &pb.UtxoInfo{
			Outpoint: &pb.TxOutpoint{
				Hash:  u.Outpoint.Hash,
				Index: int32(u.Outpoint.Index),
			},
			Value: u.Value,
			Asset: u.Asset,
		})
	}
	return res
}
