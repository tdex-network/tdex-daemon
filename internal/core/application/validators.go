package application

import (
	"errors"
	"regexp"

	"github.com/shopspring/decimal"
)

func validateAmount(satoshis decimal.Decimal) error {
	if satoshis.IsZero() || satoshis.IsNegative() {
		return errors.New("amount must be greater than zero")
	} 

	if satoshis.Cmp(decimal.NewFromInt(2099999997690000)) > 0 {
		return errors.New("amount cannot be greater than 2099999997690000")
	}

	return nil
}

func validateAssetString(asset string) error {
	const regularExpression = `[0-9A-Fa-f]{64}`	

	matched, err := regexp.Match(regularExpression, []byte(asset))
	if err != nil {
		return err
	}

	if !matched {
		return errors.New(asset + " is an invalid asset string.")
	}

	return nil
}
