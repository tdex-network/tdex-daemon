package formula

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/tdex-network/tdex-daemon/pkg/mathutil"
)

type PluggableOpts struct {
	BalanceIn  uint64
	BalanceOut uint64
	Price      decimal.Decimal
	Fee        uint64
}

var (
	ErrInvalidPluggableOptsType = fmt.Errorf("opts must be of type PluggableOpts")
)

type Pluggable struct{}

func (s Pluggable) SpotPrice(
	_ interface{},
) (spotPrice decimal.Decimal, err error) {
	return
}

func (s Pluggable) OutGivenIn(
	_opts interface{}, amountIn uint64,
) (uint64, error) {
	opts, ok := _opts.(PluggableOpts)
	if !ok {
		return 0, ErrInvalidPluggableOptsType
	}
	if amountIn == 0 {
		return 0, ErrAmountTooLow
	}

	amountR := decimal.NewFromInt(int64(amountIn)).Mul(opts.Price).BigInt().Uint64()
	amountR, _ = mathutil.LessFee(amountR, opts.Fee)
	if amountR == 0 {
		return 0, ErrAmountTooLow
	}
	return amountR, nil
}

func (s Pluggable) InGivenOut(
	_opts interface{}, amountOut uint64,
) (uint64, error) {
	opts, ok := _opts.(PluggableOpts)
	if !ok {
		return 0, ErrInvalidPluggableOptsType
	}
	if amountOut == 0 {
		return 0, ErrAmountTooLow
	}
	if amountOut >= opts.BalanceOut {
		return 0, ErrAmountTooBig
	}

	amountP := decimal.NewFromInt(int64(amountOut)).Mul(opts.Price).BigInt().Uint64()
	amountP, _ = mathutil.PlusFee(amountP, opts.Fee)
	if amountP == 0 {
		return 0, ErrAmountTooLow
	}
	return amountP, nil
}
