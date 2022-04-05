package grpchandler

import (
	"context"
	"encoding/hex"
	"errors"

	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/go/tdex/v1"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const ErrCannotServeRequest = "cannot serve request, please retry"

type tradeHandler struct {
	tdexv1.UnimplementedTradeServer
	tradeSvc application.TradeService
}

func NewTradeHandler(
	tradeSvc application.TradeService,
) tdexv1.TradeServer {
	return newTradeHandler(tradeSvc)
}

func newTradeHandler(
	tradeSvc application.TradeService,
) *tradeHandler {
	return &tradeHandler{
		tradeSvc: tradeSvc,
	}
}

func (t tradeHandler) Markets(
	ctx context.Context,
	req *tdexv1.MarketsRequest,
) (*tdexv1.MarketsReply, error) {
	return t.markets(ctx, req)
}

func (t tradeHandler) Balances(
	ctx context.Context,
	req *tdexv1.BalancesRequest,
) (*tdexv1.BalancesReply, error) {
	return t.balances(ctx, req)
}

func (t tradeHandler) MarketPrice(
	ctx context.Context,
	req *tdexv1.MarketPriceRequest,
) (*tdexv1.MarketPriceReply, error) {
	return t.marketPrice(ctx, req)
}

func (t tradeHandler) TradePropose(
	req *tdexv1.TradeProposeRequest,
	stream tdexv1.Trade_TradeProposeServer,
) error {
	return t.tradePropose(stream, req)
}

func (t tradeHandler) ProposeTrade(
	ctx context.Context, req *tdexv1.ProposeTradeRequest,
) (*tdexv1.ProposeTradeReply, error) {
	return t.proposeTrade(ctx, req)
}

func (t tradeHandler) TradeComplete(
	req *tdexv1.TradeCompleteRequest, stream tdexv1.Trade_TradeCompleteServer,
) error {
	return t.tradeComplete(stream, req)
}

func (t tradeHandler) CompleteTrade(
	ctx context.Context, req *tdexv1.CompleteTradeRequest,
) (*tdexv1.CompleteTradeReply, error) {
	return t.completeTrade(ctx, req)
}

func (t tradeHandler) markets(
	ctx context.Context,
	req *tdexv1.MarketsRequest,
) (*tdexv1.MarketsReply, error) {
	markets, err := t.tradeSvc.GetTradableMarkets(ctx)
	if err != nil {
		return nil, err
	}

	marketsWithFee := make([]*tdexv1.MarketWithFee, 0, len(markets))
	for _, v := range markets {
		m := &tdexv1.MarketWithFee{
			Market: &tdexv1.Market{
				BaseAsset:  v.BaseAsset,
				QuoteAsset: v.QuoteAsset,
			},
			Fee: &tdexv1.Fee{
				BasisPoint: v.BasisPoint,
				Fixed: &tdexv1.Fixed{
					BaseFee:  v.FixedBaseFee,
					QuoteFee: v.FixedQuoteFee,
				},
			},
		}
		marketsWithFee = append(marketsWithFee, m)
	}

	return &tdexv1.MarketsReply{Markets: marketsWithFee}, nil
}

func (t tradeHandler) balances(
	ctx context.Context,
	req *tdexv1.BalancesRequest,
) (*tdexv1.BalancesReply, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	balance, err := t.tradeSvc.GetMarketBalance(ctx, market)
	if err != nil {
		return nil, err
	}

	balancesWithFee := make([]*tdexv1.BalanceWithFee, 0)
	balancesWithFee = append(balancesWithFee, &tdexv1.BalanceWithFee{
		Balance: &tdexv1.Balance{
			BaseAmount:  balance.Balance.BaseAmount,
			QuoteAmount: balance.Balance.QuoteAmount,
		},
		Fee: &tdexv1.Fee{
			BasisPoint: balance.Fee.BasisPoint,
			Fixed: &tdexv1.Fixed{
				BaseFee:  balance.Fee.FixedBaseFee,
				QuoteFee: balance.Fee.FixedQuoteFee,
			},
		},
	})

	return &tdexv1.BalancesReply{
		Balances: balancesWithFee,
	}, nil
}

func (t tradeHandler) marketPrice(
	ctx context.Context,
	req *tdexv1.MarketPriceRequest,
) (*tdexv1.MarketPriceReply, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	tradeType := req.GetType()
	if err := validateTradeType(tradeType); err != nil {
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

	preview, err := t.tradeSvc.GetMarketPrice(
		ctx, market, int(tradeType), amount, asset,
	)
	if err != nil {
		return nil, err
	}

	basePrice, _ := preview.Price.BasePrice.Float64()
	quotePrice, _ := preview.Price.QuotePrice.Float64()

	return &tdexv1.MarketPriceReply{
		Prices: []*tdexv1.PriceWithFee{
			{
				Price: &tdexv1.Price{
					BasePrice:  basePrice,
					QuotePrice: quotePrice,
				},
				Fee: &tdexv1.Fee{
					BasisPoint: preview.Fee.BasisPoint,
					Fixed: &tdexv1.Fixed{
						BaseFee:  preview.Fee.FixedBaseFee,
						QuoteFee: preview.Fee.FixedQuoteFee,
					},
				},
				Amount: preview.Amount,
				Asset:  preview.Asset,
				Balance: &tdexv1.Balance{
					BaseAmount:  preview.Balance.BaseAmount,
					QuoteAmount: preview.Balance.QuoteAmount,
				},
			},
		},
	}, nil
}

func (t tradeHandler) proposeTrade(
	ctx context.Context, req *tdexv1.ProposeTradeRequest,
) (*tdexv1.ProposeTradeReply, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	tradeType := req.GetType()
	if err := validateTradeType(tradeType); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	swapRequest := req.GetSwapRequest()
	if err := validateSwapRequest(swapRequest); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	accept, fail, swapExpiryTime, err := t.tradeSvc.TradePropose(
		ctx, market, int(tradeType), swapRequest,
	)
	if err != nil {
		return nil, err
	}

	var swapAccept *tdexv1.SwapAccept
	var swapFail *tdexv1.SwapFail

	if accept != nil {
		swapAccept = &tdexv1.SwapAccept{
			Id:                accept.GetId(),
			RequestId:         accept.GetRequestId(),
			Transaction:       accept.GetTransaction(),
			InputBlindingKey:  accept.GetInputBlindingKey(),
			OutputBlindingKey: accept.GetOutputBlindingKey(),
		}
	}
	if fail != nil {
		swapFail = &tdexv1.SwapFail{
			Id:             fail.GetId(),
			MessageId:      fail.GetMessageId(),
			FailureCode:    fail.GetFailureCode(),
			FailureMessage: fail.GetFailureMessage(),
		}
	}

	return &tdexv1.ProposeTradeReply{
		SwapAccept:     swapAccept,
		SwapFail:       swapFail,
		ExpiryTimeUnix: swapExpiryTime,
	}, nil
}

func (t tradeHandler) completeTrade(
	ctx context.Context, req *tdexv1.CompleteTradeRequest,
) (*tdexv1.CompleteTradeReply, error) {
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

	var swapFailStub *tdexv1.SwapFail
	if fail != nil {
		swapFailStub = &tdexv1.SwapFail{
			Id:             fail.GetId(),
			MessageId:      fail.GetMessageId(),
			FailureCode:    fail.GetFailureCode(),
			FailureMessage: fail.GetFailureMessage(),
		}
	}

	return &tdexv1.CompleteTradeReply{
		Txid:     txID,
		SwapFail: swapFailStub,
	}, nil
}

func (t tradeHandler) tradePropose(
	stream tdexv1.Trade_TradeProposeServer, req *tdexv1.TradeProposeRequest,
) error {
	rr := &tdexv1.ProposeTradeRequest{
		Market:      req.GetMarket(),
		Type:        req.GetType(),
		SwapRequest: req.GetSwapRequest(),
	}
	reply, err := t.proposeTrade(stream.Context(), rr)
	if err != nil {
		return err
	}
	resp := &tdexv1.TradeProposeReply{
		SwapAccept:     reply.GetSwapAccept(),
		SwapFail:       reply.GetSwapFail(),
		ExpiryTimeUnix: reply.GetExpiryTimeUnix(),
	}

	if err := stream.Send(resp); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

func (t tradeHandler) tradeComplete(
	stream tdexv1.Trade_TradeCompleteServer, req *tdexv1.TradeCompleteRequest,
) error {
	rr := &tdexv1.CompleteTradeRequest{
		SwapComplete: req.GetSwapComplete(),
		SwapFail:     req.GetSwapFail(),
	}
	reply, err := t.completeTrade(stream.Context(), rr)
	if err != nil {
		return err
	}

	resp := &tdexv1.TradeCompleteReply{
		Txid:     reply.GetTxid(),
		SwapFail: reply.GetSwapFail(),
	}
	if err := stream.Send(resp); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

func validateTradeType(tType tdexv1.TradeType) error {
	if int(tType) < application.TradeBuy || int(tType) > application.TradeSell {
		return errors.New("trade type is unknown")
	}
	return nil
}

func validateAmount(amount uint64) error {
	if amount <= 0 {
		return errors.New("amount is too low")
	}
	return nil
}

func validateSwapRequest(swapRequest *tdexv1.SwapRequest) error {
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

func validateAsset(asset string) error {
	if buf, err := hex.DecodeString(asset); err != nil || len(buf) != 32 {
		return errors.New("invalid asset")
	}
	return nil
}
