package pricefeederinfra

import (
	"context"
	"sync"

	pricefeeder "github.com/tdex-network/tdex-daemon/pkg/price-feeder"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

var (
	sources = map[string]struct{}{
		"kraken": {},
	}
)

type priceFeederService struct {
	feederSvc           pricefeeder.PriceFeeder
	priceFeedRepository PriceFeedRepository
	feedChan            chan ports.PriceFeedChan

	marketsMtx sync.RWMutex
	markets    []pricefeeder.Market
}

func NewService(
	feederSvc pricefeeder.PriceFeeder,
	priceFeedRepository PriceFeedRepository,
) ports.PriceFeeder {
	return &priceFeederService{
		feederSvc:           feederSvc,
		priceFeedRepository: priceFeedRepository,
		feedChan:            make(chan ports.PriceFeedChan),
		marketsMtx:          sync.RWMutex{},
		markets:             make([]pricefeeder.Market, 0),
	}
}

func (p *priceFeederService) Start(
	ctx context.Context,
	markets []ports.Market,
) (chan ports.PriceFeedChan, error) {
	mkts := make([]pricefeeder.Market, len(markets))
	for _, v := range markets {
		priceFeed, err := p.priceFeedRepository.GetPriceFeedsByMarket(ctx, Market{
			BaseAsset:  v.GetBaseAsset(),
			QuoteAsset: v.GetQuoteAsset(),
		})
		if err != nil {
			return nil, err
		}

		mkts = append(mkts, pricefeeder.Market{
			BaseAsset:  v.GetBaseAsset(),
			QuoteAsset: v.GetQuoteAsset(),
			Ticker:     priceFeed.Market.Ticker,
		})
	}

	if err := p.watchMarkets(mkts); err != nil {
		return nil, err
	}

	return p.feedChan, nil
}

func (p *priceFeederService) watchMarkets(
	markets []pricefeeder.Market,
) error {
	if err := p.feederSvc.SubscribeMarkets(markets); err != nil {
		return err
	}

	if err := p.feederSvc.Start(); err != nil {
		return err
	}

	go func() {
		for priceFeed := range p.feederSvc.FeedChan() {
			p.feedChan <- Feed{
				Market: Market{
					BaseAsset:  priceFeed.Market.BaseAsset,
					QuoteAsset: priceFeed.Market.QuoteAsset,
				},
				Price: Price{
					BasePrice:  priceFeed.Price.BasePrice,
					QuotePrice: priceFeed.Price.QuotePrice,
				},
			}
		}
	}()

	return nil
}

func (p *priceFeederService) Stop(ctx context.Context) {
	p.feederSvc.Stop()
	close(p.feedChan)
}

func (p *priceFeederService) StartFeed(ctx context.Context, feedID string) error {
	p.feederSvc.Stop()

	priceFeed, err := p.priceFeedRepository.GetPriceFeed(ctx, feedID)
	if err != nil {
		return err
	}

	p.addMarkets([]pricefeeder.Market{
		{
			BaseAsset:  priceFeed.Market.BaseAsset,
			QuoteAsset: priceFeed.Market.QuoteAsset,
			Ticker:     priceFeed.Market.Ticker,
		},
	})

	return p.watchMarkets(p.getMarkets())
}

func (p *priceFeederService) StopFeed(ctx context.Context, feedID string) error {
	p.feederSvc.Stop()

	priceFeed, err := p.priceFeedRepository.GetPriceFeed(ctx, feedID)
	if err != nil {
		return err
	}

	p.removeMarkets([]pricefeeder.Market{
		{
			BaseAsset:  priceFeed.Market.BaseAsset,
			QuoteAsset: priceFeed.Market.QuoteAsset,
			Ticker:     priceFeed.Market.Ticker,
		},
	})

	return p.watchMarkets(p.getMarkets())
}

func (p *priceFeederService) AddPriceFeed(
	ctx context.Context,
	req ports.AddPriceFeedReq,
) (string, error) {
	addPriceRequest := AddPriceFeedReq(req)
	if err := addPriceRequest.Validate(); err != nil {
		return "", err
	}

	market := Market{
		BaseAsset:  req.MarketBaseAsset,
		QuoteAsset: req.MarketQuoteAsset,
		Ticker:     req.Ticker,
	}
	priceFeed := NewPriceFeed(market, req.Source)
	if err := p.priceFeedRepository.AddPriceFeed(ctx, priceFeed); err != nil {
		return "", err
	}

	return priceFeed.ID, nil
}

func (p *priceFeederService) UpdatePriceFeed(
	ctx context.Context,
	req ports.UpdatePriceFeedReq,
) error {
	updatePriceRequest := UpdatePriceFeedReq(req)
	if err := updatePriceRequest.Validate(); err != nil {
		return err
	}

	return p.priceFeedRepository.UpdatePriceFeed(
		ctx,
		req.ID, func(priceFeed *PriceFeed) (*PriceFeed, error) {
			priceFeed.Market.Ticker = req.Ticker
			priceFeed.Source = req.Source

			return priceFeed, nil
		},
	)
}

func (p *priceFeederService) RemovePriceFeed(
	ctx context.Context,
	feedID string,
) error {
	return p.priceFeedRepository.RemovePriceFeed(ctx, feedID)
}

func (p *priceFeederService) GetPriceFeed(
	ctx context.Context,
	baseAsset string,
	quoteAsset string,
) (*ports.PriceFeed, error) {
	priceFeed, err := p.priceFeedRepository.GetPriceFeedsByMarket(ctx, Market{
		BaseAsset:  baseAsset,
		QuoteAsset: quoteAsset,
	})
	if err != nil {
		return nil, err
	}

	return &ports.PriceFeed{
		ID:               priceFeed.ID,
		MarketBaseAsset:  priceFeed.Market.BaseAsset,
		MarketQuoteAsset: priceFeed.Market.QuoteAsset,
		Source:           priceFeed.Source,
		Ticker:           priceFeed.Market.Ticker,
		On:               priceFeed.On,
	}, nil
}

func (p *priceFeederService) GetPriceFeedForFeedID(
	ctx context.Context,
	feedID string,
) (*ports.PriceFeed, error) {
	priceFeed, err := p.priceFeedRepository.GetPriceFeed(ctx, feedID)
	if err != nil {
		return nil, err
	}

	return &ports.PriceFeed{
		ID:               priceFeed.ID,
		MarketBaseAsset:  priceFeed.Market.BaseAsset,
		MarketQuoteAsset: priceFeed.Market.QuoteAsset,
		Source:           priceFeed.Source,
		Ticker:           priceFeed.Market.Ticker,
		On:               priceFeed.On,
	}, nil
}

func (p *priceFeederService) ListSources(ctx context.Context) []string {
	supportedSources := make([]string, 0, len(sources))
	for src := range sources {
		supportedSources = append(supportedSources, src)
	}

	return supportedSources
}

func (p *priceFeederService) addMarkets(markets []pricefeeder.Market) {
	p.marketsMtx.Lock()
	defer p.marketsMtx.Unlock()

	p.markets = append(p.markets, markets...)
}

func (p *priceFeederService) removeMarkets(markets []pricefeeder.Market) {
	p.marketsMtx.Lock()
	defer p.marketsMtx.Unlock()

	for _, m := range markets {
		for i, v := range p.markets {
			if v.BaseAsset == m.BaseAsset && v.QuoteAsset == m.QuoteAsset {
				p.markets = append(p.markets[:i], p.markets[i+1:]...)
			}
		}
	}
}

func (p *priceFeederService) getMarkets() []pricefeeder.Market {
	p.marketsMtx.RLock()
	defer p.marketsMtx.RUnlock()

	return p.markets
}
