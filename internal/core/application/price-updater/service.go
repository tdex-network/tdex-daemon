package priceupdater

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"

	log "github.com/sirupsen/logrus"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

// Service periodically updates markets prices.
// prices are fed into service from external price provider eg. Kraken.
type Service struct {
	// priceFeeder is the external price provider.
	priceFeeder ports.PriceFeeder
	// repoManager is the used to access the db.
	repoManager ports.RepoManager
}

func NewService(
	priceFeeder ports.PriceFeeder,
) *Service {
	return &Service{
		priceFeeder: priceFeeder,
	}
}

// Start starts the service.
// It subscribes to the priceFeeder for price updates.
// It also starts a goroutine that reads from the priceFeeder feed channel
// and updates the db with the new prices.
func (s *Service) Start() error {
	markets, err := s.repoManager.MarketRepository().GetAllMarkets(
		context.Background(),
	)
	if err != nil {
		return err
	}

	mkts := make([]ports.Market, 0, len(markets))
	for _, m := range markets {
		// convert to ports.Market
		mkts = append(mkts, &Market{
			BaseAsset:    m.BaseAsset,
			QuoteAsset:   m.QuoteAsset,
			MarketTicker: "", //TODO populate ticker
		})
	}

	if err := s.priceFeeder.SubscribeMarkets([]ports.Market(mkts)); err != nil {
		return err
	}

	if err := s.priceFeeder.Start(); err != nil {
		return err
	}

	go func() {
		log.Debugln("reading price feed chan started")

		for priceFeed := range s.priceFeeder.FeedChan() {
			market, err := s.repoManager.MarketRepository().GetMarketByAssets(
				context.Background(),
				priceFeed.GetMarket().GetBaseAsset(),
				priceFeed.GetMarket().GetQuoteAsset(),
			)
			if err != nil {
				log.WithError(err).Errorf(
					"cannot get market %s-%s",
					priceFeed.GetMarket().GetBaseAsset(),
					priceFeed.GetMarket().GetQuoteAsset(),
				)
				continue
			}

			if err := s.repoManager.MarketRepository().UpdateMarketPrice(
				context.Background(),
				market.Name,
				domain.MarketPrice{
					BasePrice:  priceFeed.GetPrice().GetBasePrice().String(),
					QuotePrice: priceFeed.GetPrice().GetQuotePrice().String(),
				},
			); err != nil {
				log.WithError(err).Errorf(
					"cannot update market price %s-%s",
					priceFeed.GetMarket().GetBaseAsset(),
					priceFeed.GetMarket().GetQuoteAsset(),
				)
			}
		}

		log.Debugln("reading price feed chan stopped")
	}()

	return nil
}

// Stop stops the service.
func (s *Service) Stop() {
	s.priceFeeder.Stop()
}
