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
	feederSvc      pricefeeder.PriceFeeder
	priceFeedStore PriceFeedStore
	feedChan       chan ports.PriceFeedChan

	marketsMtx sync.RWMutex
	markets    []pricefeeder.Market
}

func NewService(
	feederSvc pricefeeder.PriceFeeder,
	priceFeedStore PriceFeedStore,
) ports.PriceFeeder {
	return &priceFeederService{
		feederSvc:      feederSvc,
		priceFeedStore: priceFeedStore,
		feedChan:       make(chan ports.PriceFeedChan),
		marketsMtx:     sync.RWMutex{},
		markets:        make([]pricefeeder.Market, 0),
	}
}

func (p *priceFeederService) Start(
	ctx context.Context,
	markets []ports.Market,
) (chan ports.PriceFeedChan, error) {
	mkts := make([]pricefeeder.Market, len(markets))
	for _, v := range markets {
		priceFeed, err := p.priceFeedStore.GetPriceFeedsByMarket(ctx, Market{
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

	if err := p.feederSvc.Start(); err != nil {
		return nil, err
	}

	if err := p.feederSvc.SubscribeMarkets(mkts); err != nil {
		return nil, err
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

	return p.feedChan, nil
}

func (p *priceFeederService) Stop(ctx context.Context) {
	p.feederSvc.Stop()
	close(p.feedChan)
}

func (p *priceFeederService) StartFeed(ctx context.Context, feedID string) error {
	priceFeed, err := p.priceFeedStore.GetPriceFeed(ctx, feedID)
	if err != nil {
		return err
	}

	market := pricefeeder.Market{
		BaseAsset:  priceFeed.Market.BaseAsset,
		QuoteAsset: priceFeed.Market.QuoteAsset,
		Ticker:     priceFeed.Market.Ticker,
	}

	p.addMarkets([]pricefeeder.Market{market})

	return p.feederSvc.SubscribeMarkets([]pricefeeder.Market{market})
}

func (p *priceFeederService) StopFeed(ctx context.Context, feedID string) error {
	priceFeed, err := p.priceFeedStore.GetPriceFeed(ctx, feedID)
	if err != nil {
		return err
	}

	market := pricefeeder.Market{
		BaseAsset:  priceFeed.Market.BaseAsset,
		QuoteAsset: priceFeed.Market.QuoteAsset,
		Ticker:     priceFeed.Market.Ticker,
	}

	p.removeMarkets([]pricefeeder.Market{market})

	return p.feederSvc.UnSubscribeMarkets([]pricefeeder.Market{market})
}

func (p *priceFeederService) AddPriceFeed(
	ctx context.Context, market ports.Market, source, ticker string,
) (string, error) {
	if err := validateAddPriceFeed(market, source, ticker); err != nil {
		return "", err
	}

	priceFeed := NewPriceFeed(
		Market{
			BaseAsset:  market.GetBaseAsset(),
			QuoteAsset: market.GetQuoteAsset(),
			Ticker:     ticker,
		},
		source,
	)
	if err := p.priceFeedStore.AddPriceFeed(ctx, priceFeed); err != nil {
		return "", err
	}

	return priceFeed.ID, nil
}

func (p *priceFeederService) UpdatePriceFeed(
	ctx context.Context, id, source, ticker string,
) error {
	if err := ValidateUpdatePriceFeed(id, source, ticker); err != nil {
		return err
	}

	return p.priceFeedStore.UpdatePriceFeed(
		ctx,
		id,
		func(priceFeed *PriceFeed) (*PriceFeed, error) {
			priceFeed.Market.Ticker = ticker
			priceFeed.Source = source

			return priceFeed, nil
		},
	)
}

func (p *priceFeederService) RemovePriceFeed(
	ctx context.Context,
	feedID string,
) error {
	return p.priceFeedStore.RemovePriceFeed(ctx, feedID)
}

func (p *priceFeederService) GetPriceFeedForMarket(
	ctx context.Context,
	baseAsset string,
	quoteAsset string,
) (ports.PriceFeed, error) {
	priceFeed, err := p.priceFeedStore.GetPriceFeedsByMarket(ctx, Market{
		BaseAsset:  baseAsset,
		QuoteAsset: quoteAsset,
	})
	if err != nil {
		return nil, err
	}

	return priceFeed, nil
}

func (p *priceFeederService) GetPriceFeed(
	ctx context.Context,
	feedID string,
) (ports.PriceFeed, error) {
	priceFeed, err := p.priceFeedStore.GetPriceFeed(ctx, feedID)
	if err != nil {
		return nil, err
	}

	return priceFeed, nil
}

func (p *priceFeederService) ListPriceFeeds(
	ctx context.Context,
) ([]ports.PriceFeed, error) {
	priceFeeds, err := p.priceFeedStore.GetAllPriceFeeds(ctx)
	if err != nil {
		return nil, err
	}

	priceFeedsResponse := make([]ports.PriceFeed, 0, len(priceFeeds))
	for _, priceFeed := range priceFeeds {
		priceFeedsResponse = append(priceFeedsResponse, priceFeed)
	}

	return priceFeedsResponse, nil
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
