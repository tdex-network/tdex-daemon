package wallet

import (
	"encoding/hex"
	"fmt"

	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/vulpemventures/go-elements/network"
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
	Network              *network.Network
}

func (o UpdateSwapTxOpts) validate() error {
	if len(o.PsetBase64) <= 0 {
		return ErrNullPset
	}
	_, err := pset.NewPsetFromBase64(o.PsetBase64)
	if err != nil {
		return err
	}

	if o.Network == nil {
		return ErrNullNetwork
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

	selectedUnspents, change, err := explorer.SelectUnspents(
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
		Network:        opts.Network,
	})
	output, _ := newTxOutput(opts.OutputAsset, opts.OutputAmount, script)

	outputsToAdd := []*transaction.TxOutput{output}
	if change > 0 {
		_, script, _ := w.DeriveConfidentialAddress(DeriveConfidentialAddressOpts{
			DerivationPath: opts.ChangeDerivationPath,
			Network:        opts.Network,
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

// UpdateTxOpts is the struct given to UpdateTx method
type UpdateTxOpts struct {
	PsetBase64           string
	Unspents             []explorer.Utxo
	Outputs              []*transaction.TxOutput
	ChangePathsByAsset   map[string]string
	MilliSatsPerBytes    int
	Network              *network.Network
	WantPrivateBlindKeys bool
}

func (o UpdateTxOpts) validate() error {
	if len(o.PsetBase64) <= 0 {
		return ErrNullPset
	}
	_, err := pset.NewPsetFromBase64(o.PsetBase64)
	if err != nil {
		return err
	}

	if o.Network == nil {
		return ErrNullNetwork
	}

	if len(o.Unspents) > 0 {
		for _, in := range o.Unspents {
			_, _, err := in.Parse()
			if err != nil {
				return err
			}
		}

		if len(o.ChangePathsByAsset) <= 0 {
			return ErrNullChangePathsByAsset
		}

		for _, out := range o.Outputs {
			asset := bufferutil.AssetHashFromBytes(out.Asset)
			if _, ok := o.ChangePathsByAsset[asset]; !ok {
				return fmt.Errorf("missing derivation path for eventual change of asset '%s'", asset)
			}
		}

		// make sure that a change path for LBTC exists. It will be used for both an
		// an eventual change and for fee change (summed together)
		lbtcAsset := o.Network.AssetID
		if _, ok := o.ChangePathsByAsset[lbtcAsset]; !ok {
			return fmt.Errorf("missing derivation path for eventual change of asset '%s'", lbtcAsset)
		}

		if o.MilliSatsPerBytes < 100 {
			return ErrInvalidMilliSatsPerBytes
		}
	}

	return nil
}

func (o UpdateTxOpts) getOutputsTotalAmountsByAsset() map[string]uint64 {
	totalAmountsByAsset := map[string]uint64{}
	for _, out := range o.Outputs {
		asset := bufferutil.AssetHashFromBytes(out.Asset)
		totalAmountsByAsset[asset] += bufferutil.ValueFromBytes(out.Value)
	}
	return totalAmountsByAsset
}

func (o UpdateTxOpts) getUnspentsUnblindingKeys(w *Wallet) [][]byte {
	keys := make([][]byte, 0, len(o.Unspents))
	for _, u := range o.Unspents {
		blindingPrvkey, _, _ := w.DeriveBlindingKeyPair(
			DeriveBlindingKeyPairOpts{
				Script: u.Script(),
			},
		)
		keys = append(keys, blindingPrvkey.Serialize())
	}
	return keys
}

func (o UpdateTxOpts) getInputAssets() []string {
	assets := make([]string, 0, len(o.ChangePathsByAsset))
	for asset := range o.ChangePathsByAsset {
		assets = append(assets, asset)
	}
	return assets
}

// UpdateTxResult is the struct returned by UpdateTx method.
// PsetBase64: the updated partial transaction with new inputs and outputs
// SelectedUnspents: the list of unspents added as inputs to the pset
// ChangeOutptusBlindingKeys: the list of blinding keys for the evnutal
// 														change(s) added to the pset
// FeeAmount: the amount in satoshi of the fee amount that can added in a
//						second moment giving the user the possibility to eventually blind
//						the pset first
type UpdateTxResult struct {
	PsetBase64                string
	SelectedUnspents          []explorer.Utxo
	ChangeOutputsBlindingKeys map[string][]byte
	FeeAmount                 uint64
}

// UpdateTx adds the provided outputs and eventual inputs to the provided
// partial transaction. The assets of the inputs to add is determined by the
// assets of the provided outputs. For each assset type a derivation path for
// an eventual change must be provided.
// Its also mandatory to provide a derivation path for the LBTC asset type
// since this method takes care of adding inputs (if necessary) for covering
// the fee amount.
// While the list of outputs is required, the list of unspents is optional.
// In case it's not empty, a coin selection is performed for each type of
// asset, adding the eventual change output to the list of outputs to add to
// the tx. In the other case, only the outputs are added to the provided
// partial transaction.
func (w *Wallet) UpdateTx(opts UpdateTxOpts) (*UpdateTxResult, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	if err := w.validate(); err != nil {
		return nil, err
	}

	ptx, _ := pset.NewPsetFromBase64(opts.PsetBase64)
	inputsToAdd := make([]explorer.Utxo, 0)
	outputsToAdd := make([]*transaction.TxOutput, len(opts.Outputs))
	changeOutputsBlindingKeys := map[string][]byte{}
	feeAmount := uint64(0)
	copy(outputsToAdd, opts.Outputs)

	if len(opts.Unspents) > 0 {
		// retrieve all the asset hashes of input to add to the pset
		inAssets := opts.getInputAssets()
		// calculate target amount of each asset for coin selection
		totalAmountsByAsset := opts.getOutputsTotalAmountsByAsset()
		// retrieve input prv blinding keys
		unspentsBlinidingKeys := opts.getUnspentsUnblindingKeys(w)

		// select unspents and update the list of inputs to add and eventually the
		// list of outputs to add by adding the change output if necessary
		for _, asset := range inAssets {
			if totalAmountsByAsset[asset] > 0 {
				selectedUnspents, change, err := explorer.SelectUnspents(
					opts.Unspents,
					unspentsBlinidingKeys,
					totalAmountsByAsset[asset],
					asset,
				)
				if err != nil {
					return nil, err
				}
				inputsToAdd = append(inputsToAdd, selectedUnspents...)

				if change > 0 {
					_, script, _ := w.DeriveConfidentialAddress(
						DeriveConfidentialAddressOpts{
							DerivationPath: opts.ChangePathsByAsset[asset],
							Network:        opts.Network,
						},
					)

					changeOutput, _ := newTxOutput(asset, change, script)
					outputsToAdd = append(outputsToAdd, changeOutput)

					prvBlindingKey, pubBlindingKey, err := w.DeriveBlindingKeyPair(
						DeriveBlindingKeyPairOpts{
							Script: script,
						})
					if err != nil {
						return nil, err
					}
					if opts.WantPrivateBlindKeys {
						changeOutputsBlindingKeys[hex.EncodeToString(script)] =
							prvBlindingKey.Serialize()
					} else {
						changeOutputsBlindingKeys[hex.EncodeToString(script)] =
							pubBlindingKey.SerializeCompressed()
					}
				}
			}
		}

		_, lbtcChangeScript, _ := w.DeriveConfidentialAddress(
			DeriveConfidentialAddressOpts{
				DerivationPath: opts.ChangePathsByAsset[opts.Network.AssetID],
				Network:        opts.Network,
			},
		)

		feeAmount = estimateTxSize(
			len(inputsToAdd)+len(ptx.Inputs),
			len(outputsToAdd)+len(ptx.Outputs),
			!anyOutputWithScript(outputsToAdd, lbtcChangeScript),
			opts.MilliSatsPerBytes,
		)

		// if a LBTC change output already exists and its value covers the
		// estimated fee amount, it's enough to add the fee output and updating
		// the change output's value by subtracting the fee amount.
		// Otherwise, another coin selection over those LBTC utxos not already
		// included is necessary and the already existing change output's value
		// will be eventually updated by adding the change amount returned by the
		// coin selection
		if anyOutputWithScript(outputsToAdd, lbtcChangeScript) {
			changeOutputIndex := outputIndexByScript(outputsToAdd, lbtcChangeScript)
			changeAmount := bufferutil.ValueFromBytes(outputsToAdd[changeOutputIndex].Value)
			if feeAmount < changeAmount {
				outputsToAdd[changeOutputIndex].Value, _ = bufferutil.ValueToBytes(changeAmount - feeAmount)
			} else {
				unspents := getRemainingUnspents(opts.Unspents, inputsToAdd)
				selectedUnspents, change, err := explorer.SelectUnspents(
					unspents,
					unspentsBlinidingKeys,
					feeAmount,
					opts.Network.AssetID,
				)
				if err != nil {
					return nil, err
				}
				inputsToAdd = append(inputsToAdd, selectedUnspents...)

				if change > 0 {
					outputsToAdd[changeOutputIndex].Value, _ = bufferutil.ValueToBytes(changeAmount + change)
				}
			}
		} else {
			// In case there's no LBTC change, it's necessary to choose some other
			// unspents from those not yet selected, add it/them to the list of
			// inputs to add to the tx and add another output for the eventual change
			// returned by the coin selection
			unspents := getRemainingUnspents(opts.Unspents, inputsToAdd)
			selectedUnspents, change, err := explorer.SelectUnspents(
				unspents,
				unspentsBlinidingKeys,
				feeAmount,
				opts.Network.AssetID,
			)
			if err != nil {
				return nil, err
			}
			inputsToAdd = append(inputsToAdd, selectedUnspents...)

			if change > 0 {
				lbtcChangeOutput, _ := newTxOutput(
					opts.Network.AssetID,
					change,
					lbtcChangeScript,
				)
				outputsToAdd = append(outputsToAdd, lbtcChangeOutput)

				lbtcChangePrvBlindingKey, lbtcChangePubBlindingKey, _ := w.DeriveBlindingKeyPair(
					DeriveBlindingKeyPairOpts{
						Script: lbtcChangeScript,
					},
				)

				if opts.WantPrivateBlindKeys {
					changeOutputsBlindingKeys[hex.EncodeToString(lbtcChangeScript)] =
						lbtcChangePrvBlindingKey.Serialize()
				} else {
					changeOutputsBlindingKeys[hex.EncodeToString(lbtcChangeScript)] =
						lbtcChangePubBlindingKey.SerializeCompressed()
				}
			}
		}
	}

	psetBase64, err := addInsAndOutsToPset(ptx, inputsToAdd, outputsToAdd)
	if err != nil {
		return nil, err
	}

	return &UpdateTxResult{
		PsetBase64:                psetBase64,
		SelectedUnspents:          inputsToAdd,
		ChangeOutputsBlindingKeys: changeOutputsBlindingKeys,
		FeeAmount:                 feeAmount,
	}, nil
}

// FinalizeAndExtractTransactionOpts is the struct given to FinalizeAndExtractTransaction method
type FinalizeAndExtractTransactionOpts struct {
	PsetBase64 string
}

func (o FinalizeAndExtractTransactionOpts) validate() error {
	if _, err := pset.NewPsetFromBase64(o.PsetBase64); err != nil {
		return err
	}
	return nil
}

// FinalizeAndExtractTransaction attempts to finalize the provided partial
// transaction and eventually extracts the final transaction and returns
// it in hex string format, along with its transaction id
func FinalizeAndExtractTransaction(opts FinalizeAndExtractTransactionOpts) (string, string, error) {
	ptx, _ := pset.NewPsetFromBase64(opts.PsetBase64)

	ok, err := ptx.ValidateAllSignatures()
	if err != nil {
		return "", "", err
	}
	if !ok {
		return "", "", ErrInvalidSignatures
	}

	if err := pset.FinalizeAll(ptx); err != nil {
		return "", "", err
	}

	tx, err := pset.Extract(ptx)
	if err != nil {
		return "", "", err
	}
	txHex, err := tx.ToHex()
	if err != nil {
		return "", "", err
	}
	return txHex, tx.TxHash().String(), nil
}
