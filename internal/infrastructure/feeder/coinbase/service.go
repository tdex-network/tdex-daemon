package coinbasefeeder

import (
	"fmt"
	"sync"
	"time"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"

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
	wellKnownMarkets = []ports.Market{
		market{
			baseAsset:  "6f0279e9ed041c3d710a9f57d0c02928416460c4b722ae3457a11eec381c526d",
			quoteAsset: "ce091c998b83c78bb71a632313ba3760f1763d9cfcffae02258ffa9865a37bd2",
			ticker:     "BTC-USDT",
		},
	}
)

type service struct {
	conn        *websocket.Conn
	writeTicker *time.Ticker
	lock        *sync.RWMutex
	chLock      *sync.Mutex

	marketByTicker      map[string]ports.Market
	latestFeedsByTicker map[string]ports.PriceFeed
	feedChan            chan ports.PriceFeed
	quitChan            chan struct{}
}

func NewCoinbasePriceFeeder(args ...interface{}) (ports.PriceFeeder, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("invalid number of args")
	}

	interval, ok := args[0].(int)
	if !ok {
		return nil, fmt.Errorf("unknown interval arg type")
	}
	writeTicker := time.NewTicker(time.Duration(interval) * time.Millisecond)

	return &service{
		writeTicker:         writeTicker,
		lock:                &sync.RWMutex{},
		chLock:              &sync.Mutex{},
		latestFeedsByTicker: make(map[string]ports.PriceFeed),
		feedChan:            make(chan ports.PriceFeed),
		quitChan:            make(chan struct{}, 1),
	}, nil
}

func (s *service) WellKnownMarkets() []ports.Market {
	return wellKnownMarkets
}

func (s *service) SubscribeMarkets(markets []ports.Market) error {
	mktTickers := make([]string, 0, len(markets))
	mktByTicker := make(map[string]ports.Market)
	for _, mkt := range markets {
		mktTickers = append(mktTickers, mkt.Ticker())
		mktByTicker[mkt.Ticker()] = mkt
	}

	conn, err := connectAndSubscribe(mktTickers)
	if err != nil {
		return err
	}

	s.conn = conn
	s.marketByTicker = mktByTicker
	return nil
}

func (s *service) Start() error {
	mustReconnect, err := s.start()
	for mustReconnect {
		log.WithError(err).Warn("connection dropped unexpectedly. Trying to reconnect...")

		tickers := make([]string, 0, len(s.marketByTicker))
		for ticker := range s.marketByTicker {
			tickers = append(tickers, ticker)
		}

		var conn *websocket.Conn
		conn, err = connectAndSubscribe(tickers)
		if err != nil {
			return err
		}
		s.conn = conn

		log.Debug("connection and subscriptions re-established. Restarting...")
		mustReconnect, err = s.start()
	}

	return err
}

func (s *service) Stop() {
	s.quitChan <- struct{}{}
}

func (s *service) FeedChan() chan ports.PriceFeed {
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
			err = s.conn.Close()
			return false, err
		default:
			// If for any reason, reading a message from the socket panics, we make
			// sure to recover and flag that a reconnection is required.
			msg := make(map[string]interface{})
			if err := s.conn.ReadJSON(&msg); err != nil {
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

			s.writePriceFeed(priceFeed.GetMarket().Ticker(), priceFeed)
		}
	}
}

func (s *service) readPriceFeeds() []ports.PriceFeed {
	s.lock.RLock()
	defer s.lock.RUnlock()

	feeds := make([]ports.PriceFeed, 0, len(s.latestFeedsByTicker))
	for _, priceFeed := range s.latestFeedsByTicker {
		feeds = append(feeds, priceFeed)
	}
	return feeds
}

func (s *service) writePriceFeed(mktTicker string, priceFeed ports.PriceFeed) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.latestFeedsByTicker[mktTicker] = priceFeed
}

func (s *service) writeToFeedChan() {
	s.chLock.Lock()
	defer s.chLock.Unlock()

	priceFeeds := s.readPriceFeeds()
	for _, priceFeed := range priceFeeds {
		s.feedChan <- priceFeed
	}
}

func (s *service) closeChannels() {
	s.chLock.Lock()
	defer s.chLock.Unlock()

	close(s.feedChan)
	close(s.quitChan)
}

func (s *service) parseFeed(msg map[string]interface{}) ports.PriceFeed {
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
	mkt := s.marketByTicker[ticker]

	return &priceFeed{
		market: mkt,
		price: &price{
			basePrice:  basePrice,
			quotePrice: quotePrice,
		},
	}
}

func connectAndSubscribe(mktTickers []string) (*websocket.Conn, error) {
	url := fmt.Sprintf("wss://%s", CoinbaseWebSocketURL)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	msg := map[string]interface{}{
		"type":        "subscribe",
		"product_ids": mktTickers,
		"channels": []string{
			"heartbeat", "ticker",
		},
	}

	if err := conn.WriteJSON(msg); err != nil {
		return nil, fmt.Errorf("cannot subscribe to given markets: %s", err)
	}

	for {
		msg := make(map[string]interface{})
		if err := conn.ReadJSON(&msg); err != nil {
			return nil, fmt.Errorf(
				"cannot read response of subscription to markets: %s", err,
			)
		}

		msgType := msg["type"].(string)
		if msgType == "error" {
			return nil, fmt.Errorf(msg["reason"].(string))
		}

		if msgType == "subscriptions" {
			break
		}
	}

	return conn, nil
}
