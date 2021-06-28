package domain

import (
	"errors"

	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/pkg/mathutil"
)

const PluggableStrategyType = 255

type PluggableStrategyOpts struct {
	BalanceIn  uint64
	BalanceOut uint64
	Price      decimal.Decimal
	Fee        uint64
}

type PluggableStrategy struct{}

func (s PluggableStrategy) SpotPrice(_opts interface{}) (spotPrice decimal.Decimal, err error) {
	return
}

func (s PluggableStrategy) OutGivenIn(_opts interface{}, amountIn uint64) (uint64, error) {
	opts, ok := _opts.(PluggableStrategyOpts)
	if !ok {
		return 0, errors.New("opts must be of type PluggableStrategyOpts")
	}
	if amountIn == 0 {
		return 0, ErrMarketPreviewAmountTooLow
	}

	amountR := decimal.NewFromInt(int64(amountIn)).Mul(opts.Price).BigInt().Uint64()
	amountR, _ = mathutil.LessFee(amountR, opts.Fee)
	return amountR, nil
}

func (s PluggableStrategy) InGivenOut(_opts interface{}, amountOut uint64) (uint64, error) {
	opts, ok := _opts.(PluggableStrategyOpts)
	if !ok {
		return 0, errors.New("opts must be of type PluggableStrategyOpts")
	}
	if amountOut == 0 {
		return 0, ErrMarketPreviewAmountTooLow
	}
	if amountOut >= opts.BalanceOut {
		return 0, ErrMarketPreviewAmountTooBig
	}

	amountP := decimal.NewFromInt(int64(amountOut)).Mul(opts.Price).BigInt().Uint64()
	amountP, _ = mathutil.PlusFee(amountP, opts.Fee)
	return amountP, nil
}

func (s PluggableStrategy) FormulaType() int {
	return PluggableStrategyType
}
