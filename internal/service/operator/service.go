package operatorservice

import (
	"context"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/domain/market"
	"github.com/tdex-network/tdex-daemon/internal/storage"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"
)

// Service is used to implement Operator service.
type Service struct {
	marketRepository market.Repository
	pb.UnimplementedOperatorServer
	crawlerSvc crawler.Service
}

// NewService returns a Operator Service
func NewService(marketRepo market.Repository) (*Service, error) {
	explorerSvc := explorer.NewService()

	markets, err := marketRepo.GetAllMarkets(context.Background())
	if err != nil {
		return nil, err
	}

	for _, m := range markets {

	}

	crawlSvc := crawler.NewService(explorerSvc)

	return &Service{
		marketRepository: marketRepo,
	}
}

func validateBaseAsset(baseAsset string) error {
	if baseAsset != config.GetString(config.BaseAssetKey) {
		return storage.ErrMarketNotExist
	}

	return nil
}
