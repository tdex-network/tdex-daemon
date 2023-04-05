package coinbasefeeder

import (
	"fmt"
	"sync"
	"time"

	pricefeeder "github.com/tdex-network/tdex-daemon/pkg/price-feeder"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

const (
	// CoinbaseWebSocketURL is the base url to open a WebSocket connection with
	// Coinbase.
	CoinbaseWebSocketURL = "ws-feed.exchange.coinbase.com"
)

var (
	wellKnownMarkets = []pricefeeder.Market{
		{
			BaseAsset:  "6f0279e9ed041c3d710a9f57d0c02928416460c4b722ae3457a11eec381c526d",
			QuoteAsset: "ce091c998b83c78bb71a632313ba3760f1763d9cfcffae02258ffa9865a37bd2",
			Ticker:     "BTC-USDT",
		},
	}
)

type service struct {
	connMtx *sync.Mutex
	conn    *websocket.Conn

	writeTicker *time.Ticker

	marketByTickerMtx *sync.RWMutex
	marketByTicker    map[string]pricefeeder.Market

	latestFeedsByTickerMtx *sync.RWMutex
	latestFeedsByTicker    map[string]pricefeeder.PriceFeed

	chLock   *sync.Mutex
	feedChan chan pricefeeder.PriceFeed

	quitChan chan struct{}
}

func NewService(args ...interface{}) (pricefeeder.PriceFeeder, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("invalid number of args")
	}

	interval, ok := args[0].(int)
	if !ok {
		return nil, fmt.Errorf("unknown interval arg type")
	}
	writeTicker := time.NewTicker(time.Duration(interval) * time.Millisecond)

	conn, err := connect()
	if err != nil {
		return nil, err
	}

	return &service{
		writeTicker:            writeTicker,
		chLock:                 &sync.Mutex{},
		latestFeedsByTickerMtx: &sync.RWMutex{},
		latestFeedsByTicker:    make(map[string]pricefeeder.PriceFeed),
		feedChan:               make(chan pricefeeder.PriceFeed),
		quitChan:               make(chan struct{}, 1),
		marketByTickerMtx:      &sync.RWMutex{},
		marketByTicker:         make(map[string]pricefeeder.Market),
		connMtx:                &sync.Mutex{},
		conn:                   conn,
	}, nil
}

func (s *service) WellKnownMarkets() []pricefeeder.Market {
	return wellKnownMarkets
}

func (s *service) SubscribeMarkets(markets []pricefeeder.Market) error {
	mktTickers := make([]string, 0, len(markets))
	mktByTicker := make(map[string]pricefeeder.Market)
	for _, mkt := range markets {
		mktTickers = append(mktTickers, mkt.Ticker)
		mktByTicker[mkt.Ticker] = mkt
	}

	if err := s.subscribe(mktTickers); err != nil {
		return err
	}

	s.addMarkets(mktByTicker)
	return nil
}

func (s *service) UnSubscribeMarkets(markets []pricefeeder.Market) error {
	mktTickers := make([]string, 0, len(markets))
	mktByTicker := make(map[string]pricefeeder.Market)
	for _, mkt := range markets {
		mktTickers = append(mktTickers, mkt.Ticker)
		mktByTicker[mkt.Ticker] = mkt
	}

	if err := s.unsubscribe(mktTickers); err != nil {
		return err
	}

	s.removeMarkets(mktByTicker)
	s.removeFeeds(mktByTicker)
	return nil
}

func (s *service) Start() error {
	mustReconnect, err := s.start()
	for mustReconnect {
		log.WithError(err).Warn("connection dropped unexpectedly. Trying to reconnect...")

		tickers := make([]string, 0, len(s.marketByTicker))
		for ticker := range s.getMarketTickers() {
			tickers = append(tickers, ticker)
		}

		conn, err := connect()
		if err != nil {
			return err
		}

		s.addConn(conn)

		if err := s.subscribe(tickers); err != nil {
			return err
		}

		log.Debug("connection and subscriptions re-established. Restarting...")
		mustReconnect, err = s.start()
	}

	return err
}

func (s *service) Stop() {
	s.quitChan <- struct{}{}
}

func (s *service) FeedChan() chan pricefeeder.PriceFeed {
	return s.feedChan
}

func (s *service) start() (mustReconnect bool, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			mustReconnect = true
		}
	}()

	go func() {
		for range s.writeTicker.C {
			s.writeToFeedChan()
		}
	}()

	for {
		select {
		case <-s.quitChan:
			s.writeTicker.Stop()
			s.closeChannels()
			err = s.getConn().Close()
			return false, err
		default:
			// If for any reason, reading a message from the socket panics, we make
			// sure to recover and flag that a reconnection is required.
			msg := make(map[string]interface{})
			if err := s.getConn().ReadJSON(&msg); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					panic(err)
				}
				log.WithError(err).Warn("could not read message from socket")
				continue
			}

			priceFeed := s.parseFeed(msg)
			if priceFeed == nil {
				continue
			}

			s.writePriceFeed(priceFeed.Market.Ticker, *priceFeed)
		}
	}
}

func (s *service) readPriceFeeds() []pricefeeder.PriceFeed {
	s.latestFeedsByTickerMtx.RLock()
	defer s.latestFeedsByTickerMtx.RUnlock()

	feeds := make([]pricefeeder.PriceFeed, 0, len(s.latestFeedsByTicker))
	for _, priceFeed := range s.latestFeedsByTicker {
		feeds = append(feeds, priceFeed)
	}
	return feeds
}

func (s *service) writePriceFeed(mktTicker string, priceFeed pricefeeder.PriceFeed) {
	s.latestFeedsByTickerMtx.Lock()
	defer s.latestFeedsByTickerMtx.Unlock()

	if mktTicker == "" {
		return
	}

	s.latestFeedsByTicker[mktTicker] = priceFeed
}

func (s *service) removeFeeds(markets map[string]pricefeeder.Market) {
	s.latestFeedsByTickerMtx.Lock()
	defer s.latestFeedsByTickerMtx.Unlock()

	for ticker := range markets {
		delete(s.latestFeedsByTicker, ticker)
	}
}

func (s *service) writeToFeedChan() {
	s.chLock.Lock()
	defer s.chLock.Unlock()

	priceFeeds := s.readPriceFeeds()
	for _, priceFeed := range priceFeeds {
		s.feedChan <- priceFeed
	}
}

func (s *service) addMarkets(markets map[string]pricefeeder.Market) {
	s.marketByTickerMtx.Lock()
	defer s.marketByTickerMtx.Unlock()

	for ticker, mkt := range markets {
		s.marketByTicker[ticker] = mkt
	}
}

func (s *service) removeMarkets(markets map[string]pricefeeder.Market) {
	s.marketByTickerMtx.Lock()
	defer s.marketByTickerMtx.Unlock()

	for ticker := range markets {
		delete(s.marketByTicker, ticker)
	}
}

func (s *service) getMarketTickers() map[string]pricefeeder.Market {
	s.marketByTickerMtx.RLock()
	defer s.marketByTickerMtx.RUnlock()

	return s.marketByTicker
}

func (s *service) getMarketByTicker(ticker string) pricefeeder.Market {
	s.marketByTickerMtx.RLock()
	defer s.marketByTickerMtx.RUnlock()

	mkt, _ := s.marketByTicker[ticker]
	return mkt
}

func (s *service) closeChannels() {
	s.chLock.Lock()
	defer s.chLock.Unlock()

	close(s.feedChan)
	close(s.quitChan)
}

func (s *service) parseFeed(msg map[string]interface{}) *pricefeeder.PriceFeed {
	if _, ok := msg["type"]; !ok {
		return nil
	}
	if e, ok := msg["type"].(string); !ok || e != "ticker" {
		return nil
	}
	if _, ok := msg["product_id"]; !ok {
		return nil
	}
	ticker, ok := msg["product_id"].(string)
	if !ok {
		return nil
	}
	if _, ok := msg["price"]; !ok {
		return nil
	}
	priceStr, ok := msg["price"].(string)
	if !ok {
		return nil
	}

	quotePrice, err := decimal.NewFromString(priceStr) // TODO: round to 8 decimals?
	if err != nil {
		return nil
	}
	basePrice := decimal.NewFromInt(1).Div(quotePrice).Round(8)
	mkt := s.getMarketByTicker(ticker)

	return &pricefeeder.PriceFeed{
		Market: mkt,
		Price: pricefeeder.Price{
			BasePrice:  basePrice,
			QuotePrice: quotePrice,
		},
	}
}

func connect() (*websocket.Conn, error) {
	url := fmt.Sprintf("wss://%s", CoinbaseWebSocketURL)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (s *service) subscribe(mktTickers []string) error {
	msg := map[string]interface{}{
		"type":        "subscribe",
		"product_ids": mktTickers,
		"channels": []string{
			"heartbeat", "ticker",
		},
	}

	if err := s.getConn().WriteJSON(msg); err != nil {
		return fmt.Errorf("cannot subscribe to given markets: %s", err)
	}

	return nil
}

func (s *service) unsubscribe(mktTickers []string) error {
	msg := map[string]interface{}{
		"type":        "unsubscribe",
		"product_ids": mktTickers,
		"channels": []string{
			"heartbeat", "ticker",
		},
	}

	if err := s.getConn().WriteJSON(msg); err != nil {
		return fmt.Errorf("cannot unsubscribe to given markets: %s", err)
	}

	return nil
}

func (s *service) addConn(c *websocket.Conn) {
	s.connMtx.Lock()
	defer s.connMtx.Unlock()

	s.conn = c
}

func (s *service) getConn() *websocket.Conn {
	s.connMtx.Lock()
	defer s.connMtx.Unlock()

	return s.conn
}