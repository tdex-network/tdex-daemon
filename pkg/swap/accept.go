package swap

import (
	"fmt"

	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v1"
	"github.com/thanhpk/randstr"
	"google.golang.org/protobuf/proto"
)

// AcceptOpts is the struct given to Accept method
type AcceptOpts struct {
	Message            []byte
	PsetBase64         string
	InputBlindingKeys  map[string][]byte
	OutputBlindingKeys map[string][]byte
	UnblindedInputs    []UnblindedInput
}

func (o AcceptOpts) validate() error {
	if isPsetV0(o.PsetBase64) {
		return checkTxAndBlindKeys(
			o.PsetBase64,
			o.InputBlindingKeys,
			o.OutputBlindingKeys,
		)
	}
	if isPsetV2(o.PsetBase64) {
		return checkTxAndUnblindedIns(o.PsetBase64, o.UnblindedInputs)
	}
	return fmt.Errorf("invalid transaction format")
}

func (o AcceptOpts) unblindedIns() []*tdexv1.UnblindedInput {
	if len(o.UnblindedInputs) <= 0 {
		return nil
	}
	list := make([]*tdexv1.UnblindedInput, 0, len(o.UnblindedInputs))
	for _, in := range o.UnblindedInputs {
		list = append(list, &tdexv1.UnblindedInput{
			Index:         in.Index,
			Asset:         in.Asset,
			Amount:        in.Amount,
			AssetBlinder:  in.AssetBlinder,
			AmountBlinder: in.AmountBlinder,
		})
	}
	return list
}

// Accept takes a AcceptOpts and returns the id of the SwapAccept entity and
// its serialized version
func Accept(opts AcceptOpts) (string, []byte, error) {
	if err := opts.validate(); err != nil {
		return "", nil, err
	}

	var msgRequest tdexv1.SwapRequest
	err := proto.Unmarshal(opts.Message, &msgRequest)
	if err != nil {
		return "", nil, fmt.Errorf("unmarshal swap request %w", err)
	}

	randomID := randstr.Hex(8)
	msgAccept := &tdexv1.SwapAccept{
		Id:                randomID,
		RequestId:         msgRequest.GetId(),
		Transaction:       opts.PsetBase64,
		InputBlindingKey:  opts.InputBlindingKeys,
		OutputBlindingKey: opts.OutputBlindingKeys,
		UnblindedInputs:   opts.unblindedIns(),
	}

	msgAcceptSerialized, err := proto.Marshal(msgAccept)
	if err != nil {
		return "", nil, err
	}
	return randomID, msgAcceptSerialized, nil
}
