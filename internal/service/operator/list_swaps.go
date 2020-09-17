package operatorservice

import (
	"context"

	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/tdex-network/tdex-daemon/internal/domain/trade"
	"github.com/tdex-network/tdex-daemon/internal/domain/market"
)

// ListSwaps returns the list of all swaps processed by the daemon that are in
// the given status
func (s *Service) ListSwaps(ctx context.Context, req *pb.ListSwapsRequest) (*pb.ListSwapsReply, error) {
	trades, err := s.tradeRepository.GetAllTrades(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	markets, err := s.getMarketsForTrades(ctx, trades)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	swaps := tradesToSwapInfo(markets, trades)
	return &pb.ListSwapsReply{
		Swaps: swaps,
	}, nil
}

func (s *Service) getMarketsForTrades(ctx context.Context, trades []*trade.Trade) ([]*market.Market, error) {
	markets := make([]*market.Market, 0, len(trades))
	for _, trade := range trades {
		market, _, err := s.marketRepository.GetMarketByAsset(ctx, trade.MarketQuoteAsset())
		if err != nil {
			return nil, err
		}
		markets = append(markets, market)
	}
	return markets, nil
}

func tradesToSwapInfo(markets []*market.Market, trades []*trade.Trade) ([]*pb.SwapInfo) {
	info := make([]*pb.SwapInfo, 0, len(trades))
	for i, trade := range trades {
		requestMsg := trade.SwapRequestMessage()
		fee := &pbtypes.Fee{
			Asset: markets[i].FeeAsset(),
			BasisPoint: markets[i].Fee(),
		}
		i := &pb.SwapInfo{
			Status: trade.Status().Code(),
			AmountP: requestMsg.GetAmountP(),
			AssetP: requestMsg.GetAssetP(),
			AmountR: requestMsg.GetAmountR(),
			AssetR: requestMsg.GetAssetR(),
			MarketFee: fee,
			RequestTimeUnix: trade.SwapRequestTime(),
			AcceptTimeUnix: trade.SwapAcceptTime(),
			CompleteTimeUnix: trade.SwapCompleteTime(),
			ExpiryTimeUnix: trade.SwapExpiryTime(),
		}
		info = append(info, i)
	}
	return info
}
