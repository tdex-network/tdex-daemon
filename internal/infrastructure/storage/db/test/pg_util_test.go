package db_test

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	postgresdb "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/pg"

	_ "github.com/jackc/pgx/v4/stdlib"
)

var (
	repoManager ports.RepoManager
	ctx         = context.Background()
)

func SetupPgDb() error {
	svc, err := postgresdb.NewService(postgresdb.DbConfig{
		DataSourceURL:      "postgresql://root:secret@127.0.0.1:5432/tdexd-test",
		MigrationSourceURL: "file://../pg/migration",
	})
	if err != nil {
		return err
	}

	repoManager = svc

	return SetupDB()
}

func TearDownPgDb() error {
	if err := TruncateDB(); err != nil {
		return err
	}

	repoManager.Close()
	return nil
}

var (
	DB     *sql.DB
	dbUser = "root"
	dbPass = "secret"
	dbHost = "127.0.0.1"
	dbPort = "5432"
	dbName = "tdexd-test"
)

func SetupDB() error {
	db, err := createDBConnection()
	if err != nil {
		return err
	}

	DB = db
	return nil
}

func ShutdownDB() error {
	if err := TruncateDB(); err != nil {
		return err
	}

	return DB.Close()
}

func TruncateDB() error {
	truncateQuery := `
          SELECT truncate_tables('%s')
 `
	formattedQuery := fmt.Sprintf(truncateQuery, dbUser)
	_, err := DB.ExecContext(context.Background(), formattedQuery)
	if err != nil {
		return err
	}

	return nil
}

func createDBConnection() (*sql.DB, error) {
	formattedURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	db, err := sql.Open("pgx", formattedURL)
	if err != nil {
		return nil, err
	}

	DB = db

	truncateFunctionQuery := `
	  CREATE OR REPLACE FUNCTION truncate_tables(username IN VARCHAR) RETURNS void AS $$
	  DECLARE
	      statements CURSOR FOR
		  SELECT tablename FROM pg_tables
		  WHERE tableowner = username AND schemaname = 'public' AND tablename NOT LIKE '%migrations';
	  BEGIN
	      FOR stmt IN statements LOOP
		  EXECUTE 'TRUNCATE TABLE ' || quote_ident(stmt.tablename) || ' CASCADE;';
	      END LOOP;
	  END;
	  $$ LANGUAGE plpgsql;
`

	_, err = db.ExecContext(context.Background(), truncateFunctionQuery)
	if err != nil {
		return nil, err
	}

	return db, nil
}
