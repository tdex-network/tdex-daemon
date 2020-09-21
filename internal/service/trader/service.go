package tradeservice

import (
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/domain/market"
	"github.com/tdex-network/tdex-daemon/internal/domain/trade"
	"github.com/tdex-network/tdex-daemon/internal/domain/unspent"
	"github.com/tdex-network/tdex-daemon/internal/domain/vault"
	"github.com/tdex-network/tdex-daemon/internal/storage"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/trade"
)

// Service is used to implement Trader service.
type Service struct {
	marketRepository  market.Repository
	unspentRepository unspent.Repository
	vaultRepository   vault.Repository
	tradeRepository   trade.Repository
	explorerService   explorer.Service
	pb.UnimplementedTradeServer
}

// NewService returns a Trade Service
func NewService(
	marketRepo market.Repository,
	unspentRepo unspent.Repository,
	vaultRepo vault.Repository,
	tradeRepo trade.Repository,
	explorerSvc explorer.Service,
) *Service {
	return &Service{
		marketRepository:  marketRepo,
		unspentRepository: unspentRepo,
		vaultRepository:   vaultRepo,
		tradeRepository:   tradeRepo,
		explorerService:   explorerSvc,
	}
}

func validateBaseAsset(baseAsset string) error {
	if baseAsset != config.GetString(config.BaseAssetKey) {
		return storage.ErrMarketNotExist
	}

	return nil
}
