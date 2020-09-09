package wallet

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/txscript"
	"github.com/vulpemventures/go-elements/payment"
	"github.com/vulpemventures/go-elements/pset"
)

// SignTransactionOpts is the struct given to SignTransaction method
type SignTransactionOpts struct {
	PsetBase64        string
	DerivationPathMap map[string]string
}

func (o SignTransactionOpts) validate() error {
	ptx, err := pset.NewPsetFromBase64(o.PsetBase64)
	if err != nil {
		return err
	}
	if len(o.DerivationPathMap) <= 0 {
		return ErrEmptyDerivationPaths
	}
	if len(ptx.Inputs) > len(o.DerivationPathMap) {
		return ErrInvalidDerivationPathsLength
	}

	for script, path := range o.DerivationPathMap {
		derivationPath, err := ParseDerivationPath(path)
		if err != nil {
			return fmt.Errorf(
				"invalid derivation path '%s' for script '%s': %v",
				path, script, err,
			)
		}
		err = checkDerivationPath(derivationPath)
		if err != nil {
			return fmt.Errorf(
				"invalid derivation path '%s' for script '%s': %v",
				path, script, err,
			)
		}
	}

	for i, in := range ptx.Inputs {
		script := in.WitnessUtxo.Script
		_, ok := o.DerivationPathMap[hex.EncodeToString(script)]
		if !ok {
			return fmt.Errorf(
				"derivation path not found in list for input %d with script '%s'",
				i,
				script,
			)
		}
	}

	return nil
}

// SignTransaction signs all inputs of a partial transaction using the keys
// derived with the help of the map script:derivation_path
func (w *Wallet) SignTransaction(opts SignTransactionOpts) (string, error) {
	if err := opts.validate(); err != nil {
		return "", err
	}
	if err := w.validate(); err != nil {
		return "", err
	}

	ptx, _ := pset.NewPsetFromBase64(opts.PsetBase64)
	for i, in := range ptx.Inputs {
		path := opts.DerivationPathMap[hex.EncodeToString(in.WitnessUtxo.Script)]
		err := w.signInput(ptx, i, path)
		if err != nil {
			return "", err
		}
	}

	return ptx.ToBase64()
}

// SignInputOpts is the struct given to SignInput method
type SignInputOpts struct {
	PsetBase64     string
	InIndex        uint32
	DerivationPath string
}

func (o SignInputOpts) validate() error {
	if len(o.PsetBase64) <= 0 {
		return ErrNullPset
	}

	ptx, err := pset.NewPsetFromBase64(o.PsetBase64)
	if err != nil {
		return err
	}
	if int(o.InIndex) >= len(ptx.Inputs) {
		return fmt.Errorf(
			"input index must be in range [0, %d]",
			len(ptx.Inputs)-1,
		)
	}
	derivationPath, err := ParseDerivationPath(o.DerivationPath)
	if err != nil {
		return err
	}
	err = checkDerivationPath(derivationPath)
	if err != nil {
		return err
	}

	if ptx.Inputs[o.InIndex].WitnessUtxo == nil {
		return ErrNullInputWitnessUtxo
	}

	return nil
}

// SignInput takes care of producing (and verifying) a signature for a
// specific input of a partial transaction with the provided private key
func (w *Wallet) SignInput(opts SignInputOpts) (string, error) {
	if err := opts.validate(); err != nil {
		return "", err
	}
	if err := w.validate(); err != nil {
		return "", err
	}

	ptx, _ := pset.NewPsetFromBase64(opts.PsetBase64)

	err := w.signInput(ptx, int(opts.InIndex), opts.DerivationPath)
	if err != nil {
		return "", err
	}

	return ptx.ToBase64()
}

func (w *Wallet) signInput(ptx *pset.Pset, inIndex int, derivationPath string) error {
	updater, err := pset.NewUpdater(ptx)
	if err != nil {
		return err
	}

	prvkey, pubkey, err := w.DeriveSigningKeyPair(DeriveSigningKeyPairOpts{
		DerivationPath: derivationPath,
	})
	if err != nil {
		return err
	}

	pay, err := payment.FromScript(ptx.Inputs[inIndex].WitnessUtxo.Script, nil, nil)
	if err != nil {
		return err
	}

	script := pay.Script
	hashForSignature := ptx.UnsignedTx.HashForWitnessV0(
		inIndex,
		script,
		ptx.Inputs[inIndex].WitnessUtxo.Value,
		txscript.SigHashAll,
	)

	signature, err := prvkey.Sign(hashForSignature[:])
	if err != nil {
		return err
	}

	if !signature.Verify(hashForSignature[:], pubkey) {
		return fmt.Errorf(
			"signature verification failed for input %d",
			inIndex,
		)
	}

	sigWithSigHashType := append(signature.Serialize(), byte(txscript.SigHashAll))
	_, err = updater.Sign(
		inIndex,
		sigWithSigHashType,
		pubkey.SerializeCompressed(),
		nil,
		nil,
	)
	if err != nil {
		return err
	}
	return nil
}
