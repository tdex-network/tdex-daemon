package grpchandler

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/core/application"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/trade"
	"github.com/tdex-network/tdex-protobuf/generated/go/types"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type traderHandler struct {
	pb.UnimplementedTradeServer
	traderSvc application.TraderService
}

func NewTraderHandler(traderSvc application.TraderService) pb.TradeServer {
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
	return &pb.BalancesReply{}, nil
}

func (t traderHandler) MarketPrice(
	ctx context.Context,
	req *pb.MarketPriceRequest,
) (*pb.MarketPriceReply, error) {
	price, err := t.traderSvc.GetMarketPrice(
		ctx,
		application.Market{
			BaseAsset:  req.Market.BaseAsset,
			QuoteAsset: req.Market.QuoteAsset,
		},
		int(req.Type),
		req.Amount,
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
	swapRequest := req.GetSwapRequest()
	tradeType := req.GetType()
	market := application.Market{
		BaseAsset:  req.GetMarket().GetBaseAsset(),
		QuoteAsset: req.GetMarket().GetQuoteAsset(),
	}
	swapAccept, swapFail, swapExpiryTime, err :=
		t.traderSvc.TradePropose(context.Background(), market, int(tradeType), swapRequest)
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
		context.Background(),
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
