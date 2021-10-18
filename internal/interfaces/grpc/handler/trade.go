package grpchandler

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	pbswap "github.com/tdex-network/tdex-protobuf/generated/go/swap"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/trade"
	"github.com/tdex-network/tdex-protobuf/generated/go/types"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const ErrCannotServeRequest = "cannot serve request, please retry"

type traderHandler struct {
	pb.UnimplementedTradeServer
	traderSvc application.TradeService
}

func NewTraderHandler(
	traderSvc application.TradeService,
) pb.TradeServer {
	return newTraderHandler(traderSvc)
}

func newTraderHandler(
	traderSvc application.TradeService,
) *traderHandler {
	return &traderHandler{
		traderSvc: traderSvc,
	}
}

func (t traderHandler) Markets(
	ctx context.Context,
	req *pb.MarketsRequest,
) (*pb.MarketsReply, error) {
	return t.markets(ctx, req)
}

func (t traderHandler) Balances(
	ctx context.Context,
	req *pb.BalancesRequest,
) (*pb.BalancesReply, error) {
	return t.balances(ctx, req)
}

func (t traderHandler) MarketPrice(
	ctx context.Context,
	req *pb.MarketPriceRequest,
) (*pb.MarketPriceReply, error) {
	return t.marketPrice(ctx, req)
}

func (t traderHandler) TradePropose(
	ctx context.Context,
	req *pb.TradeProposeRequest,
) (*pb.TradeProposeReply, error) {
	return t.tradePropose(ctx, req)
}

func (t traderHandler) TradeComplete(
	ctx context.Context,
	req *pb.TradeCompleteRequest,
) (*pb.TradeCompleteReply, error) {
	return t.tradeComplete(ctx, req)
}

func (t traderHandler) markets(
	ctx context.Context,
	req *pb.MarketsRequest,
) (*pb.MarketsReply, error) {
	markets, err := t.traderSvc.GetTradableMarkets(ctx)
	if err != nil {
		return nil, err
	}

	marketsWithFee := make([]*types.MarketWithFee, 0, len(markets))
	for _, v := range markets {
		m := &types.MarketWithFee{
			Market: &types.Market{
				BaseAsset:  v.BaseAsset,
				QuoteAsset: v.QuoteAsset,
			},
			Fee: &types.Fee{
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

func (t traderHandler) balances(
	ctx context.Context,
	req *pb.BalancesRequest,
) (*pb.BalancesReply, error) {
	market, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	balance, err := t.traderSvc.GetMarketBalance(ctx, market)
	if err != nil {
		return nil, err
	}

	balancesWithFee := make([]*types.BalanceWithFee, 0)
	balancesWithFee = append(balancesWithFee, &types.BalanceWithFee{
		Balance: &types.Balance{
			BaseAmount:  balance.Balance.BaseAmount,
			QuoteAmount: balance.Balance.QuoteAmount,
		},
		Fee: &types.Fee{
			BasisPoint: balance.Fee.BasisPoint,
			Fixed: &types.Fixed{
				BaseFee:  balance.Fee.FixedBaseFee,
				QuoteFee: balance.Fee.FixedQuoteFee,
			},
		},
	})

	return &pb.BalancesReply{
		Balances: balancesWithFee,
	}, nil
}

func (t traderHandler) marketPrice(
	ctx context.Context,
	req *pb.MarketPriceRequest,
) (*pb.MarketPriceReply, error) {
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

	preview, err := t.traderSvc.GetMarketPrice(
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
					BasePrice:  float32(basePrice),
					QuotePrice: float32(quotePrice),
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

func (t traderHandler) tradePropose(
	ctx context.Context,
	req *pb.TradeProposeRequest,
) (*pb.TradeProposeReply, error) {
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

	accept, fail, swapExpiryTime, err := t.traderSvc.TradePropose(
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

	return &pb.TradeProposeReply{
		SwapAccept:     swapAccept,
		SwapFail:       swapFail,
		ExpiryTimeUnix: swapExpiryTime,
	}, nil
}

func (t traderHandler) tradeComplete(
	ctx context.Context,
	req *pb.TradeCompleteRequest,
) (*pb.TradeCompleteReply, error) {
	var swapComplete domain.SwapComplete
	if s := req.SwapComplete; s != nil {
		swapComplete = s
	}
	var swapFail domain.SwapFail
	if s := req.SwapFail; s != nil {
		swapFail = s
	}
	txID, fail, err := t.traderSvc.TradeComplete(
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

	return &pb.TradeCompleteReply{
		Txid:     txID,
		SwapFail: swapFailStub,
	}, nil
}

func validateTradeType(tType pb.TradeType) error {
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

func validateSwapRequest(swapRequest *pbswap.SwapRequest) error {
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
