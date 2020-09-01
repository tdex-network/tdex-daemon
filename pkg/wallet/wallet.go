package wallet

import (
	"errors"
	"fmt"
)

var (
	// ErrNullNetwork ...
	ErrNullNetwork = errors.New("network params are null")
	// ErrNullInputWitnessUtxo ...
	ErrNullInputWitnessUtxo = errors.New("input witness utxo must not be null")
	// ErrNullSigningMnemonic ...
	ErrNullSigningMnemonic = errors.New("signing mnemonic is null")
	// ErrNullSigningSeed ...
	ErrNullSigningSeed = errors.New("signing seed is null")
	// ErrNullSigningMasterKey ...
	ErrNullSigningMasterKey = errors.New("signing master key is null")
	// ErrNullBlindingMasterKey ...
	ErrNullBlindingMasterKey = errors.New("blinding master key is null")
	// ErrNullBlindingSeed ...
	ErrNullBlindingSeed = errors.New("blinding seed is null")
	// ErrNullPassphrase ...
	ErrNullPassphrase = errors.New("passphrase must not be null")
	// ErrNullPlainText ...
	ErrNullPlainText = errors.New("text to encrypt must not be null")
	// ErrNullCypherText ...
	ErrNullCypherText = errors.New("cypher to decrypt must not be null")
	// ErrNullDerivationPath ...
	ErrNullDerivationPath = errors.New("derivation path must not be null")
	// ErrNullOutputDerivationPath ...
	ErrNullOutputDerivationPath = fmt.Errorf("output %v", ErrNullDerivationPath)
	// ErrNullChangeDerivationPath ...
	ErrNullChangeDerivationPath = fmt.Errorf("change %v", ErrNullDerivationPath)
	// ErrNullOutputScript ...
	ErrNullOutputScript = errors.New("output script must not be null")
	// ErrNullPset ...
	ErrNullPset = errors.New("pset base64 must not be null")

	// ErrInvalidSigningMnemonic ...
	ErrInvalidSigningMnemonic = errors.New("signing mnemonic is invalid")
	// ErrInvalidEntropySize ...
	ErrInvalidEntropySize = errors.New(
		"entropy size must be a multiple of 32 in the range [128,256]",
	)
	// ErrInvalidBlindingMnemonic ...
	ErrInvalidBlindingMnemonic = errors.New("blinding mnemonic is invalid")
	// ErrInvalidDerivationPathsLength ...
	ErrInvalidDerivationPathsLength = errors.New(
		"length of tx inputs and derivation paths must match",
	)
	// ErrInvalidCypherText ...
	ErrInvalidCypherText = errors.New("cypher must be in base64 format")
	// ErrInvalidDerivationPath ...
	ErrInvalidDerivationPath = errors.New("invalid derivation path")
	// ErrInvalidDerivationPathLength ...
	ErrInvalidDerivationPathLength = errors.New(
		"derivation path must be a relative path in the form \"account'/branch/index\"",
	)
	// ErrInvalidDerivationPathAccount ...
	ErrInvalidDerivationPathAccount = errors.New(
		"derivation path's account (first elem) must be hardened (suffix \"'\")",
	)
	// ErrInvalidInputAsset ...
	ErrInvalidInputAsset = errors.New("input asset must be a 32 byte array in hex format")
	// ErrInvalidOutputAsset ...
	ErrInvalidOutputAsset = errors.New("output asset must be a 32 byte array in hex format")
	// ErrInvalidOutputAddress ...
	ErrInvalidOutputAddress = errors.New("output address must be a valid address")
	// ErrInvalidChangeAddress ...
	ErrInvalidChangeAddress = errors.New("change address must be a valid address")

	// ErrEmptyDerivationPaths ...
	ErrEmptyDerivationPaths = errors.New("derivation path list must not be empty")
	// ErrEmptyInputs ...
	ErrEmptyInputs = errors.New("input list must not be empty")

	// ErrMalformedDerivationPath ...
	ErrMalformedDerivationPath = errors.New(
		"path must not start or end with a '/' and " +
			"can optionally start with 'm/' for absolute paths",
	)
	// ErrZeroInputAmount ...
	ErrZeroInputAmount = errors.New("input amount must not be zero")
	// ErrZeroOutputAmount ...
	ErrZeroOutputAmount = errors.New("output amount must not be zero")
)

// Wallet data structure allows to create a new wallet from seed/mnemonic,
// derive signing and blinding key pairs, and manage those keys to blind and
// sign transactions
type Wallet struct {
	SigningMnemonic   string
	SigningSeed       []byte
	SigningMasterKey  []byte
	BlindingMnemonic  string
	BlindingSeed      []byte
	BlindingMasterKey []byte
}

// NewWalletOpts is the struct given to the NewWallet method
type NewWalletOpts struct {
	EntropySize int
}

func (o NewWalletOpts) validate() error {
	if o.EntropySize < 128 || o.EntropySize > 256 || o.EntropySize%32 != 0 {
		return ErrInvalidEntropySize
	}
	return nil
}

// NewWallet creates a new wallet holding signing/blinding mnemonic and seed
func NewWallet(opts NewWalletOpts) (*Wallet, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	signingMnemonic, signingSeed, err :=
		generateMnemonicSeedAndMasterKey(opts.EntropySize)
	if err != nil {
		return nil, err
	}
	signingMasterKey, err := generateSigningMasterKey(
		signingSeed,
		DefaultBaseDerivationPath,
	)
	if err != nil {
		return nil, err
	}

	blindingMnemonic, blindingSeed, err := generateMnemonicSeedAndMasterKey(opts.EntropySize)
	if err != nil {
		return nil, err
	}
	blindingMasterKey, err := generateBlindingMasterKey(blindingSeed)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		SigningMnemonic:   signingMnemonic,
		SigningSeed:       signingSeed,
		SigningMasterKey:  signingMasterKey,
		BlindingMnemonic:  blindingMnemonic,
		BlindingSeed:      blindingSeed,
		BlindingMasterKey: blindingMasterKey,
	}, nil
}

// NewWalletFromMnemonicOpts is the struct given to the NewWalletFromMnemonicOpts method
type NewWalletFromMnemonicOpts struct {
	SigningMnemonic  string
	BlindingMnemonic string
}

func (o NewWalletFromMnemonicOpts) validate() error {
	if len(o.SigningMnemonic) <= 0 {
		return ErrNullSigningMnemonic
	}
	if !isMnemonicValid(o.SigningMnemonic) {
		return ErrInvalidSigningMnemonic
	}
	if len(o.BlindingMnemonic) > 0 {
		if !isMnemonicValid(o.BlindingMnemonic) {
			return ErrInvalidBlindingMnemonic
		}
	}
	return nil
}

// NewWalletFromMnemonic generates the sigining and (optionally) blinding seeds
// from the corresponding mnemonics provided
func NewWalletFromMnemonic(opts NewWalletFromMnemonicOpts) (*Wallet, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	signingSeed := generateSeedFromMnemonic(opts.SigningMnemonic)
	signingMasterKey, err := generateSigningMasterKey(
		signingSeed,
		DefaultBaseDerivationPath,
	)
	if err != nil {
		return nil, err
	}

	blindingSeed := make([]byte, 0)
	blindingMasterKey := make([]byte, 0)
	if len(opts.BlindingMnemonic) > 0 {
		blindingSeed = generateSeedFromMnemonic(opts.BlindingMnemonic)
		blindingMasterKey, err = generateBlindingMasterKey(blindingSeed)
		if err != nil {
			return nil, err
		}
	}

	return &Wallet{
		SigningMnemonic:   opts.SigningMnemonic,
		SigningSeed:       signingSeed,
		SigningMasterKey:  signingMasterKey,
		BlindingMnemonic:  opts.BlindingMnemonic,
		BlindingSeed:      blindingSeed,
		BlindingMasterKey: blindingMasterKey,
	}, nil
}

func (w *Wallet) validate() error {
	if len(w.SigningSeed) <= 0 {
		return ErrNullSigningSeed
	}
	if len(w.SigningMasterKey) <= 0 {
		return ErrNullSigningMasterKey
	}
	if len(w.SigningMnemonic) <= 0 {
		return ErrNullSigningMnemonic
	}
	if !isMnemonicValid(w.SigningMnemonic) {
		return ErrInvalidSigningMnemonic
	}

	if len(w.BlindingMnemonic) > 0 {
		if !isMnemonicValid(w.BlindingMnemonic) {
			return ErrInvalidBlindingMnemonic
		}
		if len(w.BlindingSeed) <= 0 {
			return ErrNullBlindingSeed
		}
		if len(w.BlindingMasterKey) <= 0 {
			return ErrNullBlindingMasterKey
		}
	}
	return nil
}
