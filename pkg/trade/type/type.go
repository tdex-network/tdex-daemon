package tradetype

import (
	"errors"

	pb "github.com/tdex-network/tdex-protobuf/generated/go/types"
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
	if t != TradeType(pb.TradeType_BUY) && t != TradeType(pb.TradeType_SELL) {
		return ErrInvalidTradeType
	}
	return nil
}

// IsBuy returns whether the current trade type is BUY
func (t TradeType) IsBuy() bool {
	return t == TradeType(pb.TradeType_BUY)
}

// IsSell returns whether the current trade type is SELL
func (t TradeType) IsSell() bool {
	return t == TradeType(pb.TradeType_SELL)
}

// String formats the type to a human-readable form
func (t TradeType) String() string {
	return pb.TradeType(t).String()
}
