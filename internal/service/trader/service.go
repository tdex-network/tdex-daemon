package tradeservice

import (
	"context"
	"fmt"

	"github.com/tdex-network/tdex-daemon/internal/domain/market"
	"github.com/tdex-network/tdex-daemon/internal/storage"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/trade"
	"github.com/thanhpk/randstr"
)

// Service is used to implement Trader service.
type Service struct {
	marketRepository market.Repository
	pb.UnimplementedTradeServer
}

// NewServer returns a Trade Server
func NewServer() *Service {
	return &Service{
		marketRepository: storage.NewInMemoryMarketRepository(),
	}
}

//AddTestMarket ...
func (s *Service) AddTestMarket() {
	_, latestAccountIndex, err := s.marketRepository.GetLatestMarket(context.Background())
	if err != nil {
		println("latest market")
		panic(fmt.Errorf("latest market: %w", err))
	}

	nextAccountIndex := latestAccountIndex + 1
	randAssetHash := randstr.Hex(32)

	if err := s.marketRepository.UpdateMarket(context.Background(), nextAccountIndex, func(m *market.Market) (*market.Market, error) {
		fundingTxs := []market.OutpointWithAsset{
			{"5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225", "abc", 1},
			{"5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225", "def", 1},
			{randAssetHash, "abc", 0},
			{randAssetHash, "cde", 0},
			{"5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225", "foobar", 1},
		}
		if err := m.FundMarket(fundingTxs); err != nil {
			return nil, err
		}

		if err := m.MakeTradable(); err != nil {
			return nil, err
		}

		return m, nil
	}); err != nil {
		panic(fmt.Errorf("update market: %w", err))
	}

	println("Created, funded and opened a market " + randAssetHash)
}
