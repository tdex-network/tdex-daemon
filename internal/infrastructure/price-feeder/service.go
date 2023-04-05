package pricefeederinfra

import (
	"context"
	"sync"

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
	feedChanMtx       *sync.Mutex
	feedChan          chan ports.PriceFeedChan

	feederSvcBySourceStartedMtx sync.RWMutex
	feederSvcBySourceStarted    map[string]bool
}

func NewService(
	feederSvcBySource map[string]pricefeeder.PriceFeeder,
	priceFeedStore PriceFeedStore,
) ports.PriceFeeder {
	return &priceFeederService{
		feederSvcBySource:           feederSvcBySource,
		priceFeedStore:              priceFeedStore,
		feedChanMtx:                 &sync.Mutex{},
		feedChan:                    make(chan ports.PriceFeedChan),
		feederSvcBySourceStartedMtx: sync.RWMutex{},
		feederSvcBySourceStarted:    make(map[string]bool),
	}
}

func (p *priceFeederService) Start(
	ctx context.Context,
) (chan ports.PriceFeedChan, error) {
	if err := p.start(ctx); err != nil {
		return nil, err
	}

	return p.feedChan, nil
}

func (p *priceFeederService) start(ctx context.Context) error {
	mktsBySource := make(map[string][]pricefeeder.Market)
	priceFeeds, err := p.priceFeedStore.GetStartedPriceFeeds(ctx)
	if err != nil {
		return err
	}

	for _, v := range priceFeeds {
		if vv, ok := mktsBySource[v.Source]; ok {
			mktsBySource[v.Source] = append(vv, pricefeeder.Market{
				BaseAsset:  v.Market.GetBaseAsset(),
				QuoteAsset: v.Market.GetQuoteAsset(),
				Ticker:     v.Market.Ticker,
			})
		} else {
			mktsBySource[v.Source] = append(
				[]pricefeeder.Market{},
				pricefeeder.Market{
					BaseAsset:  v.Market.GetBaseAsset(),
					QuoteAsset: v.Market.GetQuoteAsset(),
					Ticker:     v.Market.Ticker,
				},
			)
		}
	}

	if len(mktsBySource) == 0 {
		return nil
	}

	for source := range mktsBySource {
		go func(s string, feeder pricefeeder.PriceFeeder) {
			p.SetPriceFeederStarted(s)
			if err := feeder.Start(); err != nil {
				log.Warnf("error while starting price feeder: %v", err) //TODO handle error
			}

		}(source, p.feederSvcBySource[source])
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
				p.sendPriceFeed(priceFeed)
			}
		}(v)
	}

	log.Debugln("price feeder service started")

	return nil
}

func (p *priceFeederService) Stop(ctx context.Context) {
	for _, v := range p.feederSvcBySource {
		v.Stop()
	}

	p.closePriceFeedChan()
}

func (p *priceFeederService) StartFeed(ctx context.Context, feedID string) error {
	priceFeed, err := p.priceFeedStore.GetPriceFeed(ctx, feedID)
	if err != nil {
		return err
	}

	if priceFeed.Started {
		return nil
	}

	market := pricefeeder.Market{
		BaseAsset:  priceFeed.Market.BaseAsset,
		QuoteAsset: priceFeed.Market.QuoteAsset,
		Ticker:     priceFeed.Market.Ticker,
	}

	if err := p.priceFeedStore.UpdatePriceFeed(ctx, feedID,
		func(priceFeed *PriceFeed) (*PriceFeed, error) {
			priceFeed.Started = true
			return priceFeed, nil
		},
	); err != nil {
		return err
	}

	if !p.IsPriceFeederStarted(priceFeed.Source) {
		if err := p.start(ctx); err != nil {
			return err
		}

		return nil
	}

	return p.getFeederServiceBySource(priceFeed.GetSource()).
		SubscribeMarkets([]pricefeeder.Market{market})
}

func (p *priceFeederService) StopFeed(ctx context.Context, feedID string) error {
	priceFeed, err := p.priceFeedStore.GetPriceFeed(ctx, feedID)
	if err != nil {
		return err
	}

	if !priceFeed.Started {
		return nil
	}

	market := pricefeeder.Market{
		BaseAsset:  priceFeed.Market.BaseAsset,
		QuoteAsset: priceFeed.Market.QuoteAsset,
		Ticker:     priceFeed.Market.Ticker,
	}

	if err := p.priceFeedStore.UpdatePriceFeed(ctx, feedID,
		func(priceFeed *PriceFeed) (*PriceFeed, error) {
			priceFeed.Started = false
			return priceFeed, nil
		},
	); err != nil {
		return err
	}

	p.SetPriceFeederStopped(priceFeed.Source)

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

func (p *priceFeederService) closePriceFeedChan() {
	p.feedChanMtx.Lock()
	defer p.feedChanMtx.Unlock()

	close(p.feedChan)
}

func (p *priceFeederService) sendPriceFeed(
	priceFeed pricefeeder.PriceFeed,
) {
	p.feedChanMtx.Lock()
	defer p.feedChanMtx.Unlock()

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

func (p *priceFeederService) IsPriceFeederStarted(source string) bool {
	p.feederSvcBySourceStartedMtx.RLock()
	defer p.feederSvcBySourceStartedMtx.RUnlock()

	return p.feederSvcBySourceStarted[source]
}

func (p *priceFeederService) SetPriceFeederStarted(source string) {
	p.feederSvcBySourceStartedMtx.Lock()
	defer p.feederSvcBySourceStartedMtx.Unlock()

	p.feederSvcBySourceStarted[source] = true
}

func (p *priceFeederService) SetPriceFeederStopped(source string) {
	p.feederSvcBySourceStartedMtx.Lock()
	defer p.feederSvcBySourceStartedMtx.Unlock()

	p.feederSvcBySourceStarted[source] = false
}
