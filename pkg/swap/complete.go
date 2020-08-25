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

//Complete takes a CompleteOpts and returns a serialized SwapComplete message
func (*Swap) Complete(complete CompleteOpts) ([]byte, error) {
	var msgAccept pb.SwapAccept
	err := proto.Unmarshal(complete.Message, &msgAccept)
	if err != nil {
		return nil, fmt.Errorf("unmarshal swap accept %w", err)
	}

	_, err = pset.NewPsetFromBase64(complete.PsetBase64)
	if err != nil {
		return nil, err
	}

	//TODO check if signatures of the inputs are valid

	randomID := randstr.Hex(8)
	msgComplete := &pb.SwapComplete{
		Id:          randomID,
		AcceptId:    msgAccept.GetId(),
		Transaction: complete.PsetBase64,
	}

	return proto.Marshal(msgComplete)
}
