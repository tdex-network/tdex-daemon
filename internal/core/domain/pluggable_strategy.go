package domain

import (
	"errors"

	"github.com/shopspring/decimal"
)

var tenThousand = decimal.NewFromInt(10000)

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
