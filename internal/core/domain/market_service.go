package domain

import (
	"fmt"
	"math"

	"github.com/shopspring/decimal"
	mm "github.com/tdex-network/tdex-daemon/pkg/marketmaking"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking/formula"
	"github.com/tdex-network/tdex-daemon/pkg/mathutil"
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

func (m *Market) IsStrategyBalanced() bool {
	return m.Strategy.Type == int(StrategyTypeBalanced)
}

// IsStrategyPluggable returns true if the the startegy isn't automated.
func (m *Market) IsStrategyPluggable() bool {
	return !m.Strategy.IsZero() && m.Strategy.Type == int(StrategyTypePluggable)
}

// IsStrategyPluggableInitialized returns true if the prices have been set.
func (m *Market) IsStrategyPluggableInitialized() bool {
	return !m.Price.AreZero()
}

// VerifyMarketFunds verifies that the provided list of outpoints (each
// including the unblinded asset) are valid funds of the market, by checking
// that their assets match those of the market.
func (m *Market) VerifyMarketFunds(fundingTxs []OutpointWithAsset) error {
	assetCount := make(map[string]int)
	for _, o := range fundingTxs {
		assetCount[o.Asset]++
	}

	if len(assetCount) > 2 {
		return ErrMarketTooManyAssets
	}

	_, baseFundsOk := assetCount[m.BaseAsset]
	_, quoteFundsOk := assetCount[m.QuoteAsset]

	// balanced strategy requires funds of both assets to be non zero.
	if m.IsStrategyBalanced() {
		if !baseFundsOk {
			return ErrMarketMissingBaseAsset
		}
		if !quoteFundsOk {
			return ErrMarketMissingQuoteAsset
		}

		return nil
	}

	// for other strategies, it's ok to have single asset funds instead.
	if !baseFundsOk && !quoteFundsOk {
		return ErrMarketMissingFunds
	}

	return nil
}

// MakeTradable ...
func (m *Market) MakeTradable() error {
	if m.IsStrategyPluggable() && !m.IsStrategyPluggableInitialized() {
		return ErrMarketNotPriced
	}

	m.Tradable = true
	return nil
}

// MakeNotTradable ...
func (m *Market) MakeNotTradable() error {
	m.Tradable = false
	return nil
}

// MakeStrategyPluggable makes the current market using a given price
// (ie. set via UpdateMarketPrice rpc either manually or a price feed plugin)
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
	if m.IsTradable() {
		return ErrMarketMustBeClosed
	}

	if err := validateFixedFee(baseFee, quoteFee); err != nil {
		return err
	}

	if baseFee >= 0 {
		m.FixedFee.BaseFee = baseFee
	}
	if quoteFee >= 0 {
		m.FixedFee.QuoteFee = quoteFee
	}
	return nil
}

// ChangeBasePrice ...
func (m *Market) ChangeAssetPrecision(
	baseAssetPrecision, quoteAssetPrecision int,
) error {
	if m.IsTradable() {
		return ErrMarketMustBeClosed
	}

	if baseAssetPrecision >= 0 {
		if err := validatePrecision(uint(baseAssetPrecision)); err != nil {
			return fmt.Errorf("invalid base asset precision: %s", err)
		}
	}
	if quoteAssetPrecision >= 0 {
		if err := validatePrecision(uint(quoteAssetPrecision)); err != nil {
			return fmt.Errorf("invalid quote asset precision: %s", err)
		}
	}

	if baseAssetPrecision >= 0 {
		m.BaseAssetPrecision = uint(baseAssetPrecision)
	}
	if quoteAssetPrecision >= 0 {
		m.QuoteAssetPrecision = uint(quoteAssetPrecision)
	}
	return nil
}

// ChangeBasePrice ...
func (m *Market) ChangeBasePrice(price decimal.Decimal) error {
	zero := decimal.NewFromInt(0)
	if price.LessThanOrEqual(zero) {
		return ErrMarketInvalidBasePrice
	}

	m.Price.BasePrice = price
	return nil
}

// ChangeQuotePrice ...
func (m *Market) ChangeQuotePrice(price decimal.Decimal) error {
	zero := decimal.NewFromInt(0)
	if price.LessThanOrEqual(zero) {
		return ErrMarketInvalidQuotePrice
	}

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
	var args interface{}
	if m.IsStrategyPluggable() {
		args = m.formulaOptsForPluggable(baseBalance, quoteBalance, isBaseAsset, isBuy)
	} else {
		args = m.formulaOptsForBalanced(baseBalance, quoteBalance, isBaseAsset, isBuy)
	}

	price, err := m.priceForStrategy(baseBalance, quoteBalance)
	if err != nil {
		return nil, err
	}

	assetPrecision := math.Pow10(int(m.BaseAssetPrecision))
	if !isBaseAsset {
		assetPrecision = math.Pow10(int(m.QuoteAssetPrecision))
	}
	amountDecimal := mathutil.Div(amount, uint64(assetPrecision))

	previewAmount, err := formula(args, amountDecimal)
	if err != nil {
		return nil, err
	}

	previewAsset := m.BaseAsset
	previewAssetPrecision := math.Pow10(int(m.BaseAssetPrecision))
	if isBaseAsset {
		previewAsset = m.QuoteAsset
		previewAssetPrecision = math.Pow10(int(m.QuoteAssetPrecision))
	}

	previewAmountInSats := previewAmount.Mul(
		decimal.NewFromFloat(previewAssetPrecision),
	).BigInt().Uint64()
	previewAmountInSats, err = m.chargeFixedFees(
		baseBalance, quoteBalance, previewAmountInSats, isBaseAsset, isBuy,
	)
	if err != nil {
		return nil, err
	}
	if previewAmountInSats == 0 {
		return nil, ErrMarketPreviewAmountTooLow
	}

	return &PreviewInfo{
		Price:  price,
		Amount: previewAmountInSats,
		Asset:  previewAsset,
	}, nil
}

func (m *Market) SpotPrice(
	baseBalance, quoteBalance uint64,
) (Prices, error) {
	return m.priceForStrategy(baseBalance, quoteBalance)
}

func (m *Market) formula(
	isBaseAsset, isBuy bool,
) func(interface{}, decimal.Decimal) (decimal.Decimal, error) {
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

func (m *Market) formulaOptsForPluggable(
	baseBalance, quoteBalance uint64, isBaseAsset, isBuy bool,
) interface{} {
	bp := uint64(math.Pow10(int(m.BaseAssetPrecision)))
	qp := uint64(math.Pow10(int(m.QuoteAssetPrecision)))
	balanceIn := mathutil.Div(baseBalance, bp)
	balanceOut := mathutil.Div(quoteBalance, qp)
	if isBuy {
		balanceIn, balanceOut = balanceOut, balanceIn
	}

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

func (m *Market) formulaOptsForBalanced(
	baseBalance, quoteBalance uint64, isBaseAsset, isBuy bool,
) interface{} {
	bp := uint64(math.Pow10(int(m.BaseAssetPrecision)))
	qp := uint64(math.Pow10(int(m.QuoteAssetPrecision)))
	balanceIn := mathutil.Div(baseBalance, bp)
	balanceOut := mathutil.Div(quoteBalance, qp)
	if isBuy {
		balanceIn, balanceOut = balanceOut, balanceIn
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
	bp := uint64(math.Pow10(int(m.BaseAssetPrecision)))
	qp := uint64(math.Pow10(int(m.QuoteAssetPrecision)))
	balanceIn := mathutil.Div(quoteBalance, qp)
	balanceOut := mathutil.Div(baseBalance, bp)
	basePrice, err := m.getStrategySafe().Formula().SpotPrice(
		formula.BalancedReservesOpts{
			BalanceIn:  balanceIn,
			BalanceOut: balanceOut,
		},
	)
	if err != nil {
		return
	}
	balanceIn, balanceOut = balanceOut, balanceIn
	quotePrice, err := m.getStrategySafe().Formula().SpotPrice(
		formula.BalancedReservesOpts{
			BalanceIn:  balanceIn,
			BalanceOut: balanceOut,
		},
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
	if baseFee < -1 || quoteFee < -1 {
		return ErrInvalidFixedFee
	}

	return nil
}

func validatePrecision(precision uint) error {
	if precision > 8 {
		return ErrInvalidPrecision
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
