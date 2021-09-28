package application

import (
	"errors"
	"regexp"
)

func validateAssetString(asset string) error {
	const regularExpression = `[0-9a-f]{64}`

	matched, err := regexp.Match(regularExpression, []byte(asset))
	if err != nil {
		return err
	}

	if !matched {
		return errors.New(asset + " is an invalid asset string.")
	}

	return nil
}
