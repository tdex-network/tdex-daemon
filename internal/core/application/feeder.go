package application

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/core/application/feeder"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

type FeederService interface {
	Stop(ctx context.Context)
	StartFeed(ctx context.Context, feedID string) error
	StopFeed(ctx context.Context, feedID string) error
	AddPriceFeed(
		ctx context.Context, market ports.Market, source, ticker string,
	) (string, error)
	UpdatePriceFeed(
		ctx context.Context, id, source, ticker string,
	) error
	RemovePriceFeed(ctx context.Context, id string) error
	GetPriceFeed(
		ctx context.Context, baseAsset, quoteAsset string,
	) (ports.PriceFeed, error)
	ListSources(ctx context.Context) []string
	ListPriceFeeds(ctx context.Context) ([]ports.PriceFeed, error)
}

func NewFeederService(
	feederSvc ports.PriceFeeder,
	repoManager ports.RepoManager,
) (FeederService, error) {
	svc, err := feeder.NewService(feederSvc, repoManager)
	if err != nil {

	}

	return &svc, nil
}
