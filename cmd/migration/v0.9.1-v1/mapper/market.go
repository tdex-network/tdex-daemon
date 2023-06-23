package mapper

import (
	v091domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v091-domain"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
)

func (m *mapperService) FromV091MarketsToV1Markets(
	markets []*v091domain.Market,
) ([]domain.Market, error) {
	res := make([]domain.Market, 0, len(markets))
	for _, v := range markets {
		market, err := m.fromV091MarketToV1Market(v)
		if err != nil {
			return nil, err
		}
		res = append(res, *market)
	}

	return res, nil
}

func (m *mapperService) fromV091MarketToV1Market(
	market *v091domain.Market,
) (*domain.Market, error) {
	basePrice := ""
	if !market.Price.BasePrice.IsZero() {
		basePrice = market.Price.BasePrice.String()
	}

	quotePrice := ""
	if !market.Price.QuotePrice.IsZero() {
		basePrice = market.Price.QuotePrice.String()
	}

	return &domain.Market{
		BaseAsset:           market.BaseAsset,
		QuoteAsset:          market.QuoteAsset,
		Name:                market.AccountName(),
		BaseAssetPrecision:  market.BaseAssetPrecision,
		QuoteAssetPrecision: market.QuoteAssetPrecision,
		PercentageFee: domain.MarketFee{
			BaseAsset:  uint64(market.Fee),
			QuoteAsset: uint64(market.Fee),
		},
		FixedFee: domain.MarketFee{
			BaseAsset:  uint64(market.FixedFee.BaseFee),
			QuoteAsset: uint64(market.FixedFee.QuoteFee),
		},
		Tradable:     market.Tradable,
		StrategyType: market.Strategy.Type + 1,
		Price: domain.MarketPrice{
			BasePrice:  basePrice,
			QuotePrice: quotePrice,
		},
	}, nil
}
