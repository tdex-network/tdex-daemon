package domain

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/shopspring/decimal"
)

// Market defines the Market entity data structure for holding an asset pair state.
type Market struct {
	// Base asset in hex format.
	BaseAsset string
	// Quote asset in hex format.
	QuoteAsset string
	// Name of the market.
	Name string
	// Percentage fee expressed in basis points.
	PercentageFee uint32
	// Fixed fee amount expressed in satoshi for both assets.
	FixedFee FixedFee
	// if curretly open for trades
	Tradable bool
	// Market Making strategy type
	StrategyType int
	// Pluggable Price of the asset pair.
	Price MarketPrice
}

type FixedFee struct {
	BaseFee  uint64
	QuoteFee uint64
}

// MarketPrice represents base and quote market price
type MarketPrice struct {
	// how much 1 base asset is valued in quote asset.
	BasePrice string
	// how much 1 quote asset is valued in base asset
	QuotePrice string
}

func (mp MarketPrice) IsZero() bool {
	empty := MarketPrice{}
	if mp == empty {
		return true
	}
	bp, _ := decimal.NewFromString(mp.BasePrice)
	qp, _ := decimal.NewFromString(mp.QuotePrice)
	return bp.IsZero() && qp.IsZero()
}

// PreviewInfo contains info about a price preview based on the market's current
// strategy.
type PreviewInfo struct {
	Price  MarketPrice
	Amount uint64
	Asset  string
}

// NewMarket returns a new market with an account index, the asset pair and the
// percentage fee set.
func NewMarket(
	baseAsset, quoteAsset string, percentageFee uint32,
) (*Market, error) {
	if !isValidAsset(baseAsset) {
		return nil, ErrMarketInvalidBaseAsset
	}
	if !isValidAsset(quoteAsset) {
		return nil, ErrMarketInvalidQuoteAsset
	}
	if !isValidPercentageFee(int(percentageFee)) {
		return nil, ErrMarketInvalidPercentageFee
	}
	accountName := makeAccountName(baseAsset, quoteAsset)

	return &Market{
		BaseAsset:     baseAsset,
		QuoteAsset:    quoteAsset,
		Name:          accountName,
		PercentageFee: percentageFee,
		StrategyType:  StrategyTypeBalanced,
	}, nil
}

func makeAccountName(baseAsset, quoteAsset string) string {
	buf, _ := hex.DecodeString(baseAsset + quoteAsset)
	return hex.EncodeToString(btcutil.Hash160(buf))[:5]
}

func isValidAsset(asset string) bool {
	buf, err := hex.DecodeString(asset)
	if err != nil {
		return false
	}
	return len(buf) == 32
}

func isValidPercentageFee(basisPoint int) bool {
	return basisPoint >= 0 && basisPoint <= 9999
}

func isValidFixedFee(baseFee, quoteFee int) bool {
	return baseFee >= -1 && quoteFee >= -1
}
