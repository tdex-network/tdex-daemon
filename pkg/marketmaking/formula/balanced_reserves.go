// Package formula defines the formulas that implements the MarketFormula interface
package formula

import (
	"errors"

	"github.com/shopspring/decimal"
)

const BalancedReservesType = 1

var (
	balancedWeightIn  = decimal.NewFromInt(50)
	balancedWeightOut = decimal.NewFromInt(50)
	tenThousands      = decimal.NewFromInt(10000)
)

var (
	// ErrInvalidOptsType ...
	ErrInvalidOptsType = errors.New("opts must be of type BalancedReservesOpts")
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
	// The fee should be in satoshis and represents the calculated fee to take out from the swap.
	Fee uint64
	// Defines if the fee should be charged on the way in (ie. on )
	ChargeFeeOnTheWayIn bool
}

// BalancedReserves defines an AMM strategy with fixed 50/50 reserves
type BalancedReserves struct{}

// SpotPrice calculates the spot price (without fees) given the balances fo the two reserves. The weight reserve ratio is fixed at 50/50
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

// OutGivenIn returns the amountOut of asset that will be exchanged for the given amountIn.
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

// InGivenOut returns the amountIn of assets that will be needed for having the desired amountOut in return.
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
