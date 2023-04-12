package ports

import "context"

type PriceFeeder interface {
	// AddPriceFeed adds a new price feed for the target market.
	AddPriceFeed(
		ctx context.Context, market Market, source, ticker string,
	) (string, error)
	// StartFeed starts forwarding price feeds from the given source to the
	// target market.
	StartPriceFeed(ctx context.Context, id string) (chan PriceFeed, error)
	// StopFeed stops updating the target price feed.
	StopPriceFeed(ctx context.Context, id string) error
	// UpdatePriceFeed updates an existing price feed.
	UpdatePriceFeed(ctx context.Context, id, source, ticker string) error
	// RemovePriceFeed removes an existing price feed.
	RemovePriceFeed(ctx context.Context, id string) error
	// GetPriceFeed returns info about the target price feed.
	GetPriceFeed(ctx context.Context, id string) (PriceFeedInfo, error)
	// ListPriceFeeds returns the list of price feeds.
	ListPriceFeeds(ctx context.Context) ([]PriceFeedInfo, error)
	// ListSources returns the list of supported price sources.
	ListSources(ctx context.Context) []string

	Close()
}
