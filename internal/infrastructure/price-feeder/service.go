package pricefeeder

import (
	"context"
	"fmt"
	"sync"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	pricefeeder "github.com/tdex-network/tdex-daemon/pkg/price-feeder"
	bitfinexfeeder "github.com/tdex-network/tdex-daemon/pkg/price-feeder/bitfinex"
	coinbasefeeder "github.com/tdex-network/tdex-daemon/pkg/price-feeder/coinbase"
	krakenfeeder "github.com/tdex-network/tdex-daemon/pkg/price-feeder/kraken"
)

const (
	krakenSource   = "kraken"
	bitfinexSource = "bitfinex"
	coinbaseSource = "coinbase"
)

var (
	feederFactory = map[string]func() (pricefeeder.PriceFeeder, error){
		krakenSource:   krakenfeeder.NewService,
		bitfinexSource: bitfinexfeeder.NewService,
		coinbaseSource: coinbasefeeder.NewService,
	}
)

type service struct {
	store PriceFeedStore

	lock           *sync.Mutex
	sources        map[string]pricefeeder.PriceFeeder
	feedChBySource map[string]chan ports.PriceFeed

	feedsLock   *sync.RWMutex
	activeFeeds map[string]struct{}
}

func NewService(store PriceFeedStore) ports.PriceFeeder {
	return &service{
		store:          store,
		lock:           &sync.Mutex{},
		sources:        make(map[string]pricefeeder.PriceFeeder),
		feedChBySource: make(map[string]chan ports.PriceFeed),
		feedsLock:      &sync.RWMutex{},
		activeFeeds:    make(map[string]struct{}),
	}
}

func (s *service) AddPriceFeed(
	ctx context.Context, market ports.Market, source, ticker string,
) (string, error) {
	priceFeed, err := NewPriceFeedInfo(market, source, ticker)
	if err != nil {
		return "", err
	}

	if err := s.store.AddPriceFeed(ctx, *priceFeed); err != nil {
		return "", err
	}

	return priceFeed.GetId(), nil
}

func (s *service) StartPriceFeed(
	ctx context.Context, id string,
) (chan ports.PriceFeed, error) {
	feed, err := s.store.GetPriceFeed(ctx, id)
	if err != nil {
		return nil, err
	}

	if s.isActiveFeed(id) {
		return nil, fmt.Errorf("price feed already started")
	}

	feederSvc, feedCh, err := s.getPriceFeederBySource(feed.Source)
	if err != nil {
		return nil, err
	}

	if err := feederSvc.SubscribeMarkets(feed.toMarketList()); err != nil {
		return nil, err
	}

	feed.Started = true
	if err := s.store.UpdatePriceFeed(
		ctx, id, func(_ *PriceFeedInfo) (*PriceFeedInfo, error) {
			return feed, nil
		},
	); err != nil {
		return nil, err
	}

	s.addActiveFeed(feed.ID)

	return feedCh, nil
}

func (s *service) StopPriceFeed(ctx context.Context, id string) error {
	feed, err := s.store.GetPriceFeed(ctx, id)
	if err != nil {
		return err
	}

	if !feed.Started {
		return nil
	}

	feed.Started = false
	if err := s.store.UpdatePriceFeed(
		ctx, id, func(_ *PriceFeedInfo) (*PriceFeedInfo, error) {
			return feed, nil
		},
	); err != nil {
		return err
	}

	feederSvc, _, err := s.getPriceFeederBySource(feed.Source)
	if err != nil {
		return err
	}

	if err := feederSvc.UnsubscribeMarkets(feed.toMarketList()); err != nil {
		return err
	}

	// If there are no subscriptions for this price source, let's close the
	// connection.
	if len(feederSvc.ListSubscriptions()) <= 0 {
		s.removePriceFeederBySource(feed.Source)
	}

	s.removeActiveFeed(id)

	return nil
}

func (s *service) UpdatePriceFeed(
	ctx context.Context, id, source, ticker string,
) error {
	if source == "" && ticker == "" {
		return fmt.Errorf("missing price source and/or market ticker")
	}
	if len(source) > 0 {
		if _, ok := feederFactory[source]; !ok {
			return fmt.Errorf("unknown price source")
		}
	}

	return s.store.UpdatePriceFeed(
		ctx, id, func(priceFeed *PriceFeedInfo) (*PriceFeedInfo, error) {
			if len(ticker) > 0 {
				priceFeed.Ticker = ticker
			}
			if len(source) > 0 {
				priceFeed.Source = source
			}
			return priceFeed, nil
		},
	)
}

func (s *service) RemovePriceFeed(ctx context.Context, id string) error {
	return s.store.RemovePriceFeed(ctx, id)
}

func (s *service) GetPriceFeed(
	ctx context.Context, id string,
) (ports.PriceFeedInfo, error) {
	return s.store.GetPriceFeed(ctx, id)
}

func (s *service) ListPriceFeeds(
	ctx context.Context,
) ([]ports.PriceFeedInfo, error) {
	priceFeeds, err := s.store.GetAllPriceFeeds(ctx)
	if err != nil {
		return nil, err
	}

	list := make([]ports.PriceFeedInfo, 0, len(priceFeeds))
	for _, priceFeed := range priceFeeds {
		list = append(list, priceFeed)
	}
	return list, nil
}

func (s *service) ListSources(ctx context.Context) []string {
	supportedSources := make([]string, 0, len(feederFactory))
	for src := range feederFactory {
		supportedSources = append(supportedSources, src)
	}

	return supportedSources
}

func (s *service) Close() {
	s.store.Close()
	for _, ch := range s.feedChBySource {
		close(ch)
	}
}

func (s *service) getPriceFeederBySource(
	source string,
) (pricefeeder.PriceFeeder, chan ports.PriceFeed, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if svc, ok := s.sources[source]; ok {
		ch := s.feedChBySource[source]
		return svc, ch, nil
	}

	svcFactory := feederFactory[source]

	svc, err := svcFactory()
	if err != nil {
		return nil, nil, err
	}

	s.sources[source] = svc
	extCh := svc.Start()
	intCh := make(chan ports.PriceFeed, 20)
	go func(extCh chan pricefeeder.PriceFeed, intCh chan ports.PriceFeed) {
		for feed := range extCh {
			intCh <- priceFeedInfo(feed)
		}
	}(extCh, intCh)
	s.feedChBySource[source] = intCh
	return svc, intCh, nil
}

func (s *service) removePriceFeederBySource(source string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	svc, ok := s.sources[source]
	if !ok {
		return
	}

	svc.Stop()
	close(s.feedChBySource[source])

	delete(s.sources, source)
	delete(s.feedChBySource, source)
}

func (s *service) addActiveFeed(id string) {
	s.feedsLock.Lock()
	defer s.feedsLock.Unlock()

	s.activeFeeds[id] = struct{}{}
}

func (s *service) removeActiveFeed(id string) {
	s.feedsLock.Lock()
	defer s.feedsLock.Unlock()

	delete(s.activeFeeds, id)
}

func (s *service) isActiveFeed(id string) bool {
	s.feedsLock.RLock()
	defer s.feedsLock.RUnlock()

	_, ok := s.activeFeeds[id]
	return ok
}
