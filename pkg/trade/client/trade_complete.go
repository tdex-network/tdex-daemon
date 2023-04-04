package tradeclient

import (
	"context"
	"errors"

	tdexv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v2"

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
		if err := proto.Unmarshal(o.SwapComplete, &tdexv2.SwapComplete{}); err != nil {
			return ErrInvalidSwapCompleteMessage
		}
	}
	if len(o.SwapFail) > 0 {
		if err := proto.Unmarshal(o.SwapFail, &tdexv2.SwapFail{}); err != nil {
			return ErrInvalidSwapFailMessage
		}
	}

	return nil
}

// TradeComplete crafts the request and calls the TradeComplete rpc
func (c *Client) TradeComplete(opts TradeCompleteOpts) (*tdexv2.CompleteTradeResponse, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	var swapComplete *tdexv2.SwapComplete
	var swapFail *tdexv2.SwapFail
	if len(opts.SwapComplete) > 0 {
		swapComplete = &tdexv2.SwapComplete{}
		//nolint
		proto.Unmarshal(opts.SwapComplete, swapComplete)
	}
	if len(opts.SwapFail) > 0 {
		swapFail = &tdexv2.SwapFail{}
		//nolint
		proto.Unmarshal(opts.SwapFail, swapFail)
	}

	request := &tdexv2.CompleteTradeRequest{
		SwapComplete: swapComplete,
		SwapFail:     swapFail,
	}
	return c.client.CompleteTrade(context.Background(), request)
}
