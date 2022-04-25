package tradeclient

import (
	"context"
	"encoding/hex"
	"errors"

	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v1"

	trademarket "github.com/tdex-network/tdex-daemon/pkg/trade/market"
	tradetype "github.com/tdex-network/tdex-daemon/pkg/trade/type"
)

// Markets calls the Markets rpc and returns its response
func (c *Client) Markets() (*tdexv1.ListMarketsResponse, error) {
	return c.client.ListMarkets(context.Background(), &tdexv1.ListMarketsRequest{})
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
func (c *Client) GetMarketBalance(opts BalancesOpts) (*tdexv1.GetMarketBalanceResponse, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	request := &tdexv1.GetMarketBalanceRequest{
		Market: &tdexv1.Market{
			BaseAsset:  opts.Market.BaseAsset,
			QuoteAsset: opts.Market.QuoteAsset,
		},
	}
	return c.client.GetMarketBalance(context.Background(), request)
}

// PreviewTradeOpts is the struct given to PreviewTrade method
type PreviewTradeOpts struct {
	Market    trademarket.Market
	TradeType tradetype.TradeType
	Amount    uint64
	Asset     string
}

func (o PreviewTradeOpts) validate() error {
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

// PreviewTrade crafts the request and calls the PreviewTrade rpc
func (c *Client) PreviewTrade(opts PreviewTradeOpts) (*tdexv1.PreviewTradeResponse, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	request := &tdexv1.PreviewTradeRequest{
		Market: &tdexv1.Market{
			BaseAsset:  opts.Market.BaseAsset,
			QuoteAsset: opts.Market.QuoteAsset,
		},
		Type:   tdexv1.TradeType(opts.TradeType),
		Amount: opts.Amount,
		Asset:  opts.Asset,
	}
	return c.client.PreviewTrade(context.Background(), request)
}
