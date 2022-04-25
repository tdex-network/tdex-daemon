package grpchandler

import (
	"context"
	"encoding/hex"
	"errors"

	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v1"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
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
	ctx context.Context, req *tdexv1.ListMarketsRequest,
) (*tdexv1.ListMarketsResponse, error) {
	return t.listMarkets(ctx, req)
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
	ctx context.Context, req *tdexv1.ListMarketsRequest,
) (*tdexv1.ListMarketsResponse, error) {
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

	return &tdexv1.ListMarketsResponse{Markets: marketsWithFee}, nil
}

func (t tradeHandler) getMarketBalance(
	ctx context.Context, req *tdexv1.GetMarketBalanceRequest,
) (*tdexv1.GetMarketBalanceResponse, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	balance, err := t.tradeSvc.GetMarketBalance(ctx, market)
	if err != nil {
		return nil, err
	}

	return &tdexv1.GetMarketBalanceResponse{
		Balance: &tdexv1.BalanceWithFee{
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

	return &tdexv1.PreviewTradeResponse{
		Previews: []*tdexv1.Preview{
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
) (*tdexv1.ProposeTradeResponse, error) {
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

	return &tdexv1.ProposeTradeResponse{
		SwapAccept:     swapAccept,
		SwapFail:       swapFail,
		ExpiryTimeUnix: swapExpiryTime,
	}, nil
}

func (t tradeHandler) completeTrade(
	ctx context.Context, req *tdexv1.CompleteTradeRequest,
) (*tdexv1.CompleteTradeResponse, error) {
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

	return &tdexv1.CompleteTradeResponse{
		Txid:     txID,
		SwapFail: swapFailStub,
	}, nil
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
