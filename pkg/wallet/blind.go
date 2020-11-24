package wallet

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/vulpemventures/go-elements/pset"
)

const (
	// MaxBlindingAttempts is the max number of times the blinding of a pset
	// can be repeated in case it fails to generate valid proofs.
	MaxBlindingAttempts = 8
	// DefaultBlindingAttempts is the default number of times the blinding of a
	// pset is retried if it fails to generate valid proofs.
	DefaultBlindingAttempts = 4
)

// BlindTransactionOpts is the struct given to BlindTransaction method
type BlindTransactionOpts struct {
	PsetBase64         string
	OutputBlindingKeys [][]byte
	Attempts           int
}

func (o BlindTransactionOpts) validate() error {
	if len(o.PsetBase64) <= 0 {
		return ErrNullPset
	}
	ptx, err := pset.NewPsetFromBase64(o.PsetBase64)
	if err != nil {
		return err
	}
	for _, in := range ptx.Inputs {
		if in.WitnessUtxo == nil {
			return ErrNullInputWitnessUtxo
		}
	}

	if len(o.OutputBlindingKeys) != len(ptx.Outputs) {
		return ErrInvalidOutputBlindingKeysLen
	}

	if o.Attempts < 0 || o.Attempts > MaxBlindingAttempts {
		return ErrInvalidAttempts
	}
	return nil
}

func (o BlindTransactionOpts) maxAttempts() int {
	if o.Attempts == 0 {
		return DefaultBlindingAttempts
	}
	return o.Attempts
}

// BlindTransaction blinds the outputs of the provided partial transaction
// by deriving the blinding keys from the output scripts following SLIP-77 spec
func (w *Wallet) BlindTransaction(opts BlindTransactionOpts) (string, error) {
	if err := opts.validate(); err != nil {
		return "", err
	}
	if err := w.validate(); err != nil {
		return "", err
	}

	ptx, _ := pset.NewPsetFromBase64(opts.PsetBase64)

	inputBlindingKeys := make([][]byte, 0, len(ptx.Inputs))
	for _, in := range ptx.Inputs {
		blindingPrvkey, _, _ := w.DeriveBlindingKeyPair(DeriveBlindingKeyPairOpts{
			Script: in.WitnessUtxo.Script,
		})
		inputBlindingKeys = append(inputBlindingKeys, blindingPrvkey.Serialize())
	}

	if err := w.blindTransaction(
		ptx,
		inputBlindingKeys,
		opts.OutputBlindingKeys,
		opts.maxAttempts(),
	); err != nil {
		return "", err
	}

	return ptx.ToBase64()
}

// BlindSwapTransactionOpts is the struct given to BlindSwapTransaction method
type BlindSwapTransactionOpts struct {
	PsetBase64         string
	InputBlindingKeys  map[string][]byte
	OutputBlindingKeys map[string][]byte
	Attempts           int
}

func (o BlindSwapTransactionOpts) validate() error {
	if len(o.PsetBase64) <= 0 {
		return ErrNullPset
	}
	ptx, err := pset.NewPsetFromBase64(o.PsetBase64)
	if err != nil {
		return err
	}

	for i, in := range ptx.Inputs {
		script := hex.EncodeToString(in.WitnessUtxo.Script)
		if _, ok := o.InputBlindingKeys[script]; !ok {
			return fmt.Errorf(
				"missing blinding key for input %d with script '%s'", i, script,
			)
		}
	}
	for i, out := range ptx.UnsignedTx.Outputs {
		script := hex.EncodeToString(out.Script)
		if _, ok := o.OutputBlindingKeys[script]; !ok {
			return fmt.Errorf(
				"missing blinding key for output %d with script '%s'", i, script,
			)
		}
	}

	if o.Attempts < 0 || o.Attempts > MaxBlindingAttempts {
		return ErrInvalidAttempts
	}

	return nil
}

func (o BlindSwapTransactionOpts) maxAttempts() int {
	if o.Attempts == 0 {
		return DefaultBlindingAttempts
	}
	return o.Attempts
}

// BlindSwapTransaction blinds the outputs of a swap transaction. Since this
// type of transaciton is composed of inputs and outputs owned by 2 different
// parties, the blinding keys for inputs and outputs are provided through maps
// outputScript -> blinding key. Note that all the blinding keys provided must
// be private, thus for the outputs this function will use the provided
// blinding keys to get the list of all public keys. This of course also means
// that no blinding keys are derived internally, but these are all provided as
// function arguments.
func (w *Wallet) BlindSwapTransaction(opts BlindSwapTransactionOpts) (string, error) {
	if err := opts.validate(); err != nil {
		return "", err
	}
	if err := w.validate(); err != nil {
		return "", err
	}

	ptx, _ := pset.NewPsetFromBase64(opts.PsetBase64)

	inputBlindingKeys := make([][]byte, 0, len(ptx.Inputs))
	for _, in := range ptx.Inputs {
		script := hex.EncodeToString(in.WitnessUtxo.Script)
		inputBlindingKeys = append(inputBlindingKeys, opts.InputBlindingKeys[script])
	}

	outputBlindingKeys := make([][]byte, 0, len(ptx.Outputs))
	for _, out := range ptx.UnsignedTx.Outputs {
		script := hex.EncodeToString(out.Script)
		_, pubkey := btcec.PrivKeyFromBytes(btcec.S256(), opts.OutputBlindingKeys[script])
		outputBlindingKeys = append(outputBlindingKeys, pubkey.SerializeCompressed())
	}

	if err := w.blindTransaction(
		ptx,
		inputBlindingKeys,
		outputBlindingKeys,
		opts.maxAttempts(),
	); err != nil {
		return "", err
	}

	return ptx.ToBase64()
}

func (w *Wallet) blindTransaction(
	ptx *pset.Pset,
	inBlindingKeys, outBlindingKeys [][]byte,
	maxAttempts int,
) error {
	blinder, err := pset.NewBlinder(
		ptx,
		inBlindingKeys,
		outBlindingKeys,
		nil,
		nil,
	)
	if err != nil {
		return err
	}

	retryCount := 0
	for {
		if retryCount >= maxAttempts {
			return ErrReachedMaxBlindingAttempts
		}

		if err := blinder.Blind(); err != nil {
			if err == pset.ErrGenerateSurjectionProof {
				retryCount++
				continue
			}
			return err
		}
		return nil
	}
}
