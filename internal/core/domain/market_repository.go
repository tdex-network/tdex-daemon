package domain

import "context"

// MarketRepository defines the abstraction for Market
type MarketRepository interface {
	// Retrieves a market with the given account index. If not found, a new entry shall be created.
	GetOrCreateMarket(ctx context.Context, market *Market) (*Market, error)
	// Retrieves a market with a given account index.
	GetMarketByAccount(ctx context.Context, accountIndex int) (market *Market, err error)
	// Retrieves a market with a given a quote asset hash.
	GetMarketByAsset(ctx context.Context, quoteAsset string) (market *Market, accountIndex int, err error)
	// Retrieves the latest market sorted by account index
	GetLatestMarket(ctx context.Context) (market *Market, accountIndex int, err error)
	// Retrieves all the markets that are open for trading
	GetTradableMarkets(ctx context.Context) ([]Market, error)
	// Retrieves all the markets
	GetAllMarkets(ctx context.Context) ([]Market, error)
	// Updates the state of a market. In order to be flexible for many use case and to manage
	// at an higher level the possible errors, an update closure function shall be passed
	UpdateMarket(
		ctx context.Context,
		accountIndex int,
		updateFn func(m *Market) (*Market, error),
	) error
	// Open and close trading activities for a market with the given quote asset hash
	OpenMarket(ctx context.Context, quoteAsset string) error
	CloseMarket(ctx context.Context, quoteAsset string) error
	// Update only the price without touching market details
	UpdatePrices(ctx context.Context, accountIndex int, prices Prices) error
	// DeleteMarket deletes market for accountIndex
	DeleteMarket(ctx context.Context, accountIndex int) error
}
