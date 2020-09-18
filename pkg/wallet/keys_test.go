package wallet

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vulpemventures/go-elements/network"
)

func TestExtendedKey(t *testing.T) {
	wallet, err := newTestWallet()
	if err != nil {
		t.Fatal(err)
	}
	opts := ExtendedKeyOpts{
		Account: 0,
	}

	xprv, err := wallet.ExtendedPrivateKey(opts)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEmpty(t, xprv)

	xpub, err := wallet.ExtendedPublicKey(opts)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEmpty(t, xpub)
}

func TestDeriveSigningKeyPair(t *testing.T) {
	wallet, err := newTestWallet()
	if err != nil {
		t.Fatal(err)
	}
	opts := DeriveSigningKeyPairOpts{
		DerivationPath: "0'/0/0",
	}
	prvkey, pubkey, err := wallet.DeriveSigningKeyPair(opts)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, prvkey)
	assert.NotNil(t, pubkey)
}

func TestDeriveBlindingKeyPair(t *testing.T) {
	wallet, err := newTestWallet()
	if err != nil {
		t.Fatal(err)
	}
	script, _ := hex.DecodeString("001439397080b51ef22c59bd7469afacffbeec0da12e")
	opts := DeriveBlindingKeyPairOpts{
		Script: script,
	}
	blindingPrvkey, blindingPubkey, err := wallet.DeriveBlindingKeyPair(opts)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, blindingPrvkey)
	assert.NotNil(t, blindingPubkey)
}

func TestFailingExtendedKey(t *testing.T) {
	wallet, err := newTestWallet()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		opts ExtendedKeyOpts
		err  error
	}{
		{
			opts: ExtendedKeyOpts{
				Account: MaxHardenedValue + 1,
			},
			err: ErrOutOfRangeDerivationPathAccount,
		},
	}

	for _, tt := range tests {
		_, err := wallet.ExtendedPrivateKey(tt.opts)
		assert.Equal(t, tt.err, err)
		_, err = wallet.ExtendedPublicKey(tt.opts)
		assert.Equal(t, tt.err, err)
	}
}

func TestFailingDeriveSigningKeyPair(t *testing.T) {
	wallet, err := newTestWallet()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		opts DeriveSigningKeyPairOpts
		err  error
	}{
		{
			opts: DeriveSigningKeyPairOpts{"0/0"},
			err:  ErrInvalidDerivationPathLength,
		},
		{
			opts: DeriveSigningKeyPairOpts{"0/0/0/0"},
			err:  ErrInvalidDerivationPathLength,
		},
		{
			opts: DeriveSigningKeyPairOpts{"0'/0/0/0"},
			err:  ErrInvalidDerivationPathLength,
		},
		{
			opts: DeriveSigningKeyPairOpts{"0/0/0"},
			err:  ErrInvalidDerivationPathAccount,
		},
	}

	for _, tt := range tests {
		_, _, err := wallet.DeriveSigningKeyPair(tt.opts)
		assert.Equal(t, tt.err, err)
	}
}

func TestFailingDeriveBlindingKeyPair(t *testing.T) {
	wallet, err := newTestWallet()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		opts DeriveBlindingKeyPairOpts
		err  error
	}{
		{
			opts: DeriveBlindingKeyPairOpts{[]byte{}},
			err:  ErrNullOutputScript,
		},
	}

	for _, tt := range tests {
		_, _, err := wallet.DeriveBlindingKeyPair(tt.opts)
		assert.Equal(t, tt.err, err)
	}
}

func TestDeriveConfidentialAddress(t *testing.T) {
	wallet, err := newTestWallet()
	if err != nil {
		t.Fatal(err)
	}

	opts := DeriveConfidentialAddressOpts{
		DerivationPath: "0'/0/0",
		Network:        &network.Liquid,
	}
	ctAddress, script, err := wallet.DeriveConfidentialAddress(opts)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, len(ctAddress) > 0)
	assert.Equal(t, true, len(script) > 0)
}

func TestFailingDeriveConfidentialAddress(t *testing.T) {
	wallet, err := newTestWallet()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		opts DeriveConfidentialAddressOpts
		err  error
	}{
		{
			opts: DeriveConfidentialAddressOpts{
				DerivationPath: "",
				Network:        &network.Liquid,
			},
			err: ErrNullDerivationPath,
		},
		{
			opts: DeriveConfidentialAddressOpts{
				DerivationPath: "0'/0/0",
				Network:        nil,
			},
			err: ErrNullNetwork,
		},
	}

	for _, tt := range tests {
		_, _, err := wallet.DeriveConfidentialAddress(tt.opts)
		assert.Equal(t, tt.err, err)
	}
}
