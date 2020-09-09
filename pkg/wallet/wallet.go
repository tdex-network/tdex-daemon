package wallet

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
	EntropySize   int
	ExtraMnemonic bool
}

func (o NewWalletOpts) validate() error {
	if o.EntropySize > 0 {
		if o.EntropySize < 128 || o.EntropySize > 256 || o.EntropySize%32 != 0 {
			return ErrInvalidEntropySize
		}
	}
	if o.EntropySize < 0 {
		return ErrInvalidEntropySize
	}
	return nil
}

// NewWallet creates a new wallet holding signing/blinding mnemonic and seed
func NewWallet(opts NewWalletOpts) (*Wallet, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	if opts.EntropySize == 0 {
		opts.EntropySize = 128
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

	blindingMnemonic := signingMnemonic
	blindingSeed := signingSeed
	if opts.ExtraMnemonic {
		blindingMnemonic, blindingSeed, err = generateMnemonicSeedAndMasterKey(opts.EntropySize)
		if err != nil {
			return nil, err
		}
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
	blindingMnemonic := opts.SigningMnemonic
	if len(opts.BlindingMnemonic) > 0 {
		blindingMnemonic = opts.BlindingMnemonic
	}

	blindingSeed = generateSeedFromMnemonic(blindingMnemonic)
	blindingMasterKey, err = generateBlindingMasterKey(blindingSeed)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		signingMnemonic:   opts.SigningMnemonic,
		signingMasterKey:  signingMasterKey,
		blindingMnemonic:  blindingMnemonic,
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
