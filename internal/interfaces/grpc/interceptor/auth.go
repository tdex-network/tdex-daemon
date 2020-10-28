package interceptor

import (
	"context"
	"fmt"
	"github.com/tdex-network/tdex-daemon/pkg/macaroons"
	"google.golang.org/grpc"
	"gopkg.in/macaroon-bakery.v2/bakery"
)

var entities = []string{"operator"}

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

		err := auth(ctx, permissionMap, info.FullMethod, macaroonSvc)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func streamAuthHandler(
	permissionMap map[string][]bakery.Op,
	macaroonSvc macaroons.Service,
) grpc.StreamServerInterceptor {

	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {

		err := auth(ss.Context(), permissionMap, info.FullMethod, macaroonSvc)
		if err != nil {
			return err
		}

		return handler(srv, ss)
	}
}

func auth(
	ctx context.Context,
	permissionMap map[string][]bakery.Op,
	rpcMethod string,
	macaroonSvc macaroons.Service,
) error {
	uriPermissions, ok := permissionMap[rpcMethod]
	if !ok {
		return fmt.Errorf("%s: unknown permissions "+
			"required for method", rpcMethod)
	}

	if isForAuth(uriPermissions[0].Entity) {
		validator := macaroonSvc.GetValidator(rpcMethod)

		// Now that we know what validator to use, let it do its work.
		err := validator.ValidateMacaroon(
			ctx, uriPermissions, rpcMethod,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func isForAuth(entity string) bool {
	for _, v := range entities {
		if entity == v {
			return true
		}
	}
	return false
}
