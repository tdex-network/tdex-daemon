package domain

import (
	"encoding/hex"
	"math"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking/formula"
	"github.com/tdex-network/tdex-daemon/pkg/mathutil"
)

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
	Price  MarketPrice
	Amount uint64
	Asset  string
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

// NewMarket returns a new market with an account index, the asset pair and the
// percentage fee set.
func NewMarket(
	baseAsset, quoteAsset string, percentageFee uint32,
	baseAssetPrecision, quoteAssetPrecision uint,
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
	if !isValidPrecision(baseAssetPrecision) {
		return nil, ErrMarketInvalidBaseAssetPrecision
	}
	if !isValidPrecision(quoteAssetPrecision) {
		return nil, ErrMarketInvalidQuoteAssetPrecision
	}
	accountName := makeAccountName(baseAsset, quoteAsset)

	return &Market{
		BaseAsset:           baseAsset,
		QuoteAsset:          quoteAsset,
		Name:                accountName,
		PercentageFee:       percentageFee,
		StrategyType:        StrategyTypeBalanced,
		BaseAssetPrecision:  baseAssetPrecision,
		QuoteAssetPrecision: quoteAssetPrecision,
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
func (m *Market) ChangePercentageFee(fee uint32) error {
	if m.IsTradable() {
		return ErrMarketIsOpen
	}

	if !isValidPercentageFee(int(fee)) {
		return ErrMarketInvalidPercentageFee
	}

	m.PercentageFee = fee
	return nil
}

// ChangeFixedFee updates market's fixed fee to those given.
func (m *Market) ChangeFixedFee(baseFee, quoteFee int64) error {
	if m.IsTradable() {
		return ErrMarketIsOpen
	}

	if !isValidFixedFee(int(baseFee), int(quoteFee)) {
		return ErrMarketInvalidFixedFee
	}

	if baseFee >= 0 {
		m.FixedFee.BaseFee = uint64(baseFee)
	}
	if quoteFee >= 0 {
		m.FixedFee.QuoteFee = uint64(quoteFee)
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
	isBaseAsset, isBuy bool,
) (*PreviewInfo, error) {
	if !m.IsTradable() {
		return nil, ErrMarketIsClosed
	}

	if isBaseAsset {
		if amount < m.FixedFee.BaseFee {
			return nil, ErrMarketPreviewAmountTooLow
		}
	} else {
		if amount < m.FixedFee.QuoteFee {
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
		Fee:        uint64(m.PercentageFee),
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
		Fee:                 uint64(m.PercentageFee),
		ChargeFeeOnTheWayIn: true,
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
	balanceIn := mathutil.Div(quoteBalance, qp)
	balanceOut := mathutil.Div(baseBalance, bp)
	basePrice, err := m.strategy().SpotPrice(
		formula.BalancedReservesOpts{
			BalanceIn:  balanceIn,
			BalanceOut: balanceOut,
		},
	)
	if err != nil {
		return
	}
	balanceIn, balanceOut = balanceOut, balanceIn
	quotePrice, err := m.strategy().SpotPrice(
		formula.BalancedReservesOpts{
			BalanceIn:  balanceIn,
			BalanceOut: balanceOut,
		},
	)

	price = MarketPrice{
		BasePrice:  basePrice.String(),
		QuotePrice: quotePrice.String(),
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

func isValidPrecision(precision uint) bool {
	return int(precision) >= 0 && precision <= 8
}
