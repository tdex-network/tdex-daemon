package wallet

type NewMnemonicOpts struct {
	EntropySize int
}

func (o NewMnemonicOpts) validate() error {
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

// NewMnemonic returns a new mnemonic as a list of words
func NewMnemonic(opts NewMnemonicOpts) ([]string, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	if opts.EntropySize == 0 {
		opts.EntropySize = 128
	}

	return generateMnemonic(opts.EntropySize)
}
