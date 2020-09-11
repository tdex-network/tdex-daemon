package wallet

import (
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
