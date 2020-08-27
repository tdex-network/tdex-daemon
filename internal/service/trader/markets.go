package tradeservice

import (
	"context"

	pbtrade "github.com/tdex-network/tdex-protobuf/generated/go/trade"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Markets is the domain controller for the Markets RPC
func (s *Service) Markets(ctx context.Context, req *pbtrade.MarketsRequest) (res *pbtrade.MarketsReply, err error) {
	tradableMarkets, err := s.marketRepository.GetTradableMarkets(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	marketsReply := &pbtrade.MarketsReply{}
	for _, mkt := range tradableMarkets {
		marketsReply.Markets = append(marketsReply.Markets, &pbtypes.MarketWithFee{
			Market: &pbtypes.Market{
				BaseAsset:  mkt.BaseAssetHash(),
				QuoteAsset: mkt.QuoteAssetHash(),
			},
			Fee: mkt.Fee(),
		})
	}

	return marketsReply, nil
}
