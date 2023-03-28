package feeder

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

type Service struct {
	repoManager ports.RepoManager
	feederSvc   ports.PriceFeeder
}

func NewService(
	repoManager ports.RepoManager,
) Service {
	return Service{repoManager: repoManager}
}

func (s *Service) Start(ctx context.Context) error {
	markets, err := s.repoManager.MarketRepository().GetAllMarkets(
		context.Background(),
	)
	if err != nil {
		return err
	}

	mkts := make([]ports.Market, len(markets))
	for _, v := range markets {
		mkts = append(mkts, marketInfo{
			baseAsset:  v.BaseAsset,
			quoteAsset: v.QuoteAsset,
		})
	}

	priceFeedChan, err := s.feederSvc.Start(ctx, mkts)
	if err != nil {
		return err
	}

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

	return nil
}

func (s *Service) Stop(ctx context.Context) {
	s.feederSvc.Stop(ctx)
}

func (s *Service) StartFeed(ctx context.Context, feedID string) error {
	priceFeed, err := s.feederSvc.GetPriceFeedForFeedID(ctx, feedID)
	if err != nil {
		return err
	}

	market, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx,
		priceFeed.MarketBaseAsset,
		priceFeed.MarketQuoteAsset,
	)
	if err != nil {
		return err
	}

	if !market.IsStrategyPluggable() {
		return ErrMarketNotPluggable
	}

	return s.feederSvc.StartFeed(ctx, feedID)
}

func (s *Service) StopFeed(ctx context.Context, feedID string) error {
	return s.feederSvc.StopFeed(ctx, feedID)
}

func (s *Service) AddPriceFeed(
	ctx context.Context,
	req ports.AddPriceFeedReq,
) (string, error) {
	return s.feederSvc.AddPriceFeed(ctx, req)
}

func (s *Service) UpdatePriceFeed(
	ctx context.Context,
	req ports.UpdatePriceFeedReq,
) error {
	priceFeed, err := s.feederSvc.GetPriceFeedForFeedID(ctx, req.ID)
	if err != nil {
		return err
	}

	if priceFeed.On {
		return ErrFeedOn
	}

	return s.feederSvc.UpdatePriceFeed(ctx, req)
}

func (s *Service) RemovePriceFeed(
	ctx context.Context,
	feedID string,
) error {
	priceFeed, err := s.feederSvc.GetPriceFeedForFeedID(ctx, feedID)
	if err != nil {
		return err
	}

	if priceFeed.On {
		return ErrFeedOn
	}

	market, err := s.repoManager.MarketRepository().GetMarketByAssets(
		ctx,
		priceFeed.MarketBaseAsset,
		priceFeed.MarketQuoteAsset,
	)
	if err != nil {
		return err
	}

	if market.IsStrategyPluggable() {
		return ErrMarketPluggable
	}

	return s.feederSvc.RemovePriceFeed(ctx, feedID)
}

func (s *Service) GetPriceFeed(
	ctx context.Context,
	baseAsset string,
	quoteAsset string,
) (*ports.PriceFeed, error) {
	return s.feederSvc.GetPriceFeed(ctx, baseAsset, quoteAsset)
}

func (s *Service) ListSources(ctx context.Context) []string {
	return s.feederSvc.ListSources(ctx)
}
