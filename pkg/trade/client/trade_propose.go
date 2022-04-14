package tradeclient

import (
	"context"
	"errors"

	trademarket "github.com/tdex-network/tdex-daemon/pkg/trade/market"
	tradetype "github.com/tdex-network/tdex-daemon/pkg/trade/type"

	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/go/tdex/v1"

	"google.golang.org/protobuf/proto"
)

var (
	// ErrMalformedSwapRequestMessage ...
	ErrMalformedSwapRequestMessage = errors.New("swap request must be a valid serialized message")
)

// TradeProposeOpts is the struct given to TradePropose method
type TradeProposeOpts struct {
	Market      trademarket.Market
	SwapRequest []byte
	TradeType   tradetype.TradeType
}

func (o TradeProposeOpts) validate() error {
	if err := o.Market.Validate(); err != nil {
		return err
	}
	if err := proto.Unmarshal(o.SwapRequest, &tdexv1.SwapRequest{}); err != nil {
		return ErrMalformedSwapRequestMessage
	}
	if err := o.TradeType.Validate(); err != nil {
		return err
	}

	return nil
}

// TradePropose crafts the request and calls the TradePropose rpc
func (c *Client) TradePropose(opts TradeProposeOpts) (*tdexv1.ProposeTradeResponse, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	market := &tdexv1.Market{
		BaseAsset:  opts.Market.BaseAsset,
		QuoteAsset: opts.Market.QuoteAsset,
	}
	swapRequest := &tdexv1.SwapRequest{}
	proto.Unmarshal(opts.SwapRequest, swapRequest)

	request := &tdexv1.ProposeTradeRequest{
		Market:      market,
		SwapRequest: swapRequest,
		Type:        tdexv1.TradeType(opts.TradeType),
	}
	return c.client.ProposeTrade(context.Background(), request)
}
