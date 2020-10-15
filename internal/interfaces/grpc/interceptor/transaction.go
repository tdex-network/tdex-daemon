package interceptor

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
	"google.golang.org/grpc"
)

func unaryTransactionHandler(db ports.DbManager) grpc.
	UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		tx := db.NewTransaction()
		defer tx.Discard()

		dbContext := context.WithValue(ctx, "tx", tx)
		res, err := handler(dbContext, req)

		if err := tx.Commit(); err != nil {
			log.Error(err)
		}
		return res, err
	}
}

func streamTransactionHandler(db *dbbadger.DbManager) grpc.
	StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		////TODO implement transaction handler for stream calls
		//tx := db.NewTransaction()
		//defer tx.Discard()
		//
		//dbStreamContext := context.WithValue(stream.Context(), "tx", tx)
		//fromContext := grpc.ServerTransportStreamFromContext(dbStreamContext)
		return handler(srv, stream)
	}
}
