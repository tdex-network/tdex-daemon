package interceptor

import (
	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
	"google.golang.org/grpc"
)

// UnaryInterceptor returns the unary interceptor
func UnaryInterceptor(dbManager *dbbadger.DbManager) grpc.ServerOption {
	return grpc.UnaryInterceptor(
		middleware.ChainUnaryServer(
			unaryLogger,
		),
	)
}

// StreamInterceptor returns the stream interceptor with a logrus log
func StreamInterceptor(dbManager *dbbadger.DbManager) grpc.ServerOption {
	return grpc.StreamInterceptor(
		middleware.ChainStreamServer(
			streamLogger,
		),
	)
}
