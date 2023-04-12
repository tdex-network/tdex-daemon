package feeder

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

type Service struct {
	repoManager ports.RepoManager
	feederSvc   ports.PriceFeeder
}

func NewService(
	feederSvc ports.PriceFeeder, repoManager ports.RepoManager,
) (*Service, error) {
	svc := &Service{
		repoManager: repoManager,
		feederSvc:   feederSvc,
	}

	if err := svc.startPriceFeeds(); err != nil {
		return nil, err
	}

	return svc, nil
}

func (s *Service) AddPriceFeed(
	ctx context.Context, market ports.Market, source, ticker string,
) (string, error) {
	return s.feederSvc.AddPriceFeed(ctx, market, source, ticker)
}

func (s *Service) StartPriceFeed(ctx context.Context, id string) error {
	priceFeed, err := s.feederSvc.GetPriceFeed(ctx, id)
	if err != nil {
		return err
	}

	market, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx, priceFeed.GetMarket().GetBaseAsset(),
		priceFeed.GetMarket().GetQuoteAsset(),
	)
	if err != nil {
		return err
	}

	if !market.IsStrategyPluggable() {
		return fmt.Errorf("market strategy must be pluggable")
	}

	ch, err := s.feederSvc.StartPriceFeed(ctx, id)
	if err != nil {
		return err
	}

	go func(priceFeed ports.PriceFeedInfo) {
		log.Debugf(
			"start listening to price feeds from %s for market %s",
			priceFeed.GetSource(), priceFeed.GetTicker(),
		)

		ctx := context.Background()
		for feed := range ch {
			market, _ := s.repoManager.MarketRepository().GetMarketByAssets(
				ctx, feed.GetMarket().GetBaseAsset(), feed.GetMarket().GetQuoteAsset(),
			)
			if market == nil {
				continue
			}

			log.Debugf(
				"received price feed from %s for market %s, updating market price...",
				priceFeed.GetSource(), priceFeed.GetTicker(),
			)

			if err := s.repoManager.MarketRepository().UpdateMarketPrice(
				ctx, market.Name, marketPrice{feed.GetPrice()}.toDomain(),
			); err != nil {
				log.WithError(err).Warnf(
					"failed to update price for market %s", priceFeed.GetTicker(),
				)
			}
		}

		log.Debugf(
			"stop listening to price feeds from %s for market %s",
			priceFeed.GetSource(), priceFeed.GetTicker(),
		)
	}(priceFeed)

	return nil
}

func (s *Service) StopPriceFeed(ctx context.Context, id string) error {
	return s.feederSvc.StopPriceFeed(ctx, id)
}

func (s *Service) UpdatePriceFeed(
	ctx context.Context, id, source, ticker string,
) error {
	priceFeed, err := s.feederSvc.GetPriceFeed(ctx, id)
	if err != nil {
		return err
	}

	if priceFeed.IsStarted() {
		return fmt.Errorf("price feed must be stopped to be updated")
	}

	return s.feederSvc.UpdatePriceFeed(ctx, id, source, ticker)
}

func (s *Service) RemovePriceFeed(ctx context.Context, id string) error {
	priceFeed, err := s.feederSvc.GetPriceFeed(ctx, id)
	if err != nil {
		return err
	}

	if priceFeed.IsStarted() {
		return fmt.Errorf("price feed must be stopped to be removed")
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
		return fmt.Errorf(
			"market strategy must be changed from pluggable to remove a price feed",
		)
	}

	return s.feederSvc.RemovePriceFeed(ctx, id)
}

func (s *Service) GetPriceFeed(
	ctx context.Context, id string,
) (ports.PriceFeedInfo, error) {
	return s.feederSvc.GetPriceFeed(ctx, id)
}

func (s *Service) ListSources(ctx context.Context) []string {
	return s.feederSvc.ListSources(ctx)
}

func (s *Service) ListPriceFeeds(
	ctx context.Context,
) ([]ports.PriceFeedInfo, error) {
	return s.feederSvc.ListPriceFeeds(ctx)
}

func (s *Service) Close() {
	s.feederSvc.Close()
}

func (s *Service) startPriceFeeds() error {
	feeds, err := s.feederSvc.ListPriceFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		if feed.IsStarted() {
			if err := s.StartPriceFeed(
				context.Background(), feed.GetId(),
			); err != nil {
				return err
			}
		}
	}
	return nil
}
