package swap

import (
	"encoding/hex"
	"errors"

	"github.com/vulpemventures/go-elements/confidential"
	"github.com/vulpemventures/go-elements/elementsutil"
	"github.com/vulpemventures/go-elements/pset"
	"github.com/vulpemventures/go-elements/transaction"
)

func outputFoundInTransaction(outputs []*transaction.TxOutput, value uint64, asset string, ouptutBlindKeys map[string][]byte) (bool, error) {
	for _, output := range outputs {
		// if confidential, unblind before check
		if output.IsConfidential() {
			script := hex.EncodeToString(output.Script)
			blindingPrivateKey, ok := ouptutBlindKeys[script]
			if !ok {
				return false, errors.New("No blinding private key for script: " + script)
			}

			unblinded, err := confidential.UnblindOutputWithKey(output, blindingPrivateKey)
			if err != nil {
				return false, errors.New("Unable to unblind output with script: " + script)
			}

			// check if the unblinded output respects criteria
			if unblinded.Value == value &&
				elementsutil.AssetHashFromBytes(unblinded.Asset) == asset {
				return true, nil
			}
		}
		// unconfidential check
		outAsset := elementsutil.AssetHashFromBytes(output.Asset)
		outValue, _ := elementsutil.ValueFromBytes(output.Value)
		if outValue == value && outAsset == asset {
			return true, nil
		}
	}
	// output not found
	return false, nil
}

func countCumulativeAmount(utxos []pset.PInput, asset string, inputBlindKeys map[string][]byte) (uint64, error) {
	var amount uint64 = 0

	// filter the utxos using assetHash
	filteredUtxos, err := utxosFilteredByAssetHashAndUnblinded(utxos, asset, inputBlindKeys)
	if err != nil {
		return 0, err
	}

	// sum all the filteredUtxos' values
	for _, utxo := range filteredUtxos {
		value, _ := elementsutil.ValueFromBytes(utxo.WitnessUtxo.Value)
		amount += value
	}

	return amount, nil
}

func utxosFilteredByAssetHashAndUnblinded(utxos []pset.PInput, asset string, inputBlindKeys map[string][]byte) ([]pset.PInput, error) {
	filteredUtxos := make([]pset.PInput, 0)

	for _, utxo := range utxos {
		// if confidential, unblind before checking asset hash
		if utxo.WitnessUtxo.IsConfidential() {
			script := hex.EncodeToString(utxo.WitnessUtxo.Script)
			blindKey, ok := inputBlindKeys[script]
			if !ok {
				return nil, errors.New("No blinding private key for script: " + script)
			}

			unblinded, err := confidential.UnblindOutputWithKey(utxo.WitnessUtxo, blindKey)
			if err != nil {
				return nil, errors.New("Unable to unblind output with script: " + script)
			}

			unblindedAsset := hex.EncodeToString(elementsutil.ReverseBytes(unblinded.Asset))
			// replace Asset and Value by unblinded data before append
			if unblindedAsset == asset {
				assetBytes, _ := elementsutil.AssetHashToBytes(unblindedAsset)
				utxo.WitnessUtxo.Asset = assetBytes

				valueBytes, err := elementsutil.ValueToBytes(unblinded.Value)
				if err != nil {
					return nil, err
				}
				utxo.WitnessUtxo.Value = valueBytes

				utxo.WitnessUtxo.RangeProof = make([]byte, 0)
				utxo.WitnessUtxo.SurjectionProof = make([]byte, 0)
				utxo.WitnessUtxo.Nonce = make([]byte, 0)

				filteredUtxos = append(filteredUtxos, utxo)
			}

			continue
		}

		if elementsutil.AssetHashFromBytes(utxo.WitnessUtxo.Asset) == asset {
			filteredUtxos = append(filteredUtxos, utxo)
		}
	}

	return filteredUtxos, nil
}
