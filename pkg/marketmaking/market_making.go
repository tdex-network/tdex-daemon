package marketmaking

import (
	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking/formula"
)

// MakingFormula defines the interface for implementing the formula to derive the spot price
type MakingFormula interface {
	SpotPrice(spotPriceOpts interface{}) (spotPrice decimal.Decimal, err error)
	OutGivenIn(outGivenInOpts interface{}, amountIn uint64) (amountOut uint64, err error)
	InGivenOut(inGivenOutOpts interface{}, amountOut uint64) (amountIn uint64, err error)
}

func NewBalancedReservedFormula() MakingFormula {
	return formula.BalancedReserves{}
}
