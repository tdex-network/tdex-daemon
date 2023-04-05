package feeder

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

type Service struct {
	repoManager ports.RepoManager
	feederSvc   ports.PriceFeeder
}

func NewService(
	feederSvc ports.PriceFeeder,
	repoManager ports.RepoManager,
) (Service, error) {
	svc := Service{
		repoManager: repoManager,
		feederSvc:   feederSvc,
	}

	if err := svc.startPriceFeed(feederSvc); err != nil {
		return Service{}, err
	}

	return svc, nil
}

func (s *Service) Stop(ctx context.Context) {
	s.feederSvc.Stop(ctx)
}

func (s *Service) StartFeed(ctx context.Context, feedID string) error {
	priceFeed, err := s.feederSvc.GetPriceFeed(ctx, feedID)
	if err != nil {
		return err
	}

	market, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx,
		priceFeed.GetMarket().GetBaseAsset(),
		priceFeed.GetMarket().GetQuoteAsset(),
	)
	if err != nil {
		return err
	}

	if !market.IsStrategyPluggable() {
		return fmt.Errorf("market is not pluggable")
	}

	return s.feederSvc.StartFeed(ctx, feedID)
}

func (s *Service) StopFeed(ctx context.Context, feedID string) error {
	return s.feederSvc.StopFeed(ctx, feedID)
}

func (s *Service) AddPriceFeed(
	ctx context.Context, market ports.Market, source, ticker string,
) (string, error) {
	return s.feederSvc.AddPriceFeed(ctx, market, source, ticker)
}

func (s *Service) UpdatePriceFeed(
	ctx context.Context, id, source, ticker string,
) error {
	priceFeed, err := s.feederSvc.GetPriceFeed(ctx, id)
	if err != nil {
		return err
	}

	if priceFeed.IsStarted() {
		return fmt.Errorf("feed needs to be stopped before updating it")
	}

	return s.feederSvc.UpdatePriceFeed(ctx, id, source, ticker)
}

func (s *Service) RemovePriceFeed(ctx context.Context, id string) error {
	priceFeed, err := s.feederSvc.GetPriceFeed(ctx, id)
	if err != nil {
		return err
	}

	if priceFeed.IsStarted() {
		return fmt.Errorf("feed needs to be stopped before updating it")
	}

	market, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx,
		priceFeed.GetMarket().GetBaseAsset(),
		priceFeed.GetMarket().GetQuoteAsset(),
	)
	if err != nil {
		return err
	}

	if market.IsStrategyPluggable() {
		return fmt.Errorf("cant remove pluggable market")
	}

	return s.feederSvc.RemovePriceFeed(ctx, id)
}

func (s *Service) GetPriceFeed(
	ctx context.Context, baseAsset, quoteAsset string,
) (ports.PriceFeed, error) {
	return s.feederSvc.GetPriceFeedForMarket(ctx, baseAsset, quoteAsset)
}

func (s *Service) ListSources(ctx context.Context) []string {
	return s.feederSvc.ListSources(ctx)
}

func (s *Service) ListPriceFeeds(ctx context.Context) ([]ports.PriceFeed, error) {
	return s.feederSvc.ListPriceFeeds(ctx)
}

func (s *Service) listenPriceFeedChan(priceFeedChan chan ports.PriceFeedChan) {
	go func() {
		log.Debugln("reading price feed chan started")

		for priceFeed := range priceFeedChan {
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
}

func (s *Service) startPriceFeed(feederSvc ports.PriceFeeder) error {
	priceFeedChan, err := feederSvc.Start(context.Background())
	if err != nil {
		return err
	}

	log.Debugln("price feed svc started")

	s.listenPriceFeedChan(priceFeedChan)

	return nil
}
