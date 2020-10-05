package application

import "github.com/tdex-network/tdex-daemon/internal/core/domain"

type Market struct {
	BaseAsset  string
	QuoteAsset string
}

type MarketWithFee struct {
	Market
	Fee
}

type Fee struct {
	FeeAsset   string
	BasisPoint int64
}

type MarketWithPrice struct {
	Market
	Price
}

type Price struct {
	BasePrice  float32
	QuotePrice float32
}

type PriceWithFee struct {
	Price
	Fee
	Amount uint64
}

type MarketStrategy struct {
	Market
	Strategy domain.StrategyType
}
