package pricefeeder

import (
	"context"
)

type PriceFeedStore interface {
	// AddPriceFeed adds a new price feed to the repository.
	AddPriceFeed(ctx context.Context, info PriceFeedInfo) error
	// GetPriceFeed returns the price feed with the given ID.
	GetPriceFeed(ctx context.Context, id string) (*PriceFeedInfo, error)
	// UpdatePriceFeed updates the price feed.
	UpdatePriceFeed(
		ctx context.Context, id string,
		updateFn func(priceFeed *PriceFeedInfo) (*PriceFeedInfo, error),
	) error
	// RemovePriceFeed removes the price feed with the given ID.
	RemovePriceFeed(ctx context.Context, id string) error
	// GetAllPriceFeeds returns all price feeds of all markets.
	GetAllPriceFeeds(ctx context.Context) ([]PriceFeedInfo, error)
	Close()
}
