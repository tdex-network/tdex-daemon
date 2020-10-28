package grpchandler

import (
	"context"
	"errors"

	"github.com/tdex-network/tdex-daemon/internal/core/application"
	pbswap "github.com/tdex-network/tdex-protobuf/generated/go/swap"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/trade"
	"github.com/tdex-network/tdex-protobuf/generated/go/types"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type traderHandler struct {
	pb.UnimplementedTradeServer
	traderSvc application.TradeService
}

func NewTraderHandler(traderSvc application.TradeService) pb.TradeServer {
	return &traderHandler{
		traderSvc: traderSvc,
	}
}

func (t traderHandler) Markets(
	ctx context.Context,
	req *pb.MarketsRequest,
) (*pb.MarketsReply, error) {
	markets, err := t.traderSvc.GetTradableMarkets(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	marketsWithFee := make([]*types.MarketWithFee, 0, len(markets))
	for _, v := range markets {
		m := &types.MarketWithFee{
			Market: &types.Market{
				BaseAsset:  v.BaseAsset,
				QuoteAsset: v.QuoteAsset,
			},
			Fee: &types.Fee{
				Asset:      v.FeeAsset,
				BasisPoint: v.BasisPoint,
			},
		}
		marketsWithFee = append(marketsWithFee, m)
	}

	return &pb.MarketsReply{Markets: marketsWithFee}, nil
}

func (t traderHandler) Balances(
	ctx context.Context,
	req *pb.BalancesRequest,
) (*pb.BalancesReply, error) {

	mkt := req.GetMarket()
	if err := validateMarket(mkt); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	balance, err := t.traderSvc.GetMarketBalance(
		ctx,
		application.Market{
			BaseAsset:  req.Market.BaseAsset,
			QuoteAsset: req.Market.QuoteAsset,
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	balancesWithFee := make([]*types.BalanceWithFee, 0)
	balancesWithFee = append(balancesWithFee, &types.BalanceWithFee{
		Balance: &types.Balance{
			BaseAmount:  balance.BaseAmount,
			QuoteAmount: balance.QuoteAmount,
		},
		Fee: &types.Fee{
			Asset:      balance.FeeAsset,
			BasisPoint: balance.BasisPoint,
		},
	})

	return &pb.BalancesReply{
		Balances: balancesWithFee,
	}, nil
}

func (t traderHandler) MarketPrice(
	ctx context.Context,
	req *pb.MarketPriceRequest,
) (*pb.MarketPriceReply, error) {
	market := req.GetMarket()
	if err := validateMarket(market); err != nil {
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

	price, err := t.traderSvc.GetMarketPrice(
		ctx,
		application.Market{
			BaseAsset:  market.GetBaseAsset(),
			QuoteAsset: market.GetQuoteAsset(),
		},
		int(tradeType),
		amount,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	basePrice, _ := price.BasePrice.Float64()
	quotePrice, _ := price.QuotePrice.Float64()

	return &pb.MarketPriceReply{
		Prices: []*pbtypes.PriceWithFee{
			{
				Price: &pbtypes.Price{
					BasePrice:  float32(basePrice),
					QuotePrice: float32(quotePrice),
				},
				Fee: &pbtypes.Fee{
					Asset:      price.FeeAsset,
					BasisPoint: price.BasisPoint,
				},
				Amount: price.Amount,
			},
		},
	}, nil
}

func (t traderHandler) TradePropose(
	req *pb.TradeProposeRequest,
	stream pb.Trade_TradeProposeServer,
) error {
	mkt := req.GetMarket()
	if err := validateMarket(mkt); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	tradeType := req.GetType()
	if err := validateTradeType(tradeType); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	swapRequest := req.GetSwapRequest()
	if err := validateSwapRequest(swapRequest); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	market := application.Market{
		BaseAsset:  mkt.GetBaseAsset(),
		QuoteAsset: mkt.GetQuoteAsset(),
	}
	swapAccept, swapFail, swapExpiryTime, err := t.traderSvc.TradePropose(
		stream.Context(),
		market,
		int(tradeType),
		swapRequest,
	)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	reply := &pb.TradeProposeReply{
		SwapAccept:     swapAccept,
		SwapFail:       swapFail,
		ExpiryTimeUnix: swapExpiryTime,
	}
	if err := stream.Send(reply); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

func (t traderHandler) TradeComplete(
	req *pb.TradeCompleteRequest,
	stream pb.Trade_TradeCompleteServer,
) error {
	txID, swapFail, err := t.traderSvc.TradeComplete(
		stream.Context(),
		req.GetSwapComplete(),
		req.GetSwapFail(),
	)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	reply := &pb.TradeCompleteReply{
		Txid:     txID,
		SwapFail: swapFail,
	}
	if err := stream.Send(reply); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

func validateTradeType(tType pbtypes.TradeType) error {
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
