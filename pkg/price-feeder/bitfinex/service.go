package bitfinexfeeder

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	pricefeeder "github.com/tdex-network/tdex-daemon/pkg/price-feeder"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

const (
	// BitfinexWebSocketURL is the base url to open a WebSocket connection with
	// Bitfinex.
	BitfinexWebSocketURL = "api-pub.bitfinex.com/ws/2"
)

var (
	wellKnownMarkets = []pricefeeder.Market{
		{
			BaseAsset:  "6f0279e9ed041c3d710a9f57d0c02928416460c4b722ae3457a11eec381c526d",
			QuoteAsset: "ce091c998b83c78bb71a632313ba3760f1763d9cfcffae02258ffa9865a37bd2",
			Ticker:     "BTCUST",
		},
	}
)

type service struct {
	conn        *websocket.Conn
	writeTicker *time.Ticker
	lock        *sync.RWMutex
	chLock      *sync.Mutex

	marketByTicker      map[string]pricefeeder.Market
	latestFeedsByTicker map[string]pricefeeder.PriceFeed
	tickersByChanId     map[int]string
	feedChan            chan pricefeeder.PriceFeed
	quitChan            chan struct{}
}

func NewBitfinexPriceFeeder(args ...interface{}) (pricefeeder.PriceFeeder, error) {
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
		latestFeedsByTicker: make(map[string]pricefeeder.PriceFeed),
		feedChan:            make(chan pricefeeder.PriceFeed),
		quitChan:            make(chan struct{}, 1),
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

	conn, tickersByChanId, err := connectAndSubscribe(mktTickers)
	if err != nil {
		return err
	}

	s.conn = conn
	s.tickersByChanId = tickersByChanId
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

		conn, tickersByChanId, err := connectAndSubscribe(tickers)
		if err != nil {
			return err
		}
		s.conn = conn
		s.tickersByChanId = tickersByChanId

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
			err = s.conn.Close()
			return false, err
		default:
			// if for any reason, reading a message from the socket panics, we make
			// sure to recover and flag that a reconnection is required.
			_, message, err := s.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					panic(err)
				}
			}

			priceFeed := s.parseFeed(message)
			if priceFeed == nil {
				continue
			}

			s.writePriceFeed(priceFeed.Market.Ticker, *priceFeed)
		}
	}
}

func (s *service) readPriceFeeds() []pricefeeder.PriceFeed {
	s.lock.RLock()
	defer s.lock.RUnlock()

	feeds := make([]pricefeeder.PriceFeed, 0, len(s.latestFeedsByTicker))
	for _, priceFeed := range s.latestFeedsByTicker {
		feeds = append(feeds, priceFeed)
	}
	return feeds
}

func (s *service) writePriceFeed(mktTicker string, priceFeed pricefeeder.PriceFeed) {
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

func (s *service) parseFeed(msg []byte) *pricefeeder.PriceFeed {
	var i []interface{}
	if err := json.Unmarshal(msg, &i); err != nil {
		return nil
	}
	if len(i) != 2 {
		return nil
	}

	c, ok := i[0].(float64)
	if !ok {
		return nil
	}
	chanId := int(c)

	ticker, ok := s.tickersByChanId[chanId]
	if !ok {
		return nil
	}
	mkt, ok := s.marketByTicker[ticker]
	if !ok {
		return nil
	}

	ii, ok := i[1].([]interface{})
	if !ok {
		return nil
	}
	if len(ii) < 10 {
		return nil
	}

	p, ok := ii[6].(float64)
	if !ok {
		return nil
	}

	quotePrice := decimal.NewFromFloat(p).Round(8)
	basePrice := decimal.NewFromInt(1).Div(quotePrice) // TODO: round to 8 decimals?

	return &pricefeeder.PriceFeed{
		Market: mkt,
		Price: pricefeeder.Price{
			BasePrice:  basePrice,
			QuotePrice: quotePrice,
		},
	}
}

func connectAndSubscribe(
	mktTickers []string,
) (*websocket.Conn, map[int]string, error) {
	url := fmt.Sprintf("wss://%s", BitfinexWebSocketURL)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, nil, err
	}

	tickersByChanID := make(map[int]string)
	for _, ticker := range mktTickers {
		msg := map[string]interface{}{
			"event":   "subscribe",
			"channel": "ticker",
			"symbol":  fmt.Sprintf("t%s", ticker),
		}

		if err := conn.WriteJSON(msg); err != nil {
			return nil, nil, fmt.Errorf("cannot subscribe to market %s: %s", ticker, err)
		}

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return nil, nil, fmt.Errorf(
					"cannot read response of subscribtion for market %s: %s", ticker, err,
				)
			}

			chanId, err := parseSubscriptionResponse(msg, ticker)
			if err != nil {
				return nil, nil, err
			}
			if chanId == -1 {
				continue
			}

			tickersByChanID[chanId] = ticker
			break
		}
	}
	return conn, tickersByChanID, nil
}

func parseSubscriptionResponse(msg []byte, ticker string) (int, error) {
	m := make(map[string]interface{})
	if err := json.Unmarshal(msg, &m); err != nil {
		return -1, nil
	}
	e, ok := m["event"].(string)
	if !ok {
		return -1, nil
	}
	if e == "error" {
		return -1, fmt.Errorf("%s %s", m["pair"].(string), m["msg"].(string))
	}
	if e != "subscribed" {
		return -1, nil
	}
	if c, ok := m["channel"].(string); !ok || c != "ticker" {
		return -1, nil
	}
	if t, ok := m["pair"].(string); !ok || t != ticker {
		return -1, nil
	}
	return int(m["chanId"].(float64)), nil
}
