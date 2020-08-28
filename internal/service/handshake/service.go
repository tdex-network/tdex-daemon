package handshakeservice

import pb "github.com/tdex-network/tdex-protobuf/generated/go/handshake"

// Service is used to implement Handshake service.
type Service struct {
	pb.HandshakeServer
}
