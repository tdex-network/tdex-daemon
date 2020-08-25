package swap

import (
	"fmt"

	pb "github.com/tdex-network/tdex-protobuf/generated/go/swap"
	"github.com/thanhpk/randstr"
	"google.golang.org/protobuf/proto"
)

//AcceptOpts is the struct given to Accept method
type AcceptOpts struct {
	message    []byte
	psetBase64 string
}

//Accept takes a AcceptOpts and returns a serialized SwapAccept message
func (c *Swap) Accept(accept AcceptOpts) ([]byte, error) {
	var msgRequest pb.SwapRequest
	err := proto.Unmarshal(accept.message, &msgRequest)
	if err != nil {
		return nil, fmt.Errorf("unmarshal swap request %w", err)
	}

	randomID := randstr.Hex(8)
	msgAccept := &pb.SwapAccept{
		Id:          randomID,
		RequestId:   msgRequest.GetId(),
		Transaction: accept.psetBase64,
	}

	if err := c.compareMessagesAndTransaction(&msgRequest, msgAccept); err != nil {
		return nil, err
	}

	return proto.Marshal(msgAccept)
}
