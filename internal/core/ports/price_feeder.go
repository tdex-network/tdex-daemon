package ports

type PriceFeeder interface {
	WellKnownMarkets() []Market
	SubscribeMarkets([]Market) error

	Start() error
	Stop()

	FeedChan() chan PriceFeed
}
