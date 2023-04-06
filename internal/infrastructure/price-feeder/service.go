package pricefeederinfra

import (
	"context"
	"fmt"
	"sync"

	coinbasefeeder "github.com/tdex-network/tdex-daemon/pkg/price-feeder/coinbase"
	krakenfeeder "github.com/tdex-network/tdex-daemon/pkg/price-feeder/kraken"

	bitfinexfeeder "github.com/tdex-network/tdex-daemon/pkg/price-feeder/bitfinex"

	log "github.com/sirupsen/logrus"

	pricefeeder "github.com/tdex-network/tdex-daemon/pkg/price-feeder"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"
)

const (
	krakenSource   = "kraken"
	bitfinexSource = "bitfinex"
	coinbaseSource = "coinbase"

	priceFeedInterval = 2000
)

var (
	sources = map[string]struct{}{
		krakenSource:   {},
		bitfinexSource: {},
		coinbaseSource: {},
	}
)

type priceFeederService struct {
	priceFeedStore PriceFeedStore

	feederSvcBySourceMtx *sync.Mutex
	feederSvcBySource    map[string]feederSvcInfo

	feedChanMtx *sync.Mutex
	feedChan    chan ports.PriceFeedChan
}

func NewService(
	priceFeedStore PriceFeedStore,
) ports.PriceFeeder {
	return &priceFeederService{
		feederSvcBySourceMtx: &sync.Mutex{},
		feederSvcBySource:    make(map[string]feederSvcInfo),
		priceFeedStore:       priceFeedStore,
		feedChanMtx:          &sync.Mutex{},
		feedChan:             make(chan ports.PriceFeedChan),
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
	mktsBySource, err := p.getMarketsBySource(ctx)
	if err != nil {
		return err
	}

	if len(mktsBySource) == 0 {
		return nil
	}

	for k, v := range mktsBySource {
		svc, err := p.getFeederSvcBySource(k)
		if err != nil {
			return err
		}

		svc.start(v)

		go func() {
			for priceFeed := range svc.feederSvc.FeedChan() {
				p.sendPriceFeed(priceFeed)
			}
		}()
	}

	log.Debugln("price feeder service started")

	return nil
}

func (p *priceFeederService) getMarketsBySource(ctx context.Context,
) (map[string][]pricefeeder.Market, error) {
	mktsBySource := make(map[string][]pricefeeder.Market)
	priceFeeds, err := p.priceFeedStore.GetStartedPriceFeeds(ctx)
	if err != nil {
		return nil, err
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

	return mktsBySource, nil
}

func (p *priceFeederService) Stop(ctx context.Context) {
	for _, v := range p.getAllFeederSvc() {
		v.feederSvc.Stop()
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

	feederSvc, err := p.getFeederSvcBySource(priceFeed.Source)
	if err != nil {
		return err
	}

	if !feederSvc.isStarted() {
		if err := p.start(ctx); err != nil {
			return err
		}

		return nil
	}

	return feederSvc.subscribe([]pricefeeder.Market{market})
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

	feederSvc, err := p.getFeederSvcBySource(priceFeed.Source)
	if err != nil {
		return err
	}

	shouldStop, err := feederSvc.unSubscribe(market)
	if err != nil {
		return err
	}

	if shouldStop {
		feederSvc.stop()
		p.deleteFeederSvcBySource(priceFeed.Source)
	}

	return nil
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

type feederSvcInfo struct {
	feederSvc         pricefeeder.PriceFeeder
	subscribedMarkets []pricefeeder.Market
	started           bool
}

func (f *feederSvcInfo) start(mkts []pricefeeder.Market) {
	if f.started {
		return
	}

	go func() {
		if err := f.feederSvc.Start(); err != nil {
			log.Warnf("error while starting price feeder: %v", err) //TODO handle error
		}
	}()

	go func() {
		if err := f.subscribe(mkts); err != nil {
			log.Warnf("error while subscribing to markets: %v", err) //TODO handle error
		}
	}()

	f.started = true
}

func (f *feederSvcInfo) stop() {
	f.feederSvc.Stop()
}

func (f *feederSvcInfo) isStarted() bool {
	return f.started
}

func (f *feederSvcInfo) subscribe(mkts []pricefeeder.Market) error {
	if err := f.feederSvc.SubscribeMarkets(mkts); err != nil {
		return err
	}

	for _, mkt := range mkts {
		f.subscribedMarkets = append(f.subscribedMarkets, mkt)
	}

	return nil
}

func (f *feederSvcInfo) unSubscribe(mkt pricefeeder.Market) (bool, error) {
	var shouldStop bool
	for i, v := range f.subscribedMarkets {
		if v.Ticker == mkt.Ticker && v.BaseAsset == mkt.BaseAsset &&
			v.QuoteAsset == mkt.QuoteAsset {
			f.subscribedMarkets = append(f.subscribedMarkets[:i],
				f.subscribedMarkets[i+1:]...)
		}
	}

	if len(f.subscribedMarkets) == 0 {
		shouldStop = true
		return shouldStop, nil
	}

	return shouldStop, f.feederSvc.UnSubscribeMarkets([]pricefeeder.Market{mkt})
}

func (p *priceFeederService) deleteFeederSvcBySource(source string) {
	p.feederSvcBySourceMtx.Lock()
	defer p.feederSvcBySourceMtx.Unlock()

	delete(p.feederSvcBySource, source)
}

func (p *priceFeederService) getFeederSvcBySource(
	source string) (feederSvcInfo, error) {
	p.feederSvcBySourceMtx.Lock()
	defer p.feederSvcBySourceMtx.Unlock()

	svc, ok := p.feederSvcBySource[source]
	if !ok {
		svc, err := feederSvcInfoBySourceFactory(source)
		if err != nil {
			return feederSvcInfo{}, err
		}

		p.feederSvcBySource[source] = svc

		return svc, nil
	}

	return svc, nil
}

func (p *priceFeederService) getAllFeederSvc() []feederSvcInfo {
	p.feederSvcBySourceMtx.Lock()
	defer p.feederSvcBySourceMtx.Unlock()

	svcs := make([]feederSvcInfo, 0, len(p.feederSvcBySource))
	for _, svc := range p.feederSvcBySource {
		svcs = append(svcs, svc)
	}

	return svcs
}

func feederSvcInfoBySourceFactory(source string) (feederSvcInfo, error) {
	var (
		svc pricefeeder.PriceFeeder
		err error
	)

	switch source {
	case krakenSource:
		svc, err = bitfinexfeeder.NewService(priceFeedInterval)
	case coinbaseSource:
		svc, err = coinbasefeeder.NewService(priceFeedInterval)
	case bitfinexSource:
		svc, err = krakenfeeder.NewService(priceFeedInterval)
	default:
		return feederSvcInfo{}, fmt.Errorf("unsupported price feeder source: %s", source)
	}

	if err != nil {
		return feederSvcInfo{}, err
	}

	return feederSvcInfo{
		feederSvc:         svc,
		subscribedMarkets: make([]pricefeeder.Market, 0),
		started:           false,
	}, nil
}
