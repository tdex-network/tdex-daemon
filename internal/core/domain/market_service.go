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

func (f FixedFee) validate() error {
	if f.BaseFee < 0 || f.QuoteFee < 0 {
		return ErrInvalidFixedFee
	}

	if (f.BaseFee > 0 && f.QuoteFee == 0) || (f.QuoteFee > 0 && f.BaseFee == 0) {
		return ErrMissingFixedFee
	}

	return nil
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
// For backward compatibility it returns whether the strategy is zero-ed or,
// on the contrary, its type is extaclty PluggableStrategyType.
func (m *Market) IsStrategyPluggable() bool {
	return m.Strategy.IsZero() || m.Strategy.Type == PluggableStrategyType
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

	m.Strategy = mm.NewStrategyFromFormula(PluggableStrategy{})
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

// ChangeFeeBasisPoint ...
func (m *Market) ChangeFeeBasisPoint(fee int64) error {
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

// ChangeFixedFee ...
func (m *Market) ChangeFixedFee(baseFee, quoteFee int64) error {
	if !m.IsFunded() {
		return ErrMarketNotFunded
	}

	if m.IsTradable() {
		return ErrMarketMustBeClosed
	}

	if err := validateFixedFee(baseFee, quoteFee); err != nil {
		return err
	}

	m.FixedFee.BaseFee = baseFee
	m.FixedFee.QuoteFee = quoteFee
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

func (m *Market) Preview(
	baseBalance, quoteBalance, amount uint64,
	isBaseAsset, isBuy bool,
) (*PreviewInfo, error) {
	if !m.IsTradable() {
		return nil, ErrMarketIsClosed
	}

	if isBaseAsset {
		if amount < uint64(m.FixedFee.BaseFee) {
			return nil, ErrMarketPreviewAmountTooLow
		}
	} else {
		if amount < uint64(m.FixedFee.QuoteFee) {
			return nil, ErrMarketPreviewAmountTooLow
		}
	}

	formula := m.formula(isBaseAsset, isBuy)
	args := m.formulaOpts(baseBalance, quoteBalance, isBaseAsset, isBuy)

	price, err := m.priceForStrategy(baseBalance, quoteBalance)
	if err != nil {
		return nil, err
	}

	previewAmount, err := formula(args, amount)
	if err != nil {
		return nil, err
	}

	previewAmount, err = m.chargeFixedFees(
		baseBalance, quoteBalance, previewAmount, isBaseAsset, isBuy,
	)
	if err != nil {
		return nil, err
	}

	previewAsset := m.BaseAsset
	if isBaseAsset {
		previewAsset = m.QuoteAsset
	}

	return &PreviewInfo{
		Price:  price,
		Amount: previewAmount,
		Asset:  previewAsset,
	}, nil
}

func (m *Market) formula(
	isBaseAsset, isBuy bool,
) func(interface{}, uint64) (uint64, error) {
	formula := m.getStrategySafe().Formula()
	if isBuy {
		if isBaseAsset {
			return formula.InGivenOut
		}
		return formula.OutGivenIn
	}

	if isBaseAsset {
		return formula.OutGivenIn
	}
	return formula.InGivenOut
}

func (m *Market) formulaOpts(
	baseBalance, quoteBalance uint64, isBaseAsset, isBuy bool,
) interface{} {
	balanceIn := baseBalance
	balanceOut := quoteBalance
	if isBuy {
		balanceIn = quoteBalance
		balanceOut = baseBalance
	}

	if m.IsStrategyPluggable() {
		price := m.BaseAssetPrice()
		if isBaseAsset {
			price = m.QuoteAssetPrice()
		}

		return PluggableStrategyOpts{
			BalanceIn:  balanceIn,
			BalanceOut: balanceOut,
			Price:      price,
			Fee:        uint64(m.Fee),
		}
	}

	return formula.BalancedReservesOpts{
		BalanceIn:           balanceIn,
		BalanceOut:          balanceOut,
		Fee:                 uint64(m.Fee),
		ChargeFeeOnTheWayIn: true,
	}
}

func (m *Market) priceForStrategy(baseBalance, quoteBalance uint64) (Prices, error) {
	if m.IsStrategyPluggable() {
		return m.Price, nil
	}

	return m.priceFromBalances(baseBalance, quoteBalance)
}

func (m *Market) priceFromBalances(
	baseBalance, quoteBalance uint64,
) (price Prices, err error) {
	toIface := func(o formula.BalancedReservesOpts) interface{} {
		return o
	}

	basePrice, err := m.getStrategySafe().Formula().SpotPrice(
		toIface(formula.BalancedReservesOpts{
			BalanceIn:  quoteBalance,
			BalanceOut: baseBalance,
		}),
	)
	if err != nil {
		return
	}
	quotePrice, err := m.getStrategySafe().Formula().SpotPrice(
		toIface(formula.BalancedReservesOpts{
			BalanceIn:  baseBalance,
			BalanceOut: quoteBalance,
		}),
	)
	if err != nil {
		return
	}

	price = Prices{
		BasePrice:  basePrice,
		QuotePrice: quotePrice,
	}
	return
}

func (m *Market) chargeFixedFees(
	baseBalance, quoteBalance, amount uint64,
	isBaseAsset, isBuy bool,
) (uint64, error) {
	if isBuy {
		if isBaseAsset {
			return amount + uint64(m.FixedFee.QuoteFee), nil
		}
		return safeSubtractFeesFromAmount(
			amount, uint64(m.FixedFee.BaseFee), baseBalance,
		)
	}

	if isBaseAsset {
		return safeSubtractFeesFromAmount(
			amount, uint64(m.FixedFee.QuoteFee), quoteBalance,
		)
	}
	return amount + uint64(m.FixedFee.BaseFee), nil
}

// getStrategySafe is a backward compatible method that returns the current
// strategy as the implementation of the mm.MakingFormula interface.
func (m *Market) getStrategySafe() mm.MakingStrategy {
	if m.IsStrategyPluggable() {
		return mm.NewStrategyFromFormula(PluggableStrategy{})
	}
	return m.Strategy
}

func validateFee(basisPoint int64) error {
	if basisPoint < 0 {
		return ErrMarketFeeTooLow
	}
	if basisPoint > 9999 {
		return ErrMarketFeeTooHigh
	}

	return nil
}

func validateFixedFee(baseFee, quoteFee int64) error {
	if baseFee < 0 || quoteFee < 0 {
		return ErrInvalidFixedFee
	}
	if (baseFee > 0 && quoteFee == 0) || (quoteFee > 0 && baseFee == 0) {
		return ErrMissingFixedFee
	}

	return nil
}

func getLatestPrice(pt Prices) (decimal.Decimal, decimal.Decimal) {
	if pt.IsZero() || pt.AreZero() {
		return decimal.NewFromInt(0), decimal.NewFromInt(0)
	}

	return pt.BasePrice, pt.QuotePrice
}

func safeSubtractFeesFromAmount(amount, fee, balance uint64) (uint64, error) {
	amountLessFees := amount
	if amountLessFees <= uint64(fee) {
		return 0, ErrMarketPreviewAmountTooLow
	}
	amountLessFees -= uint64(fee)
	if amountLessFees >= balance {
		return 0, ErrMarketPreviewAmountTooBig
	}
	return amountLessFees, nil
}
