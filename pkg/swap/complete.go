package swap

import (
	"fmt"

	pb "github.com/tdex-network/tdex-protobuf/generated/go/swap"
	"github.com/thanhpk/randstr"
	"github.com/vulpemventures/go-elements/pset"
	"google.golang.org/protobuf/proto"
)

//CompleteOpts is the struct given to the Complete method
type CompleteOpts struct {
	Message    []byte
	PsetBase64 string
}

// Complete takes a CompleteOpts and returns the id of the SwapComplete entity
// and its serialized version
func Complete(complete CompleteOpts) (string, []byte, error) {
	var msgAccept pb.SwapAccept
	err := proto.Unmarshal(complete.Message, &msgAccept)
	if err != nil {
		return "", nil, fmt.Errorf("unmarshal swap accept %w", err)
	}

	_, err = pset.NewPsetFromBase64(complete.PsetBase64)
	if err != nil {
		return "", nil, err
	}

	//TODO check if signatures of the inputs are valid

	randomID := randstr.Hex(8)
	msgComplete := &pb.SwapComplete{
		Id:          randomID,
		AcceptId:    msgAccept.GetId(),
		Transaction: complete.PsetBase64,
	}

	msgCompleteSerialized, err := proto.Marshal(msgComplete)
	if err != nil {
		return "", nil, err
	}

	return randomID, msgCompleteSerialized, nil
}
