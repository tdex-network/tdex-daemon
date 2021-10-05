package application

import (
	"encoding/hex"
	"errors"
)

func validateAssetString(asset string) error {
	buf, err := hex.DecodeString(asset)
	if err != nil {
		return errors.New("asset is not in hex format")
	}

	if len(buf) != 32 {
		return errors.New("asset length is invalid")
	}

	return nil
}
