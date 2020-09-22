package mathutil

import (
	"math"
	"math/big"

	"github.com/shopspring/decimal"
)

var (
	//BigOne represents a single unit of an asset with precision 8
	BigOne = uint64(math.Pow10(8))
	//BigOneDecimal represents a single unit of an asset with precision 8 as decimal.Decimal
	BigOneDecimal = decimal.NewFromInt(int64(BigOne))
)

func init() {
	decimal.DivisionPrecision = 8
}

//BigAdd takes two int64 numbers and sum them x + y and returns the result
func BigAdd(x, y int64) (z int64) {
	X, Y := big.NewInt(x), big.NewInt(y)
	z = new(big.Int).Add(X, Y).Int64()
	return
}

//BigSub takes two int64 numbers and sum them x - y and returns the result
func BigSub(x, y int64) (z int64) {
	X, Y := big.NewInt(x), big.NewInt(y)
	z = new(big.Int).Sub(X, Y).Int64()
	return
}

//BigMul takes two int64 numbers and sum them x - y and returns the result
func BigMul(x, y int64) (z int64) {
	X, Y := big.NewInt(x), big.NewInt(y)
	z = new(big.Int).Mul(X, Y).Int64()
	return
}

//BigDiv takes two int64 numbers and sum them x - y and returns the result
func BigDiv(x, y int64) (z int64) {
	X, Y := big.NewInt(x), big.NewInt(y)
	z = new(big.Int).Div(X, Y).Int64()
	return
}

//Add takes two uint64 numbers and sum them x + y and returns the result as decimal.Decimal
func Add(x, y uint64) (z decimal.Decimal) {
	X, Y := decimal.NewFromBigInt(new(big.Int).SetUint64(x), 0), decimal.NewFromBigInt(new(big.Int).SetUint64(y), 0)
	z = AddDecimal(X, Y)
	return
}

//AddDecimal takes two decimal.Decimal numbers and sum them x + y and returns the result as decimal.Decimal
func AddDecimal(X, Y decimal.Decimal) (z decimal.Decimal) {
	z = X.Add(Y)
	return
}

//Sub takes two uint64 numbers and subtract them x - y and returns the result as decimal.Decimal
func Sub(x, y uint64) (z decimal.Decimal) {
	X, Y := decimal.NewFromBigInt(new(big.Int).SetUint64(x), 0), decimal.NewFromBigInt(new(big.Int).SetUint64(y), 0)
	z = SubDecimal(X, Y)
	return
}

//SubDecimal takes two decimal.Decimal numbers and subtract them x - y and returns the result as decimal.Decimal
func SubDecimal(X, Y decimal.Decimal) (z decimal.Decimal) {
	z = X.Sub(Y)
	return
}

// Mul takes two int64 numbers and multiply them x * y and returns the result as decimal.Decimal
func Mul(x, y uint64) (z decimal.Decimal) {
	X, Y := decimal.NewFromBigInt(new(big.Int).SetUint64(x), 0), decimal.NewFromBigInt(new(big.Int).SetUint64(y), 0)
	z = MulDecimal(X, Y)
	return
}

// MulDecimal takes two decimal.Decimal numbers and multiply them x * y and returns the result as decimal.Decimal
func MulDecimal(X, Y decimal.Decimal) (z decimal.Decimal) {
	z = X.Mul(Y)
	return
}

// Div takes two uint64 numbers and divides them x / y and returns the result as decimal.Decimal
func Div(x, y uint64) (z decimal.Decimal) {
	X, Y := decimal.NewFromBigInt(new(big.Int).SetUint64(x), 0), decimal.NewFromBigInt(new(big.Int).SetUint64(y), 0)
	z = DivDecimal(X, Y)
	return
}

// DivDecimal takes two decimal.Decimal numbers and divides them x / y and returns the result as decimal.Decimal
func DivDecimal(X, Y decimal.Decimal) (z decimal.Decimal) {
	z = X.Div(Y)
	return
}
