package v091domain

import (
	"encoding/hex"
	"errors"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/shopspring/decimal"
)

const (
	FeeAccount = iota
	PersonalAccount
	FeeFragmenterAccount
	MarketFragmenterAccount

	BalancedReservesType = 1
)

type Market struct {
	AccountIndex        int
	BaseAsset           string
	QuoteAsset          string
	BaseAssetPrecision  uint
	QuoteAssetPrecision uint
	Fee                 int64
	FixedFee            FixedFee
	Tradable            bool
	Strategy            MakingStrategy
	Price               Prices
}

func (m Market) AccountName() string {
	buf, _ := hex.DecodeString(m.BaseAsset + m.QuoteAsset)
	return hex.EncodeToString(btcutil.Hash160(buf))[:5]
}

type Prices struct {
	BasePrice  decimal.Decimal
	QuotePrice decimal.Decimal
}

type FixedFee struct {
	BaseFee  int64
	QuoteFee int64
}

type MakingStrategy struct {
	Type    int
	formula MakingFormula
}

type MakingFormula interface {
	SpotPrice(spotPriceOpts interface{}) (spotPrice decimal.Decimal, err error)
	OutGivenIn(outGivenInOpts interface{}, amountIn decimal.Decimal) (amountOut decimal.Decimal, err error)
	InGivenOut(inGivenOutOpts interface{}, amountOut decimal.Decimal) (amountIn decimal.Decimal, err error)
	FormulaType() int
}

func NewStrategyFromFormula(
	formula MakingFormula,
) MakingStrategy {
	strategy := MakingStrategy{
		Type:    formula.FormulaType(),
		formula: formula,
	}

	return strategy
}

type StrategyType int32

const (
	StrategyTypePluggable StrategyType = 0
)

var (
	ErrInvalidOptsType = errors.New("opts must be of type BalancedReservesOpts")
	ErrAmountTooLow    = errors.New("provided amount is too low")
	ErrAmountTooBig    = errors.New("provided amount is too big")
	ErrBalanceTooLow   = errors.New("reserve balance amount is too low")

	tenThousands = decimal.NewFromInt(10000)
)

type BalancedReservesOpts struct {
	BalanceIn           decimal.Decimal
	BalanceOut          decimal.Decimal
	WeightIn            uint64
	WeightOut           uint64
	Fee                 uint64
	ChargeFeeOnTheWayIn bool
}

type BalancedReserves struct{}

func (BalancedReserves) SpotPrice(_opts interface{}) (spotPrice decimal.Decimal, err error) {
	opts, ok := _opts.(BalancedReservesOpts)
	if !ok {
		err = ErrInvalidOptsType
		return
	}

	if opts.BalanceIn.Equals(decimal.Zero) || opts.BalanceOut.Equals(decimal.Zero) {
		err = ErrBalanceTooLow
		return
	}

	spotPrice = opts.BalanceOut.Div(opts.BalanceIn)
	return
}

func (BalancedReserves) OutGivenIn(
	_opts interface{}, amountIn decimal.Decimal,
) (amountOut decimal.Decimal, err error) {
	opts, ok := _opts.(BalancedReservesOpts)
	if !ok {
		err = ErrInvalidOptsType
		return
	}
	if opts.BalanceIn.Equal(decimal.Zero) || opts.BalanceOut.Equal(decimal.Zero) {
		err = ErrBalanceTooLow
		return
	}

	percentageFee := decimal.NewFromInt(int64(opts.Fee)).Div(tenThousands)
	if opts.ChargeFeeOnTheWayIn {
		amountIn = amountIn.Mul(decimal.NewFromInt(1).Sub(percentageFee))
	}
	if amountIn.LessThanOrEqual(decimal.Zero) {
		err = ErrAmountTooLow
		return
	}

	amount := opts.BalanceOut.Mul(decimal.NewFromInt(1).Sub(
		opts.BalanceIn.Div(opts.BalanceIn.Add(amountIn)),
	))
	if !opts.ChargeFeeOnTheWayIn {
		amount = amount.Mul(decimal.NewFromInt(1).Sub(percentageFee))
	}
	amount = amount.Round(8)
	if amount.LessThanOrEqual(decimal.Zero) {
		err = ErrAmountTooLow
		return
	}
	if amount.GreaterThanOrEqual(opts.BalanceOut) {
		err = ErrAmountTooBig
		return
	}

	amountOut = amount
	return
}

func (BalancedReserves) InGivenOut(
	_opts interface{}, amountOut decimal.Decimal,
) (amountIn decimal.Decimal, err error) {
	opts, ok := _opts.(BalancedReservesOpts)
	if !ok {
		err = ErrInvalidOptsType
		return
	}
	if opts.BalanceIn.Equals(decimal.Zero) || opts.BalanceOut.Equals(decimal.Zero) {
		err = ErrBalanceTooLow
		return
	}

	percentageFee := decimal.NewFromInt(int64(opts.Fee)).Div(tenThousands)
	if !opts.ChargeFeeOnTheWayIn {
		amountOut = amountOut.Mul(decimal.NewFromInt(1).Add(percentageFee))
	}
	if amountOut.Equals(decimal.Zero) {
		err = ErrAmountTooLow
		return
	}
	if amountOut.GreaterThanOrEqual(opts.BalanceOut) {
		err = ErrAmountTooBig
		return
	}

	amount := opts.BalanceIn.Mul(
		opts.BalanceOut.Div(opts.BalanceOut.Sub(amountOut)).Sub(
			decimal.NewFromInt(1),
		),
	)

	if opts.ChargeFeeOnTheWayIn {
		amount = amount.Mul(decimal.NewFromInt(1).Add(percentageFee))
	}
	amount = amount.Round(8)
	if amount.LessThanOrEqual(decimal.Zero) {
		err = ErrAmountTooLow
		return
	}

	amountIn = amount
	return
}

func (BalancedReserves) FormulaType() int {
	return BalancedReservesType
}

var (
	ErrMarketPreviewAmountTooBig = errors.New("provided amount is too big")
	ErrMarketPreviewAmountTooLow = errors.New("provided amount is too low")

	tenThousand = decimal.NewFromInt(10000)
)

type PluggableStrategyOpts struct {
	BalanceIn  decimal.Decimal
	BalanceOut decimal.Decimal
	Price      decimal.Decimal
	Fee        uint64
}

type PluggableStrategy struct{}

func (s PluggableStrategy) SpotPrice(_ interface{}) (decimal.Decimal, error) {
	return decimal.Zero, nil
}

func (s PluggableStrategy) OutGivenIn(
	_opts interface{}, amountIn decimal.Decimal,
) (amountOut decimal.Decimal, err error) {
	opts, ok := _opts.(PluggableStrategyOpts)
	if !ok {
		err = errors.New("opts must be of type PluggableStrategyOpts")
		return
	}
	if amountIn.Equals(decimal.Zero) {
		err = ErrMarketPreviewAmountTooLow
		return
	}

	percentageFee := decimal.NewFromInt(int64(opts.Fee)).Div(tenThousand)
	amount := amountIn.Mul(opts.Price).Mul(decimal.NewFromInt(1).Sub(percentageFee))
	amount = amount.Round(8)
	if amount.LessThanOrEqual(decimal.Zero) {
		err = ErrMarketPreviewAmountTooLow
		return
	}
	if amount.GreaterThanOrEqual(opts.BalanceOut) {
		err = ErrMarketPreviewAmountTooBig
		return
	}

	amountOut = amount
	return
}

func (s PluggableStrategy) InGivenOut(
	_opts interface{}, amountOut decimal.Decimal,
) (amountIn decimal.Decimal, err error) {
	opts, ok := _opts.(PluggableStrategyOpts)
	if !ok {
		err = errors.New("opts must be of type PluggableStrategyOpts")
		return
	}
	if amountOut.Equals(decimal.Zero) {
		err = ErrMarketPreviewAmountTooLow
		return
	}
	if amountOut.GreaterThanOrEqual(opts.BalanceOut) {
		err = ErrMarketPreviewAmountTooBig
		return
	}

	percentageFee := decimal.NewFromInt(int64(opts.Fee)).Div(tenThousand)
	amount := amountOut.Mul(opts.Price).Mul(decimal.NewFromInt(1).Add(percentageFee))
	amount = amount.Round(8)
	if amount.LessThanOrEqual(decimal.Zero) {
		err = ErrMarketPreviewAmountTooLow
		return
	}

	amountIn = amount
	return
}

func (s PluggableStrategy) FormulaType() int {
	return int(StrategyTypePluggable)
}
