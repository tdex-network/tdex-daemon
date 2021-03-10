package domain

import (
	"github.com/shopspring/decimal"
	mm "github.com/tdex-network/tdex-daemon/pkg/marketmaking"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking/formula"
)

// IsZero ...
func (p Prices) IsZero() bool {
	return p == Prices{}
}

// AreZero ...
func (p Prices) AreZero() bool {
	if p.IsZero() {
		return true
	}

	return decimal.Decimal(p.BasePrice).Equal(decimal.NewFromInt(0)) && decimal.Decimal(p.QuotePrice).Equal(decimal.NewFromInt(0))
}

// IsTradable returns true if the market is available for trading
func (m *Market) IsTradable() bool {
	return m.Tradable
}

// IsFunded method returns true if the market contains a non empty funding tx outpoint for each asset
func (m *Market) IsFunded() bool {
	return m.BaseAsset != "" && m.QuoteAsset != ""
}

// BaseAssetPrice returns the latest price for the base asset
func (m *Market) BaseAssetPrice() decimal.Decimal {
	basePrice, _ := getLatestPrice(m.Price)

	return decimal.Decimal(basePrice)
}

// QuoteAssetPrice returns the latest price for the quote asset
func (m *Market) QuoteAssetPrice() decimal.Decimal {
	_, quotePrice := getLatestPrice(m.Price)

	return decimal.Decimal(quotePrice)
}

// IsStrategyPluggable returns true if the the startegy isn't automated.
func (m *Market) IsStrategyPluggable() bool {
	return m.Strategy.IsZero()
}

// IsStrategyPluggableInitialized returns true if the prices have been set.
func (m *Market) IsStrategyPluggableInitialized() bool {
	return !m.Price.AreZero()
}

// FundMarket adds the assets of market from the given array of outpoints.
// Since the list of outpoints can contain an infinite number of utxos with
// different  assets, they're fistly indexed by their asset, then the market's
// base asset is updated if found in the list, otherwise only the very first
// asset type is used as the market's quote asset, discarding the others that
// should be manually transferred to some other address because they won't be
// used by the daemon.
func (m *Market) FundMarket(fundingTxs []OutpointWithAsset, baseAssetHash string) error {
	if m.IsFunded() {
		return nil
	}

	assetCount := make(map[string]int)
	for _, o := range fundingTxs {
		assetCount[o.Asset]++
	}

	if _, ok := assetCount[baseAssetHash]; !ok {
		return ErrMarketMissingBaseAsset
	}
	if len(assetCount) < 2 {
		return ErrMarketMissingQuoteAsset
	}

	if len(assetCount) > 2 {
		return ErrMarketTooManyAssets
	}

	for asset := range assetCount {
		if asset == baseAssetHash {
			m.BaseAsset = baseAssetHash
		} else {
			m.QuoteAsset = asset
		}
	}

	return nil
}

// MakeTradable ...
func (m *Market) MakeTradable() error {
	if !m.IsFunded() {
		return ErrMarketNotFunded
	}

	if m.IsStrategyPluggable() && !m.IsStrategyPluggableInitialized() {
		return ErrMarketNotPriced
	}

	m.Tradable = true
	return nil
}

// MakeNotTradable ...
func (m *Market) MakeNotTradable() error {
	if !m.IsFunded() {
		return ErrMarketNotFunded
	}

	m.Tradable = false
	return nil
}

// MakeStrategyPluggable makes the current market using a given price
//(ie. set via UpdateMarketPrice rpc either manually or a price feed plugin)
func (m *Market) MakeStrategyPluggable() error {
	if m.IsTradable() {
		// We need the market be switched off before making this change
		return ErrMarketMustBeClosed
	}

	m.Strategy = mm.MakingStrategy{}
	m.ChangeBasePrice(decimal.NewFromInt(0))
	m.ChangeQuotePrice(decimal.NewFromInt(0))

	return nil
}

// MakeStrategyBalanced makes the current market using a balanced AMM formula 50/50
func (m *Market) MakeStrategyBalanced() error {
	if m.IsTradable() {
		// We need the market be switched off before making this change
		return ErrMarketMustBeClosed
	}

	m.Strategy = mm.NewStrategyFromFormula(formula.BalancedReserves{})

	return nil
}

// ChangeFee ...
func (m *Market) ChangeFee(fee int64) error {
	if !m.IsFunded() {
		return ErrMarketNotFunded
	}

	if m.IsTradable() {
		return ErrMarketMustBeClosed
	}

	if err := validateFee(fee); err != nil {
		return err
	}

	m.Fee = fee
	return nil
}

// ChangeBasePrice ...
func (m *Market) ChangeBasePrice(price decimal.Decimal) error {
	if !m.IsFunded() {
		return ErrMarketNotFunded
	}

	// TODO add logic to be sure that the price do not change to much from the latest one

	m.Price.BasePrice = price
	return nil
}

// ChangeQuotePrice ...
func (m *Market) ChangeQuotePrice(price decimal.Decimal) error {
	if !m.IsFunded() {
		return ErrMarketNotFunded
	}

	//TODO check if the previous price is changing too much as security measure

	m.Price.QuotePrice = price
	return nil
}

func validateFee(basisPoint int64) error {
	if basisPoint < 1 {
		return ErrMarketFeeTooLow
	}
	if basisPoint > 9999 {
		return ErrMarketFeeTooHigh
	}

	return nil
}

func getLatestPrice(pt Prices) (decimal.Decimal, decimal.Decimal) {
	if pt.IsZero() || pt.AreZero() {
		return decimal.NewFromInt(0), decimal.NewFromInt(0)
	}

	return pt.BasePrice, pt.QuotePrice
}
