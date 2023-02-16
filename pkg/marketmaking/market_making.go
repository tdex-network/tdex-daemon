package marketmaking

import (
	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking/formula"
)

// MakingFormula defines the interface for implementing the formula to derive the spot price
type MakingFormula interface {
	SpotPrice(spotPriceOpts interface{}) (spotPrice decimal.Decimal, err error)
	OutGivenIn(outGivenInOpts interface{}, amountIn decimal.Decimal) (amountOut decimal.Decimal, err error)
	InGivenOut(inGivenOutOpts interface{}, amountOut decimal.Decimal) (amountIn decimal.Decimal, err error)
}

func NewBalancedReservedFormula() MakingFormula {
	return formula.BalancedReserves{}
}

func NewPluggableFormula() MakingFormula {
	return formula.Pluggable{}
}
