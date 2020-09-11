package wallet

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWallet(t *testing.T) {
	tests := []struct {
		opts         NewWalletOpts
		sameMnemonic bool
	}{
		{
			opts:         NewWalletOpts{},
			sameMnemonic: true,
		},
		{
			opts:         NewWalletOpts{ExtraMnemonic: true},
			sameMnemonic: false,
		},
	}
	for _, tt := range tests {
		wallet, err := NewWallet(tt.opts)
		if err != nil {
			t.Fatal(err)
		}
		signingMnemonic, err := wallet.SigningMnemonic()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, isMnemonicValid(signingMnemonic))
		blindingMnemonic, err := wallet.BlindingMnemonic()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, isMnemonicValid(blindingMnemonic))
		if tt.sameMnemonic {
			assert.Equal(t, signingMnemonic, blindingMnemonic)
		} else {
			assert.NotEqual(t, signingMnemonic, blindingMnemonic)
		}
	}
}

func TestFailingNewMnemonic(t *testing.T) {
	tests := []int{-1, 127, 257, 130}
	for _, tt := range tests {
		opts := NewMnemonicOpts{
			EntropySize: tt,
		}
		_, err := NewMnemonic(opts)
		assert.NotNil(t, err)
	}
}

func TestNewWalletFromMnemonic(t *testing.T) {
	wallet, err := newTestWallet()
	if err != nil {
		t.Fatal(err)
	}
	signingMnemonic, _ := wallet.SigningMnemonic()
	opts := NewWalletFromMnemonicOpts{
		SigningMnemonic: signingMnemonic,
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
				SigningMnemonic: nil,
			},
			err: ErrNullSigningMnemonic,
		},
		{
			opts: NewWalletFromMnemonicOpts{
				SigningMnemonic: strings.Split("legal winner thank year wave sausage worth useful legal winner thank yellow yellow", " "),
			},
			err: ErrInvalidSigningMnemonic,
		},
		{
			opts: NewWalletFromMnemonicOpts{
				SigningMnemonic:  strings.Split("letter advice cage absurd amount doctor acoustic avoid letter advice cage absurd amount doctor acoustic avoid letter always", " "),
				BlindingMnemonic: strings.Split("legal winner thank year wave sausage worth useful legal winner thank yellow yellow", " "),
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
	return NewWallet(NewWalletOpts{})
}
