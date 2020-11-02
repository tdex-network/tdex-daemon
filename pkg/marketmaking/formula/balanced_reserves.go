// Package formula defines the formulas that implements the MarketFormula interface
package formula

import (
	"errors"

	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking"
	"github.com/tdex-network/tdex-daemon/pkg/mathutil"
)

const (
	balancedWeightIn     = 50
	balancedWeightOut    = 50
	BalancedReservesType = 1
)

var (
	// ErrAmountTooLow ...
	ErrAmountTooLow = errors.New("provided amount is too low")
	// ErrAmountTooBig ...
	ErrAmountTooBig = errors.New("provided amount is too big")
	// ErrBalanceTooLow
	ErrBalanceTooLow = errors.New("reserve balance amount is too low")
)

//BalancedReserves defines an AMM strategy with fixed 50/50 reserves
type BalancedReserves struct{}

// SpotPrice calculates the spot price (without fees) given the balances fo the two reserves. The weight reserve ratio is fixed at 50/50
func (BalancedReserves) SpotPrice(opts *marketmaking.FormulaOpts) (spotPrice decimal.Decimal) {
	// 2 : 20k = 1 : x
	// BI : BO = OneInput : SpotPrice
	numer := mathutil.Div(opts.BalanceOut, balancedWeightOut)
	denom := mathutil.Div(opts.BalanceIn, balancedWeightIn)
	spotPrice = mathutil.DivDecimal(numer, denom)
	return
}

// OutGivenIn returns the amountOut of asset that will be exchanged for the given amountIn
func (BalancedReserves) OutGivenIn(opts *marketmaking.FormulaOpts, amountIn uint64) (amountOut uint64, err error) {
	if amountIn == 0 {
		err = ErrAmountTooLow
		return
	}

	if opts.BalanceIn == 0 || opts.BalanceOut == 0 {
		err = ErrBalanceTooLow
		return
	}

	invariant := mathutil.Mul(opts.BalanceIn, opts.BalanceOut)
	nextInBalance := mathutil.Add(opts.BalanceIn, amountIn)
	nextOutBalance := mathutil.DivDecimal(invariant, nextInBalance).BigInt().Uint64()
	amountOutWithoutFees := mathutil.Sub(opts.BalanceOut, nextOutBalance).BigInt().Uint64()

	if opts.ChargeFeeOnTheWayIn {
		amountOut, _ = mathutil.LessFee(amountOutWithoutFees, opts.Fee)
	} else {
		amountOut, _ = mathutil.PlusFee(amountOutWithoutFees, opts.Fee)
	}

	return
}

// InGivenOut returns the amountIn of assets that will be needed for having the desired amountOut in return
func (BalancedReserves) InGivenOut(opts *marketmaking.FormulaOpts, amountOut uint64) (amountIn uint64, err error) {
	if amountOut == 0 {
		err = ErrAmountTooLow
		return
	}
	if amountOut >= opts.BalanceOut {
		err = ErrAmountTooBig
		return
	}

	if opts.BalanceIn == 0 || opts.BalanceOut == 0 {
		err = ErrBalanceTooLow
		return
	}

	invariant := mathutil.Mul(opts.BalanceIn, opts.BalanceOut)
	nextOutBalance := mathutil.Sub(opts.BalanceOut, amountOut)
	nextInBalance := mathutil.DivDecimal(invariant, nextOutBalance)
	amountInWithoutFees := mathutil.Sub(nextInBalance.BigInt().Uint64(), opts.BalanceIn).BigInt().Uint64()

	if opts.ChargeFeeOnTheWayIn {
		amountIn, _ = mathutil.PlusFee(amountInWithoutFees, opts.Fee)
	} else {
		amountIn, _ = mathutil.LessFee(amountInWithoutFees, opts.Fee)
	}

	return
}

func (BalancedReserves) FormulaType() int {
	return BalancedReservesType
}
