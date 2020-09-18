package wallet

import (
	"testing"

	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/stretchr/testify/assert"
)

func TestParseDerivationPath(t *testing.T) {
	tests := []struct {
		input  string
		output DerivationPath
		err    error
	}{
		// Plain absolute derivation paths
		{"m/84'/0'/0'/0", DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, 0}, nil},
		{"m/84'/0'/0'/128", DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, 128}, nil},
		{"m/84'/0'/0'/0'", DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart}, nil},
		{"m/84'/0'/0'/128'", DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart + 0, hdkeychain.HardenedKeyStart + 128}, nil},
		{"m/2147483732/2147483648/2147483648/0", DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, 0}, nil},
		{"m/2147483732/2147483648/2147483648/2147483648", DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart}, nil},

		// Hexadecimal absolute derivation paths
		{"m/0x54'/0x00'/0x00'/0x00", DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, 0}, nil},
		{"m/0x54'/0x00'/0x00'/0x80", DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, 128}, nil},
		{"m/0x54'/0x00'/0x00'/0x00'", DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart}, nil},
		{"m/0x54'/0x00'/0x00'/0x80'", DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart + 128}, nil},
		{"m/0x80000054/0x80000000/0x80000000/0x00", DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, 0}, nil},
		{"m/0x80000054/0x80000000/0x80000000/0x80000000", DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart}, nil},

		// Weird inputs just to ensure they work
		{"	m  /   84			'\n/\n   00	\n\n\t'   /\n0 ' /\t\t	0", DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, hdkeychain.HardenedKeyStart, 0}, nil},

		// Relative derivation paths
		{"84'/0'/0/0", DerivationPath{hdkeychain.HardenedKeyStart + 84, hdkeychain.HardenedKeyStart, 0, 0}, nil},
		{"0'/0/0", DerivationPath{hdkeychain.HardenedKeyStart, 0, 0}, nil},
		{"0/0", DerivationPath{0, 0}, nil},

		// Invalid derivation paths
		{"", nil, ErrNullDerivationPath},                  // Empty relative derivation path
		{"m", nil, ErrMalformedDerivationPath},            // Empty absolute derivation path
		{"m/", nil, ErrMalformedDerivationPath},           // Missing last derivation component
		{"/84'/0'/0'/0", nil, ErrMalformedDerivationPath}, // Absolute path without m prefix, might be user error
		{"m/2147483648'", nil, nil},                       // Overflows 32 bit integer (dynamic values on error, not constant)
		{"m/-1'", nil, nil},                               // Cannot contain negative number (dynamic values on error, not constant)
		{"0", nil, ErrMalformedDerivationPath},            // Bad derivation path
	}
	for _, tt := range tests {
		path, err := ParseDerivationPath(tt.input)
		if err != nil {
			if tt.err != nil {
				assert.Equal(t, tt.err, err)
			}
		}
		assert.Equal(t, tt.output, path)
	}
}
