package interceptor

import (
	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/pkg/macaroons"
	"google.golang.org/grpc"
)

// UnaryInterceptor returns the unary interceptor
func UnaryInterceptor(
	macaroonSvc *macaroons.Service,
	walletUnlockerSvc application.UnlockerService,
) grpc.ServerOption {
	return grpc.UnaryInterceptor(
		middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(),
			unaryWalletLockerAuthHandler(walletUnlockerSvc),
			unaryMacaroonAuthHandler(macaroonSvc),
			unaryLogger,
		),
	)
}

// StreamInterceptor returns the stream interceptor with a logrus log
func StreamInterceptor(
	macaroonSvc *macaroons.Service,
	walletUnlockerSvc application.UnlockerService,
) grpc.ServerOption {
	return grpc.StreamInterceptor(
		middleware.ChainStreamServer(
			grpc_recovery.StreamServerInterceptor(),
			streamWalletLockerAuthHandler(walletUnlockerSvc),
			streamMacaroonAuthHandler(macaroonSvc),
			streamLogger,
		),
	)
}
