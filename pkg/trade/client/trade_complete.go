package tradeclient

import (
	"context"
	"errors"

	pbswap "github.com/tdex-network/tdex-protobuf/generated/go/swap"
	pbtrade "github.com/tdex-network/tdex-protobuf/generated/go/trade"

	"google.golang.org/protobuf/proto"
)

// TradeCompleteOpts is the struct given to TradeComplete method
type TradeCompleteOpts struct {
	SwapComplete []byte
	SwapFail     []byte
}

var (
	// ErrNullTradeCompleteOpts ...
	ErrNullTradeCompleteOpts = errors.New("swap complete and swap fail messages must not be both null")
	// ErrInvalidTradeCompleteOpts ...
	ErrInvalidTradeCompleteOpts = errors.New("swap complete and swap fail messages must not be both defined")
	// ErrInvalidSwapCompleteMessage ...
	ErrInvalidSwapCompleteMessage = errors.New("swap complete must be a valid serialized message")
	// ErrInvalidSwapFailMessage ...
	ErrInvalidSwapFailMessage = errors.New("swap fail must be a valid serialized message")
)

func (o TradeCompleteOpts) validate() error {
	if len(o.SwapComplete) <= 0 && len(o.SwapFail) <= 0 {
		return ErrNullTradeCompleteOpts
	}
	if len(o.SwapComplete) > 0 && len(o.SwapFail) > 0 {
		return ErrInvalidTradeCompleteOpts
	}
	if len(o.SwapComplete) > 0 {
		if err := proto.Unmarshal(o.SwapComplete, &pbswap.SwapComplete{}); err != nil {
			return ErrInvalidSwapCompleteMessage
		}
	}
	if len(o.SwapFail) > 0 {
		if err := proto.Unmarshal(o.SwapFail, &pbswap.SwapFail{}); err != nil {
			return ErrInvalidSwapFailMessage
		}
	}

	return nil
}

// TradeComplete crafts the request and calls the TradeComplete rpc
func (c *Client) TradeComplete(opts TradeCompleteOpts) (*pbtrade.TradeCompleteReply, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	var swapComplete *pbswap.SwapComplete
	var swapFail *pbswap.SwapFail
	if len(opts.SwapComplete) > 0 {
		swapComplete = &pbswap.SwapComplete{}
		proto.Unmarshal(opts.SwapComplete, swapComplete)
	}
	if len(opts.SwapFail) > 0 {
		swapFail = &pbswap.SwapFail{}
		proto.Unmarshal(opts.SwapFail, swapFail)
	}

	request := &pbtrade.TradeCompleteRequest{
		SwapComplete: swapComplete,
		SwapFail:     swapFail,
	}
	return c.client.TradeCompleteUnary(context.Background(), request)
}
