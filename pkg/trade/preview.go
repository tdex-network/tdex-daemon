package trade

import (
	"encoding/hex"
	"errors"

	tradeclient "github.com/tdex-network/tdex-daemon/pkg/trade/client"
	trademarket "github.com/tdex-network/tdex-daemon/pkg/trade/market"
	tradetype "github.com/tdex-network/tdex-daemon/pkg/trade/type"
)

var (
	// ErrInvalidMarket ...
	ErrInvalidMarket = errors.New(
		"market must be a pair of valid 32-bytes assets encoded in hex format",
	)
	// ErrInvalidTradeType ...
	ErrInvalidTradeType = errors.New("trade type must be either BUY or SELL")
	// ErrInvalidAmount ...
	ErrInvalidAmount = errors.New("amount must be a positive satoshi amount")
)

// PreviewOpts is the struct given to Preview method
type PreviewOpts struct {
	Market    trademarket.Market
	TradeType int
	Amount    uint64
	Asset     string
}

func (o PreviewOpts) validate() error {
	if err := o.Market.Validate(); err != nil {
		return err
	}
	if err := tradetype.TradeType(o.TradeType).Validate(); err != nil {
		return err
	}
	if o.Amount <= 0 {
		return ErrInvalidAmount
	}
	if buf, err := hex.DecodeString(o.Asset); err != nil || len(buf) != 32 {
		return ErrInvalidAsset
	}
	return nil
}

// PreviewResult is the struct returned by Preview method
type PreviewResult struct {
	AssetToSend     string
	AmountToSend    uint64
	AssetToReceive  string
	AmountToReceive uint64
}

// Preview queries the gRPC server to get the latest price for the given market,
// then calculates the amount to send or to receive depending on the given type.
func (t *Trade) Preview(opts PreviewOpts) (*PreviewResult, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	tradeType := tradetype.TradeType(opts.TradeType)
	reply, err := t.client.MarketPrice(tradeclient.MarketPriceOpts{
		Market:    opts.Market,
		TradeType: tradeType,
		Amount:    opts.Amount,
		Asset:     opts.Asset,
	})
	if err != nil {
		return nil, err
	}
	preview := reply.GetPrices()[0]

	assetToSend := preview.GetAsset()
	amountToSend := preview.GetAmount()
	assetToReceive := opts.Asset
	amountToReceive := opts.Amount

	if tradeType.IsBuy() {
		if opts.Asset == opts.Market.QuoteAsset {
			assetToSend, assetToReceive = assetToReceive, assetToSend
			amountToSend, amountToReceive = amountToReceive, amountToSend
		}
	} else {
		if opts.Asset == opts.Market.BaseAsset {
			assetToSend, assetToReceive = assetToReceive, assetToSend
			amountToSend, amountToReceive = amountToReceive, amountToSend
		}
	}

	return &PreviewResult{
		AssetToSend:     assetToSend,
		AmountToSend:    amountToSend,
		AssetToReceive:  assetToReceive,
		AmountToReceive: amountToReceive,
	}, nil
}
