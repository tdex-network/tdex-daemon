package tradetype

import (
	"errors"

	tdexv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v2"
)

const (
	// Buy type
	Buy TradeType = iota
	// Sell type
	Sell
)

var (
	// ErrInvalidTradeType ...
	ErrInvalidTradeType = errors.New("trade type must be either BUY or SELL")
)

type TradeType int

// Validate makes sure that the current trade type is either BUY or SELL
func (t TradeType) Validate() error {
	if t != TradeType(tdexv2.TradeType_TRADE_TYPE_BUY) && t != TradeType(tdexv2.TradeType_TRADE_TYPE_SELL) {
		return ErrInvalidTradeType
	}
	return nil
}

// IsBuy returns whether the current trade type is BUY
func (t TradeType) IsBuy() bool {
	return t == TradeType(tdexv2.TradeType_TRADE_TYPE_BUY)
}

// IsSell returns whether the current trade type is SELL
func (t TradeType) IsSell() bool {
	return t == TradeType(tdexv2.TradeType_TRADE_TYPE_SELL)
}

// String formats the type to a human-readable form
func (t TradeType) String() string {
	return tdexv2.TradeType(t).String()
}
