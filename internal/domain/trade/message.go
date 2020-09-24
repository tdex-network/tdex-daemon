package trade

import (
	pb "github.com/tdex-network/tdex-protobuf/generated/go/swap"
	"google.golang.org/protobuf/proto"
)

type swapMessage struct {
	request  []byte
	accept   []byte
	complete []byte
}

func (m swapMessage) Request() *pb.SwapRequest {
	s := &pb.SwapRequest{}
	proto.Unmarshal(m.request, s)
	return s
}

func (m swapMessage) Accept() *pb.SwapAccept {
	s := &pb.SwapAccept{}
	proto.Unmarshal(m.accept, s)
	return s
}

func (m swapMessage) Complete() *pb.SwapComplete {
	s := &pb.SwapComplete{}
	proto.Unmarshal(m.complete, s)
	return s
}
