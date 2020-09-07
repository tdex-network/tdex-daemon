package operatorservice

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/domain/market"
	"github.com/tdex-network/tdex-daemon/internal/domain/unspent"
	"github.com/tdex-network/tdex-daemon/internal/storage"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"
)

// Service is used to implement Operator service.
type Service struct {
	marketRepository  market.Repository
	unspentRepository unspent.Repository
	pb.UnimplementedOperatorServer
	crawlerSvc crawler.Service
}

// NewService returns a Operator Service
func NewService(
	marketRepository market.Repository,
	unspentRepository unspent.Repository,
	crawlerSvc crawler.Service,
) (*Service, error) {
	svc := &Service{
		marketRepository:  marketRepository,
		unspentRepository: unspentRepository,
		crawlerSvc:        crawlerSvc,
	}

	return svc, nil
}

func (s *Service) ObserveBlockchain() {
	go s.crawlerSvc.Start()
	go s.handleBlockChainEvents()
}

func (s *Service) handleBlockChainEvents() {
events:
	for event := range s.crawlerSvc.GetEventChannel() {
		switch event.EventType {
		case crawler.FeeAccountDeposit:
			unspents := make([]unspent.Unspent, 0)
		utxoLoop:
			for _, utxo := range event.Utxos {
				isTrxConfirmed, err := explorer.IsTransactionConfirmed(
					utxo.Hash(),
				)
				if err != nil {
					log.Error(err)
					continue utxoLoop
				}
				if isTrxConfirmed {
					u := unspent.NewUnspent(
						utxo.Hash(),
						utxo.Asset(),
						event.Address,
						utxo.Index(),
						utxo.Value(),
						false,
						false,
						nil, //TODO populate this
					)
					unspents = append(unspents, u)
				}
			}
			s.unspentRepository.AddUnspent(unspents)
			markets, err := s.marketRepository.GetTradableMarkets(context.Background())
			if err != nil {
				fmt.Println(err)
				continue events
			}

			balance := s.unspentRepository.GetBalance(event.Address, event.AssetHash)
			if balance < uint64(config.GetInt(config.FeeAccountBalanceThresholdKey)) {
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
		utxoLoop1:
			for _, utxo := range event.Utxos {
				isTrxConfirmed, err := explorer.IsTransactionConfirmed(
					utxo.Hash(),
				)
				if err != nil {
					log.Error(err)
					continue utxoLoop1
				}
				if isTrxConfirmed {
					u := unspent.NewUnspent(
						utxo.Hash(),
						utxo.Asset(),
						event.Address,
						utxo.Index(),
						utxo.Value(),
						false,
						false,
						nil, //TODO populate this
					)
					unspents = append(unspents, u)
				}
			}
			s.unspentRepository.AddUnspent(unspents)

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
