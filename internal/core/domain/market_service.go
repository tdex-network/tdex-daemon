package domain

import (
	"errors"
	"sort"
	"time"

	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/config"
	mm "github.com/tdex-network/tdex-daemon/pkg/marketmaking"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking/formula"
)

// AccountIndex returns the account index
func (m *Market) AccountIndex() int {
	return m.accountIndex
}

// BaseAssetHash returns the base asset hash
func (m *Market) BaseAssetHash() string {
	return m.baseAsset.assetHash
}

// QuoteAssetHash returns the quote asset hash
func (m *Market) QuoteAssetHash() string {
	return m.quoteAsset.assetHash
}

// Fee returns the selected fee
func (m *Market) Fee() int64 {
	return m.fee
}

// FeeAsset returns the selected asset to be used for market fee collection
func (m *Market) FeeAsset() string {
	return m.feeAsset
}

// MakeTradable ...
func (m *Market) MakeTradable() error {
	if !m.IsFunded() {
		return ErrNotFunded
	}

	if m.IsStrategyPluggable() && !m.IsStrategyPluggableInitialized() {
		return ErrNotPriced
	}

	m.tradable = true
	return nil
}

// MakeNotTradable ...
func (m *Market) MakeNotTradable() error {
	if !m.IsFunded() {
		return ErrNotFunded
	}

	m.tradable = false
	return nil
}

// IsTradable returns true if the market is available for trading
func (m *Market) IsTradable() bool {
	return m.tradable
}

func validateFee(basisPoint int64) error {
	if basisPoint < 1 || basisPoint > 9999 {
		return errors.New("percentage of the fee on each swap must be > 0.01 and < 99")
	}

	return nil
}

// IsNotZero ...
func (d depositedAsset) IsNotZero() bool {
	return d != depositedAsset{}
}

// IsFunded method returns true if the market contains a non empty funding tx outpoint for each asset
func (m *Market) IsFunded() bool {
	return m.baseAsset.IsNotZero() && m.quoteAsset.IsNotZero()
}

// FundMarket adds funding details given an array of outpoints and recognize quote asset
func (m *Market) FundMarket(fundingTxs []OutpointWithAsset) error {
	if m.IsFunded() {
		return nil
	}

	var baseAssetHash = config.GetString(config.BaseAssetKey)
	var otherAssetHash string

	var baseAssetTxs []OutpointWithAsset
	var otherAssetTxs []OutpointWithAsset

	assetCount := make(map[string]int)
	for _, o := range fundingTxs {
		assetCount[o.Asset]++
		if o.Asset == baseAssetHash {
			baseAssetTxs = append(baseAssetTxs, o)
		} else {
			// Potentially here could be different assets mixed
			// We chek if unique quote asset later after the loop
			otherAssetTxs = append(otherAssetTxs, o)
			otherAssetHash = o.Asset
		}
	}

	if _, ok := assetCount[baseAssetHash]; !ok {
		return errors.New("base asset is missing")
	}

	if keysNumber := len(assetCount); keysNumber != 2 {
		return errors.New("must be deposited 2 unique assets")
	}

	m.baseAsset = &depositedAsset{
		assetHash: baseAssetHash,
	}

	m.quoteAsset = &depositedAsset{
		assetHash: otherAssetHash,
	}

	return nil
}

// ChangeFee ...
func (m *Market) ChangeFee(fee int64) error {

	if !m.IsFunded() {
		return ErrNotFunded
	}

	if m.IsTradable() {
		return ErrMarketMustBeClose
	}

	if err := validateFee(fee); err != nil {
		return err
	}

	m.fee = fee
	return nil
}

// ChangeFeeAsset ...
func (m *Market) ChangeFeeAsset(asset string) error {
	// In case of empty asset hash, no updates happens and therefore it exit without error
	if asset == "" {
		return nil
	}

	if !m.IsFunded() {
		return ErrNotFunded
	}

	if m.IsTradable() {
		return ErrMarketMustBeClose
	}

	if asset != m.BaseAssetHash() && asset != m.QuoteAssetHash() {
		return errors.New("the given asset must be either the base or quote" +
			" asset in the pair")
	}

	m.feeAsset = asset
	return nil
}

// BaseAssetPrice returns the latest price for the base asset
func (m *Market) BaseAssetPrice() decimal.Decimal {
	_, price := getLatestPrice(m.basePrice)

	return decimal.Decimal(price)
}

// QuoteAssetPrice returns the latest price for the quote asset
func (m *Market) QuoteAssetPrice() decimal.Decimal {
	_, price := getLatestPrice(m.quotePrice)

	return decimal.Decimal(price)
}

// ChangeBasePrice ...
func (m *Market) ChangeBasePrice(price decimal.Decimal) error {
	if !m.IsFunded() {
		return ErrNotFunded
	}

	// TODO add logic to be sure that the price do not change to much from the latest one

	timestamp := uint64(time.Now().Unix())
	if _, ok := m.basePrice[timestamp]; ok {
		return ErrPriceExists
	}

	m.basePrice[timestamp] = Price(price)
	return nil
}

// ChangeQuotePrice ...
func (m *Market) ChangeQuotePrice(price decimal.Decimal) error {
	if !m.IsFunded() {
		return ErrNotFunded
	}

	//TODO check if the previous price is changing too much as security measure

	timestamp := uint64(time.Now().Unix())
	if _, ok := m.quotePrice[timestamp]; ok {
		return ErrPriceExists
	}

	m.quotePrice[timestamp] = Price(price)
	return nil
}

// IsZero ...
func (pt PriceByTime) IsZero() bool {
	return len(pt) == 0
}

// IsZero ...
func (p Price) IsZero() bool {
	return decimal.Decimal(p).Equal(decimal.NewFromInt(0))
}

func getLatestPrice(keyValue PriceByTime) (uint64, Price) {
	if keyValue.IsZero() {
		return uint64(0), Price(decimal.NewFromInt(0))
	}

	keys := make([]uint64, 0, len(keyValue))
	for k := range keyValue {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	latestKey := keys[len(keys)-1]
	latestValue := keyValue[latestKey]
	return latestKey, latestValue
}

// Strategy ...
func (m *Market) Strategy() mm.MakingStrategy {
	return m.strategy
}

// IsStrategyPluggable returns true if the the startegy isn't automated.
func (m *Market) IsStrategyPluggable() bool {
	return m.strategy.IsZero()
}

// IsStrategyPluggableInitialized returns true if the prices have been set.
func (m *Market) IsStrategyPluggableInitialized() bool {
	return !m.basePrice.IsZero() && !m.quotePrice.IsZero()
}

// MakeStrategyPluggable makes the current market using a given price (ie. set via UpdateMarketPrice rpc either manually or a price feed plugin)
func (m *Market) MakeStrategyPluggable() error {
	if m.IsTradable() {
		// We need the market be switched off before making this change
		return ErrMarketMustBeClose
	}

	m.strategy = mm.MakingStrategy{}

	return nil
}

// MakeStrategyBalanced makes the current market using a balanced AMM formula 50/50
func (m *Market) MakeStrategyBalanced() error {
	if m.IsTradable() {
		// We need the market be switched off before making this change
		return ErrMarketMustBeClose
	}

	m.strategy = mm.NewStrategyFromFormula(formula.BalancedReserves{})

	return nil
}
