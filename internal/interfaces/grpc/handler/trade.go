package grpchandler

import (
	"context"
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	pbswap "github.com/tdex-network/tdex-protobuf/generated/go/swap"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/trade"
	pbtrade "github.com/tdex-network/tdex-protobuf/generated/go/trade"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const ErrCannotServeRequest = "cannot serve request, please retry"

type traderHandler struct {
	pb.UnimplementedTradeServer
	traderSvc application.TradeService
	dbManager ports.DbManager
}

func NewTraderHandler(
	traderSvc application.TradeService,
	dbManager ports.DbManager,
) pb.TradeServer {
	return newTraderHandler(traderSvc, dbManager)
}

func newTraderHandler(
	traderSvc application.TradeService,
	dbManager ports.DbManager,
) *traderHandler {
	return &traderHandler{
		traderSvc: traderSvc,
		dbManager: dbManager,
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
	req *pb.TradeProposeRequest,
	stream pb.Trade_TradeProposeServer,
) error {
	return t.tradePropose(req, stream)
}

func (t traderHandler) TradeComplete(
	req *pb.TradeCompleteRequest,
	stream pb.Trade_TradeCompleteServer,
) error {
	return t.tradeComplete(req, stream)
}

func (t traderHandler) markets(
	reqCtx context.Context,
	req *pb.MarketsRequest,
) (*pb.MarketsReply, error) {
	res, err := t.dbManager.RunTransaction(
		reqCtx,
		readOnlyTx,
		func(ctx context.Context) (interface{}, error) {
			markets, err := t.traderSvc.GetTradableMarkets(ctx)
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
						Asset:      v.FeeAsset,
						BasisPoint: v.BasisPoint,
					},
				}
				marketsWithFee = append(marketsWithFee, m)
			}

			return &pb.MarketsReply{Markets: marketsWithFee}, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res.(*pb.MarketsReply), nil
}

func (t traderHandler) balances(
	reqCtx context.Context,
	req *pb.BalancesRequest,
) (*pb.BalancesReply, error) {
	mkt := req.GetMarket()
	if err := validateMarket(mkt); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	res, err := t.dbManager.RunTransaction(
		reqCtx,
		readOnlyTx,
		func(ctx context.Context) (interface{}, error) {
			balance, err := t.traderSvc.GetMarketBalance(
				ctx,
				application.Market{
					BaseAsset:  req.Market.BaseAsset,
					QuoteAsset: req.Market.QuoteAsset,
				},
			)
			if err != nil {
				return nil, err
			}

			balancesWithFee := make([]*pbtypes.BalanceWithFee, 0)
			balancesWithFee = append(balancesWithFee, &pbtypes.BalanceWithFee{
				Balance: &pbtypes.Balance{
					BaseAmount:  balance.BaseAmount,
					QuoteAmount: balance.QuoteAmount,
				},
				Fee: &pbtypes.Fee{
					Asset:      balance.FeeAsset,
					BasisPoint: balance.BasisPoint,
				},
			})

			return &pb.BalancesReply{
				Balances: balancesWithFee,
			}, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res.(*pb.BalancesReply), nil
}

func (t traderHandler) marketPrice(
	reqCtx context.Context,
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

	res, err := t.dbManager.RunTransaction(
		reqCtx,
		readOnlyTx,
		func(ctx context.Context) (interface{}, error) {
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
				return nil, err
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
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res.(*pb.MarketPriceReply), nil
}

func (t traderHandler) tradePropose(
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

	var market = application.Market{
		BaseAsset:  mkt.GetBaseAsset(),
		QuoteAsset: mkt.GetQuoteAsset(),
	}

	res, err := t.dbManager.RunTransaction(
		stream.Context(),
		!readOnlyTx,
		func(ctx context.Context) (interface{}, error) {
			swapAccept, swapFail, swapExpiryTime, err := t.traderSvc.TradePropose(
				ctx,
				market,
				int(tradeType),
				swapRequest,
			)
			if err != nil {
				return nil, err
			}

			return &pb.TradeProposeReply{
				SwapAccept:     swapAccept,
				SwapFail:       swapFail,
				ExpiryTimeUnix: swapExpiryTime,
			}, nil
		},
	)
	if err != nil {
		log.Debug("trying to process trade proposal: ", err)
		return status.Error(codes.Internal, ErrCannotServeRequest)
	}

	if err := stream.Send(res.(*pb.TradeProposeReply)); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

func (t traderHandler) tradeComplete(
	req *pb.TradeCompleteRequest,
	stream pb.Trade_TradeCompleteServer,
) error {
	res, err := t.dbManager.RunTransaction(
		stream.Context(),
		!readOnlyTx,
		func(ctx context.Context) (interface{}, error) {
			txID, swapFail, err := t.traderSvc.TradeComplete(
				ctx,
				req.GetSwapComplete(),
				req.GetSwapFail(),
			)
			if err != nil {
				return nil, err
			}

			return &pb.TradeCompleteReply{
				Txid:     txID,
				SwapFail: swapFail,
			}, nil
		},
	)
	if err != nil {
		log.Debug("trying to complete trade: ", err)
		return status.Error(codes.Internal, ErrCannotServeRequest)
	}

	if err := stream.Send(res.(*pb.TradeCompleteReply)); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func validateTradeType(tType pbtrade.TradeType) error {
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
