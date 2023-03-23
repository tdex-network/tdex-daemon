package grpchandler

import (
	"context"
	"strings"

	tdexv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v2"
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
) tdexv2.TradeServiceServer {
	return newTradeHandler(tradeSvc)
}

func newTradeHandler(tradeSvc application.TradeService) *tradeHandler {
	return &tradeHandler{
		tradeSvc: tradeSvc,
	}
}

func (t tradeHandler) ListMarkets(
	ctx context.Context, _ *tdexv2.ListMarketsRequest,
) (*tdexv2.ListMarketsResponse, error) {
	return t.listMarkets(ctx)
}

func (t tradeHandler) GetMarketBalance(
	ctx context.Context, req *tdexv2.GetMarketBalanceRequest,
) (*tdexv2.GetMarketBalanceResponse, error) {
	return t.getMarketBalance(ctx, req)
}

func (t tradeHandler) GetMarketPrice(
	ctx context.Context, req *tdexv2.GetMarketPriceRequest,
) (*tdexv2.GetMarketPriceResponse, error) {
	return t.getMarketPrice(ctx, req)
}

func (t tradeHandler) PreviewTrade(
	ctx context.Context, req *tdexv2.PreviewTradeRequest,
) (*tdexv2.PreviewTradeResponse, error) {
	return t.previewTrade(ctx, req)
}

func (t tradeHandler) ProposeTrade(
	ctx context.Context, req *tdexv2.ProposeTradeRequest,
) (*tdexv2.ProposeTradeResponse, error) {
	return t.proposeTrade(ctx, req)
}

func (t tradeHandler) CompleteTrade(
	ctx context.Context, req *tdexv2.CompleteTradeRequest,
) (*tdexv2.CompleteTradeResponse, error) {
	return t.completeTrade(ctx, req)
}

func (t tradeHandler) listMarkets(
	ctx context.Context,
) (*tdexv2.ListMarketsResponse, error) {
	markets, err := t.tradeSvc.GetTradableMarkets(ctx)
	if err != nil {
		return nil, err
	}

	marketsWithFee := make([]*tdexv2.MarketWithFee, 0, len(markets))
	for _, v := range markets {
		m := &tdexv2.MarketWithFee{
			Market: market{v.GetMarket()}.toProto(),
			Fee: marketFeeInfo{
				v.GetPercentageFee(), v.GetFixedFee(),
			}.toProto(),
		}
		marketsWithFee = append(marketsWithFee, m)
	}

	return &tdexv2.ListMarketsResponse{Markets: marketsWithFee}, nil
}

func (t tradeHandler) getMarketBalance(
	ctx context.Context, req *tdexv2.GetMarketBalanceRequest,
) (*tdexv2.GetMarketBalanceResponse, error) {
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
	return &tdexv2.GetMarketBalanceResponse{
		Balance: &tdexv2.Balance{
			BaseAmount:  baseBalance,
			QuoteAmount: quoteBalance,
		},
		Fee: marketFeeInfo{
			info.GetPercentageFee(), info.GetFixedFee(),
		}.toProto(),
	}, nil
}

func (t tradeHandler) getMarketPrice(
	ctx context.Context, req *tdexv2.GetMarketPriceRequest,
) (*tdexv2.GetMarketPriceResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	price, minAmount, err := t.tradeSvc.GetMarketPrice(ctx, market)
	if err != nil {
		return nil, err
	}

	spotPrice, _ := price.Float64()
	return &tdexv2.GetMarketPriceResponse{
		SpotPrice:         spotPrice,
		MinTradableAmount: minAmount,
	}, nil
}

func (t tradeHandler) previewTrade(
	ctx context.Context, req *tdexv2.PreviewTradeRequest,
) (*tdexv2.PreviewTradeResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	tradeType, err := parseTradeType(req.GetType())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	amount, err := parseAmount(req.GetAmount())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	asset, err := parseAsset(req.GetAsset())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	feeAsset, err := parseAsset(req.GetFeeAsset())
	if err != nil {
		// Change 'invalid asset' message to 'invalid fee asset'.
		errMsg := strings.Replace(err.Error(), " ", " fee ", -1)
		return nil, status.Error(codes.InvalidArgument, errMsg)
	}

	preview, err := t.tradeSvc.TradePreview(
		ctx, market, tradeType, amount, asset, feeAsset,
	)
	if err != nil {
		return nil, err
	}

	return &tdexv2.PreviewTradeResponse{
		Previews: []*tdexv2.Preview{
			{
				Price: marketPriceInfo{preview.GetMarketPrice()}.toProto(),
				Fee: marketFeeInfo{
					preview.GetMarketPercentageFee(), preview.GetMarketFixedFee(),
				}.toProto(),
				Amount:    preview.GetAmount(),
				Asset:     preview.GetAsset(),
				FeeAmount: preview.GetFeeAmount(),
				FeeAsset:  preview.GetFeeAsset(),
			},
		},
	}, nil
}

func (t tradeHandler) proposeTrade(
	ctx context.Context, req *tdexv2.ProposeTradeRequest,
) (*tdexv2.ProposeTradeResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	tradeType, err := parseTradeType(req.GetType())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	swapRequest, err := parseSwapRequest(
		req.GetSwapRequest(), req.GetFeeAsset(), req.GetFeeAmount(),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	accept, fail, swapExpiryTime, err := t.tradeSvc.TradePropose(
		ctx, market, tradeType, swapRequest,
	)
	if err != nil {
		return nil, err
	}

	return &tdexv2.ProposeTradeResponse{
		SwapAccept:     swapAcceptInfo{accept}.toProto(),
		SwapFail:       swapFailInfo{fail}.toProto(),
		ExpiryTimeUnix: uint64(swapExpiryTime),
	}, nil
}

func (t tradeHandler) completeTrade(
	ctx context.Context, req *tdexv2.CompleteTradeRequest,
) (*tdexv2.CompleteTradeResponse, error) {
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

	var swapFailStub *tdexv2.SwapFail
	if fail != nil {
		swapFailStub = &tdexv2.SwapFail{
			Id:             fail.GetId(),
			MessageId:      fail.GetMessageId(),
			FailureCode:    fail.GetFailureCode(),
			FailureMessage: fail.GetFailureMessage(),
		}
	}

	return &tdexv2.CompleteTradeResponse{
		Txid:     txID,
		SwapFail: swapFailStub,
	}, nil
}
