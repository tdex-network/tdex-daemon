package swap

import (
	pb "github.com/tdex-network/tdex-protobuf/generated/go/swap"
	"github.com/thanhpk/randstr"
	"google.golang.org/protobuf/proto"
)

// RequestOpts is the struct to be given to the Request method
type RequestOpts struct {
	Id                 string
	AssetToSend        string
	AmountToSend       uint64
	AssetToReceive     string
	AmountToReceive    uint64
	PsetBase64         string
	InputBlindingKeys  map[string][]byte
	OutputBlindingKeys map[string][]byte
}

// Request takes a RequestOpts struct and returns a serialized protobuf message.
func Request(opts RequestOpts) ([]byte, error) {
	id := opts.Id
	if len(id) <= 0 {
		id = randstr.Hex(8)
	}
	msg := &pb.SwapRequest{
		Id: id,
		// Proposer
		AssetP:  opts.AssetToSend,
		AmountP: opts.AmountToSend,
		// Receiver
		AssetR:  opts.AssetToReceive,
		AmountR: opts.AmountToReceive,
		//PSET
		Transaction: opts.PsetBase64,
		// Blinding keys
		InputBlindingKey:  opts.InputBlindingKeys,
		OutputBlindingKey: opts.InputBlindingKeys,
	}

	return proto.Marshal(msg)
}
