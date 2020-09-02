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
	// ErrNullBlindingMnemonic ...
	ErrNullBlindingMnemonic = errors.New("blinding mnemonic is null")
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

	// ErrNotConfidentialWallet ...
	ErrNotConfidentialWallet = errors.New(
		"wallet must have valid blinding mnemonic and master key for operations " +
			"such blinding a transaction",
	)
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
	signingMnemonic   string
	signingMasterKey  []byte
	blindingMnemonic  string
	blindingMasterKey []byte
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
		signingMnemonic:   signingMnemonic,
		signingMasterKey:  signingMasterKey,
		blindingMnemonic:  blindingMnemonic,
		blindingMasterKey: blindingMasterKey,
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
		signingMnemonic:   opts.SigningMnemonic,
		signingMasterKey:  signingMasterKey,
		blindingMnemonic:  opts.BlindingMnemonic,
		blindingMasterKey: blindingMasterKey,
	}, nil
}

func (w *Wallet) validate() error {
	if len(w.signingMasterKey) <= 0 {
		return ErrNullSigningMasterKey
	}
	if len(w.signingMnemonic) <= 0 {
		return ErrNullSigningMnemonic
	}
	if !isMnemonicValid(w.signingMnemonic) {
		return ErrInvalidSigningMnemonic
	}

	if len(w.blindingMnemonic) > 0 {
		if !isMnemonicValid(w.blindingMnemonic) {
			return ErrInvalidBlindingMnemonic
		}
		if len(w.blindingMasterKey) <= 0 {
			return ErrNullBlindingMasterKey
		}
	}
	return nil
}

// SigningMnemonic is getter for signing mnemonic
func (w *Wallet) SigningMnemonic() (string, error) {
	if err := w.validate(); err != nil {
		return "", err
	}
	return w.signingMnemonic, nil
}

// BlindingMnemonic is getter for blinding mnemonic
func (w *Wallet) BlindingMnemonic() (string, error) {
	if err := w.validate(); err != nil {
		return "", err
	}
	if len(w.blindingMnemonic) <= 0 {
		return "", ErrNullBlindingMnemonic
	}
	return w.blindingMnemonic, nil
}

// IsConfidential returns whether the blinding mnemonic/master key are set
// for the current wallet
func (w *Wallet) IsConfidential() bool {
	return len(w.blindingMnemonic) > 0 && len(w.blindingMasterKey) > 0
}
