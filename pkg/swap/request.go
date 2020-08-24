package swap

import (
	pb "github.com/tdex-network/tdex-protobuf/generated/go/swap"
	"github.com/thanhpk/randstr"
	"google.golang.org/protobuf/proto"
)

// RequestOpts is the struct to be given to the Request method
type RequestOpts struct {
	AssetToBeSent   string
	AmountToBeSent  uint64
	AssetToReceive  string
	AmountToReceive uint64
	PsetBase64      string
}

// Request takes a RequestOpts struct and returns a serialized protobuf message.
func (c *Core) Request(opts RequestOpts) ([]byte, error) {
	randomID := randstr.Hex(8)
	msg := &pb.SwapRequest{
		Id: randomID,
		// Proposer
		AssetP:  opts.AssetToBeSent,
		AmountP: opts.AmountToBeSent,
		// Receiver
		AssetR:  opts.AssetToReceive,
		AmountR: opts.AmountToReceive,
		//PSET
		Transaction: opts.PsetBase64,
	}

	if err := c.compareMessagesAndTransaction(msg, nil); err != nil {
		return nil, err
	}

	return proto.Marshal(msg)
}
