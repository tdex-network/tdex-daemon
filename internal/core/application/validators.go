package application

import (
	"errors"
)

func validateAmount(satoshis int64) error {
	const maxSatoshis = 2099999997690000

	if satoshis <= 0 {
		errors.New("amount must be greater than to zero")
	} 

	if satoshis > maxSatoshis {
		errors.New("amount cannot be greater than 2099999997690000")
	}

	return nil
}