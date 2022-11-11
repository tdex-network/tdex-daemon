package domain

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking/formula"
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

// IsTradable returns true if the market is available for trading
func (m *Market) IsTradable() bool {
	return m.Tradable
}

// BaseAssetPrice returns the latest price for the base asset
func (m *Market) BaseAssetPrice() decimal.Decimal {
	if m.Price.IsZero() {
		return decimal.Zero
	}
	p, _ := decimal.NewFromString(m.Price.BasePrice)
	return p
}

// QuoteAssetPrice returns the latest price for the quote asset
func (m *Market) QuoteAssetPrice() decimal.Decimal {
	if m.Price.IsZero() {
		return decimal.Zero
	}
	p, _ := decimal.NewFromString(m.Price.QuotePrice)
	return p
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
) func(interface{}, uint64) (uint64, error) {
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
	balanceIn := baseBalance
	balanceOut := quoteBalance
	if isBuy {
		balanceIn = quoteBalance
		balanceOut = baseBalance
	}

	price := m.BaseAssetPrice()
	if isBaseAsset {
		price = m.QuoteAssetPrice()
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
	balanceIn := baseBalance
	balanceOut := quoteBalance
	if isBuy {
		balanceIn = quoteBalance
		balanceOut = baseBalance
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
	toIface := func(o formula.BalancedReservesOpts) interface{} {
		return o
	}
	strategy := m.strategy()
	basePrice, err := strategy.SpotPrice(
		toIface(formula.BalancedReservesOpts{
			BalanceIn:  quoteBalance,
			BalanceOut: baseBalance,
		}),
	)
	if err != nil {
		return
	}
	quotePrice, err := strategy.SpotPrice(
		toIface(formula.BalancedReservesOpts{
			BalanceIn:  baseBalance,
			BalanceOut: quoteBalance,
		}),
	)
	if err != nil {
		return
	}

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
