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
	// ErrBalanceTooLow ...
	ErrBalanceTooLow = errors.New("reserve balance amount is too low")
)

//BalancedReserves defines an AMM strategy with fixed 50/50 reserves
type BalancedReserves struct{}

// SpotPrice calculates the spot price (without fees) given the balances fo the two reserves. The weight reserve ratio is fixed at 50/50
func (BalancedReserves) SpotPrice(opts *marketmaking.FormulaOpts) (spotPrice decimal.Decimal, err error) {
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

// OutGivenIn returns the amountOut of asset that will be exchanged for the given amountIn
func (BalancedReserves) OutGivenIn(opts *marketmaking.FormulaOpts, amountIn uint64) (amountOut uint64, err error) {
	if opts.BalanceIn == 0 || opts.BalanceOut == 0 {
		err = ErrBalanceTooLow
		return
	}

	if amountIn == 0 {
		err = ErrAmountTooLow
		return
	}
	if amountIn >= opts.BalanceIn {
		err = ErrAmountTooBig
		return
	}

	amountInWithFees, _ := mathutil.LessFee(amountIn, opts.Fee)
	if !opts.ChargeFeeOnTheWayIn {
		amountInWithFees, _ = mathutil.PlusFee(amountIn, opts.Fee)
	}

	balanceIn := decimal.NewFromInt(int64(opts.BalanceIn))
	balanceOut := decimal.NewFromInt(int64(opts.BalanceOut))
	amountOutDecimal := balanceOut.Mul(
		decimal.NewFromInt(1).Sub(
			balanceIn.Div(
				balanceIn.Add(decimal.NewFromInt(int64(amountInWithFees))),
			),
		),
	)
	amountOut = amountOutDecimal.BigInt().Uint64()

	return
}

// InGivenOut returns the amountIn of assets that will be needed for having the desired amountOut in return
func (BalancedReserves) InGivenOut(opts *marketmaking.FormulaOpts, amountOut uint64) (amountIn uint64, err error) {
	if opts.BalanceIn == 0 || opts.BalanceOut == 0 {
		err = ErrBalanceTooLow
		return
	}

	if amountOut == 0 {
		err = ErrAmountTooLow
		return
	}
	if amountOut >= opts.BalanceOut {
		err = ErrAmountTooBig
		return
	}

	one := decimal.NewFromInt(1)
	feeDecimal := decimal.NewFromInt(int64(opts.Fee)).Div(decimal.NewFromInt(mathutil.TenThousands))
	amountPercentage := one.Add(feeDecimal)
	if !opts.ChargeFeeOnTheWayIn {
		amountPercentage = one.Sub(feeDecimal)
	}
	balanceIn := decimal.NewFromInt(int64(opts.BalanceIn))
	balanceOut := decimal.NewFromInt(int64(opts.BalanceOut))

	amountInDecimal := balanceIn.Mul(
		balanceOut.Div(
			balanceOut.Sub(decimal.NewFromInt(int64(amountOut))),
		).Sub(one),
	).Mul(one.Div(amountPercentage))

	amountIn = amountInDecimal.BigInt().Uint64()

	return
}

func (BalancedReserves) FormulaType() int {
	return BalancedReservesType
}
