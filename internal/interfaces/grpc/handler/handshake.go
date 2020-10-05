package grpchandler

import (
	"context"
	pb "github.com/tdex-network/tdex-protobuf/generated/go/handshake"
)

type handshakeHandler struct {
	pb.UnimplementedHandshakeServer
}

func NewHandshakeHandler() pb.HandshakeServer {
	return &handshakeHandler{}
}

func (h handshakeHandler) Connect(
	ctx context.Context,
	init *pb.Init,
) (*pb.Ack, error) {
	return &pb.Ack{}, nil
}

func (h handshakeHandler) UnarySecret(
	ctx context.Context,
	message *pb.SecretMessage,
) (*pb.SecretMessage, error) {
	return &pb.SecretMessage{}, nil
}

func (h handshakeHandler) StreamSecret(
	message *pb.SecretMessage,
	stream pb.Handshake_StreamSecretServer,
) error {
	if err := stream.Send(&pb.SecretMessage{}); err != nil {
		return err
	}
	return nil
}
