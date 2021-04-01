package wallet

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEstimateTxSize(t *testing.T) {
	tests := []struct {
		inScriptTypes        []int
		outScriptTypes       []int
		inAuxiliaryP2ShSize  []int
		outAuxiliaryP2ShSize []int
		expectedSize         int
	}{
		// https://blockstream.info/liquid/tx/3bf5b21f9b5785de089be6dc4963058b4734bf86a9434c9910ad739dbf742eb0
		{
			inScriptTypes:  []int{P2SH_P2WPKH},
			outScriptTypes: []int{P2SH_P2WPKH, P2SH_P2WPKH},
			expectedSize:   2516,
		},
		// https://blockstream.info/liquid/tx/06d4897d60128cccc588ccd5e1d62eba3d23b154ce8954e6b8057356c9eb9fa0
		{
			inScriptTypes:  []int{P2SH_P2WPKH, P2SH_P2WPKH},
			outScriptTypes: []int{P2WPKH, P2WPKH},
			expectedSize:   2621,
		},
		// https://blockstream.info/liquid/tx/34941db50a2128008451304200e396b64b68120f411f0a4fe0c2f9cef1f9864f
		{
			inScriptTypes:  []int{P2WPKH, P2WPKH, P2WPKH},
			outScriptTypes: []int{P2WPKH, P2WPKH, P2WPKH, P2WPKH, P2WPKH},
			expectedSize:   6258,
		},
	}
	for _, tt := range tests {
		size := EstimateTxSize(
			tt.inScriptTypes, tt.outScriptTypes,
			tt.inAuxiliaryP2ShSize, tt.outAuxiliaryP2ShSize,
		)
		fmt.Println(tt.expectedSize, size)
		assert.GreaterOrEqual(t, size, tt.expectedSize)
	}
}
