package wallet

import (
	"encoding/hex"
	"fmt"

	"github.com/vulpemventures/go-elements/pset"
)

// BlindTransactionOpts is the struct given to BlindTransaction method
type BlindTransactionOpts struct {
	PsetBase64         string
	OutputBlindingKeys [][]byte
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
	return nil
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

	blinder, err := pset.NewBlinder(
		ptx,
		inputBlindingKeys,
		opts.OutputBlindingKeys,
		nil,
		nil,
	)
	if err != nil {
		return "", err
	}

	err = blinder.Blind()
	if err != nil {
		return "", err
	}
	return ptx.ToBase64()
}

// BlindSwapTransactionOpts is the struct given to BlindSwapTransaction method
type BlindSwapTransactionOpts struct {
	PsetBase64         string
	InputBlindingKeys  map[string][]byte
	OutputBlindingKeys map[string][]byte
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

	return nil
}

// BlindSwapTransaction blinds the outputs of a swap transaction. Since this
// type of transaciton is composed of inputs and outputs owned by 2 different
// parties, the blinding keys for inputs and outputs are provided through maps
// outputScript -> blinding key. Thus, the wallet won't derive any key from the
// scripts of inputs and outputs of the provided transaction.
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
		outputBlindingKeys = append(outputBlindingKeys, opts.OutputBlindingKeys[script])
	}

	blinder, err := pset.NewBlinder(
		ptx,
		inputBlindingKeys,
		outputBlindingKeys,
		nil,
		nil,
	)
	if err != nil {
		return "", err
	}

	err = blinder.Blind()
	if err != nil {
		return "", err
	}
	return ptx.ToBase64()
}
