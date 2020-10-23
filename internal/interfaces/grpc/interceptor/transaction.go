package interceptor

import (
	"context"
	"errors"

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
	) (reply interface{}, err error) {
		tx := db.NewTransaction()
		defer func() {
			if err != nil {
				tx.Discard()
			}
			if rec := recover(); rec != nil {
				log.Error(rec)
				reply = nil
				err = errors.New("not able to serve request")
			}
		}()

		dbContext := context.WithValue(ctx, "tx", tx)
		res, err := handler(dbContext, req)
		if err != nil {
			return
		}

		if err = tx.Commit(); err != nil {
			log.Error(err)
			return
		}

		reply = res

		return
	}
}

func streamTransactionHandler(db *dbbadger.DbManager) grpc.
	StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) (err error) {
		tx := db.NewTransaction()
		defer func() {
			if err != nil {
				tx.Discard()
			}
			if rec := recover(); rec != nil {
				log.Error(rec)
				err = errors.New("not able to serve request")
			}
		}()

		streamContextWithTx := context.WithValue(stream.Context(), "tx", tx)
		newStream := WrapServerStream(stream, streamContextWithTx)

		if err = handler(srv, newStream); err != nil {
			return
		}

		if err = tx.Commit(); err != nil {
			log.Error(err)
		}

		return
	}
}
