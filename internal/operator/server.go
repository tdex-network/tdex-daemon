package operatorservice

import pb "github.com/tdex-network/tdex-protobuf/generated/go/operator"

// Server is used to implement Trader service.
type Server struct {
	pb.OperatorServer
}
