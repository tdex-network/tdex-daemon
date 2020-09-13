package formula

import "math/big"

// PlusFee calculates an amount with a fee added given a bigInt amount and a fee expressed in basis point (ie. 0.25 = 25)
func PlusFee(amount *big.Int, feeAsBasisPoint *big.Int) (withFee *big.Int, calculatedFee *big.Int) {
	amountDividedByTenThousands := big.NewInt(0)
	amountDividedByTenThousands.Div(amount, big.NewInt(10000))

	calculatedFee = big.NewInt(0)
	calculatedFee.Mul(amountDividedByTenThousands, feeAsBasisPoint)

	withFee = big.NewInt(0)
	withFee.Add(amount, calculatedFee)

	return withFee, calculatedFee

}

// LessFee calculates an amount with a fee subtracted given a bigInt amount and a bigInt fee expressed in basis point (ie. 0.25 = 25)
func LessFee(amount *big.Int, feeAsBasisPoint *big.Int) (withFee *big.Int, calculatedFee *big.Int) {

	amountDividedByTenThousands := big.NewInt(0)
	amountDividedByTenThousands.Div(amount, big.NewInt(10000))

	calculatedFee = big.NewInt(0)
	calculatedFee.Mul(amountDividedByTenThousands, feeAsBasisPoint)

	withFee = big.NewInt(0)
	withFee.Sub(amount, calculatedFee)

	return withFee, calculatedFee

}
