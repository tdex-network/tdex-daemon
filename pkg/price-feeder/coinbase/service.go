package coinbasefeeder

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	pricefeeder "github.com/tdex-network/tdex-daemon/pkg/price-feeder"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	baseURL                 = "ws-feed.exchange.coinbase.com"
	maxReconnectionAttempts = 3
)

var unexpectedErrors = []int{
	websocket.CloseNormalClosure,
	websocket.CloseGoingAway,
	websocket.CloseProtocolError,
	websocket.CloseUnsupportedData,
	websocket.CloseNoStatusReceived,
	websocket.CloseAbnormalClosure,
	websocket.CloseInvalidFramePayloadData,
	websocket.ClosePolicyViolation,
	websocket.CloseMessageTooBig,
	websocket.CloseMandatoryExtension,
	websocket.CloseInternalServerErr,
	websocket.CloseServiceRestart,
	websocket.CloseTryAgainLater,
	websocket.CloseTLSHandshake,
}

type service struct {
	conn *websocket.Conn

	marketLock      *sync.RWMutex
	marketsByTicker map[string]pricefeeder.Market

	feedLock         *sync.RWMutex
	lastFeedByTicker map[string]pricefeeder.PriceFeed

	feedCh chan pricefeeder.PriceFeed
}

func NewService() (pricefeeder.PriceFeeder, error) {
	conn, err := connect()
	if err != nil {
		return nil, err
	}
	marketLock, feedLock := &sync.RWMutex{}, &sync.RWMutex{}
	marketsByTicker := make(map[string]pricefeeder.Market)
	lastFeedByTicker := make(map[string]pricefeeder.PriceFeed)
	feedCh := make(chan pricefeeder.PriceFeed, 20)

	return &service{
		conn, marketLock, marketsByTicker, feedLock, lastFeedByTicker, feedCh,
	}, nil
}

func (s *service) Start() chan pricefeeder.PriceFeed {
	go s.start()
	return s.feedCh
}

func (s *service) Stop() {
	s.conn.Close()
	close(s.feedCh)
}

func (s *service) SubscribeMarkets(markets []pricefeeder.Market) error {
	tickers := make([]string, 0, len(markets))
	marketsToAdd := make([]pricefeeder.Market, 0, len(markets))
	for _, mkt := range markets {
		if _, ok := s.getMarketByTicker(mkt.Ticker); !ok {
			tickers = append(tickers, mkt.Ticker)
			marketsToAdd = append(marketsToAdd, mkt)
		}
	}

	if err := s.subscribe(tickers); err != nil {
		return err
	}

	s.addMarkets(marketsToAdd)
	return nil
}

func (s *service) UnsubscribeMarkets(markets []pricefeeder.Market) error {
	tickers := make([]string, 0, len(markets))
	marketsToRemove := make([]pricefeeder.Market, 0, len(markets))
	for _, mkt := range markets {
		if _, ok := s.getMarketByTicker(mkt.Ticker); ok {
			tickers = append(tickers, mkt.Ticker)
			marketsToRemove = append(marketsToRemove, mkt)
		}
	}

	if err := s.unsubscribe(tickers); err != nil {
		return err
	}

	s.removeFeeds(tickers)
	s.removeMarkets(marketsToRemove)
	return nil
}

func (s *service) ListSubscriptions() []pricefeeder.Market {
	return s.getMarkets()
}

func (s *service) start() {
	defer func(s *service) {
		if rec := recover(); rec != nil {
			log.Debug(
				"connection with kraken server dropped, attempting to reconnect...",
			)
			s.reconnect()
		}
	}(s)

	for {
		_, message, err := s.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, unexpectedErrors...) {
				panic(err)
			}
			return
		}

		priceFeed := s.parseFeed(message)
		if priceFeed == nil {
			continue
		}

		lastFeed, ok := s.getPriceFeed(priceFeed.Market.Ticker)
		// Prevent updating a feed if it hasn't changed.
		if ok && priceFeed.Price.BasePrice.Equal(lastFeed.Price.BasePrice) {
			continue
		}

		s.updatePriceFeed(priceFeed.Market.Ticker, *priceFeed)
		s.feedCh <- *priceFeed
	}
}

func (s *service) reconnect() {
	var conn *websocket.Conn
	var err error
	for attempt := 0; attempt < maxReconnectionAttempts; attempt++ {
		conn, err = connect()
		if err == nil {
			break
		}
		log.WithError(err).Debugf("reconnection attempt %d failed", attempt)
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		log.Fatal("failed to reconnect to kraken server")
	}

	s.conn = conn

	go s.start()

	if err := s.resubscribe(); err != nil {
		log.WithError(err).Fatal(
			"failed to restore subscriptions after reconnection",
		)
	}

	log.Debug("kraken: connection with server restored")
}

func (s *service) resubscribe() error {
	tickers := s.getMarketTickers()
	if len(tickers) <= 0 {
		return nil
	}

	return s.subscribe(tickers)
}

func (s *service) subscribe(mktTickers []string) error {
	msg := map[string]interface{}{
		"type":        "subscribe",
		"product_ids": mktTickers,
		"channels": []string{
			"heartbeat", "ticker",
		},
	}

	if err := s.conn.WriteJSON(msg); err != nil {
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

	if err := s.conn.WriteJSON(msg); err != nil {
		return fmt.Errorf("cannot unsubscribe to given markets: %s", err)
	}

	return nil
}

func (s *service) parseFeed(buf []byte) *pricefeeder.PriceFeed {
	msg := make(map[string]interface{})
	if err := json.Unmarshal(buf, &msg); err != nil {
		return nil
	}
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
	mkt, ok := s.getMarketByTicker(ticker)
	if !ok {
		return nil
	}

	return &pricefeeder.PriceFeed{
		Market: mkt,
		Price: pricefeeder.Price{
			BasePrice:  basePrice,
			QuotePrice: quotePrice,
		},
	}
}

func (s *service) addMarkets(markets []pricefeeder.Market) {
	s.marketLock.Lock()
	defer s.marketLock.Unlock()

	for _, mkt := range markets {
		s.marketsByTicker[mkt.Ticker] = mkt
	}
}

func (s *service) removeMarkets(markets []pricefeeder.Market) {
	s.marketLock.Lock()
	defer s.marketLock.Unlock()

	for _, mkt := range markets {
		delete(s.marketsByTicker, mkt.Ticker)
	}
}

func (s *service) getMarkets() []pricefeeder.Market {
	s.marketLock.RLock()
	defer s.marketLock.RUnlock()

	markets := make([]pricefeeder.Market, 0, len(s.marketsByTicker))
	for _, mkt := range s.marketsByTicker {
		markets = append(markets, mkt)
	}
	return markets
}

func (s *service) getMarketByTicker(ticker string) (pricefeeder.Market, bool) {
	s.marketLock.RLock()
	defer s.marketLock.RUnlock()

	mkt, ok := s.marketsByTicker[ticker]
	return mkt, ok
}

func (s *service) getMarketTickers() []string {
	s.marketLock.RLock()
	defer s.marketLock.RUnlock()

	tickers := make([]string, 0, len(s.marketsByTicker))
	for ticker := range s.marketsByTicker {
		tickers = append(tickers, ticker)
	}
	return tickers
}

func (s *service) updatePriceFeed(ticker string, feed pricefeeder.PriceFeed) {
	if ticker == "" {
		return
	}

	s.feedLock.Lock()
	defer s.feedLock.Unlock()

	s.lastFeedByTicker[ticker] = feed
}

func (s *service) getPriceFeed(ticker string) (pricefeeder.PriceFeed, bool) {
	s.feedLock.RLock()
	defer s.feedLock.RUnlock()

	feed, ok := s.lastFeedByTicker[ticker]
	return feed, ok
}

func (s *service) removeFeeds(tickers []string) {
	s.feedLock.Lock()
	defer s.feedLock.Unlock()

	for _, ticker := range tickers {
		delete(s.lastFeedByTicker, ticker)
	}
}

func connect() (*websocket.Conn, error) {
	url := fmt.Sprintf("wss://%s", baseURL)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
