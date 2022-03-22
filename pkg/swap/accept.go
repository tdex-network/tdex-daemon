package swap

import (
	"fmt"

	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/go/tdex/v1"
	"github.com/thanhpk/randstr"
	"google.golang.org/protobuf/proto"
)

// AcceptOpts is the struct given to Accept method
type AcceptOpts struct {
	Message            []byte
	PsetBase64         string
	InputBlindingKeys  map[string][]byte
	OutputBlindingKeys map[string][]byte
}

func (o AcceptOpts) validate() error {
	return checkTxAndBlindKeys(
		o.PsetBase64,
		o.InputBlindingKeys,
		o.OutputBlindingKeys,
	)
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
	}

	msgAcceptSerialized, err := proto.Marshal(msgAccept)
	if err != nil {
		return "", nil, err
	}
	return randomID, msgAcceptSerialized, nil
}
