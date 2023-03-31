package ports

import "context"

type PriceFeeder interface {
	// Start starts the price feeder service which is periodically fetching
	//and updating market prices from external price provide
	Start(ctx context.Context, markets []Market) (chan PriceFeedChan, error)
	// Stop stops the price feeder service
	Stop(ctx context.Context)
	// StartFeed starts forwarding price feeds from the given source to the
	//target market.
	StartFeed(ctx context.Context, id string) error
	// StopFeed stops updating specific price feed for the given feedID
	StopFeed(ctx context.Context, feedID string) error
	// AddPriceFeed adds a new price feed
	//it is necessary to start the feed after adding it in order to start
	//updating prices
	AddPriceFeed(
		ctx context.Context, market Market, source, ticker string,
	) (string, error)
	// UpdatePriceFeed updates an existing price feed
	UpdatePriceFeed(ctx context.Context, id, source, ticker string) error
	// RemovePriceFeed removes an existing price feed
	RemovePriceFeed(ctx context.Context, feedID string) error
	// GetPriceFeedForMarket returns the price feed for the given market
	GetPriceFeedForMarket(
		ctx context.Context, baseAsset string, quoteAsset string,
	) (PriceFeed, error)
	// GetPriceFeed returns the price feed for the given feedID
	GetPriceFeed(ctx context.Context, id string) (PriceFeed, error)
	// ListPriceFeeds returns the list of price feeds of all markets
	ListPriceFeeds(ctx context.Context) ([]PriceFeed, error)
	// ListSources returns the list of supported price sources
	ListSources(ctx context.Context) []string
}
