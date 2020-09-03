package market

import (
	mm "github.com/tdex-network/tdex-daemon/pkg/marketmaking"
	"github.com/tdex-network/tdex-daemon/pkg/marketmaking/formula"
)

// Strategy ...
func (m *Market) Strategy() *mm.MakingStrategy {
	return m.strategy
}

// IsStrategyPluggable returns true if the the startegy isn't automated.
func (m *Market) IsStrategyPluggable() bool {
	return m.strategy.IsZero()
}

// IsStrategyPluggableInitialized returns true if the the startegy isn't automated and the prices have been set.
func (m *Market) IsStrategyPluggableInitialized() bool {
	return m.IsStrategyPluggable() && !m.basePrice.IsZero() && !m.quotePrice.IsZero()
}

// MakeStrategyPluggable makes the current market using a given price (ie. set via UpdateMarketPrice rpc either manually or a price feed plugin)
func (m *Market) MakeStrategyPluggable() error {
	if !m.IsFunded() {
		return ErrNotFunded
	}

	if m.IsTradable() {
		// We need the market be switched off before making this change
		return ErrTradable
	}

	m.strategy = &mm.MakingStrategy{}

	return nil
}

// MakeStrategyBalanced makes the current market using a balanced AMM formula 50/50
func (m *Market) MakeStrategyBalanced() error {
	if !m.IsFunded() {
		return ErrNotFunded
	}

	if m.IsTradable() {
		// We need the market be switched off before making this change
		return ErrTradable
	}

	m.strategy = mm.NewStrategyFromFormula("balanced", "50/50 asset reserves", formula.BalancedReserves{})

	return nil
}
