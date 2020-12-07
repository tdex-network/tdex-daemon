package mathutil

import (
	"math/big"

	"github.com/shopspring/decimal"
)

// TenThousands ...
const TenThousands = 10000

// PlusFee calculates an amount with a fee added given a int64 amount and a fee expressed in basis point (ie. 0.25 = 25)
func PlusFee(amount, feeAsBasisPoint uint64) (withFee, calculatedFee uint64) {
	feeDecimal := decimal.NewFromInt(int64(feeAsBasisPoint)).Div(decimal.NewFromInt(TenThousands))
	amountDecimal := decimal.NewFromBigInt(new(big.Int).SetUint64(amount), 0)
	percentage := decimal.NewFromInt(1).Add(feeDecimal)

	calculatedFee = amountDecimal.Mul(feeDecimal).BigInt().Uint64()
	withFee = amountDecimal.Mul(percentage).BigInt().Uint64()
	return
}

// LessFee calculates an amount with a subtracted given a int64 amount and a fee expressed in basis point (ie. 0.25 = 25)
func LessFee(amount, feeAsBasisPoint uint64) (withFee, calculatedFee uint64) {
	feeDecimal := decimal.NewFromInt(int64(feeAsBasisPoint)).Div(decimal.NewFromInt(TenThousands))
	amountDecimal := decimal.NewFromBigInt(new(big.Int).SetUint64(amount), 0)
	percentage := decimal.NewFromInt(1).Sub(feeDecimal)

	calculatedFee = amountDecimal.Mul(feeDecimal).BigInt().Uint64()
	withFee = amountDecimal.Mul(percentage).BigInt().Uint64()
	return
}
