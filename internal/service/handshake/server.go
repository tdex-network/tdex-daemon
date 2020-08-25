package handshakeservice

import pb "github.com/tdex-network/tdex-protobuf/generated/go/handshake"

// Server is used to implement Trader service.
type Server struct {
	pb.HandshakeServer
}
