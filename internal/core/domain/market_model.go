package domain

import (
	"encoding/hex"
	"fmt"

	"github.com/shopspring/decimal"
	mm "github.com/tdex-network/tdex-daemon/pkg/marketmaking"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking/formula"
)

type FixedFee struct {
	BaseFee  int64
	QuoteFee int64
}

// Market defines the Market entity data structure for holding an asset pair state
type Market struct {
	// AccountIndex links a market to a HD wallet account derivation.
	AccountIndex int
	BaseAsset    string
	QuoteAsset   string
	// Precision of the market assets.
	BaseAssetPrecision  uint
	QuoteAssetPrecision uint
	// Each Market has a different fee expressed in basis point of each swap.
	Fee      int64
	FixedFee FixedFee
	// if curretly open for trades
	Tradable bool
	// Market Making strategy
	Strategy mm.MakingStrategy
	// Pluggable Price of the asset pair.
	Price Prices
}

// OutpointWithAsset contains the transaction outpoint (tx hash and vout) along with the asset hash
type OutpointWithAsset struct {
	Asset string
	Txid  string
	Vout  int
}

// Prices ...
type Prices struct {
	// how much 1 base asset is valued in quote asset.
	BasePrice decimal.Decimal
	// how much 1 quote asset is valued in base asset
	QuotePrice decimal.Decimal
}

// StrategyType is the Market making strategy type
type StrategyType int32

// PreviewInfo contains info about a price preview based on the market's current
// strategy.
type PreviewInfo struct {
	Price  Prices
	Amount uint64
	Asset  string
}

// NewMarket returns a new market with an account index, the asset pair and the
// percentage fee set.
func NewMarket(
	accountIndex int, baseAsset, quoteAsset string, feeInBasisPoint int64,
	baseAssetPrecision, quoteAssetPrecision uint,
) (*Market, error) {
	if !isValidAsset(baseAsset) {
		return nil, ErrMarketInvalidBaseAsset
	}
	if !isValidAsset(quoteAsset) {
		return nil, ErrMarketInvalidQuoteAsset
	}
	if err := validateAccountIndex(accountIndex); err != nil {
		return nil, err
	}
	if err := validateFee(feeInBasisPoint); err != nil {
		return nil, err
	}
	if err := validatePrecision(baseAssetPrecision); err != nil {
		return nil, fmt.Errorf("invalid base asset precision: %s", err)
	}
	if err := validatePrecision(quoteAssetPrecision); err != nil {
		return nil, fmt.Errorf("invalid quote asset precision: %s", err)
	}

	return &Market{
		AccountIndex: accountIndex,
		BaseAsset:    baseAsset,
		QuoteAsset:   quoteAsset,
		Fee:          feeInBasisPoint,
		Strategy:     mm.NewStrategyFromFormula(formula.BalancedReserves{}),
	}, nil
}

func isValidAsset(asset string) bool {
	buf, err := hex.DecodeString(asset)
	if err != nil {
		return false
	}
	return len(buf) == 32
}
