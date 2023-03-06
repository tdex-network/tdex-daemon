package swap

import (
	"fmt"

	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v1"
	tdexv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v2"
	"github.com/thanhpk/randstr"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// AcceptOpts is the struct given to Accept method
type AcceptOpts struct {
	Message            []byte
	Transaction        string
	InputBlindingKeys  map[string][]byte
	OutputBlindingKeys map[string][]byte
	UnblindedInputs    []UnblindedInput
}

func (o AcceptOpts) validate() error {
	if isPsetV0(o.Transaction) {
		return validateSwapTxV0(
			o.Transaction, o.InputBlindingKeys, o.OutputBlindingKeys,
		)
	}
	if isPsetV2(o.Transaction) {
		return validateSwapAcceptTx(o.Transaction, o.UnblindedInputs)
	}
	return fmt.Errorf("invalid swap transaction format")
}

func (o AcceptOpts) unblindedIns() []*tdexv2.UnblindedInput {
	if len(o.UnblindedInputs) <= 0 {
		return nil
	}
	list := make([]*tdexv2.UnblindedInput, 0, len(o.UnblindedInputs))
	for _, in := range o.UnblindedInputs {
		list = append(list, &tdexv2.UnblindedInput{
			Index:         in.Index,
			Asset:         in.Asset,
			Amount:        in.Amount,
			AssetBlinder:  in.AssetBlinder,
			AmountBlinder: in.AmountBlinder,
		})
	}
	return list
}

func (o AcceptOpts) forV1() bool {
	return isPsetV0(o.Transaction)
}

func (o AcceptOpts) forV2() bool {
	return isPsetV2(o.Transaction)
}

// Accept takes a AcceptOpts and returns the id of the SwapAccept entity and
// its serialized version
func Accept(opts AcceptOpts) (string, []byte, error) {
	if err := opts.validate(); err != nil {
		return "", nil, err
	}

	var message protoreflect.ProtoMessage
	randomID := randstr.Hex(8)
	switch {
	case opts.forV1():
		var msgRequest tdexv1.SwapRequest
		if err := proto.Unmarshal(opts.Message, &msgRequest); err != nil {
			return "", nil, fmt.Errorf("unmarshal swap request %w", err)
		}

		message = &tdexv1.SwapAccept{
			Id:                randomID,
			RequestId:         msgRequest.GetId(),
			Transaction:       opts.Transaction,
			InputBlindingKey:  opts.InputBlindingKeys,
			OutputBlindingKey: opts.OutputBlindingKeys,
		}
	case opts.forV2():
		fallthrough
	default:
		var msgRequest tdexv2.SwapRequest
		if err := proto.Unmarshal(opts.Message, &msgRequest); err != nil {
			return "", nil, fmt.Errorf("unmarshal swap request %w", err)
		}

		message = &tdexv2.SwapAccept{
			Id:              randomID,
			RequestId:       msgRequest.GetId(),
			Transaction:     opts.Transaction,
			UnblindedInputs: opts.unblindedIns(),
		}
	}

	msgAcceptSerialized, err := proto.Marshal(message)
	if err != nil {
		return "", nil, err
	}
	return randomID, msgAcceptSerialized, nil
}
