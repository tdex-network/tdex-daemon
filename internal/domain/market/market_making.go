package market

import "github.com/pkg/errors"

var (
	// ConstantProduct ...
	ConstantProduct = MakingStrategy{"product", "constant product", "x*y=k"}
	// ConstantValueFunction ...
	ConstantValueFunction = MakingStrategy{"value", "constant value function", "∏B**W"}
)

var strategies = []MakingStrategy{
	ConstantProduct,
	ConstantValueFunction,
}

// MakingStrategy defines the automated market making strategy, usingi a formula to be applied to calculate the price of next trade.
type MakingStrategy struct {
	name        string
	description string
	formula     string
}

// NewStrategyFromName returns the startegy struct with the name
func NewStrategyFromName(strategyName string) (MakingStrategy, error) {
	strategy, ok := getStrategyByName(strategyName)
	if !ok {
		return MakingStrategy{}, errors.Errorf("unknown '%s' strategy", strategyName)
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
func (ms MakingStrategy) Formula() string {
	return ms.formula
}

// Strategy ...
func (m *Market) Strategy() MakingStrategy {
	return m.strategy
}

// ChangeStrategyWithName ...
func (m *Market) ChangeStrategyWithName(strategyName string) error {
	strategy, ok := getStrategyByName(strategyName)
	if !ok {
		return errors.Errorf("unknown '%s' strategy", strategyName)
	}

	m.strategy = strategy
	return nil
}

func getStrategyByName(name string) (MakingStrategy, bool) {
	for _, strategy := range strategies {
		if strategy.Name() == name {
			return strategy, true
		}
	}

	return MakingStrategy{}, false
}
