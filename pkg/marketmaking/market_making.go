package marketmaking

// MakingStrategy defines the automated market making strategy, usingi a formula to be applied to calculate the price of next trade.
type MakingStrategy struct {
	formula MakingFormula
}

//FormulaOpts defines the parameters needed to calculate the spot price
type FormulaOpts struct {
	BalanceIn  int64
	BalanceOut int64
	WeightIn   int64
	WeightOut  int64
	// The fee should be in satoshis and represents the calculated fee to take out from the swap.
	Fee int64
	// Defines id the fee should be cahrged on the way in (ie. on )
	ChargeFeeOnTheWayIn bool
}

// MakingFormula defines the interface for implmenting the formula to derive the spot price
type MakingFormula interface {
	SpotPrice(spotPriceOpts *FormulaOpts) (spotPrice int64)
	OutGivenIn(outGivenInOpts *FormulaOpts, amountIn int64) (amountOut int64)
	InGivenOut(inGivenOutOpts *FormulaOpts, amountOut int64) (amountIn int64)
}

// NewStrategyFromFormula returns the startegy struct with the name
func NewStrategyFromFormula(formula MakingFormula) *MakingStrategy {
	strategy := &MakingStrategy{
		formula: formula,
	}

	return strategy
}

// IsZero checks if the given startegy is the zero value
func (ms MakingStrategy) IsZero() bool {
	return ms == MakingStrategy{}
}

// Formula returns the mathematical formula of the MM strategy
func (ms *MakingStrategy) Formula() MakingFormula {
	return ms.formula
}
