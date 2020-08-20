package tradeservice

import (
	"context"

	pbhandshake "github.com/tdex-network/tdex-protobuf/generated/go/handshake"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/trade"
)

//UnarySecret is the domain controller for the UnarySecret RPC
func (s *Server) UnarySecret(ctx context.Context, req *pbhandshake.SecretMessage) (res *pbhandshake.SecretMessage, err error) {
	return &pbhandshake.SecretMessage{}, nil
}

// StreamSecret is the domain controller for the StreamSecret RPC
func (s *Server) StreamSecret(req *pbhandshake.SecretMessage, stream pb.Trade_StreamSecretServer) error {
	if err := stream.Send(&pbhandshake.SecretMessage{}); err != nil {
		return err
	}
	return nil
}
