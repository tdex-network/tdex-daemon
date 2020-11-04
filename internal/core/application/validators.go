package application

import (
	"errors"
	"regexp"
)

func validateAmount(satoshis int64) error {
	const maxSatoshis = 2099999997690000

	if satoshis <= 0 {
		return errors.New("amount must be greater than to zero")
	} 

	if satoshis > maxSatoshis {
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
