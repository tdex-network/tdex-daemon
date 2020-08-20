package market

import "context"

// Repository defines the abstraction for Market
type Repository interface {
	GetMarketByAsset(ctx context.Context, quoteAsset string) (market *Market, accountIndex int, err error)
	GetOrCreateMarket(ctx context.Context, accountIndex int) (market *Market, err error)
	GetTradableMarkets(ctx context.Context) ([]*Market, error)
	UpdateMarket(
		ctx context.Context,
		accountIndex int,
		updateFn func(m *Market) (*Market, error),
	) error
	OpenMarket(ctx context.Context, quoteAsset string) error
	CloseMarket(ctx context.Context, quoteAsset string) error
}
