// Package formula defines the formulas that implements the MarketFormula interface
package formula

import (
	"errors"

	"github.com/shopspring/decimal"
)

var (
	// ErrInvalidBalancedReservesOptsType ...
	ErrInvalidBalancedReservesOptsType = errors.New(
		"opts must be of type BalancedReservesOpts",
	)
	// ErrAmountTooLow ...
	ErrAmountTooLow = errors.New("provided amount is too low")
	// ErrAmountTooBig ...
	ErrAmountTooBig = errors.New("provided amount is too big")
	// ErrBalanceTooLow ...
	ErrBalanceTooLow = errors.New("reserve balance amount is too low")
)

// BalancedReservesOpts defines the parameters needed to calculate the spot price
type BalancedReservesOpts struct {
	BalanceIn  decimal.Decimal
	BalanceOut decimal.Decimal
	WeightIn   uint64
	WeightOut  uint64
}

// BalancedReserves defines an AMM strategy with fixed 50/50 reserves
type BalancedReserves struct{}

// SpotPrice calculates the spot price (without fees) given the balances fo the two reserves. The weight reserve ratio is fixed at 50/50
func (BalancedReserves) SpotPrice(_opts interface{}) (spotPrice decimal.Decimal, err error) {
	opts, ok := _opts.(BalancedReservesOpts)
	if !ok {
		err = ErrInvalidBalancedReservesOptsType
		return
	}

	if opts.BalanceIn.Equals(decimal.Zero) || opts.BalanceOut.Equals(decimal.Zero) {
		err = ErrBalanceTooLow
		return
	}

	spotPrice = opts.BalanceOut.Div(opts.BalanceIn)
	return
}

// OutGivenIn returns the amountOut of asset that will be exchanged for the given amountIn.
func (BalancedReserves) OutGivenIn(
	_opts interface{}, amountIn decimal.Decimal,
) (amountOut decimal.Decimal, err error) {
	opts, ok := _opts.(BalancedReservesOpts)
	if !ok {
		err = ErrInvalidBalancedReservesOptsType
		return
	}
	if opts.BalanceIn.Equal(decimal.Zero) || opts.BalanceOut.Equal(decimal.Zero) {
		err = ErrBalanceTooLow
		return
	}
	if amountIn.LessThanOrEqual(decimal.Zero) {
		err = ErrAmountTooLow
		return
	}

	amount := opts.BalanceOut.Mul(decimal.NewFromInt(1).Sub(
		opts.BalanceIn.Div(opts.BalanceIn.Add(amountIn)),
	)).Round(8)

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

// InGivenOut returns the amountIn of assets that will be needed for having the desired amountOut in return.
func (BalancedReserves) InGivenOut(
	_opts interface{}, amountOut decimal.Decimal,
) (amountIn decimal.Decimal, err error) {
	opts, ok := _opts.(BalancedReservesOpts)
	if !ok {
		err = ErrInvalidBalancedReservesOptsType
		return
	}
	if opts.BalanceIn.Equals(decimal.Zero) || opts.BalanceOut.Equals(decimal.Zero) {
		err = ErrBalanceTooLow
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

	amount := opts.BalanceIn.Mul(
		opts.BalanceOut.Div(opts.BalanceOut.Sub(amountOut)).Sub(
			decimal.NewFromInt(1),
		),
	).Round(8)

	if amount.LessThanOrEqual(decimal.Zero) {
		err = ErrAmountTooLow
		return
	}

	amountIn = amount
	return
}
