package operatorservice

import (
	"context"
	"fmt"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/domain/market"
	"github.com/tdex-network/tdex-daemon/internal/domain/unspent"
	"github.com/tdex-network/tdex-daemon/internal/storage"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"
)

// Service is used to implement Operator service.
type Service struct {
	marketRepository market.Repository
	unspentRepo      unspent.Repository
	pb.UnimplementedOperatorServer
	crawlerSvc crawler.Service
}

// NewService returns a Operator Service
func NewService(
	marketRepo market.Repository,
	unspentRepo unspent.Repository,
	crawlerSvc crawler.Service,
) (*Service, error) {
	svc := &Service{
		marketRepository: marketRepo,
		unspentRepo:      unspentRepo,
		crawlerSvc:       crawlerSvc,
	}

	return svc, nil
}

func (s *Service) ObserveBlockChain() {
	go s.crawlerSvc.Start()
	go s.handleBlockChainEvents()
}

func (s *Service) handleBlockChainEvents() {
events:
	for event := range s.crawlerSvc.GetEventChannel() {
		switch event.EventType {
		case crawler.FeeAccountDeposit:
			unspents := make([]unspent.Unspent, 0)
			for _, utxo := range event.Utxos {
				u := unspent.Unspent{
					Txid:      utxo.Hash(),
					Vout:      utxo.Index(),
					Value:     utxo.Value(),
					AssetHash: utxo.Asset(),
					Address:   event.Address,
					Spent:     false,
				}
				unspents = append(unspents, u)
			}
			s.unspentRepo.AddUnspent(unspents)
			markets, err := s.marketRepository.GetTradableMarkets(context.Background())
			if err != nil {
				fmt.Println(err)
				continue events
			}

			balance := s.unspentRepo.GetBalance(event.Address, event.AssetHash)
			if balance < uint64(config.GetInt(config.FeeAccountBalanceTresholdKey)) {
				fmt.Println("fee account balance too low - Trades and" +
					" deposits will be disabled")
				for _, m := range markets {
					err := s.marketRepository.CloseMarket(context.Background(), m.QuoteAssetHash())
					if err != nil {
						fmt.Println(err)
						continue events
					}
				}
				continue events
			}

			for _, m := range markets {
				err := s.marketRepository.OpenMarket(context.Background(),
					m.QuoteAssetHash())
				if err != nil {
					fmt.Println(err)
					continue events
				}
			}

		case crawler.MarketAccountDeposit:
			unspents := make([]unspent.Unspent, 0)
			for _, utxo := range event.Utxos {
				u := unspent.Unspent{
					Txid:      utxo.Hash(),
					Vout:      utxo.Index(),
					Value:     utxo.Value(),
					AssetHash: utxo.Asset(),
					Address:   event.Address,
					Spent:     false,
				}
				unspents = append(unspents, u)
			}
			s.unspentRepo.AddUnspent(unspents)

			m, _, err := s.marketRepository.GetMarketByAsset(
				context.Background(),
				event.AssetHash,
			)
			if err != nil {
				fmt.Println(err)
				continue events
			}

			fundingTxs := make([]market.OutpointWithAsset, 0)
			for _, u := range event.Utxos {
				tx := market.OutpointWithAsset{
					Asset: u.Asset(),
					Txid:  u.Hash(),
					Vout:  int(u.Index()),
				}
				fundingTxs = append(fundingTxs, tx)
			}

			err = m.FundMarket(fundingTxs)
			if err != nil {
				fmt.Println(err)
				continue events
			}
		}
	}
}

func validateBaseAsset(baseAsset string) error {
	if baseAsset != config.GetString(config.BaseAssetKey) {
		return storage.ErrMarketNotExist
	}

	return nil
}
