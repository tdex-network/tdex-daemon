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
		tx := db.NewTransaction()
		defer tx.Discard()

		streamContextWithTx := context.WithValue(stream.Context(), "tx", tx)
		newStream := WrapServerStream(stream, streamContextWithTx)

		err := handler(srv, newStream)

		if err := tx.Commit(); err != nil {
			log.Error(err)
		}

		return err
	}
}
