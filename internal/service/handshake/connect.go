package handshakeservice

import (
	"context"

	pb "github.com/tdex-network/tdex-protobuf/generated/go/handshake"
)

// Connect is the domain controller for the Connect RPC
func (s *Service) Connect(ctx context.Context, req *pb.Init) (res *pb.Ack, err error) {
	return &pb.Ack{}, nil
}
