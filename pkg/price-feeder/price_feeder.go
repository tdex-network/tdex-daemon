package pricefeeder

import (
	"github.com/shopspring/decimal"
)

type PriceFeed struct {
	Market Market
	Price  Price
}

type Market struct {
	BaseAsset  string
	QuoteAsset string
	Ticker     string
}

type Price struct {
	BasePrice  decimal.Decimal
	QuotePrice decimal.Decimal
}

type PriceFeeder interface {
	Start() chan PriceFeed
	Stop()

	SubscribeMarkets([]Market) error
	UnsubscribeMarkets([]Market) error
	ListSubscriptions() []Market
}
