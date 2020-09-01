package wallet

import (
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/go-elements/payment"
	"github.com/vulpemventures/go-elements/pset"
	"github.com/vulpemventures/go-elements/transaction"
)

// CreateTx crafts a new empty partial transaction
func (w *Wallet) CreateTx() (string, error) {
	ptx, err := pset.New([]*transaction.TxInput{}, []*transaction.TxOutput{}, 2, 0)
	if err != nil {
		return "", err
	}
	return ptx.ToBase64()
}

// UpdateTxOpts is the struct given to UpdateTx method
type UpdateTxOpts struct {
	PsetBase64           string
	Inputs               []explorer.Utxo
	InputAmount          uint64
	InputAsset           string
	OutputAmount         uint64
	OutputAsset          string
	OutputDerivationPath string
	ChangeDerivationPath string
}

func (o UpdateTxOpts) validate() error {
	if len(o.PsetBase64) <= 0 {
		return ErrNullPset
	}
	_, err := pset.NewPsetFromBase64(o.PsetBase64)
	if err != nil {
		return err
	}

	// check input args
	if o.InputAmount == 0 {
		return ErrZeroInputAmount
	}
	if _, err := valueToBytes(o.InputAmount); err != nil {
		return err
	}
	if len(o.InputAsset)/2 != 32 {
		return ErrInvalidInputAsset
	}
	if _, err := assetHashToBytes(o.InputAsset); err != nil {
		return ErrInvalidInputAsset
	}

	// check input list
	if len(o.Inputs) <= 0 {
		return ErrEmptyInputs
	}
	for _, in := range o.Inputs {
		_, _, err := in.Parse()
		if err != nil {
			return err
		}
	}

	// check output args
	if o.OutputAmount == 0 {
		return ErrZeroOutputAmount
	}
	if _, err := valueToBytes(o.OutputAmount); err != nil {
		return err
	}
	if len(o.OutputAsset)/2 != 32 {
		return ErrInvalidOutputAsset
	}
	if _, err := assetHashToBytes(o.OutputAsset); err != nil {
		return ErrInvalidOutputAsset
	}

	if len(o.OutputDerivationPath) <= 0 {
		return ErrNullOutputDerivationPath
	}
	outputDerivationPath, err := ParseDerivationPath(o.OutputDerivationPath)
	if err != nil {
		return err
	}
	if err := checkDerivationPath(outputDerivationPath); err != nil {
		return err
	}

	if len(o.ChangeDerivationPath) <= 0 {
		return ErrNullChangeDerivationPath
	}
	changeDerivationPath, err := ParseDerivationPath(o.ChangeDerivationPath)
	if err != nil {
		return err
	}
	if err := checkDerivationPath(changeDerivationPath); err != nil {
		return err
	}

	return nil
}

// UpdateTx takes care of adding inputs and output(s) to the provided partial
// transaction. Inputs are selected so that the minimum number of them is used
// to reach the target InputAmount. The subset of selected inputs is returned
// along with the updated partial transaction
func (w *Wallet) UpdateTx(opts UpdateTxOpts) (string, []explorer.Utxo, error) {
	if err := opts.validate(); err != nil {
		return "", nil, err
	}

	ptx, _ := pset.NewPsetFromBase64(opts.PsetBase64)

	unspentsBlinidingKeys := make([][]byte, 0, len(opts.Inputs))
	for _, u := range opts.Inputs {
		blindingPrvkey, _, err := w.DeriveBlindingKeyPair(
			DeriveBlindingKeyPairOpts{
				Script: u.Script(),
			},
		)
		if err != nil {
			return "", nil, err
		}
		unspentsBlinidingKeys = append(unspentsBlinidingKeys, blindingPrvkey.Serialize())
	}

	selectedUnspents, change, err := explorer.SelectUnSpents(
		opts.Inputs,
		unspentsBlinidingKeys,
		opts.InputAmount,
		opts.InputAsset,
	)
	if err != nil {
		return "", nil, err
	}

	updater, err := pset.NewUpdater(ptx)
	if err != nil {
		return "", nil, err
	}

	for i, unspent := range selectedUnspents {
		in, out, _ := unspent.Parse()
		updater.AddInput(in)
		err := updater.AddInWitnessUtxo(out, i)
		if err != nil {
			return "", nil, err
		}
	}

	outAsset, _ := assetHashToBytes(opts.OutputAsset)
	outValue, _ := valueToBytes(opts.OutputAmount)
	_, pubkey, _ := w.DeriveSigningKeyPair(DeriveSigningKeyPairOpts{
		DerivationPath: opts.OutputDerivationPath,
	})
	script := payment.FromPublicKey(pubkey, &network.Liquid, nil).WitnessScript

	output := transaction.NewTxOutput(outAsset, outValue, script)
	updater.AddOutput(output)

	if change > 0 {
		changeValue, _ := valueToBytes(change)
		_, pubkey, _ := w.DeriveSigningKeyPair(DeriveSigningKeyPairOpts{
			DerivationPath: opts.ChangeDerivationPath,
		})
		changeScript :=
			payment.FromPublicKey(pubkey, &network.Liquid, nil).WitnessScript
		inAsset, _ := assetHashToBytes(opts.InputAsset)

		changeOutput := transaction.NewTxOutput(inAsset, changeValue, changeScript)
		updater.AddOutput(changeOutput)
	}

	psetBase64, err := ptx.ToBase64()
	if err != nil {
		return "", nil, err
	}

	return psetBase64, selectedUnspents, nil
}
