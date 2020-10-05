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
		return nil, err
	}

	marketsReply := new(pb.MarketsReply)
	marketsWithFee := make([]*types.MarketWithFee, 0)
	marketsReply.Markets = marketsWithFee

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

	return marketsReply, nil
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
			QuoteAsset: req.Market.BaseAsset,
		},
	)
	if err != nil {
		return nil, status.Error(
			codes.Internal,
			err.Error(),
		)
	}

	return &pb.MarketPriceReply{
		Prices: []*pbtypes.PriceWithFee{
			{
				Price: &pbtypes.Price{
					BasePrice:  price.BasePrice,
					QuotePrice: price.QuotePrice,
				},
				Fee: &pbtypes.Fee{
					Asset:      price.FeeAsset,
					BasisPoint: price.BasisPoint,
				},
			},
		},
	}, nil
}

func (t traderHandler) TradePropose(
	req *pb.TradeProposeRequest,
	server pb.Trade_TradeProposeServer,
) error {
	return t.traderSvc.TradePropose(req, server)
}

func (t traderHandler) TradeComplete(
	req *pb.TradeCompleteRequest,
	server pb.Trade_TradeCompleteServer,
) error {
	return t.traderSvc.TradeComplete(req, server)
}
