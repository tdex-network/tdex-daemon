package interceptor

import (
	"context"
	"fmt"
	"github.com/tdex-network/tdex-daemon/pkg/macaroons"
	"google.golang.org/grpc"
	"gopkg.in/macaroon-bakery.v2/bakery"
)

func unaryAuthHandler(
	permissionMap map[string][]bakery.Op,
	macaroonSvc macaroons.Service,
) grpc.UnaryServerInterceptor {

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		uriPermissions, ok := permissionMap[info.FullMethod]
		if !ok {
			return nil, fmt.Errorf("%s: unknown permissions "+
				"required for method", info.FullMethod)
		}

		if uriPermissions[0].Entity == "operator" {
			validator := macaroonSvc.GetValidator(info.FullMethod)

			// Now that we know what validator to use, let it do its work.
			err := validator.ValidateMacaroon(
				ctx, uriPermissions, info.FullMethod,
			)
			if err != nil {
				return nil, err
			}
		}

		return handler(ctx, req)
	}
}

func streamAuthHandler(
	permissionMap map[string][]bakery.Op,
	macaroonSvc macaroons.Service,
) grpc.StreamServerInterceptor {

	return func(srv interface{}, ss grpc.ServerStream,
		info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

		uriPermissions, ok := permissionMap[info.FullMethod]
		if !ok {
			return fmt.Errorf("%s: unknown permissions required "+
				"for method", info.FullMethod)
		}

		if uriPermissions[0].Entity == "operator" {
			validator := macaroonSvc.GetValidator(info.FullMethod)

			// Now that we know what validator to use, let it do its work.
			err := validator.ValidateMacaroon(
				ss.Context(), uriPermissions, info.FullMethod,
			)
			if err != nil {
				return err
			}
		}

		return handler(srv, ss)
	}
}
