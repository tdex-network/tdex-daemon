package tradeclient

import (
	"context"

	pbtrade "github.com/tdex-network/tdex-protobuf/generated/go/trade"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"
)

// Markets calls the Markets rpc and returns its response
func (c *Client) Markets() (*pbtrade.MarketsReply, error) {
	return c.client.Markets(context.Background(), &pbtrade.MarketsRequest{})
}

// BalancesOpts is the struct given to Balances method
type BalancesOpts struct {
	MarketBaseAsset  string
	MarketQuoteAsset string
}

func (o BalancesOpts) validate() error {
	if !isValidAsset(o.MarketBaseAsset) {
		return ErrInvalidBaseAsset
	}
	if !isValidAsset(o.MarketQuoteAsset) {
		return ErrInvalidQuoteAsset
	}
	return nil
}

// Balances crafts the request and calls the Balances rpc
func (c *Client) Balances(opts BalancesOpts) (*pbtrade.BalancesReply, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	request := &pbtrade.BalancesRequest{
		Market: &pbtypes.Market{
			BaseAsset:  opts.MarketBaseAsset,
			QuoteAsset: opts.MarketQuoteAsset,
		},
	}
	return c.client.Balances(context.Background(), request)
}

// MarketPriceOpts is the struct given to MarketPrice method
type MarketPriceOpts struct {
	MarketBaseAsset  string
	MarketQuoteAsset string
	TradeType        int
}

func (o MarketPriceOpts) validate() error {
	if !isValidAsset(o.MarketBaseAsset) {
		return ErrInvalidBaseAsset
	}
	if !isValidAsset(o.MarketQuoteAsset) {
		return ErrInvalidQuoteAsset
	}
	if isValidTradeType(o.TradeType) {
		return ErrInvalidTradeType
	}
	return nil
}

// MarketPrice crafts the request and calls the MarketPrice rpc
func (c *Client) MarketPrice(opts MarketPriceOpts) (*pbtrade.MarketPriceReply, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	request := &pbtrade.MarketPriceRequest{
		Market: &pbtypes.Market{
			BaseAsset:  opts.MarketBaseAsset,
			QuoteAsset: opts.MarketQuoteAsset,
		},
		Type: pbtypes.TradeType(opts.TradeType),
	}
	return c.client.MarketPrice(context.Background(), request)
}
