package formula

import (
	"github.com/shopspring/decimal"
)

func init() {
	decimal.DivisionPrecision = 8
}

//Sub takes two int64 numbers and subtract them x - y and returns the result as decimal.Decimal
func Sub(x, y int64) (z decimal.Decimal) {
	X, Y := decimal.NewFromInt(x), decimal.NewFromInt(y)
	z = SubDecimal(X, Y)
	return
}

//SubDecimal takes two int64 numbers and subtract them x - y and returns the result as decimal.Decimal
func SubDecimal(X, Y decimal.Decimal) (z decimal.Decimal) {
	z = X.Sub(Y)
	return
}

// Div takes two int64 numbers and divides them x / y and returns the result as decimal.Decimal
func Div(x, y int64) (z decimal.Decimal) {
	X, Y := decimal.NewFromInt(x), decimal.NewFromInt(y)
	z = DivDecimal(X, Y)
	return
}

// DivDecimal takes two decimal.Decimal numbers and divides them x / y and returns the result as decimal.Decimal
func DivDecimal(X, Y decimal.Decimal) (z decimal.Decimal) {
	z = X.Div(Y)
	return
}

// Mul takes two int64 numbers and multiply them x * y and returns the result as decimal.Decimal
func Mul(x, y int64) (z decimal.Decimal) {
	X, Y := decimal.NewFromInt(x), decimal.NewFromInt(y)
	z = MulDecimal(X, Y)
	return
}

// MulDecimal takes two decimal.Decimal numbers and multiply them x * y and returns the result as decimal.Decimal
func MulDecimal(X, Y decimal.Decimal) (z decimal.Decimal) {
	z = X.Mul(Y)
	return
}
