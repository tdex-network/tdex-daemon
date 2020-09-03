package wallet

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil/base58"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/payment"
	"github.com/vulpemventures/go-elements/slip77"
)

// ExtendedKeyOpts is the struct given to
// ExtendedPrivateKey and ExtendedPublicKey methods
type ExtendedKeyOpts struct {
	Account uint32
}

func (o ExtendedKeyOpts) validate() error {
	if o.Account > (MaxHardenedValue) {
		return ErrOutOfRangeDerivationPathAccount
	}
	return nil
}

// ExtendedPrivateKey returns the signing extended private key in base58 format
// for the provided account index
func (w *Wallet) ExtendedPrivateKey(opts ExtendedKeyOpts) (string, error) {
	if err := opts.validate(); err != nil {
		return "", err
	}
	if err := w.validate(); err != nil {
		return "", err
	}

	masterKey, err := hdkeychain.NewKeyFromString(
		base58.Encode(w.signingMasterKey),
	)
	if err != nil {
		return "", err
	}

	xprv, err := masterKey.Child(opts.Account)
	if err != nil {
		return "", err
	}

	return xprv.String(), nil
}

// ExtendedPublicKey returns the signing extended public key in base58 format
// for the provided account index
func (w *Wallet) ExtendedPublicKey(opts ExtendedKeyOpts) (string, error) {
	if err := opts.validate(); err != nil {
		return "", err
	}
	if err := w.validate(); err != nil {
		return "", err
	}

	masterKey, err := hdkeychain.NewKeyFromString(
		base58.Encode(w.signingMasterKey),
	)
	if err != nil {
		return "", err
	}

	xprv, err := masterKey.Child(opts.Account)
	if err != nil {
		return "", err
	}

	xpub, err := xprv.Neuter()
	if err != nil {
		return "", err
	}
	return xpub.String(), nil
}

// DeriveSigningKeyPairOpts is the struct given to DeriveSigningKeyPair method
type DeriveSigningKeyPairOpts struct {
	DerivationPath string
}

func (o DeriveSigningKeyPairOpts) validate() error {
	derivationPath, err := ParseDerivationPath(o.DerivationPath)
	if err != nil {
		return err
	}

	err = checkDerivationPath(derivationPath)
	if err != nil {
		return err
	}

	return nil
}

// DeriveSigningKeyPair derives the key pair of the provided derivation path
func (w *Wallet) DeriveSigningKeyPair(opts DeriveSigningKeyPairOpts) (
	*btcec.PrivateKey,
	*btcec.PublicKey,
	error,
) {
	if err := opts.validate(); err != nil {
		return nil, nil, err
	}
	if err := w.validate(); err != nil {
		return nil, nil, err
	}

	hdNode, err := hdkeychain.NewKeyFromString(
		base58.Encode(w.signingMasterKey),
	)
	if err != nil {
		return nil, nil, err
	}

	derivationPath, _ := ParseDerivationPath(opts.DerivationPath)
	for _, step := range derivationPath {
		hdNode, err = hdNode.Child(step)
		if err != nil {
			return nil, nil, err
		}
	}

	privateKey, err := hdNode.ECPrivKey()
	if err != nil {
		return nil, nil, err
	}
	publicKey, err := hdNode.ECPubKey()
	if err != nil {
		return nil, nil, err
	}

	return privateKey, publicKey, nil
}

// DeriveBlindingKeyPairOpts is the struct given to DeriveBlindingKeyPair method
type DeriveBlindingKeyPairOpts struct {
	Script []byte
}

func (o DeriveBlindingKeyPairOpts) validate() error {
	if len(o.Script) <= 0 {
		return ErrNullOutputScript
	}
	return nil
}

// DeriveBlindingKeyPair derives the SLIP77 blinding key pair from the provided
// output script
func (w *Wallet) DeriveBlindingKeyPair(opts DeriveBlindingKeyPairOpts) (
	*btcec.PrivateKey,
	*btcec.PublicKey,
	error,
) {
	if err := opts.validate(); err != nil {
		return nil, nil, err
	}
	if err := w.validate(); err != nil {
		return nil, nil, err
	}
	if len(w.blindingMasterKey) <= 0 {
		return nil, nil, ErrNullBlindingMasterKey
	}
	slip77Node, err := slip77.FromMasterKey(w.blindingMasterKey)
	if err != nil {
		return nil, nil, err
	}
	return slip77Node.DeriveKey(opts.Script)
}

// DeriveConfidentialAddressOpts is the struct given to DeriveConfidentialAddress method
type DeriveConfidentialAddressOpts struct {
	DerivationPath string
	Network        *network.Network
}

func (o DeriveConfidentialAddressOpts) validate() error {
	derivationPath, err := ParseDerivationPath(o.DerivationPath)
	if err != nil {
		return err
	}

	err = checkDerivationPath(derivationPath)
	if err != nil {
		return err
	}

	if o.Network == nil {
		return ErrNullNetwork
	}

	return nil
}

// DeriveConfidentialAddress derives both the signing and blinding pubkeys to
// then generate the corresponding confidential address
func (w *Wallet) DeriveConfidentialAddress(
	opts DeriveConfidentialAddressOpts,
) (string, error) {
	if err := opts.validate(); err != nil {
		return "", err
	}
	if err := w.validate(); err != nil {
		return "", err
	}

	_, pubkey, err := w.DeriveSigningKeyPair(DeriveSigningKeyPairOpts{
		DerivationPath: opts.DerivationPath,
	})
	if err != nil {
		return "", err
	}

	script := payment.FromPublicKey(pubkey, opts.Network, nil).WitnessScript

	_, blindingPubkey, err := w.DeriveBlindingKeyPair(DeriveBlindingKeyPairOpts{
		Script: script,
	})
	if err != nil {
		return "", err
	}

	p2wpkh := payment.FromPublicKey(pubkey, opts.Network, blindingPubkey)

	return p2wpkh.ConfidentialWitnessPubKeyHash()
}

func checkDerivationPath(path DerivationPath) error {
	if len(path) != 3 {
		return ErrInvalidDerivationPathLength
	}
	// first elem must be hardened!
	if path[0] < hdkeychain.HardenedKeyStart {
		return ErrInvalidDerivationPathAccount
	}
	return nil
}
