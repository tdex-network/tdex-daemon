package postgresdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/pg/sqlc/queries"

	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	postgresDriver             = "pgx"
	insecureDataSourceTemplate = "postgresql://%s:%s@%s:%d/%s?sslmode=disable"

	uniqueViolation = "23505"
	pgxNoRows       = "no rows in result set"
)

type repoManager struct {
	pgxPool *pgxpool.Pool
	querier *queries.Queries

	marketRepository     domain.MarketRepository
	tradeRepository      domain.TradeRepository
	depositRepository    domain.DepositRepository
	withdrawalRepository domain.WithdrawalRepository
}

func NewService(dbConfig DbConfig) (ports.RepoManager, error) {
	dataSource := insecureDataSourceStr(dbConfig)

	pgxPool, err := connect(dataSource)
	if err != nil {
		return nil, err
	}

	if err = migrateDb(dataSource, dbConfig.MigrationSourceURL); err != nil {
		return nil, err
	}

	rm := &repoManager{
		pgxPool: pgxPool,
		querier: queries.New(pgxPool),
	}

	marketRepository := NewMarketRepositoryImpl(rm.querier, rm.execTx)
	tradeRepository := NewTradeRepositoryImpl(rm.querier, rm.execTx)
	depositRepository := NewDepositRepositoryImpl(rm.querier, rm.execTx)
	withdrawalRepository := NewWithdrawalRepositoryImpl(rm.querier, rm.execTx)

	rm.marketRepository = marketRepository
	rm.tradeRepository = tradeRepository
	rm.depositRepository = depositRepository
	rm.withdrawalRepository = withdrawalRepository

	return rm, nil
}

func (r *repoManager) MarketRepository() domain.MarketRepository {
	return r.marketRepository
}

func (r *repoManager) TradeRepository() domain.TradeRepository {
	return r.tradeRepository
}

func (r *repoManager) DepositRepository() domain.DepositRepository {
	return r.depositRepository
}

func (r *repoManager) WithdrawalRepository() domain.WithdrawalRepository {
	return r.withdrawalRepository
}

func (r *repoManager) Close() {
	r.pgxPool.Close()
}

func (r *repoManager) execTx(
	ctx context.Context,
	txBody func(*queries.Queries) error,
) error {
	conn, err := r.pgxPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}

	// Rollback is safe to call even if the tx is already closed, so if
	// the tx commits successfully, this is a no-op.
	defer func() {
		err := tx.Rollback(ctx)
		switch {
		// If the tx was already closed (it was successfully executed)
		// we do not need to log that error.
		case errors.Is(err, pgx.ErrTxClosed):
			return

		// If this is an unexpected error, log it.
		case err != nil:
			log.Errorf("unable to rollback db tx: %v", err)
		}
	}()

	if err := txBody(r.querier.WithTx(tx)); err != nil {
		return err
	}

	// Commit transaction.
	return tx.Commit(ctx)
}

type DbConfig struct {
	DbUser             string
	DbPassword         string
	DbHost             string
	DbPort             int
	DbName             string
	MigrationSourceURL string
}

func connect(dataSource string) (*pgxpool.Pool, error) {
	return pgxpool.Connect(context.Background(), dataSource)
}

func migrateDb(dataSource, migrationSourceUrl string) error {
	pg := postgres.Postgres{}

	d, err := pg.Open(dataSource)
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationSourceUrl,
		postgresDriver,
		d,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func insecureDataSourceStr(dbConfig DbConfig) string {
	return fmt.Sprintf(
		insecureDataSourceTemplate,
		dbConfig.DbUser,
		dbConfig.DbPassword,
		dbConfig.DbHost,
		dbConfig.DbPort,
		dbConfig.DbName,
	)
}
