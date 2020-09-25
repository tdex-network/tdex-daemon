package tradeclient

import (
	"context"
	"errors"

	trademarket "github.com/tdex-network/tdex-daemon/pkg/trade/market"
	tradetype "github.com/tdex-network/tdex-daemon/pkg/trade/type"

	pbswap "github.com/tdex-network/tdex-protobuf/generated/go/swap"
	pbtrade "github.com/tdex-network/tdex-protobuf/generated/go/trade"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"

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
	if err := proto.Unmarshal(o.SwapRequest, &pbswap.SwapRequest{}); err != nil {
		return ErrMalformedSwapRequestMessage
	}
	if err := o.TradeType.Validate(); err != nil {
		return err
	}

	return nil
}

// TradePropose crafts the request and calls the TradePropose rpc
func (c *Client) TradePropose(opts TradeProposeOpts) (*pbtrade.TradeProposeReply, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	market := &pbtypes.Market{
		BaseAsset:  opts.Market.BaseAsset,
		QuoteAsset: opts.Market.QuoteAsset,
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
