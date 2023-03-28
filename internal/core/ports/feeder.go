package ports

import "context"

type PriceFeeder interface {
	// Start starts the price feeder service which is periodically fetching
	//and updating market prices from external price provide
	Start(ctx context.Context, markets []Market) (chan PriceFeedChan, error)
	// Stop stops the price feeder service
	Stop(ctx context.Context)
	// StartFeed starts updating specific price feed for the given feedID
	StartFeed(ctx context.Context, feedID string) error
	// StopFeed stops updating specific price feed for the given feedID
	StopFeed(ctx context.Context, feedID string) error
	// AddPriceFeed adds a new price feed
	//it is necessary to start the feed after adding it in order to start updating prices
	AddPriceFeed(ctx context.Context, req AddPriceFeedReq) (string, error)
	// UpdatePriceFeed updates an existing price feed
	UpdatePriceFeed(ctx context.Context, req UpdatePriceFeedReq) error
	// RemovePriceFeed removes an existing price feed
	RemovePriceFeed(ctx context.Context, feedID string) error
	// GetPriceFeed returns the price feed for the given market
	GetPriceFeed(ctx context.Context, baseAsset string, quoteAsset string) (*PriceFeed, error)
	// GetPriceFeedForFeedID returns the price feed for the given feedID
	GetPriceFeedForFeedID(ctx context.Context, feedID string) (*PriceFeed, error)
	// ListSources returns the list of supported price sources
	ListSources(ctx context.Context) []string
}
