package formula

import (
	"math/big"

	"github.com/tdex-network/tdex-daemon/pkg/marketmaking"
)

//BalancedReserves defines an AMM strategy with fixed 50/50 reserves
type BalancedReserves struct{}

// SpotPrice calculates the spot price given the balances
func (BalancedReserves) SpotPrice(opts *marketmaking.SpotPriceOpts) (spotPrice int64) {
	BalanceIn := big.NewInt(opts.BalanceIn)
	WeightIn := big.NewInt(opts.WeightIn)
	Numer := big.NewInt(0)
	Numer.Div(BalanceIn, WeightIn)

	BalanceOut := big.NewInt(opts.BalanceOut)
	WeightOut := big.NewInt(opts.WeightOut)
	Denom := big.NewInt(0)
	Denom.Div(BalanceOut, WeightOut)

	ratio := big.NewInt(0)
	ratio.Div(Numer, Denom)

	withFee, _ := plusFee(big.NewInt(100000000), big.NewInt(opts.Fee))
	return withFee.Int64()
}

func (BalancedReserves) OutGivenIn(opts *marketmaking.SpotPriceOpts, amountIn int64) int64 {
	return 0
}

func (BalancedReserves) InGivenOut(opts *marketmaking.SpotPriceOpts, amountOut int64) int64 {
	return 0
}

func plusFee(amount *big.Int, fee *big.Int) (withFee *big.Int, calculatedFee *big.Int) {
	amountDividedByTenThousands := big.NewInt(0)
	amountDividedByTenThousands.Div(amount, big.NewInt(10000))

	calculatedFee = big.NewInt(0)
	calculatedFee.Mul(amountDividedByTenThousands, fee)

	withFee = big.NewInt(0)
	withFee.Add(amount, calculatedFee)

	return withFee, calculatedFee

}
