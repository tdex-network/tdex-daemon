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

func (h handshakeHandler) Info(
	ctx context.Context,
	request *pb.InfoRequest,
) (*pb.InfoReply, error) {
	panic("implement me")
}

func (h handshakeHandler) Connect(
	ctx context.Context,
	init *pb.Init,
) (*pb.Ack, error) {
	panic("implement me")
}

func (h handshakeHandler) UnarySecret(
	ctx context.Context,
	message *pb.SecretMessage,
) (*pb.SecretMessage, error) {
	panic("implement me")
}

func (h handshakeHandler) StreamSecret(
	message *pb.SecretMessage,
	server pb.Handshake_StreamSecretServer,
) error {
	panic("implement me")
}
