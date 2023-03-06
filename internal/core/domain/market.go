package domain

import (
	"encoding/hex"
	"fmt"
	"math"

	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking/formula"
	"github.com/tdex-network/tdex-daemon/pkg/mathutil"
)

type MarketFee struct {
	BaseAsset  uint64
	QuoteAsset uint64
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
	return mp.GetBasePrice().IsZero() && mp.GetQuotePrice().IsZero()
}
func (mp MarketPrice) GetBasePrice() decimal.Decimal {
	p, _ := decimal.NewFromString(mp.BasePrice)
	return p
}
func (mp MarketPrice) GetQuotePrice() decimal.Decimal {
	p, _ := decimal.NewFromString(mp.QuotePrice)
	return p
}

// PreviewInfo contains info about a price preview based on the market's current
// strategy.
type PreviewInfo struct {
	Price     MarketPrice
	Amount    uint64
	Asset     string
	FeeAsset  string
	FeeAmount uint64
}

// Market defines the Market entity data structure for holding an asset pair state.
type Market struct {
	// Base asset in hex format.
	BaseAsset string
	// Quote asset in hex format.
	QuoteAsset string
	// Name of the market.
	Name string
	// Precison of the base asset.
	BaseAssetPrecision uint
	// Precison of the quote asset.
	QuoteAssetPrecision uint
	// Percentage fee expressed in basis points for both assets.
	PercentageFee MarketFee
	// Fixed fee amount expressed in satoshis for both assets.
	FixedFee MarketFee
	// if curretly open for trades
	Tradable bool
	// Market Making strategy type
	StrategyType int
	// Pluggable Price of the asset pair.
	Price MarketPrice
}

// NewMarket returns a new market with an account index, the asset pair and the
// percentage fee set.
func NewMarket(
	baseAsset, quoteAsset, name string,
	basePercentageFee, quotePercentageFee uint64,
	baseAssetPrecision, quoteAssetPrecision uint,
) (*Market, error) {
	if !isValidAsset(baseAsset) {
		return nil, ErrMarketInvalidBaseAsset
	}
	if !isValidAsset(quoteAsset) {
		return nil, ErrMarketInvalidQuoteAsset
	}
	if !isValidPercentageFee(
		int64(basePercentageFee), int64(quotePercentageFee),
	) {
		return nil, ErrMarketInvalidPercentageFee
	}
	if !isValidPrecision(baseAssetPrecision) {
		return nil, ErrMarketInvalidBaseAssetPrecision
	}
	if !isValidPrecision(quoteAssetPrecision) {
		return nil, ErrMarketInvalidQuoteAssetPrecision
	}

	return &Market{
		BaseAsset:           baseAsset,
		QuoteAsset:          quoteAsset,
		Name:                name,
		StrategyType:        StrategyTypeBalanced,
		BaseAssetPrecision:  baseAssetPrecision,
		QuoteAssetPrecision: quoteAssetPrecision,
		PercentageFee: MarketFee{
			BaseAsset:  basePercentageFee,
			QuoteAsset: quotePercentageFee,
		},
	}, nil
}

// IsTradable returns true if the market is available for trading
func (m *Market) IsTradable() bool {
	return m.Tradable
}

func (m *Market) IsStrategyBalanced() bool {
	return m.StrategyType == StrategyTypeBalanced
}

// IsStrategyPluggable returns true if the strategy isn't automated.
func (m *Market) IsStrategyPluggable() bool {
	return m.StrategyType == StrategyTypePluggable
}

// MakeTradable updates the status of the market to tradable.
func (m *Market) MakeTradable() error {
	if m.IsStrategyPluggable() && m.Price.IsZero() {
		return ErrMarketNotPriced
	}

	m.Tradable = true
	return nil
}

// MakeNotTradable updates the status of the market to not tradable.
func (m *Market) MakeNotTradable() {
	m.Tradable = false
}

// MakeStrategyPluggable makes the current market using a given price
// (ie. set via UpdateMarketPrice rpc either manually or a price feed plugin)
func (m *Market) MakeStrategyPluggable() error {
	if m.IsTradable() {
		// We need the market be switched off before making this change
		return ErrMarketIsOpen
	}

	m.StrategyType = StrategyTypePluggable
	m.Price = MarketPrice{}

	return nil
}

// MakeStrategyBalanced makes the current market using a balanced AMM formula 50/50
func (m *Market) MakeStrategyBalanced() error {
	if m.IsTradable() {
		// We need the market be switched off before making this change
		return ErrMarketIsOpen
	}

	m.StrategyType = StrategyTypeBalanced

	return nil
}

// ChangePercentageFee updates market's perentage fee to the given one.
func (m *Market) ChangePercentageFee(baseFee, quoteFee int64) error {
	if m.IsTradable() {
		return ErrMarketIsOpen
	}

	if !isValidPercentageFee(baseFee, quoteFee) {
		return ErrMarketInvalidPercentageFee
	}

	if baseFee >= 0 {
		m.PercentageFee.BaseAsset = uint64(baseFee)
	}
	if quoteFee >= 0 {
		m.PercentageFee.QuoteAsset = uint64(quoteFee)
	}
	return nil
}

// ChangeFixedFee updates market's fixed fee to those given.
func (m *Market) ChangeFixedFee(baseFee, quoteFee int64) error {
	if m.IsTradable() {
		return ErrMarketIsOpen
	}

	if !isValidFixedFee(baseFee, quoteFee) {
		return ErrMarketInvalidFixedFee
	}

	if baseFee >= 0 {
		m.FixedFee.BaseAsset = uint64(baseFee)
	}
	if quoteFee >= 0 {
		m.FixedFee.QuoteAsset = uint64(quoteFee)
	}
	return nil
}

// ChangeBasePrice updates the price of market's base asset.
func (m *Market) ChangePrice(basePrice, quotePrice decimal.Decimal) error {
	if basePrice.LessThanOrEqual(decimal.Zero) {
		return ErrMarketInvalidBasePrice
	}
	if quotePrice.LessThanOrEqual(decimal.Zero) {
		return ErrMarketInvalidQuotePrice
	}

	m.Price.BasePrice = basePrice.String()
	m.Price.QuotePrice = quotePrice.String()
	return nil
}

func (m *Market) ChangeAssetPrecision(
	baseAssetPrecision, quoteAssetPrecision int,
) error {
	if m.IsTradable() {
		return ErrMarketIsOpen
	}

	if baseAssetPrecision >= 0 {
		if !isValidPrecision(uint(baseAssetPrecision)) {
			return ErrMarketInvalidBaseAssetPrecision
		}
	}
	if quoteAssetPrecision >= 0 {
		if !isValidPrecision(uint(quoteAssetPrecision)) {
			return ErrMarketInvalidQuoteAssetPrecision
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

func (m *Market) Preview(
	baseBalance, quoteBalance, amount uint64,
	asset, feeAsset string, isBuy bool,
) (*PreviewInfo, error) {
	if !m.IsTradable() {
		return nil, ErrMarketIsClosed
	}
	if asset != m.BaseAsset && asset != m.QuoteAsset {
		return nil, fmt.Errorf("asset must be either base or quote asset")
	}
	if feeAsset != m.BaseAsset && feeAsset != m.QuoteAsset {
		return nil, fmt.Errorf("fee asset must be either base or quote asset")
	}

	isBaseAsset := asset == m.BaseAsset

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

	if previewAmountInSats == 0 {
		return nil, ErrMarketPreviewAmountTooLow
	}

	amountsByAsset := map[string]uint64{
		asset:        amount,
		previewAsset: previewAmountInSats,
	}
	amountForFees := amountsByAsset[feeAsset]
	previewFeeAmount, err := m.previewFees(amountForFees, feeAsset, isBuy)
	if err != nil {
		return nil, err
	}

	return &PreviewInfo{
		Price:     price,
		Amount:    previewAmountInSats,
		Asset:     previewAsset,
		FeeAsset:  feeAsset,
		FeeAmount: previewFeeAmount,
	}, nil
}

func (m *Market) SpotPrice(
	baseBalance, quoteBalance uint64,
) (MarketPrice, error) {
	return m.priceForStrategy(baseBalance, quoteBalance)
}

func (m *Market) strategy() marketmaking.MakingFormula {
	switch m.StrategyType {
	case StrategyTypePluggable:
		return marketmaking.NewPluggableFormula()
	case StrategyTypeBalanced:
		fallthrough
	default:
		return marketmaking.NewBalancedReservedFormula()
	}
}

func (m *Market) formula(
	isBaseAsset, isBuy bool,
) func(interface{}, decimal.Decimal) (decimal.Decimal, error) {
	strategy := m.strategy()
	if isBuy {
		if isBaseAsset {
			return strategy.InGivenOut
		}
		return strategy.OutGivenIn
	}

	if isBaseAsset {
		return strategy.OutGivenIn
	}
	return strategy.InGivenOut
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

	price := m.Price.GetBasePrice()
	if isBaseAsset {
		price = m.Price.GetQuotePrice()
	}

	return formula.PluggableOpts{
		BalanceIn:  balanceIn,
		BalanceOut: balanceOut,
		Price:      price,
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
		BalanceIn:  balanceIn,
		BalanceOut: balanceOut,
	}
}

func (m *Market) priceForStrategy(
	baseBalance, quoteBalance uint64,
) (MarketPrice, error) {
	if m.IsStrategyPluggable() {
		return m.Price, nil
	}

	return m.priceFromBalances(baseBalance, quoteBalance)
}

func (m *Market) priceFromBalances(
	baseBalance, quoteBalance uint64,
) (price MarketPrice, err error) {
	bp := uint64(math.Pow10(int(m.BaseAssetPrecision)))
	qp := uint64(math.Pow10(int(m.QuoteAssetPrecision)))
	balanceIn := mathutil.Div(baseBalance, bp)
	balanceOut := mathutil.Div(quoteBalance, qp)
	quotePrice, err := m.strategy().SpotPrice(
		formula.BalancedReservesOpts{
			BalanceIn:  balanceIn,
			BalanceOut: balanceOut,
		},
	)
	if err != nil {
		return
	}
	basePrice := decimal.NewFromInt(1).Div(quotePrice)

	price = MarketPrice{
		BasePrice:  basePrice.String(),
		QuotePrice: quotePrice.String(),
	}
	return
}

func (m *Market) previewFees(amount uint64, asset string, isBuy bool) (uint64, error) {
	percentageFee := m.PercentageFee.BaseAsset
	fixedFee := m.FixedFee.BaseAsset
	if asset == m.QuoteAsset {
		percentageFee = m.PercentageFee.QuoteAsset
		fixedFee = m.FixedFee.QuoteAsset
	}

	fee := decimal.NewFromInt(int64(percentageFee)).Div(decimal.NewFromInt(10000))
	feeAmount := decimal.NewFromInt(int64(amount)).Mul(fee).BigInt().Uint64()
	feeAmount += fixedFee
	// Fees must always be added on the amount sent by the proposer and
	// subtracted on the amount he receives, ie. fees are going to be subtracted
	// if he's BUYing and wants fees on base amount of the preview, or if he's
	// SELLing and wants fees on quote amount.
	feesToSubtract := (isBuy && asset == m.BaseAsset) ||
		(!isBuy && asset == m.QuoteAsset)

	if feesToSubtract && feeAmount >= amount {
		return 0, ErrMarketPreviewAmountTooLow
	}

	return feeAmount, nil
}

func isValidAsset(asset string) bool {
	buf, err := hex.DecodeString(asset)
	if err != nil {
		return false
	}
	return len(buf) == 32
}

func isValidPercentageFee(baseFee, quoteFee int64) bool {
	isValid := func(v int64) bool {
		return (v >= 0 && v <= 9999)
	}
	return isValid(baseFee) && isValid(quoteFee)
}

func isValidFixedFee(baseFee, quoteFee int64) bool {
	isValid := func(v int64) bool {
		return v >= -1
	}
	return isValid(baseFee) && isValid(quoteFee)
}

func isValidPrecision(precision uint) bool {
	return int(precision) >= 0 && precision <= 8
}
