package mapper

import (
	v091domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v091-domain"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	swap_parser "github.com/tdex-network/tdex-daemon/internal/infrastructure/swap-parser"
)

func (m *mapperService) FromV091TradesToV1Trades(
	trades []*v091domain.Trade,
) ([]*domain.Trade, error) {
	res := make([]*domain.Trade, 0, len(trades))
	for _, v := range trades {
		trade, err := m.fromV091TradeToV1Trade(v)
		if err != nil {
			return nil, err
		}
		res = append(res, trade)
	}

	return res, nil
}

func (m *mapperService) fromV091TradeToV1Trade(
	trade *v091domain.Trade,
) (*domain.Trade, error) {
	market, err := m.v091RepoManager.MarketRepository().GetMarketByAssets(
		trade.MarketBaseAsset, trade.MarketQuoteAsset,
	)
	if err != nil {
		return nil, err
	}

	swapParser := swap_parser.NewService()
	swapRequest := swapParser.DeserializeRequest(
		trade.SwapRequest.Message, "", 0,
	)

	tradeType := domain.TradeSell
	if swapRequest.AssetR == trade.MarketBaseAsset {
		tradeType = domain.TradeBuy
	}

	return &domain.Trade{
		Id:               trade.ID.String(),
		Type:             tradeType,
		MarketName:       market.AccountName(),
		MarketBaseAsset:  trade.MarketBaseAsset,
		MarketQuoteAsset: trade.MarketQuoteAsset,
		MarketPrice: domain.MarketPrice{
			BasePrice:  trade.MarketPrice.BasePrice.String(),
			QuotePrice: trade.MarketPrice.QuotePrice.String(),
		},
		MarketPercentageFee: domain.MarketFee{
			BaseAsset:  uint64(trade.MarketFee),
			QuoteAsset: uint64(trade.MarketFee),
		},
		MarketFixedFee: domain.MarketFee{
			BaseAsset:  uint64(trade.MarketFixedBaseFee),
			QuoteAsset: uint64(trade.MarketFixedQuoteFee),
		},
		FeeAsset:     "", // TODO
		FeeAmount:    0,  // TODO
		TraderPubkey: trade.TraderPubkey,
		Status: domain.TradeStatus{
			Code:   trade.Status.Code,
			Failed: trade.Status.Failed,
		},
		PsetBase64:     trade.PsetBase64,
		TxId:           trade.TxID,
		TxHex:          trade.TxHex,
		ExpiryTime:     int64(trade.ExpiryTime),
		SettlementTime: int64(trade.SettlementTime),
		SwapRequest: &domain.Swap{
			Id:        trade.SwapRequest.ID,
			Message:   trade.SwapRequest.Message,
			Timestamp: int64(trade.SwapRequest.Timestamp),
		},
		SwapAccept: &domain.Swap{
			Id:        trade.SwapAccept.ID,
			Message:   trade.SwapAccept.Message,
			Timestamp: int64(trade.SwapAccept.Timestamp),
		},
		SwapComplete: &domain.Swap{
			Id:        trade.SwapComplete.ID,
			Message:   trade.SwapComplete.Message,
			Timestamp: int64(trade.SwapComplete.Timestamp),
		},
		SwapFail: &domain.Swap{
			Id:        trade.SwapFail.ID,
			Message:   trade.SwapFail.Message,
			Timestamp: int64(trade.SwapFail.Timestamp),
		},
	}, nil
}
