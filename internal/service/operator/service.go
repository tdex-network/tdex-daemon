package operatorservice

import (
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/domain/market"
	"github.com/tdex-network/tdex-daemon/internal/storage"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"
)

// Service is used to implement Operator service.
type Service struct {
	marketRepository market.Repository
	pb.UnimplementedOperatorServer
}

// NewService returns a Operator Service
func NewService(marketRepo market.Repository) *Service {
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
