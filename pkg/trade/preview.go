package trade

import (
	"errors"

	tradeclient "github.com/tdex-network/tdex-daemon/pkg/trade/client"
	trademarket "github.com/tdex-network/tdex-daemon/pkg/trade/market"
	tradetype "github.com/tdex-network/tdex-daemon/pkg/trade/type"

	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"

	"github.com/shopspring/decimal"
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
	marketPrice, err := t.client.MarketPrice(tradeclient.MarketPriceOpts{
		Market:    opts.Market,
		TradeType: tradeType,
	})
	if err != nil {
		return nil, err
	}
	priceWithFee := marketPrice.GetPrices()[0]

	if tradeType.IsBuy() {
		return &PreviewResult{
			AssetToSend:     opts.Market.QuoteAsset,
			AmountToSend:    calcProposeAmount(priceWithFee, opts.Amount, opts.Market.QuoteAsset),
			AssetToReceive:  opts.Market.BaseAsset,
			AmountToReceive: opts.Amount,
		}, nil
	}

	return &PreviewResult{
		AssetToSend:     opts.Market.BaseAsset,
		AmountToSend:    opts.Amount,
		AssetToReceive:  opts.Market.QuoteAsset,
		AmountToReceive: calcExpectedAmount(priceWithFee, opts.Amount, opts.Market.QuoteAsset),
	}, nil
}

func calcProposeAmount(
	priceWithFee *pbtypes.PriceWithFee,
	amountToReceive uint64,
	assetToSend string,
) uint64 {
	return calcAmount(priceWithFee, amountToReceive, assetToSend, true)
}

func calcExpectedAmount(
	priceWithFee *pbtypes.PriceWithFee,
	amountToSend uint64,
	assetToReceive string,
) uint64 {
	return calcAmount(priceWithFee, amountToSend, assetToReceive, false)
}

func calcAmount(priceWithFee *pbtypes.PriceWithFee, amountP uint64, assetP string, isBuy bool) uint64 {
	price := getPrice(priceWithFee, isBuy)
	fee := priceWithFee.GetFee()
	feePercentage := decimal.NewFromInt(fee.GetBasisPoint()).Div(decimal.NewFromInt(100))
	amount := decimal.NewFromInt(int64(amountP))

	if fee.GetAsset() == assetP {
		totAmount := amount.Mul(price)
		totAmount = addOrSubFeeAmount(totAmount, totAmount.Mul(feePercentage), isBuy)
		amountR, _ := totAmount.Float64()
		return uint64(amountR)
	}

	feeAmount := amount.Mul(feePercentage)
	totAmount := addOrSubFeeAmount(amount, feeAmount, isBuy).Mul(price)
	amountR, _ := totAmount.Float64()
	return uint64(amountR)
}

func getPrice(priceWithFee *pbtypes.PriceWithFee, isBuy bool) decimal.Decimal {
	if isBuy {
		return decimal.NewFromFloat32(priceWithFee.GetPrice().GetQuotePrice())
	}
	return decimal.NewFromFloat32(priceWithFee.GetPrice().GetBasePrice())
}

func addOrSubFeeAmount(amountA, amountB decimal.Decimal, toAdd bool) decimal.Decimal {
	if toAdd {
		return amountA.Add(amountB)
	}
	return amountA.Sub(amountB)
}
