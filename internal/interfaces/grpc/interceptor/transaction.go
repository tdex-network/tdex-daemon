package interceptor

import (
	"context"
	log "github.com/sirupsen/logrus"
	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
	"google.golang.org/grpc"
)

func unaryTransactionHandler(db *dbbadger.DbManager) grpc.
	UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		tx := db.Store.Badger().NewTransaction(true)
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
		//TODO implement transaction handler for stream calls
		return handler(srv, stream)
	}
}
