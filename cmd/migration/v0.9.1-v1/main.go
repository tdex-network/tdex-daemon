package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/mapper"
	v091domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v091-domain"

	v1domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v1-domain"

	log "github.com/sirupsen/logrus"
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

	v091DataDir := fmt.Sprintf("%v/%v", currentDir, "cmd/migration/v0.9.1-v1/v091-datadir")
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
	v091RepoManager, err := v091domain.NewRepositoryImpl(dbDataDir, nil)
	if err != nil {
		return err
	}

	v091Vault, err := v091RepoManager.GetVaultRepository().GetVault(ctx)
	if err != nil {
		return err
	}

	v1RepoManager, err := v1domain.NewRepositoryImpl(toDir, nil)
	if err != nil {
		return err
	}

	return v1RepoManager.GetWalletRepository().InsertWallet(
		context.Background(), mapper.FromV091VaultToV1Wallet(*v091Vault),
	)
}
