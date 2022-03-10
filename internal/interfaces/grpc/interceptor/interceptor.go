package interceptor

import (
	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/tdex-network/tdex-daemon/pkg/macaroons"
	"google.golang.org/grpc"
)

// UnaryInterceptor returns the unary interceptor
func UnaryInterceptor(svc *macaroons.Service) grpc.ServerOption {
	return grpc.UnaryInterceptor(
		middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(),
			unaryMacaroonAuthHandler(svc),
			unaryLogger,
		),
	)
}

// StreamInterceptor returns the stream interceptor with a logrus log
func StreamInterceptor(svc *macaroons.Service) grpc.ServerOption {
	return grpc.StreamInterceptor(
		middleware.ChainStreamServer(
			grpc_recovery.StreamServerInterceptor(),
			streamMacaroonAuthHandler(svc),
			streamLogger,
		),
	)
}
