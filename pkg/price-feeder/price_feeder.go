package pricefeeder

import (
	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

var (
	WebSocketCloseErrors = []int{
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
)

type PriceFeeder interface {
	WellKnownMarkets() []Market
	SubscribeMarkets([]Market) error
	UnSubscribeMarkets([]Market) error

	Start() error //TODO Start is blocking in impl, should be async
	Stop()

	FeedChan() chan PriceFeed
}

type PriceFeed struct {
	Market Market
	Price  Price
}

type Market struct {
	BaseAsset  string
	QuoteAsset string
	Ticker     string
}

type Price struct {
	BasePrice  decimal.Decimal
	QuotePrice decimal.Decimal
}
