package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/mapper"
	v091domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v091-domain"

	v1domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v1-domain"

	log "github.com/sirupsen/logrus"
)

const (
	dbDir = "db"
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

	network := "regtest"
	v091DataDir := path.Join(currentDir, "cmd/migration/v0.9.1-v1/v091-datadir")
	v1OceanDataDir := path.Join(currentDir, "cmd/migration/v0.9.1-v1/v1-oceandatadir", network)
	v1TdexdDataDir := path.Join(currentDir, "cmd/migration/v0.9.1-v1/v1-tdexddatadir")
	v091VaultPassword := "ciaociao"

	migrateTls()
	migrateMacaroons()
	migrateStats()
	migrateWebhooks()
	if err := migrateDomain(
		v091DataDir, v1OceanDataDir, v1TdexdDataDir, v091VaultPassword,
	); err != nil {
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

func migrateDomain(fromDir, oceanToDir, tdexdToDir, vaultPass string) error {
	v091RepoManager, err := v091domain.NewRepositoryImpl(filepath.Join(fromDir, dbDir), nil)
	if err != nil {
		return err
	}

	vault, err := v091RepoManager.GetVaultRepository().GetVault()
	if err != nil {
		return err
	}

	if !vault.IsValidPassword(vaultPass) {
		return errors.New("invalid vault password")
	}

	oceanToDbDir := filepath.Join(oceanToDir, dbDir)
	tdexdToDbDir := filepath.Join(tdexdToDir, dbDir)
	v1RepoManager, err := v1domain.NewRepositoryImpl(oceanToDbDir, tdexdToDbDir, nil)
	if err != nil {
		return err
	}

	mapperSvc := mapper.NewService(v091RepoManager)

	if err := migrateV091VaultToOceanWallet(
		v091RepoManager, v1RepoManager, mapperSvc, vaultPass,
	); err != nil {
		return err
	}

	if err := migrateV091TradesToOceanTrades(
		v091RepoManager, v1RepoManager, mapperSvc,
	); err != nil {
		return err
	}

	if err := migrateV091DepositsToOceanDeposits(
		v091RepoManager, v1RepoManager, mapperSvc,
	); err != nil {
		return err
	}

	if err := migrateV091WithdrawalsToOceanWithdrawals(
		v091RepoManager, v1RepoManager, mapperSvc,
	); err != nil {
		return err
	}

	return nil
}

func migrateV091VaultToOceanWallet(
	v091RepoManager v091domain.Repository,
	v1RepoManager v1domain.Repository,
	mapperSvc mapper.Service,
	vaultPass string,
) error {
	v091Vault, err := v091RepoManager.GetVaultRepository().GetVault()
	if err != nil {
		return err
	}

	wallet, err := mapperSvc.FromV091VaultToV1Wallet(*v091Vault, vaultPass)
	if err != nil {
		return err
	}

	return v1RepoManager.GetWalletRepository().InsertWallet(wallet)
}

func migrateV091TradesToOceanTrades(
	v091RepoManager v091domain.Repository,
	v1RepoManager v1domain.Repository,
	mapperSvc mapper.Service,
) error {
	v091Trades, err := v091RepoManager.GetTradeRepository().GetAllTrades()
	if err != nil {
		return err
	}

	v1Trades, err := mapperSvc.FromV091TradesToV1Trades(v091Trades)
	if err != nil {
		return err
	}

	return v1RepoManager.GetTradeRepository().InsertTrades(v1Trades)
}

func migrateV091DepositsToOceanDeposits(
	v091RepoManager v091domain.Repository,
	v1RepoManager v1domain.Repository,
	mapperSvc mapper.Service,
) error {
	v091Deposits, err := v091RepoManager.GetDepositRepository().GetAllDeposits()
	if err != nil {
		return err
	}

	v1Deposits, err := mapperSvc.FromV091DepositsToV1Deposits(v091Deposits)
	if err != nil {
		return err
	}

	return v1RepoManager.GetDepositRepository().InsertDeposits(v1Deposits)
}

func migrateV091WithdrawalsToOceanWithdrawals(
	v091RepoManager v091domain.Repository,
	v1RepoManager v1domain.Repository,
	mapperSvc mapper.Service,
) error {
	v091Withdrawals, err := v091RepoManager.GetWithdrawalRepository().GetAllWithdrawals()
	if err != nil {
		return err
	}

	v1Withdrawals, err := mapperSvc.FromV091WithdrawalsToV1Withdrawals(v091Withdrawals)
	if err != nil {
		return err
	}

	return v1RepoManager.GetWithdrawalRepository().InsertWithdrawals(v1Withdrawals)
}
