package pricefeederinfra

import (
	"context"

	log "github.com/sirupsen/logrus"

	pricefeeder "github.com/tdex-network/tdex-daemon/pkg/price-feeder"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

var (
	sources = map[string]struct{}{
		"kraken":   {},
		"bitfinex": {},
		"coinbase": {},
	}
)

type priceFeederService struct {
	feederSvcBySource map[string]pricefeeder.PriceFeeder
	priceFeedStore    PriceFeedStore
	feedChan          chan ports.PriceFeedChan
}

func NewService(
	feederSvcBySource map[string]pricefeeder.PriceFeeder,
	priceFeedStore PriceFeedStore,
) ports.PriceFeeder {
	return &priceFeederService{
		feederSvcBySource: feederSvcBySource,
		priceFeedStore:    priceFeedStore,
		feedChan:          make(chan ports.PriceFeedChan),
	}
}

func (p *priceFeederService) Start(
	ctx context.Context,
	markets []ports.Market,
) (chan ports.PriceFeedChan, error) {
	mktsBySource := make(map[string][]pricefeeder.Market)
	for _, v := range markets {
		priceFeed, err := p.priceFeedStore.GetPriceFeedsByMarket(ctx, Market{
			BaseAsset:  v.GetBaseAsset(),
			QuoteAsset: v.GetQuoteAsset(),
		})
		if err != nil {
			return nil, err
		}

		if priceFeed != nil {
			if vv, ok := mktsBySource[priceFeed.Source]; ok {
				mktsBySource[priceFeed.Source] = append(vv, pricefeeder.Market{
					BaseAsset:  v.GetBaseAsset(),
					QuoteAsset: v.GetQuoteAsset(),
					Ticker:     priceFeed.Market.Ticker,
				})
			} else {
				mktsBySource[priceFeed.Source] = append(
					[]pricefeeder.Market{},
					pricefeeder.Market{
						BaseAsset:  v.GetBaseAsset(),
						QuoteAsset: v.GetQuoteAsset(),
						Ticker:     priceFeed.Market.Ticker,
					},
				)
			}
		}
	}

	for _, v := range p.feederSvcBySource {
		go func(feeder pricefeeder.PriceFeeder) {
			if err := feeder.Start(); err != nil {
				log.Warnf("error while starting price feeder: %v", err) //TODO handle error
			}
		}(v)
	}

	for k, v := range mktsBySource {
		go func(source string, markets []pricefeeder.Market) {
			if err := p.feederSvcBySource[source].SubscribeMarkets(markets); err != nil {
				log.Warnf("error while starting price feeder: %v", err) //TODO handle error
			}
		}(k, v)
	}

	feedChanBySource := make(map[string]chan pricefeeder.PriceFeed)
	for k, v := range p.feederSvcBySource {
		feedChanBySource[k] = v.FeedChan()
	}

	for _, v := range feedChanBySource {
		go func(feedChan chan pricefeeder.PriceFeed) {
			for priceFeed := range feedChan {
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
		}(v)
	}

	return p.feedChan, nil
}

func (p *priceFeederService) Stop(ctx context.Context) {
	for _, v := range p.feederSvcBySource {
		v.Stop()
	}

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

	return p.getFeederServiceBySource(priceFeed.GetSource()).
		SubscribeMarkets([]pricefeeder.Market{market})
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

	return p.getFeederServiceBySource(priceFeed.GetSource()).
		UnSubscribeMarkets([]pricefeeder.Market{market})
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

func (p *priceFeederService) getFeederServiceBySource(
	source string) pricefeeder.PriceFeeder {
	v, _ := p.feederSvcBySource[source]

	return v
}
