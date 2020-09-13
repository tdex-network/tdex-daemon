// Package formula defines the formulas that implements the MarketFormula interface
package formula

import (
	"math"

	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking"
)

const (
	weightIn  = 50
	weightOut = 50
)

var (
	//BigOne represents a single unit of an asset with precision 8
	BigOne = int64(math.Pow10(8))
	//BigOneDecimal represents a single unit of an asset with precision 8 as decimal.Decimal
	BigOneDecimal = decimal.NewFromInt(BigOne)
)

//BalancedReserves defines an AMM strategy with fixed 50/50 reserves
type BalancedReserves struct{}

// SpotPrice calculates the spot price without fee given the balances. The weight reserve ratio is fixed at 50/50
func (BalancedReserves) SpotPrice(opts *marketmaking.FormulaOpts) (spotPrice int64) {
	// TODO sanitize numbers
	numer := Div(opts.BalanceIn, weightIn)
	denom := Div(opts.BalanceOut, weightOut)

	ratio := DivDecimal(numer, denom)
	scale := Div(BigOne, BigOne-opts.Fee)

	spotPrice = MulDecimal(MulDecimal(ratio, scale), BigOneDecimal).IntPart()
	return
}

// OutGivenIn returns the amountOut of asset that will be exchanged for the given amountIn
func (BalancedReserves) OutGivenIn(opts *marketmaking.FormulaOpts, amountIn int64) (amountOut int64) {
	weightRatio := Div(weightIn, weightOut)

	if opts.ChargeFeeOnTheWayIn {
		adjustedIn := amountIn * (BigOne - opts.Fee)

		y := Div(opts.BalanceIn, (opts.BalanceIn + adjustedIn))
		foo := y.Pow(weightRatio)
		bar := SubDecimal(BigOneDecimal, foo)

		amountOut = MulDecimal(Mul(opts.BalanceOut, 1), bar).IntPart()
		return
	}

	return
}

// InGivenOut returns the amountIn of assets that will be needed for the exchanging the desired amountOut of asset
func (BalancedReserves) InGivenOut(opts *marketmaking.FormulaOpts, amountOut int64) int64 {
	return 0
}
