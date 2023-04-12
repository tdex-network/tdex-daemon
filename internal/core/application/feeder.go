package application

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/core/application/feeder"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

type FeederService interface {
	AddPriceFeed(
		ctx context.Context, market ports.Market, source, ticker string,
	) (string, error)
	StartPriceFeed(ctx context.Context, id string) error
	StopPriceFeed(ctx context.Context, id string) error
	UpdatePriceFeed(ctx context.Context, id, source, ticker string) error
	RemovePriceFeed(ctx context.Context, id string) error
	GetPriceFeed(ctx context.Context, id string) (ports.PriceFeedInfo, error)
	ListSources(ctx context.Context) []string
	ListPriceFeeds(ctx context.Context) ([]ports.PriceFeedInfo, error)
	Close()
}

func NewFeederService(
	feederSvc ports.PriceFeeder, repoManager ports.RepoManager,
) (FeederService, error) {
	svc, err := feeder.NewService(feederSvc, repoManager)
	if err != nil {
		return nil, err
	}

	return svc, nil
}
