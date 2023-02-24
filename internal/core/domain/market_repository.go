package domain

import "context"

// MarketRepository is the abstraction for any kind of database intended to
// persist Markets.
type MarketRepository interface {
	// AddMarket adds a new market to the repository.
	AddMarket(ctx context.Context, market *Market) error
	// GetMarketByName returns the market with the given name.
	GetMarketByName(
		ctx context.Context, marketName string,
	) (*Market, error)
	// GetMarketByAssets returns the market with a given asset pair.
	GetMarketByAssets(
		ctx context.Context, baseAsset, quoteAsset string,
	) (*Market, error)
	// GetTradableMarkets returns all markets that are open for trading.
	GetTradableMarkets(ctx context.Context) ([]Market, error)
	// GetAllMarkets returns all markets.
	GetAllMarkets(ctx context.Context) ([]Market, error)
	// UpdateMarket updates the state of a market. The closure function let's to
	// commit multiple changes to a certain market in a transactional way.
	UpdateMarket(
		ctx context.Context,
		marketName string, updateFn func(m *Market) (*Market, error),
	) error
	// OpenMarket makes a market open for trading.
	OpenMarket(ctx context.Context, marketName string) error
	// CloseMarket puts a market in pause and not available for trading.
	CloseMarket(ctx context.Context, marketName string) error
	// DeleteMarket removes a market from the repository.
	DeleteMarket(ctx context.Context, marketName string) error
	// UpdateMarketPrice updates the price of a given market.
	UpdateMarketPrice(
		ctx context.Context, marketName string, price MarketPrice,
	) error
}
