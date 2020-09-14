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

// ParseSwapRequest checks whether the given swap request is well formed and
// returns its byte serialization
func ParseSwapRequest(request *pb.SwapRequest) ([]byte, error) {
	if err := compareMessagesAndTransaction(request, nil); err != nil {
		return nil, err
	}
	return proto.Marshal(request)
}

// Request takes a RequestOpts struct and returns a serialized protobuf message.
func Request(opts RequestOpts) ([]byte, error) {
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

	return ParseSwapRequest(msg)
}
