package pricefeederinfra

import (
	"context"
)

type PriceFeedRepository interface {
	// AddPriceFeed adds a new price feed to the repository.
	AddPriceFeed(ctx context.Context, priceFeed PriceFeed) error
	// GetPriceFeed returns the price feed with the given ID.
	GetPriceFeed(ctx context.Context, id string) (*PriceFeed, error)
	// GetPriceFeedsByMarket returns all price feed for a given market.
	GetPriceFeedsByMarket(
		ctx context.Context,
		market Market,
	) (*PriceFeed, error)
	// UpdatePriceFeed updates the price feed.
	UpdatePriceFeed(
		ctx context.Context,
		ID string,
		updateFn func(priceFeed *PriceFeed) (*PriceFeed, error),
	) error
	// RemovePriceFeed removes the price feed with the given ID.
	RemovePriceFeed(ctx context.Context, id string) error
}
