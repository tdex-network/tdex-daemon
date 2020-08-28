package swap

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/novalagung/gubrak/v2"
	"github.com/tdex-network/tdex-daemon/pkg/bufferutil"
	"github.com/tdex-network/tdex-daemon/pkg/transactionutil"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/swap"
	"github.com/vulpemventures/go-elements/pset"
	"github.com/vulpemventures/go-elements/transaction"
)

//Swap defines the initial parameter
type Swap struct {
	Verbose bool
}

func (*Swap) compareMessagesAndTransaction(request *pb.SwapRequest, accept *pb.SwapAccept) error {
	decodedFromRequest, err := pset.NewPsetFromBase64(request.GetTransaction())
	if err != nil {
		return err
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

func outputFoundInTransaction(outs []*transaction.TxOutput, value uint64, asset string, ouptutBlindKeys map[string][]byte) (bool, error) {
	found, err := gubrak.From(outs).
		Find(func(each *transaction.TxOutput) bool {

			if each.IsConfidential() {
				blindKey, ok := ouptutBlindKeys[hex.EncodeToString(each.Script)]
				if !ok {
					return false
				}

				unblinded, ok := transactionutil.UnblindOutput(each, blindKey)
				if !ok {
					return false
				}

				return unblinded.Value == value && unblinded.AssetHash == asset
			}

			return bufferutil.ValueFromBytes(each.Value) == value && bufferutil.AssetHashFromBytes(each.Asset) == asset
		}).ResultAndError()

	if err != nil {
		return false, fmt.Errorf("gubrak: %w", err)
	}

	return found != nil, nil
}

func countCumulativeAmount(utxos []pset.PInput, asset string, inputBlindKeys map[string][]byte) (uint64, error) {
	result, err := gubrak.From(utxos).
		Filter(func(each pset.PInput) bool {
			// TODO check if a nonWitnessUtxo is given

			if each.WitnessUtxo.IsConfidential() {

				blindKey, ok := inputBlindKeys[hex.EncodeToString(each.WitnessUtxo.Script)]
				if !ok {
					return false
				}

				unblinded, ok := transactionutil.UnblindOutput(each.WitnessUtxo, blindKey)
				if !ok {
					return false
				}

				return unblinded.AssetHash == asset
			}

			return bufferutil.AssetHashFromBytes(each.WitnessUtxo.Asset) == asset
		}).
		Map(func(each pset.PInput) uint64 {

			if each.WitnessUtxo.IsConfidential() {

				blindKey, _ := inputBlindKeys[hex.EncodeToString(each.WitnessUtxo.Script)]
				unblinded, _ := transactionutil.UnblindOutput(each.WitnessUtxo, blindKey)

				return unblinded.Value
			}

			return bufferutil.ValueFromBytes(each.WitnessUtxo.Value)
		}).
		Reduce(func(accumulator, value uint64) uint64 {
			return accumulator + value
		}, uint64(0)).
		ResultAndError()

	if err != nil {
		return 0, fmt.Errorf("gubrak: %w", err)
	}

	return result.(uint64), nil
}
