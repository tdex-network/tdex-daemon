package formula

import (
	"github.com/shopspring/decimal"
)

// TenThousands ...
var TenThousands = 10000

// TenThousandsDecimal ...
var TenThousandsDecimal = decimal.NewFromInt(int64(TenThousands))

// PlusFee calculates an amount with a fee added given a int64 amount and a fee expressed in basis point (ie. 0.25 = 25)
func PlusFee(amount, feeAsBasisPoint uint64) (withFee, calculatedFee uint64) {

	amountDividedByTenThousands := amount / uint64(TenThousands)
	calculatedFee = amountDividedByTenThousands * feeAsBasisPoint
	withFee = amount + calculatedFee

	return withFee, calculatedFee
}

// LessFee calculates an amount with a subtracted given a int64 amount and a fee expressed in basis point (ie. 0.25 = 25)
func LessFee(amount, feeAsBasisPoint uint64) (withFee, calculatedFee uint64) {

	amountDividedByTenThousands := amount / uint64(TenThousands)
	calculatedFee = amountDividedByTenThousands * feeAsBasisPoint
	withFee = amount - calculatedFee

	return withFee, calculatedFee
}
