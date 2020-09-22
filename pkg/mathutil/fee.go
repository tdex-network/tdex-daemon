package mathutil

import (
	"math/big"

	"github.com/shopspring/decimal"
)

// TenThousands ...
var TenThousands = uint64(10000)

// PlusFee calculates an amount with a fee added given a int64 amount and a fee expressed in basis point (ie. 0.25 = 25)
func PlusFee(amount, feeAsBasisPoint uint64) (withFee, calculatedFee uint64) {

	feeDecimal := decimal.NewFromBigInt(new(big.Int).SetUint64(feeAsBasisPoint), 0)
	amountDecimal := decimal.NewFromBigInt(new(big.Int).SetUint64(amount), 0)

	amountDividedByTenThousands := Div(amount, TenThousands)
	calculatedFeeDecimal := MulDecimal(amountDividedByTenThousands, feeDecimal)
	withFeeDecimal := AddDecimal(amountDecimal, calculatedFeeDecimal)

	return withFeeDecimal.BigInt().Uint64(), calculatedFeeDecimal.BigInt().Uint64()
}

// LessFee calculates an amount with a subtracted given a int64 amount and a fee expressed in basis point (ie. 0.25 = 25)
func LessFee(amount, feeAsBasisPoint uint64) (withFee, calculatedFee uint64) {

	feeDecimal := decimal.NewFromBigInt(new(big.Int).SetUint64(feeAsBasisPoint), 0)
	amountDecimal := decimal.NewFromBigInt(new(big.Int).SetUint64(amount), 0)

	amountDividedByTenThousands := Div(amount, TenThousands)
	calculatedFeeDecimal := MulDecimal(amountDividedByTenThousands, feeDecimal)
	withFeeDecimal := SubDecimal(amountDecimal, calculatedFeeDecimal)

	return withFeeDecimal.BigInt().Uint64(), calculatedFeeDecimal.BigInt().Uint64()
}
