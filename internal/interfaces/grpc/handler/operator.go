package grpchandler

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/operator"
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

func (o operatorHandler) UpdateMarketPercentageFee(
	ctx context.Context,
	req *pb.UpdateMarketPercentageFeeRequest,
) (*pb.UpdateMarketFeeReply, error) {
	return o.updateMarketPercentageFee(ctx, req)
}

func (o operatorHandler) UpdateMarketFixedFee(
	ctx context.Context,
	req *pb.UpdateMarketFixedFeeRequest,
) (*pb.UpdateMarketFeeReply, error) {
	return o.updateMarketFixedFee(ctx, req)
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

func (o operatorHandler) ListTrades(
	ctx context.Context,
	req *pb.ListTradesRequest,
) (*pb.ListTradesReply, error) {
	return o.listTrades(ctx, req)
}

func (o operatorHandler) WithdrawMarket(
	ctx context.Context,
	req *pb.WithdrawMarketRequest,
) (*pb.WithdrawMarketReply, error) {
	return o.withdrawMarket(ctx, req)
}

func (o operatorHandler) WithdrawFee(
	ctx context.Context, req *pb.WithdrawFeeRequest,
) (*pb.WithdrawFeeReply, error) {
	return o.withdrawFee(ctx, req)
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

func (o operatorHandler) AddWebhook(
	ctx context.Context,
	req *pb.AddWebhookRequest,
) (*pb.AddWebhookReply, error) {
	return o.addWebhook(ctx, req)
}

func (o operatorHandler) RemoveWebhook(
	ctx context.Context,
	req *pb.RemoveWebhookRequest,
) (*pb.RemoveWebhookReply, error) {
	return o.removeWebhook(ctx, req)
}
func (o operatorHandler) ListWebhooks(
	ctx context.Context,
	req *pb.ListWebhooksRequest,
) (*pb.ListWebhooksReply, error) {
	return o.listWebhooks(ctx, req)
}

func (o operatorHandler) ListDeposits(
	ctx context.Context,
	req *pb.ListDepositsRequest,
) (*pb.ListDepositsReply, error) {
	return o.listDeposits(ctx, req)
}

func (o operatorHandler) ListWithdrawals(
	ctx context.Context,
	req *pb.ListWithdrawalsRequest,
) (*pb.ListWithdrawalsReply, error) {
	return o.listWithdrawals(ctx, req)
}

func (o operatorHandler) listDeposits(
	ctx context.Context,
	req *pb.ListDepositsRequest,
) (*pb.ListDepositsReply, error) {
	var page *application.Page
	if pg := req.GetPage(); pg != nil {
		page = &application.Page{
			Number: int(pg.GetPageNumber()),
			Size:   int(pg.GetPageSize()),
		}
	}
	deposits, err := o.operatorSvc.ListDeposits(
		ctx, int(req.GetAccountIndex()), page,
	)

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
	ctx context.Context,
	req *pb.ListWithdrawalsRequest,
) (*pb.ListWithdrawalsReply, error) {
	var page *application.Page
	if pg := req.GetPage(); pg != nil {
		page = &application.Page{
			Number: int(pg.GetPageNumber()),
			Size:   int(pg.GetPageSize()),
		}
	}

	withdrawals, err := o.operatorSvc.ListWithdrawals(
		ctx,
		int(req.GetAccountIndex()),
		page,
	)

	withdrawalsProto := make([]*pb.Withdrawal, 0, len(withdrawals))
	for _, v := range withdrawals {
		withdrawalsProto = append(withdrawalsProto, &pb.Withdrawal{
			TxId:            v.TxID,
			BaseAmount:      v.BaseAmount,
			QuoteAmount:     v.QuoteAmount,
			MillisatPerByte: v.MillisatPerByte,
			Address:         v.Address,
		})
	}

	return &pb.ListWithdrawalsReply{
		AccountIndex: req.GetAccountIndex(),
		Withdrawals:  withdrawalsProto,
	}, err
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
	addresses := make([]string, 0, len(addressesAndKeys))
	for _, a := range addressesAndKeys {
		addresses = append(addresses, a.Address)
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

func (o operatorHandler) updateMarketPercentageFee(
	ctx context.Context,
	req *pb.UpdateMarketPercentageFeeRequest,
) (*pb.UpdateMarketFeeReply, error) {
	market := req.GetMarket()
	if err := validateMarket(market); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	fee := req.GetBasisPoint()

	mwf := application.MarketWithFee{
		Market: application.Market{
			BaseAsset:  market.GetBaseAsset(),
			QuoteAsset: market.GetQuoteAsset(),
		},
		Fee: application.Fee{
			BasisPoint: fee,
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
	ctx context.Context,
	req *pb.UpdateMarketFixedFeeRequest,
) (*pb.UpdateMarketFeeReply, error) {
	market := req.GetMarket()
	if err := validateMarket(market); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	fee := req.GetFixed()
	if err := validateFixedFee(fee); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	mwf := application.MarketWithFee{
		Market: application.Market{
			BaseAsset:  market.GetBaseAsset(),
			QuoteAsset: market.GetQuoteAsset(),
		},
		Fee: application.Fee{
			FixedBaseFee:  fee.GetBaseFee(),
			FixedQuoteFee: fee.GetQuoteFee(),
		},
	}
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

func (o operatorHandler) listTrades(
	ctx context.Context,
	req *pb.ListTradesRequest,
) (*pb.ListTradesReply, error) {
	var page *application.Page
	if pg := req.GetPage(); pg != nil {
		page = &application.Page{
			Number: int(pg.PageNumber),
			Size:   int(pg.PageSize),
		}
	}

	var tradeInfo []application.TradeInfo
	var err error
	if mkt := req.GetMarket(); mkt == nil {
		tradeInfo, err = o.operatorSvc.ListTrades(ctx, page)
	} else {
		market := application.Market{
			BaseAsset:  mkt.GetBaseAsset(),
			QuoteAsset: mkt.GetQuoteAsset(),
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

func (o operatorHandler) withdrawMarket(
	ctx context.Context,
	req *pb.WithdrawMarketRequest,
) (*pb.WithdrawMarketReply, error) {
	var baseAsset, quoteAsset string
	if mkt := req.GetMarket(); mkt != nil {
		baseAsset = mkt.GetBaseAsset()
		quoteAsset = mkt.GetQuoteAsset()
	}
	var baseAmount, quoteAmount uint64
	if bal := req.GetBalanceToWithdraw(); bal != nil {
		baseAmount = bal.GetBaseAmount()
		quoteAmount = bal.GetQuoteAmount()
	}

	args := application.WithdrawMarketReq{
		Market: application.Market{
			BaseAsset:  baseAsset,
			QuoteAsset: quoteAsset,
		},
		BalanceToWithdraw: application.Balance{
			BaseAmount:  baseAmount,
			QuoteAmount: quoteAmount,
		},
		MillisatPerByte: req.GetMillisatPerByte(),
		Address:         req.GetAddress(),
		Push:            true,
	}
	if err := args.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	rawTx, txid, err := o.operatorSvc.WithdrawMarketFunds(ctx, args)
	if err != nil {
		return nil, err
	}

	return &pb.WithdrawMarketReply{RawTx: rawTx, Txid: txid}, nil
}

func (o operatorHandler) withdrawFee(
	ctx context.Context, req *pb.WithdrawFeeRequest,
) (*pb.WithdrawFeeReply, error) {
	args := application.WithdrawFeeReq{
		Amount:          req.GetAmount(),
		Address:         req.GetAddress(),
		MillisatPerByte: req.GetMillisatsPerByte(),
		Push:            true,
	}
	if err := args.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	rawTx, txid, err := o.operatorSvc.WithdrawFeeFunds(ctx, args)
	if err != nil {
		return nil, err
	}

	return &pb.WithdrawFeeReply{RawTx: rawTx, Txid: txid}, nil
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

	return &pb.ListMarketReply{Markets: pbMarketInfos}, nil
}

func (o operatorHandler) reportMarketFee(
	ctx context.Context,
	req *pb.ReportMarketFeeRequest,
) (*pb.ReportMarketFeeReply, error) {
	market := req.GetMarket()
	if err := validateMarket(market); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var page *application.Page
	if pg := req.GetPage(); pg != nil {
		page = &application.Page{
			Number: int(pg.PageNumber),
			Size:   int(pg.PageSize),
		}
	}

	report, err := o.operatorSvc.GetCollectedMarketFee(
		ctx,
		application.Market{
			BaseAsset:  market.GetBaseAsset(),
			QuoteAsset: market.GetQuoteAsset(),
		},
		page,
	)
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

	return &pb.ReportMarketFeeReply{
		CollectedFees:              collectedFees,
		TotalCollectedFeesPerAsset: report.TotalCollectedFeesPerAsset,
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
	ctx context.Context,
	req *pb.ListWebhooksRequest,
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

func validateFixedFee(fee *pbtypes.Fixed) error {
	if fee == nil {
		return errors.New("fixed fee is null")
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
