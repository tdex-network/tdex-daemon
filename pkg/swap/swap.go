package swap

import (
	"encoding/hex"
	"errors"

	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/go/tdex/v1"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/transactionutil"
	"github.com/vulpemventures/go-elements/pset"
	"github.com/vulpemventures/go-elements/transaction"
)

func compareMessagesAndTransaction(request *tdexv1.SwapRequest, accept *tdexv1.SwapAccept) error {
	decodedFromRequest, err := pset.NewPsetFromBase64(request.GetTransaction())
	if err != nil {
		return err
	}

	for index, input := range decodedFromRequest.Inputs {
		if input.WitnessUtxo == nil && input.NonWitnessUtxo != nil {
			inputVout := decodedFromRequest.UnsignedTx.Inputs[index].Index
			decodedFromRequest.Inputs[index].WitnessUtxo = input.NonWitnessUtxo.Outputs[inputVout]
		}
	}

	totalP, err := countCumulativeAmount(decodedFromRequest.Inputs, request.GetAssetP(), request.GetInputBlindingKey())
	if err != nil {
		return err
	}
	if totalP < request.GetAmountP() {
		return errors.New("cumulative utxos count is not enough to cover SwapRequest.amount_p")
	}

	outputRFound, err := outputFoundInTransaction(
		decodedFromRequest.UnsignedTx.Outputs,
		request.GetAmountR(),
		request.GetAssetR(),
		request.GetOutputBlindingKey(),
	)
	if err != nil {
		return err
	}
	if !outputRFound {
		return errors.New("either SwapRequest.amount_r or SwapRequest.asset_r do not match the provided pset")
	}

	if accept != nil {
		decodedFromAccept, err := pset.NewPsetFromBase64(accept.GetTransaction())
		if err != nil {
			return err
		}

		if request.GetId() != accept.GetRequestId() {
			return errors.New("id mismatch: SwapRequest.id and SwapAccept.request_id are not the same")
		}

		totalR, err := countCumulativeAmount(decodedFromAccept.Inputs, request.GetAssetR(), accept.GetInputBlindingKey())
		if err != nil {
			return err
		}
		if totalR < request.GetAmountR() {
			return errors.New("cumulative utxos count is not enough to cover SwapRequest.amount_r")
		}

		outputPFound, err := outputFoundInTransaction(
			decodedFromAccept.UnsignedTx.Outputs,
			request.GetAmountP(),
			request.GetAssetP(),
			accept.GetOutputBlindingKey(),
		)
		if !outputPFound {
			return errors.New("either SwapRequest.amount_p or SwapRequest.asset_p do not match the provided pset")
		}
	}

	return nil
}

func outputFoundInTransaction(outputs []*transaction.TxOutput, value uint64, asset string, ouptutBlindKeys map[string][]byte) (bool, error) {
	for _, output := range outputs {
		// if confidential, unblind before check
		if output.IsConfidential() {
			script := hex.EncodeToString(output.Script)
			blindingPrivateKey, ok := ouptutBlindKeys[script]
			if !ok {
				return false, errors.New("No blinding private key for script: " + script)
			}

			unblinded, ok := transactionutil.UnblindOutput(output, blindingPrivateKey)
			if !ok {
				return false, errors.New("Unable to unblind output with script: " + script)
			}

			// check if the unblinded output respect criterias
			if unblinded.Value == value && unblinded.AssetHash == asset {
				return true, nil
			}
		}
		// unconfidential check
		if bufferutil.ValueFromBytes(output.Value) == value && bufferutil.AssetHashFromBytes(output.Asset) == asset {
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
		value := bufferutil.ValueFromBytes(utxo.WitnessUtxo.Value)
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

			unblinded, ok := transactionutil.UnblindOutput(utxo.WitnessUtxo, blindKey)
			if !ok {
				return nil, errors.New("Unable to unblind output with script: " + script)
			}

			// replace Asset and Value by unblinded data before append
			if unblinded.AssetHash == asset {
				assetBytes, err := bufferutil.AssetHashToBytes(unblinded.AssetHash)
				if err != nil {
					return nil, err
				}
				utxo.WitnessUtxo.Asset = assetBytes

				valueBytes, err := bufferutil.ValueToBytes(unblinded.Value)
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

		if bufferutil.AssetHashFromBytes(utxo.WitnessUtxo.Asset) == asset {
			filteredUtxos = append(filteredUtxos, utxo)
		}
	}

	return filteredUtxos, nil
}
