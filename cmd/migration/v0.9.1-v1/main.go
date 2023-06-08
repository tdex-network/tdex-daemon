package main

import (
	"context"
	"fmt"
	"os"

	v1domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v1-domain"

	log "github.com/sirupsen/logrus"

	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
)

const (
	dbDir = "db"
)

var (
	ctx = context.Background()
)

func main() {
	//v091 := flag.String("v091", "", "The v0.9.1 data directory that will be migrated to v1")
	//v1 := flag.String("v1", "", "The v1 data directory location")
	//
	//flag.Parse()
	//
	//if *v091 == "" || *v1 == "" {
	//	fmt.Println("Both 'v091' and 'v1' arguments are required")
	//	flag.PrintDefaults()
	//	return
	//}
	// v091DataDir := *v091
	// v1DataDir := *v1

	currentDir, err := os.Getwd()
	if err != nil {
		log.Error(err)
	}

	v091DataDir := fmt.Sprintf("%v/%v", currentDir, "cmd/migration/v0.9.1-v1/v091-testdatadir")
	v1DataDir := fmt.Sprintf("%v/%v", currentDir, "cmd/migration/v0.9.1-v1/v1-datadir")

	migrateTls()
	migrateMacaroons()
	migrateStats()
	migrateWebhooks()
	if err := migrateDomain(v091DataDir, v1DataDir); err != nil {
		log.Error(err)
	}
}

func migrateTls() {
	fmt.Println("tls migration not implemented")
}

func migrateMacaroons() {
	fmt.Println("macaroons migration not implemented")
}

func migrateStats() {
	fmt.Println("stats migration not implemented")
}

func migrateWebhooks() {
	fmt.Println("webhooks migration not implemented")
}

func migrateDomain(fromDir, toDir string) error {
	if err := migrateMasterVaultToOceanWallet(fromDir, toDir); err != nil {
		return err
	}

	return nil
}

func migrateMasterVaultToOceanWallet(fromDir, toDir string) error {
	dbDataDir := fmt.Sprintf("%v/%v", fromDir, dbDir)
	repoManager, err := dbbadger.NewRepoManager(dbDataDir, nil)
	if err != nil {
		return err
	}

	masterVault, err := repoManager.VaultRepository().GetOrCreateVault(
		ctx, nil, "", nil,
	)
	if err != nil {
		return err
	}

	v1WalletRepo, err := v1domain.NewWalletRepository(toDir, nil)
	if err != nil {
		return err
	}

	return v1WalletRepo.InsertWallet(
		context.Background(), v1domain.FromV091VaultToV1Wallet(*masterVault),
	)
}
