package tradeservice

import (
	"context"
	"fmt"
	"strconv"

	"github.com/tdex-network/tdex-daemon/internal/domain/market"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/trade"
	"github.com/thanhpk/randstr"
)

// Service is used to implement Trader service.
type Service struct {
	marketRepository market.Repository
	pb.UnimplementedTradeServer
}

// NewService returns a Trade Service
func NewService(marketRepo market.Repository) *Service {
	return &Service{
		marketRepository: marketRepo,
	}
}

//AddTestMarket ...
func (s *Service) AddTestMarket(makeItTradable bool) {
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

		if makeItTradable {
			err = m.MakeTradable()
		} else {
			err = m.MakeNotTradable()
		}
		if err != nil {
			return nil, err
		}

		return m, nil
	}); err != nil {
		panic(fmt.Errorf("update market: %w", err))
	}

	println("Created market | tradable: " + strconv.FormatBool(makeItTradable) + " | " + randAssetHash)
}
