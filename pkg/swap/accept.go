package swap

import (
	"fmt"

	pb "github.com/tdex-network/tdex-protobuf/generated/go/swap"
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

// Accept takes a AcceptOpts and returns the id of the SwapAccept entity and
//its serialized version
func Accept(accept AcceptOpts) (string, []byte, error) {
	var msgRequest pb.SwapRequest
	err := proto.Unmarshal(accept.Message, &msgRequest)
	if err != nil {
		return "", nil, fmt.Errorf("unmarshal swap request %w", err)
	}

	randomID := randstr.Hex(8)
	msgAccept := &pb.SwapAccept{
		Id:                randomID,
		RequestId:         msgRequest.GetId(),
		Transaction:       accept.PsetBase64,
		InputBlindingKey:  accept.InputBlindingKeys,
		OutputBlindingKey: accept.OutputBlindingKeys,
	}

	if err := compareMessagesAndTransaction(&msgRequest, msgAccept); err != nil {
		return "", nil, err
	}

	msgAcceptSerialized, err := proto.Marshal(msgAccept)
	if err != nil {
		return "", nil, err
	}
	return randomID, msgAcceptSerialized, nil
}
