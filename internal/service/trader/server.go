package tradeservice

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/domain/market"
	"github.com/tdex-network/tdex-daemon/internal/storage"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/trade"
	"github.com/thanhpk/randstr"
)

// Server is used to implement Trader service.
type Server struct {
	marketRepository market.Repository
	pb.UnimplementedTradeServer
}

// NewServer returns a Trade Server
func NewServer() *Server {
	return &Server{
		marketRepository: storage.NewInMemoryMarketRepository(),
	}
}

//AddTestMarket ...
func (s *Server) AddTestMarket() {
	if err := s.marketRepository.UpdateMarket(context.Background(), 0, func(m *market.Market) (*market.Market, error) {
		randAssetHash := randstr.Hex(32)
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
		panic(err)
	}

	println("Created, funded and opened a market")
}
