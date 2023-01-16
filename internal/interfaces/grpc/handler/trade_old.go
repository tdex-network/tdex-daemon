package grpchandler

import (
	"context"
	"errors"

	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	pbswap "github.com/tdex-network/tdex-protobuf/generated/go/swap"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/trade"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type tradeOldHandler struct {
	pb.UnimplementedTradeServer
	tradeSvc application.TradeService
}

func NewTradeOldHandler(
	tradeSvc application.TradeService,
) pb.TradeServer {
	return newTradeOldHandler(tradeSvc)
}

func newTradeOldHandler(
	tradeSvc application.TradeService,
) *tradeOldHandler {
	return &tradeOldHandler{
		tradeSvc: tradeSvc,
	}
}

func (t tradeOldHandler) Markets(
	ctx context.Context,
	req *pb.MarketsRequest,
) (*pb.MarketsReply, error) {
	return t.markets(ctx, req)
}

func (t tradeOldHandler) Balances(
	ctx context.Context,
	req *pb.BalancesRequest,
) (*pb.BalancesReply, error) {
	return t.balances(ctx, req)
}

func (t tradeOldHandler) MarketPrice(
	ctx context.Context,
	req *pb.MarketPriceRequest,
) (*pb.MarketPriceReply, error) {
	return t.marketPrice(ctx, req)
}

func (t tradeOldHandler) TradePropose(
	req *pb.TradeProposeRequest,
	stream pb.Trade_TradeProposeServer,
) error {
	return t.tradePropose(stream, req)
}

func (t tradeOldHandler) ProposeTrade(
	ctx context.Context, req *pb.ProposeTradeRequest,
) (*pb.ProposeTradeReply, error) {
	return t.proposeTrade(ctx, req)
}

func (t tradeOldHandler) TradeComplete(
	req *pb.TradeCompleteRequest, stream pb.Trade_TradeCompleteServer,
) error {
	return t.tradeComplete(stream, req)
}

func (t tradeOldHandler) CompleteTrade(
	ctx context.Context, req *pb.CompleteTradeRequest,
) (*pb.CompleteTradeReply, error) {
	return t.completeTrade(ctx, req)
}

func (t tradeOldHandler) markets(
	ctx context.Context,
	req *pb.MarketsRequest,
) (*pb.MarketsReply, error) {
	markets, err := t.tradeSvc.GetTradableMarkets(ctx)
	if err != nil {
		return nil, err
	}

	marketsWithFee := make([]*pbtypes.MarketWithFee, 0, len(markets))
	for _, v := range markets {
		m := &pbtypes.MarketWithFee{
			Market: &pbtypes.Market{
				BaseAsset:  v.BaseAsset,
				QuoteAsset: v.QuoteAsset,
			},
			Fee: &pbtypes.Fee{
				BasisPoint: v.BasisPoint,
				Fixed: &pbtypes.Fixed{
					BaseFee:  v.FixedBaseFee,
					QuoteFee: v.FixedQuoteFee,
				},
			},
		}
		marketsWithFee = append(marketsWithFee, m)
	}

	return &pb.MarketsReply{Markets: marketsWithFee}, nil
}

func (t tradeOldHandler) balances(
	ctx context.Context,
	req *pb.BalancesRequest,
) (*pb.BalancesReply, error) {
	market, err := parseMarketOld(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	balance, err := t.tradeSvc.GetMarketBalance(ctx, market)
	if err != nil {
		return nil, err
	}

	balancesWithFee := make([]*pbtypes.BalanceWithFee, 0)
	balancesWithFee = append(balancesWithFee, &pbtypes.BalanceWithFee{
		Balance: &pbtypes.Balance{
			BaseAmount:  balance.Balance.BaseAmount,
			QuoteAmount: balance.Balance.QuoteAmount,
		},
		Fee: &pbtypes.Fee{
			BasisPoint: balance.Fee.BasisPoint,
			Fixed: &pbtypes.Fixed{
				BaseFee:  balance.Fee.FixedBaseFee,
				QuoteFee: balance.Fee.FixedQuoteFee,
			},
		},
	})

	return &pb.BalancesReply{
		Balances: balancesWithFee,
	}, nil
}

func (t tradeOldHandler) marketPrice(
	ctx context.Context,
	req *pb.MarketPriceRequest,
) (*pb.MarketPriceReply, error) {
	market, err := parseMarketOld(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	tradeType := req.GetType()
	if err := validateTradeTypeOld(tradeType); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	amount := req.GetAmount()
	if err := validateAmount(amount); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	asset := req.GetAsset()
	if err := validateAsset(asset); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	preview, err := t.tradeSvc.GetTradePreview(
		ctx, market, int(tradeType), amount, asset,
	)
	if err != nil {
		return nil, err
	}

	basePrice, _ := preview.Price.BasePrice.Float64()
	quotePrice, _ := preview.Price.QuotePrice.Float64()

	return &pb.MarketPriceReply{
		Prices: []*pbtypes.PriceWithFee{
			{
				Price: &pbtypes.Price{
					BasePrice:  basePrice,
					QuotePrice: quotePrice,
				},
				Fee: &pbtypes.Fee{
					BasisPoint: preview.Fee.BasisPoint,
					Fixed: &pbtypes.Fixed{
						BaseFee:  preview.Fee.FixedBaseFee,
						QuoteFee: preview.Fee.FixedQuoteFee,
					},
				},
				Amount: preview.Amount,
				Asset:  preview.Asset,
				Balance: &pbtypes.Balance{
					BaseAmount:  preview.Balance.BaseAmount,
					QuoteAmount: preview.Balance.QuoteAmount,
				},
			},
		},
	}, nil
}

func (t tradeOldHandler) proposeTrade(
	ctx context.Context, req *pb.ProposeTradeRequest,
) (*pb.ProposeTradeReply, error) {
	market, err := parseMarketOld(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	tradeType := req.GetType()
	if err := validateTradeTypeOld(tradeType); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	swapRequest := req.GetSwapRequest()
	if err := validateSwapRequestOld(swapRequest); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	accept, fail, swapExpiryTime, err := t.tradeSvc.TradePropose(
		ctx, market, int(tradeType), swapRequest,
	)
	if err != nil {
		return nil, err
	}

	var swapAccept *pbswap.SwapAccept
	var swapFail *pbswap.SwapFail

	if accept != nil {
		swapAccept = &pbswap.SwapAccept{
			Id:                accept.GetId(),
			RequestId:         accept.GetRequestId(),
			Transaction:       accept.GetTransaction(),
			InputBlindingKey:  accept.GetInputBlindingKey(),
			OutputBlindingKey: accept.GetOutputBlindingKey(),
		}
	}
	if fail != nil {
		swapFail = &pbswap.SwapFail{
			Id:             fail.GetId(),
			MessageId:      fail.GetMessageId(),
			FailureCode:    fail.GetFailureCode(),
			FailureMessage: fail.GetFailureMessage(),
		}
	}

	return &pb.ProposeTradeReply{
		SwapAccept:     swapAccept,
		SwapFail:       swapFail,
		ExpiryTimeUnix: swapExpiryTime,
	}, nil
}

func (t tradeOldHandler) completeTrade(
	ctx context.Context, req *pb.CompleteTradeRequest,
) (*pb.CompleteTradeReply, error) {
	var swapComplete domain.SwapComplete
	if s := req.SwapComplete; s != nil {
		swapComplete = s
	}
	var swapFail domain.SwapFail
	if s := req.SwapFail; s != nil {
		swapFail = s
	}
	txID, fail, err := t.tradeSvc.TradeComplete(
		ctx, swapComplete, swapFail,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var swapFailStub *pbswap.SwapFail
	if fail != nil {
		swapFailStub = &pbswap.SwapFail{
			Id:             fail.GetId(),
			MessageId:      fail.GetMessageId(),
			FailureCode:    fail.GetFailureCode(),
			FailureMessage: fail.GetFailureMessage(),
		}
	}

	return &pb.CompleteTradeReply{
		Txid:     txID,
		SwapFail: swapFailStub,
	}, nil
}

func (t tradeOldHandler) tradePropose(
	stream pb.Trade_TradeProposeServer, req *pb.TradeProposeRequest,
) error {
	rr := &pb.ProposeTradeRequest{
		Market:      req.GetMarket(),
		Type:        req.GetType(),
		SwapRequest: req.GetSwapRequest(),
	}
	reply, err := t.proposeTrade(stream.Context(), rr)
	if err != nil {
		return err
	}
	resp := &pb.TradeProposeReply{
		SwapAccept:     reply.GetSwapAccept(),
		SwapFail:       reply.GetSwapFail(),
		ExpiryTimeUnix: reply.GetExpiryTimeUnix(),
	}

	if err := stream.Send(resp); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

func (t tradeOldHandler) tradeComplete(
	stream pb.Trade_TradeCompleteServer, req *pb.TradeCompleteRequest,
) error {
	rr := &pb.CompleteTradeRequest{
		SwapComplete: req.GetSwapComplete(),
		SwapFail:     req.GetSwapFail(),
	}
	reply, err := t.completeTrade(stream.Context(), rr)
	if err != nil {
		return err
	}

	resp := &pb.TradeCompleteReply{
		Txid:     reply.GetTxid(),
		SwapFail: reply.GetSwapFail(),
	}
	if err := stream.Send(resp); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

func parseMarketOld(mkt *pbtypes.Market) (market application.Market, err error) {
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

func validateTradeTypeOld(tType pb.TradeType) error {
	if int(tType) < application.TradeBuy || int(tType) > application.TradeSell {
		return errors.New("trade type is unknown")
	}
	return nil
}

func validateSwapRequestOld(swapRequest *pbswap.SwapRequest) error {
	if swapRequest == nil {
		return errors.New("swap request is null")
	}
	if swapRequest.GetAmountP() <= 0 ||
		len(swapRequest.GetAssetP()) <= 0 ||
		swapRequest.GetAmountR() <= 0 ||
		len(swapRequest.GetAssetR()) <= 0 ||
		len(swapRequest.GetTransaction()) <= 0 ||
		len(swapRequest.GetInputBlindingKey()) <= 0 ||
		len(swapRequest.GetOutputBlindingKey()) <= 0 {
		return errors.New("swap request is malformed")
	}
	return nil
}
