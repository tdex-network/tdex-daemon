package operatorservice

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/domain/market"
	"github.com/tdex-network/tdex-daemon/internal/domain/trade"
	"github.com/tdex-network/tdex-daemon/internal/domain/unspent"
	"github.com/tdex-network/tdex-daemon/internal/domain/vault"
	"github.com/tdex-network/tdex-daemon/internal/storage"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"
)

// Service is used to implement Operator service.
type Service struct {
	marketRepository  market.Repository
	unspentRepository unspent.Repository
	vaultRepository   vault.Repository
	tradeRepository   trade.Repository
	pb.UnimplementedOperatorServer
	crawlerSvc  crawler.Service
	explorerSvc explorer.Service
}

// NewService returns a Operator Service
func NewService(
	marketRepository market.Repository,
	unspentRepository unspent.Repository,
	vaultRepository vault.Repository,
	tradeRepository trade.Repository,
	crawlerSvc crawler.Service,
	explorerSvc explorer.Service,
) (*Service, error) {
	svc := &Service{
		marketRepository:  marketRepository,
		unspentRepository: unspentRepository,
		vaultRepository:   vaultRepository,
		tradeRepository:   tradeRepository,
		crawlerSvc:        crawlerSvc,
		explorerSvc:       explorerSvc,
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
		switch event.Type() {
		case crawler.FeeAccountDeposit:
			e := event.(crawler.AddressEvent)
			unspents := make([]unspent.Unspent, 0)
			if len(e.Utxos) > 0 {

			utxoLoop:
				for _, utxo := range e.Utxos {
					isTrxConfirmed, err := s.explorerSvc.IsTransactionConfirmed(
						utxo.Hash(),
					)
					if err != nil {
						log.Warn(err)
						continue utxoLoop
					}
					if isTrxConfirmed {
						u := unspent.NewUnspent(
							utxo.Hash(),
							utxo.Asset(),
							e.Address,
							utxo.Index(),
							utxo.Value(),
							false,
							false,
							nil, //TODO should this be populated
							nil,
						)
						unspents = append(unspents, u)
					}
				}
				err := s.unspentRepository.AddUnspents(context.Background(), unspents)
				if err != nil {
					log.Warn(err)
					continue events
				}

				markets, err := s.marketRepository.GetTradableMarkets(context.Background())
				if err != nil {
					log.Warn(err)
					continue events
				}

				addresses, _, err := s.vaultRepository.GetAllDerivedAddressesAndBlindingKeysForAccount(
					context.Background(),
					vault.FeeAccount,
				)
				if err != nil {
					log.Warn(err)
					continue events
				}

				var feeAccountBalance uint64
				for _, a := range addresses {
					feeAccountBalance += s.unspentRepository.GetBalance(
						context.Background(),
						a,
						config.GetString(config.BaseAssetKey),
					)
				}

				if feeAccountBalance < uint64(config.GetInt(config.FeeAccountBalanceThresholdKey)) {
					log.Debug("fee account balance too low - Trades and" +
						" deposits will be disabled")
					for _, m := range markets {
						err := s.marketRepository.CloseMarket(context.Background(), m.QuoteAssetHash())
						if err != nil {
							log.Warn(err)
							continue events
						}
					}
					continue events
				}

				for _, m := range markets {
					err := s.marketRepository.OpenMarket(
						context.Background(),
						m.QuoteAssetHash(),
					)
					if err != nil {
						log.Warn(err)
						continue events
					}
					log.Debug(fmt.Sprintf(
						"market %v, opened",
						m.AccountIndex(),
					))
				}
			}

		case crawler.MarketAccountDeposit:
			e := event.(crawler.AddressEvent)
			unspents := make([]unspent.Unspent, 0)
			if len(e.Utxos) > 0 {
			utxoLoop1:
				for _, utxo := range e.Utxos {
					isTrxConfirmed, err := s.explorerSvc.IsTransactionConfirmed(
						utxo.Hash(),
					)
					if err != nil {
						log.Warn(err)
						continue utxoLoop1
					}
					if isTrxConfirmed {
						u := unspent.NewUnspent(
							utxo.Hash(),
							utxo.Asset(),
							e.Address,
							utxo.Index(),
							utxo.Value(),
							false,
							false,
							nil, //TODO should this be populated
							nil,
						)
						unspents = append(unspents, u)
					}
				}
				err := s.unspentRepository.AddUnspents(context.Background(), unspents)
				if err != nil {
					log.Warn(err)
					continue events
				}

				fundingTxs := make([]market.OutpointWithAsset, 0)
				for _, u := range e.Utxos {
					tx := market.OutpointWithAsset{
						Asset: u.Asset(),
						Txid:  u.Hash(),
						Vout:  int(u.Index()),
					}
					fundingTxs = append(fundingTxs, tx)
				}

				m, err := s.marketRepository.GetOrCreateMarket(
					context.Background(),
					e.AccountIndex,
				)
				if err != nil {
					log.Error(err)
					continue events
				}

				if err := s.marketRepository.UpdateMarket(context.Background(), m.AccountIndex(), func(m *market.Market) (*market.Market, error) {

					if m.IsFunded() {
						return m, nil
					}

					if err := m.FundMarket(fundingTxs); err != nil {
						return nil, err
					}

					log.Info("deposit: funding market with quote asset ", m.QuoteAssetHash())

					return m, nil
				}); err != nil {
					log.Warn(err)
					continue events
				}

			}

		case crawler.TransactionConfirmed:
			//TODO
		}
	}
}

func validateBaseAsset(baseAsset string) error {
	if baseAsset != config.GetString(config.BaseAssetKey) {
		return storage.ErrMarketNotExist
	}

	return nil
}
