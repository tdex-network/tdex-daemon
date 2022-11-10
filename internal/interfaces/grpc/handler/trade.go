package grpchandler

import (
	"context"

	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v1"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const ErrCannotServeRequest = "cannot serve request, please retry"

type tradeHandler struct {
	tradeSvc application.TradeService
}

func NewTradeHandler(
	tradeSvc application.TradeService,
) tdexv1.TradeServiceServer {
	return newTradeHandler(tradeSvc)
}

func newTradeHandler(tradeSvc application.TradeService) *tradeHandler {
	return &tradeHandler{
		tradeSvc: tradeSvc,
	}
}

func (t tradeHandler) ListMarkets(
	ctx context.Context, _ *tdexv1.ListMarketsRequest,
) (*tdexv1.ListMarketsResponse, error) {
	return t.listMarkets(ctx)
}

func (t tradeHandler) GetMarketBalance(
	ctx context.Context, req *tdexv1.GetMarketBalanceRequest,
) (*tdexv1.GetMarketBalanceResponse, error) {
	return t.getMarketBalance(ctx, req)
}

func (t tradeHandler) PreviewTrade(
	ctx context.Context, req *tdexv1.PreviewTradeRequest,
) (*tdexv1.PreviewTradeResponse, error) {
	return t.previewTrade(ctx, req)
}

func (t tradeHandler) ProposeTrade(
	ctx context.Context, req *tdexv1.ProposeTradeRequest,
) (*tdexv1.ProposeTradeResponse, error) {
	return t.proposeTrade(ctx, req)
}

func (t tradeHandler) CompleteTrade(
	ctx context.Context, req *tdexv1.CompleteTradeRequest,
) (*tdexv1.CompleteTradeResponse, error) {
	return t.completeTrade(ctx, req)
}

func (t tradeHandler) listMarkets(
	ctx context.Context,
) (*tdexv1.ListMarketsResponse, error) {
	markets, err := t.tradeSvc.GetTradableMarkets(ctx)
	if err != nil {
		return nil, err
	}

	marketsWithFee := make([]*tdexv1.MarketWithFee, 0, len(markets))
	for _, v := range markets {
		m := &tdexv1.MarketWithFee{
			Market: market{v.GetMarket()}.toProto(),
			Fee:    marketFeeInfo{v.GetFee()}.toProto(),
		}
		marketsWithFee = append(marketsWithFee, m)
	}

	return &tdexv1.ListMarketsResponse{Markets: marketsWithFee}, nil
}

func (t tradeHandler) getMarketBalance(
	ctx context.Context, req *tdexv1.GetMarketBalanceRequest,
) (*tdexv1.GetMarketBalanceResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	info, err := t.tradeSvc.GetMarketBalance(ctx, market)
	if err != nil {
		return nil, err
	}

	var baseBalance, quoteBalance uint64
	if balance := info.GetBalance(); len(balance) > 0 {
		baseBalance = balance[market.GetBaseAsset()].GetConfirmedBalance()
		quoteBalance = balance[market.GetQuoteAsset()].GetConfirmedBalance()
	}
	return &tdexv1.GetMarketBalanceResponse{
		Balance: &tdexv1.BalanceWithFee{
			Balance: &tdexv1.Balance{
				BaseAmount:  baseBalance,
				QuoteAmount: quoteBalance,
			},
			Fee: marketFeeInfo{info.GetFee()}.toProto(),
		},
	}, nil
}

func (t tradeHandler) previewTrade(
	ctx context.Context, req *tdexv1.PreviewTradeRequest,
) (*tdexv1.PreviewTradeResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	tradeType, err := parseTradeType(req.GetType())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	amount, err := parseTradeAmount(req.GetAmount())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	asset, err := parseTradeAsset(req.GetAsset())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	preview, err := t.tradeSvc.TradePreview(
		ctx, market, tradeType, amount, asset,
	)
	if err != nil {
		return nil, err
	}

	var baseBalance, quoteBalance uint64
	if info := preview.GetMarketBalance(); len(info) > 0 {
		baseBalance = info[market.GetBaseAsset()].GetConfirmedBalance()
		quoteBalance = info[market.GetQuoteAsset()].GetConfirmedBalance()
	}

	return &tdexv1.PreviewTradeResponse{
		Previews: []*tdexv1.Preview{
			{
				Price:  marketPriceInfo{preview.GetMarketPrice()}.toProto(),
				Fee:    marketFeeInfo{preview.GetMarketFee()}.toProto(),
				Amount: preview.GetAmount(),
				Asset:  preview.GetAsset(),
				Balance: &tdexv1.Balance{
					BaseAmount:  baseBalance,
					QuoteAmount: quoteBalance,
				},
			},
		},
	}, nil
}

func (t tradeHandler) proposeTrade(
	ctx context.Context, req *tdexv1.ProposeTradeRequest,
) (*tdexv1.ProposeTradeResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	tradeType, err := parseTradeType(req.GetType())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	swapRequest, err := parseSwapRequest(req.GetSwapRequest())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	accept, fail, swapExpiryTime, err := t.tradeSvc.TradePropose(
		ctx, market, tradeType, swapRequest,
	)
	if err != nil {
		return nil, err
	}

	return &tdexv1.ProposeTradeResponse{
		SwapAccept:     swapAcceptInfo{accept}.toProto(),
		SwapFail:       swapFailInfo{fail}.toProto(),
		ExpiryTimeUnix: uint64(swapExpiryTime),
	}, nil
}

func (t tradeHandler) completeTrade(
	ctx context.Context, req *tdexv1.CompleteTradeRequest,
) (*tdexv1.CompleteTradeResponse, error) {
	var swapComplete ports.SwapComplete
	if s := req.SwapComplete; s != nil {
		swapComplete = s
	}
	var swapFail ports.SwapFail
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

	return &tdexv1.CompleteTradeResponse{
		Txid:     txID,
		SwapFail: swapFailStub,
	}, nil
}
