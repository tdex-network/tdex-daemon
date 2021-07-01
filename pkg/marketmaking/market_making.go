package marketmaking

import (
	"github.com/shopspring/decimal"
)

// MakingStrategy defines the automated market making strategy, usingi a formula to be applied to calculate the price of next trade.
type MakingStrategy struct {
	Type    int
	formula MakingFormula
}

// MakingFormula defines the interface for implementing the formula to derive the spot price
type MakingFormula interface {
	SpotPrice(spotPriceOpts interface{}) (spotPrice decimal.Decimal, err error)
	OutGivenIn(outGivenInOpts interface{}, amountIn uint64) (amountOut uint64, err error)
	InGivenOut(inGivenOutOpts interface{}, amountOut uint64) (amountIn uint64, err error)
	FormulaType() int
}

// NewStrategyFromFormula returns the strategy struct with the name
func NewStrategyFromFormula(
	formula MakingFormula,
) MakingStrategy {
	strategy := MakingStrategy{
		Type:    formula.FormulaType(),
		formula: formula,
	}

	return strategy
}

// IsZero checks if the given strategy is the zero value
func (ms MakingStrategy) IsZero() bool {
	return ms.Type == 0
}

// Formula returns the mathematical formula of the MM strategy
func (ms MakingStrategy) Formula() MakingFormula {
	return ms.formula
}
