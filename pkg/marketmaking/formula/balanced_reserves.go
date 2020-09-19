// Package formula defines the formulas that implements the MarketFormula interface
package formula

import (
	"math"

	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking"
)

const (
	balancedWeightIn  = 50
	balancedWeightOut = 50
)

var (
	//BigOne represents a single unit of an asset with precision 8
	BigOne = uint64(math.Pow10(8))
	//BigOneDecimal represents a single unit of an asset with precision 8 as decimal.Decimal
	BigOneDecimal = decimal.NewFromInt(int64(BigOne))
)

//BalancedReserves defines an AMM strategy with fixed 50/50 reserves
type BalancedReserves struct{}

// SpotPrice calculates the spot price (without fees) given the balances fo the two reserves. The weight reserve ratio is fixed at 50/50
func (BalancedReserves) SpotPrice(opts *marketmaking.FormulaOpts) (spotPrice uint64) {
	// 2 : 20k = 1 : x
	// BI : BO = OneInput : SpotPrice
	numer := Div(opts.BalanceOut, balancedWeightOut)
	denom := Div(opts.BalanceIn, balancedWeightIn)
	ratio := DivDecimal(numer, denom)
	return ratio.BigInt().Uint64()
}

// OutGivenIn returns the amountOut of asset that will be exchanged for the given amountIn
func (BalancedReserves) OutGivenIn(opts *marketmaking.FormulaOpts, amountIn uint64) (amountOut uint64) {

	invariant := Mul(opts.BalanceIn, opts.BalanceOut)
	nextInBalance := Add(opts.BalanceIn, amountIn)
	nextOutBalance := DivDecimal(invariant, nextInBalance)
	amountOutWithoutFees := Sub(opts.BalanceOut, nextOutBalance.BigInt().Uint64()).BigInt().Uint64()

	if opts.ChargeFeeOnTheWayIn {
		amountOut, _ = LessFee(amountOutWithoutFees, opts.Fee)
	} else {
		amountOut, _ = PlusFee(amountOutWithoutFees, opts.Fee)
	}

	return
}

// InGivenOut returns the amountIn of assets that will be needed for having the desired amountOut in return
func (BalancedReserves) InGivenOut(opts *marketmaking.FormulaOpts, amountOut uint64) (amountIn uint64) {

	invariant := Mul(opts.BalanceIn, opts.BalanceOut)
	nextOutBalance := Sub(opts.BalanceOut, amountOut)
	nextInBalance := DivDecimal(invariant, nextOutBalance)
	amountInWithoutFees := Sub(nextInBalance.BigInt().Uint64(), opts.BalanceIn).BigInt().Uint64()

	if opts.ChargeFeeOnTheWayIn {
		amountIn, _ = PlusFee(amountInWithoutFees, opts.Fee)
	} else {
		amountIn, _ = LessFee(amountInWithoutFees, opts.Fee)
	}

	return
}
