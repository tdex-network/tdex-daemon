package domain

import (
	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking/formula"
)

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

// IsStrategyPluggable returns true if the the startegy isn't automated.
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
