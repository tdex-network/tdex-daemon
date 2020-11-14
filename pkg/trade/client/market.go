package tradeclient

import (
	"context"
	"errors"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"

	trademarket "github.com/tdex-network/tdex-daemon/pkg/trade/market"
	tradetype "github.com/tdex-network/tdex-daemon/pkg/trade/type"

	pbtrade "github.com/tdex-network/tdex-protobuf/generated/go/trade"
)

// Markets calls the Markets rpc and returns its response
func (c *Client) Markets() (*pbtrade.MarketsReply, error) {
	return c.client.Markets(context.Background(), &pbtrade.MarketsRequest{})
}

// BalancesOpts is the struct given to Balances method
type BalancesOpts struct {
	Market trademarket.Market
}

func (o BalancesOpts) validate() error {
	if err := o.Market.Validate(); err != nil {
		return err
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
			BaseAsset:  opts.Market.BaseAsset,
			QuoteAsset: opts.Market.QuoteAsset,
		},
	}
	return c.client.Balances(context.Background(), request)
}

// MarketPriceOpts is the struct given to MarketPrice method
type MarketPriceOpts struct {
	Market    trademarket.Market
	TradeType tradetype.TradeType
	Amount    uint64
}

func (o MarketPriceOpts) validate() error {
	if err := o.Market.Validate(); err != nil {
		return err
	}
	if err := o.TradeType.Validate(); err != nil {
		return err
	}
	if o.Amount == 0 {
		return errors.New("amount must be greater than 0")
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
			BaseAsset:  opts.Market.BaseAsset,
			QuoteAsset: opts.Market.QuoteAsset,
		},
		Type:   pbtrade.TradeType(opts.TradeType),
		Amount: opts.Amount,
	}
	return c.client.MarketPrice(context.Background(), request)
}
