package wallet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWallet(t *testing.T) {
	wallet, err := newTestWallet()
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, wallet)
}

func TestFailingNewWallet(t *testing.T) {
	tests := []int{-1, 0, 127, 257, 130}
	for _, tt := range tests {
		opts := NewWalletOpts{
			EntropySize: tt,
		}
		_, err := NewWallet(opts)
		assert.NotNil(t, err)
	}
}

func TestNewWalletFromMnemonic(t *testing.T) {
	wallet, err := newTestWallet()
	if err != nil {
		t.Fatal(err)
	}
	opts := NewWalletFromMnemonicOpts{
		SigningMnemonic:  wallet.SigningMnemonic,
		BlindingMnemonic: wallet.BlindingMnemonic,
	}
	otherWallet, err := NewWalletFromMnemonic(opts)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, *wallet, *otherWallet)
}

func TestFailingNewWalletFromMnemonic(t *testing.T) {
	tests := []struct {
		opts NewWalletFromMnemonicOpts
		err  error
	}{
		{
			opts: NewWalletFromMnemonicOpts{
				SigningMnemonic: "",
			},
			err: ErrNullSigningMnemonic,
		},
		{
			opts: NewWalletFromMnemonicOpts{
				SigningMnemonic: "legal winner thank year wave sausage worth useful legal winner thank yellow yellow",
			},
			err: ErrInvalidSigningMnemonic,
		},
		{
			opts: NewWalletFromMnemonicOpts{
				SigningMnemonic:  "letter advice cage absurd amount doctor acoustic avoid letter advice cage absurd amount doctor acoustic avoid letter always",
				BlindingMnemonic: "legal winner thank year wave sausage worth useful legal winner thank yellow yellow",
			},
			err: ErrInvalidBlindingMnemonic,
		},
	}
	for _, tt := range tests {
		_, err := NewWalletFromMnemonic(tt.opts)
		assert.Equal(t, tt.err, err)
	}
}

func newTestWallet() (*Wallet, error) {
	opts := NewWalletOpts{
		EntropySize: 256,
	}
	return NewWallet(opts)
}
