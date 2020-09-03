package marketmaking

// MakingStrategy defines the automated market making strategy, usingi a formula to be applied to calculate the price of next trade.
type MakingStrategy struct {
	name        string
	description string
	formula     MakingFormula
}

//SpotPriceOpts defines the parameters needed to calculate the spot price
type SpotPriceOpts struct {
	BalanceIn  int64
	BalanceOut int64
	WeightIn   int64
	WeightOut  int64
	Fee        int64
	// Defines id the fee should be cahrged on the way in (ie. on )
	ChargeFeeOnTheWayIn bool
}

// MakingFormula defines the interface for implmenting the formula to derive the spot price
type MakingFormula interface {
	SpotPrice(spotPriceOpts *SpotPriceOpts) (spotPrice int64)
	OutGivenIn(spotPriceOpts *SpotPriceOpts, amountIn int64) (amountOut int64)
	InGivenOut(spotPriceOpts *SpotPriceOpts, amountOut int64) (amountIn int64)
}

// NewStrategyFromFormula returns the startegy struct with the name
func NewStrategyFromFormula(name, description string, formula MakingFormula) *MakingStrategy {
	strategy := &MakingStrategy{
		name:        name,
		description: description,
		formula:     formula,
	}

	return strategy
}

// IsZero checks if the given startegy is the zero value
func (ms MakingStrategy) IsZero() bool {
	return ms == MakingStrategy{}
}

// Name returns the short name of the MM strategy
func (ms *MakingStrategy) Name() string {
	return ms.name
}

// Description returns the long description of the MM strategy
func (ms *MakingStrategy) Description() string {
	return ms.description
}

// Formula returns the mathematical formula of the MM strategy
func (ms *MakingStrategy) Formula() MakingFormula {
	return ms.formula
}
