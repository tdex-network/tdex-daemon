package marketmaking

var (
	// ConstantProduct ...
	ConstantProduct = MakingStrategy{"product", "constant product", nil}
	// ConstantValueFunction ...
	ConstantValueFunction = MakingStrategy{"value", "constant value function", nil}
)

var strategies = []MakingStrategy{
	ConstantProduct,
	ConstantValueFunction,
}

// MakingStrategy defines the automated market making strategy, usingi a formula to be applied to calculate the price of next trade.
type MakingStrategy struct {
	name        string
	description string
	formula     MakingFormula
}

// NewStrategyFromFormula returns the startegy struct with the name
func NewStrategyFromFormula(formula MakingFormula, name, description string) (MakingStrategy, error) {
	strategy := MakingStrategy{
		name:        name,
		description: description,
		formula:     formula,
	}

	return strategy, nil
}

// IsZero checks if the given startegy is the zero value
func (ms MakingStrategy) IsZero() bool {
	return ms == MakingStrategy{}
}

// Name returns the short name of the MM strategy
func (ms MakingStrategy) Name() string {
	return ms.name
}

// Description returns the long description of the MM strategy
func (ms MakingStrategy) Description() string {
	return ms.description
}

// Formula returns the mathematical formula of the MM strategy
func (ms MakingStrategy) Formula() MakingFormula {
	return ms.formula
}
