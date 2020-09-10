package wallet

import (
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
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

// UpdateSwapTxOpts is the struct given to UpdateTx method
type UpdateSwapTxOpts struct {
	PsetBase64           string
	Unspents             []explorer.Utxo
	InputAmount          uint64
	InputAsset           string
	OutputAmount         uint64
	OutputAsset          string
	OutputDerivationPath string
	ChangeDerivationPath string
}

func (o UpdateSwapTxOpts) validate() error {
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
	if _, err := bufferutil.ValueToBytes(o.InputAmount); err != nil {
		return err
	}
	if len(o.InputAsset)/2 != 32 {
		return ErrInvalidInputAsset
	}
	if _, err := bufferutil.AssetHashToBytes(o.InputAsset); err != nil {
		return ErrInvalidInputAsset
	}

	// check input list
	if len(o.Unspents) <= 0 {
		return ErrEmptyUnspents
	}
	for _, in := range o.Unspents {
		_, _, err := in.Parse()
		if err != nil {
			return err
		}
	}

	// check output args
	if o.OutputAmount == 0 {
		return ErrZeroOutputAmount
	}
	if _, err := bufferutil.ValueToBytes(o.OutputAmount); err != nil {
		return err
	}
	if len(o.OutputAsset)/2 != 32 {
		return ErrInvalidOutputAsset
	}
	if _, err := bufferutil.AssetHashToBytes(o.OutputAsset); err != nil {
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

func (o UpdateSwapTxOpts) getUnspentsUnblindingKeys(w *Wallet) ([][]byte, error) {
	keys := make([][]byte, 0, len(o.Unspents))
	for _, u := range o.Unspents {
		blindingPrvkey, _, err := w.DeriveBlindingKeyPair(
			DeriveBlindingKeyPairOpts{
				Script: u.Script(),
			},
		)
		if err != nil {
			return nil, err
		}
		keys = append(keys, blindingPrvkey.Serialize())
	}
	return keys, nil
}

// UpdateSwapTx takes care of adding inputs and output(s) to the provided partial
// transaction. Inputs are selected so that the minimum number of them is used
// to reach the target InputAmount. The subset of selected inputs is returned
// along with the updated partial transaction
func (w *Wallet) UpdateSwapTx(opts UpdateSwapTxOpts) (string, []explorer.Utxo, error) {
	if err := opts.validate(); err != nil {
		return "", nil, err
	}
	if err := w.validate(); err != nil {
		return "", nil, err
	}

	ptx, _ := pset.NewPsetFromBase64(opts.PsetBase64)

	unspentsBlinidingKeys, err := opts.getUnspentsUnblindingKeys(w)
	if err != nil {
		return "", nil, err
	}

	selectedUnspents, change, err := explorer.SelectUnSpents(
		opts.Unspents,
		unspentsBlinidingKeys,
		opts.InputAmount,
		opts.InputAsset,
	)
	if err != nil {
		return "", nil, err
	}

	_, script, _ := w.DeriveConfidentialAddress(DeriveConfidentialAddressOpts{
		DerivationPath: opts.OutputDerivationPath,
		Network:        config.GetNetwork(),
	})
	output, _ := newTxOutput(opts.OutputAsset, opts.OutputAmount, script)

	outputsToAdd := []*transaction.TxOutput{output}
	if change > 0 {
		_, script, _ := w.DeriveConfidentialAddress(DeriveConfidentialAddressOpts{
			DerivationPath: opts.ChangeDerivationPath,
			Network:        config.GetNetwork(),
		})

		changeOutput, _ := newTxOutput(opts.InputAsset, change, script)
		outputsToAdd = append(outputsToAdd, changeOutput)
	}

	psetBase64, err := addInsAndOutsToPset(ptx, selectedUnspents, outputsToAdd)
	if err != nil {
		return "", nil, err
	}

	return psetBase64, selectedUnspents, nil
}
