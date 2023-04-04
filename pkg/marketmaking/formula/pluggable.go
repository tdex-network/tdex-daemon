package formula

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type PluggableOpts struct {
	BalanceIn  decimal.Decimal
	BalanceOut decimal.Decimal
	Price      decimal.Decimal
}

var (
	ErrInvalidPluggableOptsType = fmt.Errorf("opts must be of type PluggableOpts")
)

type Pluggable struct{}

func (s Pluggable) SpotPrice(
	_ interface{},
) (spotPrice decimal.Decimal, err error) {
	return
}

func (s Pluggable) OutGivenIn(
	_opts interface{}, amountIn decimal.Decimal,
) (amountOut decimal.Decimal, err error) {
	opts, ok := _opts.(PluggableOpts)
	if !ok {
		err = ErrInvalidPluggableOptsType
		return
	}
	if amountIn.Equals(decimal.Zero) {
		err = ErrAmountTooLow
		return
	}

	amount := amountIn.Mul(opts.Price).Round(8)
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

func (s Pluggable) InGivenOut(
	_opts interface{}, amountOut decimal.Decimal,
) (amountIn decimal.Decimal, err error) {
	opts, ok := _opts.(PluggableOpts)
	if !ok {
		err = ErrInvalidPluggableOptsType
		return
	}
	if amountOut.Equals(decimal.Zero) {
		err = ErrAmountTooLow
		return
	}
	if amountOut.GreaterThanOrEqual(opts.BalanceOut) {
		err = ErrAmountTooBig
		return
	}

	amount := amountOut.Mul(opts.Price).Round(8)
	if amount.LessThanOrEqual(decimal.Zero) {
		err = ErrAmountTooLow
		return
	}

	amountIn = amount
	return
}
