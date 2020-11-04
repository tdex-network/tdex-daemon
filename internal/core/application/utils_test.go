package application

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	validAsset = "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225"
	invalidHexAsset = "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751bzzz"
	invalidLengthAsset = "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca2900"
)

func TestValidateAssetString(t *testing.T) {
	t.Run("should return nil if the asset string is valid", func(t *testing.T) {
		err := validateAssetString(validAsset)
		assert.Equal(t, nil, err)
	})

	t.Run("should return an error if the asset's length != 64", func(t *testing.T) {
		err := validateAssetString(invalidLengthAsset)
		assert.NotEqual(t, nil, err)
	})

	t.Run("should return an error if the asset's string is not a valid hex string", func(t *testing.T) {
		err := validateAssetString(invalidHexAsset)
		assert.NotEqual(t, nil, err)
	})
}