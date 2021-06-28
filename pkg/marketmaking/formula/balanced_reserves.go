// Package formula defines the formulas that implements the MarketFormula interface
package formula

import (
	"errors"

	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/pkg/mathutil"
)

const (
	balancedWeightIn     = 50
	balancedWeightOut    = 50
	BalancedReservesType = 1
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

//BalancedReservesOpts defines the parameters needed to calculate the spot price
type BalancedReservesOpts struct {
	BalanceIn  uint64
	BalanceOut uint64
	WeightIn   uint64
	WeightOut  uint64
	// The fee should be in satoshis and represents the calculated fee to take out from the swap.
	Fee uint64
	// Defines if the fee should be charged on the way in (ie. on )
	ChargeFeeOnTheWayIn bool
}

//BalancedReserves defines an AMM strategy with fixed 50/50 reserves
type BalancedReserves struct{}

// SpotPrice calculates the spot price (without fees) given the balances fo the two reserves. The weight reserve ratio is fixed at 50/50
func (BalancedReserves) SpotPrice(_opts interface{}) (spotPrice decimal.Decimal, err error) {
	opts, ok := _opts.(BalancedReservesOpts)
	if !ok {
		err = ErrInvalidOptsType
		return
	}

	if opts.BalanceIn == 0 || opts.BalanceOut == 0 {
		err = ErrBalanceTooLow
		return
	}
	// 2 : 20k = 1 : x
	// BI : BO = OneInput : SpotPrice
	numer := mathutil.Div(opts.BalanceOut, balancedWeightOut)
	denom := mathutil.Div(opts.BalanceIn, balancedWeightIn)
	spotPrice = mathutil.DivDecimal(numer, denom)
	return
}

// OutGivenIn returns the amountOut of asset that will be exchanged for the given amountIn.
func (BalancedReserves) OutGivenIn(_opts interface{}, amountIn uint64) (uint64, error) {
	opts, ok := _opts.(BalancedReservesOpts)
	if !ok {
		return 0, ErrInvalidOptsType
	}

	if opts.BalanceIn == 0 || opts.BalanceOut == 0 {
		return 0, ErrBalanceTooLow
	}

	if amountIn == 0 {
		return 0, ErrAmountTooLow
	}

	amount := amountIn
	if opts.ChargeFeeOnTheWayIn {
		amount, _ = mathutil.LessFee(amountIn, opts.Fee)
	}
	if amount == 0 {
		return 0, ErrAmountTooLow
	}

	balanceIn := decimal.NewFromInt(int64(opts.BalanceIn))
	balanceOut := decimal.NewFromInt(int64(opts.BalanceOut))
	amountOutDecimal := balanceOut.Mul(
		decimal.NewFromInt(1).Sub(
			balanceIn.Div(
				balanceIn.Add(decimal.NewFromInt(int64(amount))),
			),
		),
	)

	amountOut := amountOutDecimal.BigInt().Uint64()
	if !opts.ChargeFeeOnTheWayIn {
		amountOut, _ = mathutil.LessFee(amountOut, opts.Fee)
	}
	if amountOut == 0 {
		return 0, ErrAmountTooLow
	}

	return amountOut, nil
}

// InGivenOut returns the amountIn of assets that will be needed for having the desired amountOut in return.
func (BalancedReserves) InGivenOut(_opts interface{}, amountOut uint64) (uint64, error) {
	opts, ok := _opts.(BalancedReservesOpts)
	if !ok {
		return 0, ErrInvalidOptsType
	}

	if opts.BalanceIn == 0 || opts.BalanceOut == 0 {
		return 0, ErrBalanceTooLow
	}

	if amountOut == 0 {
		return 0, ErrAmountTooLow
	}
	if amountOut >= opts.BalanceOut {
		return 0, ErrAmountTooBig
	}

	amount := amountOut
	if !opts.ChargeFeeOnTheWayIn {
		amount, _ = mathutil.PlusFee(amountOut, opts.Fee)
	}

	if amount == 0 {
		return 0, ErrAmountTooLow
	}
	if amount >= opts.BalanceOut {
		return 0, ErrAmountTooBig
	}

	balanceIn := decimal.NewFromInt(int64(opts.BalanceIn))
	balanceOut := decimal.NewFromInt(int64(opts.BalanceOut))
	amountInDecimal := balanceIn.Mul(
		balanceOut.Div(
			balanceOut.Sub(decimal.NewFromInt(int64(amount))),
		).Sub(decimal.NewFromInt(1)),
	)

	amountIn := amountInDecimal.BigInt().Uint64()
	if opts.ChargeFeeOnTheWayIn {
		amountIn, _ = mathutil.PlusFee(amountIn, opts.Fee)
	}
	if amountIn == 0 {
		return 0, ErrAmountTooLow
	}

	return amountIn, nil
}

func (BalancedReserves) FormulaType() int {
	return BalancedReservesType
}
