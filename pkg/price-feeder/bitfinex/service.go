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

	jsonObject jsonType = "json-object"
	jsonArray  jsonType = "array"
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

	marketByTickerMtx *sync.RWMutex
	marketByTicker    map[string]pricefeeder.Market

	latestFeedsByTickerMtx *sync.RWMutex
	latestFeedsByTicker    map[string]pricefeeder.PriceFeed

	tickersByChanIdMtx *sync.RWMutex
	tickersByChanId    map[int]string

	chanIDsByTickerMtx *sync.RWMutex
	chanIDsByTicker    map[string]int

	chLock   *sync.Mutex
	feedChan chan pricefeeder.PriceFeed

	quitChan chan struct{}
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
		tickersByChanIdMtx:     &sync.RWMutex{},
		tickersByChanId:        make(map[int]string),
		chanIDsByTickerMtx:     &sync.RWMutex{},
		chanIDsByTicker:        make(map[string]int),
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

	s.addMarketsByTicker(mktByTicker)
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
	s.removeTickerChanIds(mktByTicker)
	return nil
}

func (s *service) Start() error {
	mustReconnect, err := s.start()
	for mustReconnect {
		log.WithError(err).Warn("connection dropped unexpectedly. Trying to reconnect...")

		tickers := make([]string, 0, len(s.marketByTicker))
		for ticker := range s.getMarkets() {
			tickers = append(tickers, ticker)
		}

		conn, err := connect()
		if err != nil {
			return err
		}
		s.conn = conn

		if err = s.subscribe(tickers); err != nil {
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

			s.processMsg(message)
		}
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

func (s *service) closeChannels() {
	s.chLock.Lock()
	defer s.chLock.Unlock()

	close(s.feedChan)
	close(s.quitChan)
}

func (s *service) subscribe(
	mktTickers []string,
) error {
	for _, ticker := range mktTickers {
		msg := map[string]interface{}{
			"event":   "subscribe",
			"channel": "ticker",
			"symbol":  fmt.Sprintf("t%s", ticker),
		}

		if err := s.conn.WriteJSON(msg); err != nil {
			return fmt.Errorf("cannot subscribe to market %s: %s", ticker, err)
		}
	}

	return nil
}

func (s *service) unsubscribe(mktTickers []string) error {
	for _, ticker := range mktTickers {
		v, ok := s.getChanIdByTicker(ticker)
		if !ok {
			continue
		}

		msg := map[string]interface{}{
			"event":  "unsubscribe",
			"chanId": v,
		}
		if err := s.conn.WriteJSON(msg); err != nil {
			return fmt.Errorf("cannot unsubscribe to given markets: %s", err)
		}
	}

	return nil
}

func connect() (*websocket.Conn, error) {
	url := fmt.Sprintf("wss://%s", BitfinexWebSocketURL)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (s *service) processMsg(msg []byte) {
	switch identifyJsonType(msg) {
	case jsonObject:
		if err := s.parseSubscriptionResponse(msg); err != nil {
			log.Warnf("error parsing subscription response: %s", err)
		}
	case jsonArray:
		priceFeed := s.parseFeed(msg)
		if priceFeed != nil {
			s.writePriceFeed(priceFeed.Market.Ticker, *priceFeed)
		}
	}
}

func (s *service) parseSubscriptionResponse(msg []byte) error {
	m := make(map[string]interface{})
	if err := json.Unmarshal(msg, &m); err != nil {
		return nil
	}
	e, ok := m["event"].(string)
	if !ok {
		return nil
	}
	if e == "error" {
		return fmt.Errorf("%s %s", m["pair"].(string), m["msg"].(string))
	}
	if e != "subscribed" {
		return nil
	}
	if c, ok := m["channel"].(string); !ok || c != "ticker" {
		return nil
	}
	if _, ok := m["pair"].(string); !ok {
		return nil
	}
	ticker := m["pair"].(string)
	if _, ok := s.getMarketByTicker(ticker); !ok {
		return nil
	}

	tickerByChanID := map[int]string{
		int(m["chanId"].(float64)): ticker,
	}

	s.addTickerChanIds(tickerByChanID)
	s.addChanIdsByTicker(tickerByChanID)

	return nil
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

	ticker, ok := s.getTickerByChanId(chanId)
	if !ok {
		return nil
	}
	mkt, ok := s.getMarketByTicker(ticker)
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

type jsonType string

func identifyJsonType(msg []byte) jsonType {
	var obj map[string]interface{}
	var arr []interface{}

	if err := json.Unmarshal(msg, &obj); err == nil {
		return jsonObject
	}

	if err := json.Unmarshal(msg, &arr); err == nil {
		return jsonArray
	}

	return "unknown"
}

func (s *service) removeMarkets(markets map[string]pricefeeder.Market) {
	s.marketByTickerMtx.Lock()
	defer s.marketByTickerMtx.Unlock()

	for ticker := range markets {
		delete(s.marketByTicker, ticker)
	}
}
func (s *service) getMarkets() map[string]pricefeeder.Market {
	s.marketByTickerMtx.RLock()
	defer s.marketByTickerMtx.RUnlock()

	return s.marketByTicker
}

func (s *service) getMarketByTicker(ticker string) (pricefeeder.Market, bool) {
	s.marketByTickerMtx.RLock()
	defer s.marketByTickerMtx.RUnlock()

	mkt, ok := s.marketByTicker[ticker]
	return mkt, ok
}

func (s *service) addMarketsByTicker(markets map[string]pricefeeder.Market) {
	s.marketByTickerMtx.Lock()
	defer s.marketByTickerMtx.Unlock()

	for ticker, mkt := range markets {
		s.marketByTicker[ticker] = mkt
	}
}

func (s *service) removeFeeds(markets map[string]pricefeeder.Market) {
	s.latestFeedsByTickerMtx.Lock()
	defer s.latestFeedsByTickerMtx.Unlock()

	for ticker := range markets {
		delete(s.latestFeedsByTicker, ticker)
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

func (s *service) removeTickerChanIds(markets map[string]pricefeeder.Market) {
	s.tickersByChanIdMtx.Lock()
	defer s.tickersByChanIdMtx.Unlock()

	for ticker := range markets {
		for chanId, t := range s.tickersByChanId {
			if t == ticker {
				delete(s.tickersByChanId, chanId)
			}
		}
	}
}

func (s *service) addTickerChanIds(tickersByChanID map[int]string) {
	s.tickersByChanIdMtx.Lock()
	defer s.tickersByChanIdMtx.Unlock()

	for chanId, ticker := range tickersByChanID {
		s.tickersByChanId[chanId] = ticker
	}
}

func (s *service) getTickerByChanId(chanId int) (string, bool) {
	s.tickersByChanIdMtx.RLock()
	defer s.tickersByChanIdMtx.RUnlock()

	ticker, ok := s.tickersByChanId[chanId]
	return ticker, ok
}

func (s *service) addChanIdsByTicker(tickersByChanID map[int]string) {
	s.chanIDsByTickerMtx.Lock()
	defer s.chanIDsByTickerMtx.Unlock()

	for chanId, ticker := range tickersByChanID {
		s.chanIDsByTicker[ticker] = chanId
	}
}

func (s *service) getChanIdByTicker(ticker string) (int, bool) {
	s.chanIDsByTickerMtx.RLock()
	defer s.chanIDsByTickerMtx.RUnlock()

	chanId, ok := s.chanIDsByTicker[ticker]
	return chanId, ok
}
