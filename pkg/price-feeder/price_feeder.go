package pricefeeder

import "github.com/shopspring/decimal"

type PriceFeeder interface {
	WellKnownMarkets() []Market
	SubscribeMarkets([]Market) error

	Start() error
	Stop()

	FeedChan() chan PriceFeed
}

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
