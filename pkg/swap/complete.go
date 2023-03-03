package swap

import (
	"encoding/hex"
	"fmt"

	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v1"
	tdexv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v2"
	"github.com/thanhpk/randstr"
	"github.com/vulpemventures/go-elements/pset"
	"github.com/vulpemventures/go-elements/psetv2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// CompleteOpts is the struct given to the Complete method
type CompleteOpts struct {
	Message     []byte
	Transaction string
}

func (o CompleteOpts) validate() error {
	if len(o.Transaction) <= 0 {
		return fmt.Errorf("missing swap transaction")
	}
	if !isPsetV0(o.Transaction) && !isPsetV2(o.Transaction) && !isHex(o.Transaction) {
		return fmt.Errorf("invalid swap transaction format")
	}

	if len(o.Message) <= 0 {
		return fmt.Errorf("missing swap accept message")
	}
	v1Err := proto.Unmarshal(o.Message, &tdexv1.SwapAccept{})
	v2Err := proto.Unmarshal(o.Message, &tdexv2.SwapAccept{})
	if isPsetV0(o.Transaction) && v1Err != nil {
		return fmt.Errorf("invalid swap accept message")
	}
	if isPsetV2(o.Transaction) && v2Err != nil {
		return fmt.Errorf("invalid swap accept message")
	}
	if isHex(o.Transaction) && (v1Err != nil || v2Err != nil) {
		return fmt.Errorf("invalid swap accept message")
	}

	return nil
}

func (o CompleteOpts) forV1() bool {
	if isPsetV0(o.Transaction) {
		return true
	}
	if isHex(o.Transaction) {
		err := proto.Unmarshal(o.Message, &tdexv1.SwapAccept{})
		return err == nil
	}
	return false
}

func (o CompleteOpts) forV2() bool {
	if isPsetV2(o.Transaction) {
		return true
	}
	if isHex(o.Transaction) {
		err := proto.Unmarshal(o.Message, &tdexv1.SwapAccept{})
		return err == nil
	}
	return false
}

// Complete takes a CompleteOpts and returns the id of the SwapComplete entity
// and its serialized version
func Complete(opts CompleteOpts) (string, []byte, error) {
	if err := opts.validate(); err != nil {
		return "", nil, err
	}

	var message protoreflect.ProtoMessage
	randomID := randstr.Hex(8)
	switch {
	case opts.forV1():
		var msgAccept tdexv1.SwapAccept
		proto.Unmarshal(opts.Message, &msgAccept)

		if ptx, _ := pset.NewPsetFromBase64(opts.Transaction); ptx != nil {
			ok, err := ptx.ValidateAllSignatures()
			if err != nil {
				return "", nil, err
			}
			if !ok {
				return "", nil, fmt.Errorf("transaction contains invalid signatures")
			}
		} // TODO: validate sigs of raw tx input

		message = &tdexv1.SwapComplete{
			Id:          randomID,
			AcceptId:    msgAccept.GetId(),
			Transaction: opts.Transaction,
		}
	case opts.forV2():
		fallthrough
	default:
		var msgAccept tdexv2.SwapAccept
		proto.Unmarshal(opts.Message, &msgAccept)

		if ptx, _ := psetv2.NewPsetFromBase64(opts.Transaction); ptx != nil {
			ok, err := ptx.ValidateAllSignatures()
			if err != nil {
				return "", nil, err
			}
			if !ok {
				return "", nil, fmt.Errorf("transaction contains invalid signatures")
			}
		} // TODO: validate sigs of raw tx input

		message = &tdexv2.SwapComplete{
			Id:          randomID,
			AcceptId:    msgAccept.GetId(),
			Transaction: opts.Transaction,
		}
	}

	msgCompleteSerialized, err := proto.Marshal(message)
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
	if err != nil {
		return err
	}
	if !outputPFound {
		return fmt.Errorf("either SwapRequest.amount_p or SwapRequest.asset_p do not match the provided pset")
	}
	return nil
}

func isHex(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil
}
