package mapper

import (
	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	swap_parser "github.com/tdex-network/tdex-daemon/internal/infrastructure/swap-parser"
	v0domain "github.com/tdex-network/tdex-daemon/old-v0"
)

var statuses = map[int]int{
	v0domain.Empty:     domain.TradeStatusCodeUndefined,
	v0domain.Undefined: domain.TradeStatusCodeUndefined,
	v0domain.Proposal:  domain.TradeStatusCodeProposal,
	v0domain.Accepted:  domain.TradeStatusCodeAccepted,
	v0domain.Completed: domain.TradeStatusCodeCompleted,
	v0domain.Settled:   domain.TradeStatusCodeSettled,
	v0domain.Expired:   domain.TradeStatusCodeExpired,
}

func (m *mapperService) FromV0TradesToV1Trades(
	trades []*v0domain.Trade, net string,
) ([]*domain.Trade, error) {
	res := make([]*domain.Trade, 0, len(trades))
	for _, v := range trades {
		trade, err := m.fromV0TradeToV1Trade(v, net)
		if err != nil {
			return nil, err
		}
		if trade != nil {
			res = append(res, trade)
		}
	}

	return res, nil
}

func (m *mapperService) fromV0TradeToV1Trade(
	trade *v0domain.Trade, net string,
) (*domain.Trade, error) {
	baseAsset := trade.MarketBaseAsset
	if len(baseAsset) <= 0 {
		baseAsset = lbtcByNetwork[net]
	}
	market, err := m.v0RepoManager.MarketRepository().GetMarketByAssets(
		baseAsset, trade.MarketQuoteAsset,
	)
	if err != nil {
		return nil, err
	}
	if market == nil {
		return nil, nil
	}

	swapParser := swap_parser.NewService()
	swapRequest := swapParser.DeserializeRequest(
		trade.SwapRequest.Message, "", 0,
	)

	tradeType := domain.TradeSell
	if swapRequest.AssetR == baseAsset {
		tradeType = domain.TradeBuy
	}

	// In v0, the fee is always charged on the amount received by the trader,
	// by first deducting the percentage fee and then the fixed one from the
	// amount he should receive.
	// Given the formula to charge fees as:
	// amountR * (100 - percentageFee) - fixedFee
	// this is the proportion between percentages and amounts:
	// (amountR + fixedFee) : (100 - percentageFee) = x : percentageFee
	// To find the total fee amount (x + fixedFee), the formula is:
	// (amountR + fixedFee) * percentageFee / (100 - fixedFee) + fixedFee
	feeAsset := swapRequest.AssetR
	fixedFee := trade.MarketFixedBaseFee
	if feeAsset == trade.MarketQuoteAsset {
		fixedFee = trade.MarketFixedQuoteFee
	}
	oneHundred := decimal.NewFromInt(100)
	bp := decimal.NewFromInt(int64(trade.MarketFee)).Div(oneHundred)
	feeAmount := decimal.NewFromInt(int64(swapRequest.AmountR)+fixedFee).
		Mul(bp).Div(oneHundred.Sub(bp)).BigInt().Uint64() + uint64(fixedFee)

	return &domain.Trade{
		Id:               trade.ID.String(),
		Type:             tradeType,
		MarketName:       market.AccountName(),
		MarketBaseAsset:  baseAsset,
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
		FeeAsset:     feeAsset,
		FeeAmount:    feeAmount,
		TraderPubkey: trade.TraderPubkey,
		Status: domain.TradeStatus{
			Code:   statuses[trade.Status.Code],
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
