package swap

import (
	"encoding/hex"
	"fmt"

	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/go/tdex/v1"
	"github.com/thanhpk/randstr"
	"github.com/vulpemventures/go-elements/pset"
	"github.com/vulpemventures/go-elements/transaction"
	"google.golang.org/protobuf/proto"
)

// CompleteOpts is the struct given to the Complete method
type CompleteOpts struct {
	Message     []byte
	Transaction string
}

// Complete takes a CompleteOpts and returns the id of the SwapComplete entity
// and its serialized version
func Complete(complete CompleteOpts) (string, []byte, error) {
	var msgAccept tdexv1.SwapAccept
	err := proto.Unmarshal(complete.Message, &msgAccept)
	if err != nil {
		return "", nil, fmt.Errorf("unmarshal swap accept %w", err)
	}

	if !isHex(complete.Transaction) {
		ptx, err := pset.NewPsetFromBase64(complete.Transaction)
		if err != nil {
			return "", nil, fmt.Errorf("transaction in base64 format is invalid")
		}

		ok, err := ptx.ValidateAllSignatures()
		if err != nil {
			return "", nil, err
		}
		if !ok {
			return "", nil, fmt.Errorf("transaction contains invalid signatures")
		}
	} else {
		if _, err := transaction.NewTxFromHex(complete.Transaction); err != nil {
			return "", nil, fmt.Errorf("transaction in hex format is invalid")
		}
	}

	randomID := randstr.Hex(8)
	msgComplete := &tdexv1.SwapComplete{
		Id:          randomID,
		AcceptId:    msgAccept.GetId(),
		Transaction: complete.Transaction,
	}

	msgCompleteSerialized, err := proto.Marshal(msgComplete)
	if err != nil {
		return "", nil, err
	}

	return randomID, msgCompleteSerialized, nil
}

// ValidateCompletePsetOpts is the struct given to the ValidateCompletePset method
type ValidateCompletePsetOpts struct {
	PsetBase64         string
	InputBlindingKeys  map[string][]byte
	OutputBlindingKeys map[string][]byte
	SwapRequest        *tdexv1.SwapRequest
}

// ValidateCompletePset takes a VerifyCompeltePsetOpts and returns whether the
// final signed pset matches the original SwapRequest message
func ValidateCompletePset(opts ValidateCompletePsetOpts) error {
	ptx, err := pset.NewPsetFromBase64(opts.PsetBase64)
	if err != nil {
		return err
	}

	swapRequest := opts.SwapRequest

	totalP, err := countCumulativeAmount(ptx.Inputs, swapRequest.GetAssetP(), opts.InputBlindingKeys)
	if err != nil {
		return err
	}
	if totalP < swapRequest.GetAmountP() {
		return fmt.Errorf("cumulative utxos count is not enough to cover SwapRequest.amount_p")
	}

	outputRFound, err := outputFoundInTransaction(
		ptx.UnsignedTx.Outputs,
		swapRequest.GetAmountR(),
		swapRequest.GetAssetR(),
		opts.OutputBlindingKeys,
	)
	if err != nil {
		return err
	}
	if !outputRFound {
		return fmt.Errorf("either SwapRequest.amount_r or SwapRequest.asset_r do not match the provided pset")
	}

	totalR, err := countCumulativeAmount(ptx.Inputs, swapRequest.GetAssetR(), opts.InputBlindingKeys)
	if err != nil {
		return err
	}
	if totalR < swapRequest.GetAmountR() {
		return fmt.Errorf("cumulative utxos count is not enough to cover SwapRequest.amount_r")
	}

	outputPFound, err := outputFoundInTransaction(
		ptx.UnsignedTx.Outputs,
		swapRequest.GetAmountP(),
		swapRequest.GetAssetP(),
		opts.OutputBlindingKeys,
	)
	if !outputPFound {
		return fmt.Errorf("either SwapRequest.amount_p or SwapRequest.asset_p do not match the provided pset")
	}
	return nil
}

func isHex(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil
}
