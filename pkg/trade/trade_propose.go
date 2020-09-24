package tradeclient

import (
	"context"
	"errors"

	pbswap "github.com/tdex-network/tdex-protobuf/generated/go/swap"
	pbtrade "github.com/tdex-network/tdex-protobuf/generated/go/trade"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"

	"google.golang.org/protobuf/proto"
)

var (
	// ErrInvalidBaseAsset ...
	ErrInvalidBaseAsset = errors.New("base asset must be a 32-byte array in hex format")
	// ErrInvalidQuoteAsset ...
	ErrInvalidQuoteAsset = errors.New("quote asset must be a 32-byte array in hex format")
	// ErrMalformedSwapRequestMessage ...
	ErrMalformedSwapRequestMessage = errors.New("swap request must be a valid serialized message")
	// ErrInvalidTradeType ...
	ErrInvalidTradeType = errors.New("trade type must be either BUY or SELL")
)

// TradeProposeOpts is the struct given to TradePropose method
type TradeProposeOpts struct {
	MarketBaseAsset  string
	MarketQuoteAsset string
	SwapRequest      []byte
	TradeType        int
}

func (o TradeProposeOpts) validate() error {
	if isValidAsset(o.MarketBaseAsset) {
		return ErrInvalidBaseAsset
	}
	if isValidAsset(o.MarketQuoteAsset) {
		return ErrInvalidQuoteAsset
	}
	if err := proto.Unmarshal(o.SwapRequest, &pbswap.SwapRequest{}); err != nil {
		return ErrMalformedSwapRequestMessage
	}
	if isValidTradeType(o.TradeType) {
		return ErrInvalidTradeType
	}

	return nil
}

// TradePropose crafts the request and calls the TradePropose rpc
func (c *Client) TradePropose(opts TradeProposeOpts) (*pbtrade.TradeProposeReply, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	market := &pbtypes.Market{
		BaseAsset:  opts.MarketBaseAsset,
		QuoteAsset: opts.MarketQuoteAsset,
	}
	swapRequest := &pbswap.SwapRequest{}
	proto.Unmarshal(opts.SwapRequest, swapRequest)

	request := &pbtrade.TradeProposeRequest{
		Market:      market,
		SwapRequest: swapRequest,
		Type:        pbtypes.TradeType(opts.TradeType),
	}
	stream, err := c.client.TradePropose(context.Background(), request)
	if err != nil {
		return nil, err
	}
	return stream.Recv()
}
