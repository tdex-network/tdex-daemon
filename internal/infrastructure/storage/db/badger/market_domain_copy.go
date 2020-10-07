package dbbadger

import (
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	mm "github.com/tdex-network/tdex-daemon/pkg/marketmaking"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking/formula"
)

const (
	BalancedStrategyType = 1
)

type Market struct {
	AccountIndex int `badgerhold:"AccountIndex"`
	BaseAsset    string
	QuoteAsset   string
	Fee          int64
	FeeAsset     string
	Tradable     bool
	Strategy     int
	BasePrice    PriceByTime
	QuotePrice   PriceByTime
}

type Price float32

type PriceByTime map[uint64]Price

type StrategyType int32

func MapDomainMarketToInfraMarket(market domain.Market) *Market {
	basePrice := market.GetBasePrice()
	basePriceCopy := make(map[uint64]Price)
	for k, v := range basePrice {
		basePriceCopy[k] = Price(v)
	}

	quotePrice := market.GetQuotePrice()
	quotePriceCopy := make(map[uint64]Price)
	for k, v := range quotePrice {
		quotePriceCopy[k] = Price(v)
	}

	var strategy int
	switch market.GetStrategy().Formula().(type) {
	case formula.BalancedReserves:
		strategy = 1
	}

	return &Market{
		AccountIndex: market.AccountIndex(),
		BaseAsset:    market.BaseAssetHash(),
		QuoteAsset:   market.QuoteAssetHash(),
		Fee:          market.Fee(),
		FeeAsset:     market.FeeAsset(),
		Tradable:     market.IsTradable(),
		Strategy:     strategy,
		BasePrice:    basePriceCopy,
		QuotePrice:   quotePriceCopy,
	}
}

func MapInfraMarketToDomainMarket(market Market) *domain.Market {
	domainMarket := &domain.Market{}
	domainMarket.SetAccountIndex(market.AccountIndex)
	domainMarket.SetBaseAsset(market.BaseAsset)
	domainMarket.SetQuoteAsset(market.QuoteAsset)
	domainMarket.SetFee(market.Fee)
	domainMarket.SetFeeAsset(market.FeeAsset)
	domainMarket.SetTradable(market.Tradable)

	var f mm.MakingFormula
	switch market.Strategy {
	case BalancedStrategyType:
		f = formula.BalancedReserves{}
	}

	domainMarket.SetStrategy(mm.NewStrategyFromFormula(f))

	basePrice := market.BasePrice
	basePriceCopy := make(map[uint64]float32)
	for k, v := range basePrice {
		basePriceCopy[k] = float32(v)
	}
	domainMarket.SetBasePrice(basePriceCopy)

	quotePrice := market.QuotePrice
	quotePriceCopy := make(map[uint64]float32)
	for k, v := range quotePrice {
		quotePriceCopy[k] = float32(v)
	}
	domainMarket.SetQuotePrice(basePriceCopy)

	return domainMarket
}
