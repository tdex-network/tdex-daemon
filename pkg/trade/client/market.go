package tradeclient

import (
	"context"
	"encoding/hex"
	"errors"

	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/go/tdex/v1"

	trademarket "github.com/tdex-network/tdex-daemon/pkg/trade/market"
	tradetype "github.com/tdex-network/tdex-daemon/pkg/trade/type"
)

// Markets calls the Markets rpc and returns its response
func (c *Client) Markets() (*tdexv1.MarketsReply, error) {
	return c.client.Markets(context.Background(), &tdexv1.MarketsRequest{})
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
func (c *Client) Balances(opts BalancesOpts) (*tdexv1.BalancesReply, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	request := &tdexv1.BalancesRequest{
		Market: &tdexv1.Market{
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
	Asset     string
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
	if buf, err := hex.DecodeString(o.Asset); err != nil || len(buf) != 32 {
		return errors.New("asset must be a 32-byte array in hex format")
	}
	return nil
}

// MarketPrice crafts the request and calls the MarketPrice rpc
func (c *Client) MarketPrice(opts MarketPriceOpts) (*tdexv1.MarketPriceReply, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	request := &tdexv1.MarketPriceRequest{
		Market: &tdexv1.Market{
			BaseAsset:  opts.Market.BaseAsset,
			QuoteAsset: opts.Market.QuoteAsset,
		},
		Type:   tdexv1.TradeType(opts.TradeType),
		Amount: opts.Amount,
		Asset:  opts.Asset,
	}
	return c.client.MarketPrice(context.Background(), request)
}
