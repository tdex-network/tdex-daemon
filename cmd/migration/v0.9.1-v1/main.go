package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	v1subscription "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v1-subscription"

	v091webhook "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v091-webhook"

	"github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/mapper"
	v091domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v091-domain"

	v1domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0.9.1-v1/v1-domain"

	log "github.com/sirupsen/logrus"

	"github.com/btcsuite/btcd/btcutil"
)

const (
	dbDir                  = "db"
	tlsDir                 = "tls"
	macaroonsDbFile        = "macaroons.db"
	macaroonsPermissionDir = "macaroons"
	tdexdDataDirName       = "tdex-daemon"
	oceanDataDirName       = "oceand"
	legacyTdexdDataDirName = "tdex-daemon-v0"
	tempTdexdV1DataDirName = "tdex-daemon-v1-tmp"
)

var (
	defaultTdexdDataDir = btcutil.AppDataDir(tdexdDataDirName, false)
	defaultOceanDataDir = btcutil.AppDataDir(oceanDataDirName, false)
	tempTdexdV1DataDir  = btcutil.AppDataDir(tempTdexdV1DataDirName, false)
	legacyTdexdDataDir  = btcutil.AppDataDir(legacyTdexdDataDirName, false)
)

func main() {
	v091DataDirFlag := flag.String("v091DataDir", "", "The v0.9.1 data directory that will be migrated to v1")
	v1OceanDataDirFlag := flag.String("v1OceanDataDir", "", "The v1 ocean data directory")
	v1TdexdDataDirFlag := flag.String("v1TdexdDataDir", "", "The v1 tdexd data directory")
	v091VaultPasswordFlag := flag.String("v091VaultPassword", "", "The v0.9.1 vault password")
	esploraUrlFlag := flag.String("esploraUrl", "", "The esplora url")

	flag.Parse()

	if *v091VaultPasswordFlag == "" {
		log.Fatal(errors.New("missing required v091VaultPassword flag"))
	}
	v091VaultPassword := *v091VaultPasswordFlag

	v091DataDir := defaultTdexdDataDir
	if *v091DataDirFlag != "" {
		v091DataDir = *v091DataDirFlag
	}

	v1OceanDataDir := defaultOceanDataDir
	if *v1OceanDataDirFlag != "" {
		v1OceanDataDir = *v1OceanDataDirFlag
	}

	v1TdexdDataDir := tempTdexdV1DataDir
	renameV1DataDir := true
	if *v1TdexdDataDirFlag != "" {
		renameV1DataDir = false
		v1TdexdDataDir = *v1TdexdDataDirFlag
	}

	esploraUrl := "https://blockstream.info/liquid/api"
	if *esploraUrlFlag != "" {
		esploraUrl = *esploraUrlFlag
	}

	if _, err := os.Stat(v091DataDir); os.IsNotExist(err) {
		log.Fatalf("v0.9.1 data directory does not exist: %s", v091DataDir)
	}

	if v091DataDir == v1TdexdDataDir {
		log.Fatal("v0.9.1 data directory and v1 data directory cannot be the same")
	}

	if err := migrate(
		v091DataDir, v1OceanDataDir, v1TdexdDataDir, v091VaultPassword, esploraUrl,
	); err != nil {
		log.Fatal(err)
	}

	err := os.Rename(defaultTdexdDataDir, legacyTdexdDataDir)
	if err != nil {
		log.Fatal(err)
	}

	if renameV1DataDir {
		err = os.Rename(v1TdexdDataDir, defaultTdexdDataDir)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Info("migration completed")
}

func migrate(
	v091DataDir, v1OceanDataDir, v1TdexdDataDir, v091VaultPassword, esploraUrl string,
) error {
	log.Info("tls migration started")
	if err := migrateTls(v091DataDir, v1TdexdDataDir); err != nil {
		log.Errorf("error while migrating tls: %s", err)
	}
	log.Info("tls migration completed")

	log.Info("macaroons migration started")
	if err := migrateMacaroons(v091DataDir, v1TdexdDataDir); err != nil {
		log.Errorf("error while migrating macaroons: %s", err)
	}
	log.Info("macaroons migration completed")

	log.Info("webhook migration started")
	if err := migrateWebhooks(
		v091DataDir, v1TdexdDataDir, v091VaultPassword, esploraUrl,
	); err != nil {
		log.Errorf("error while migrating webhooks: %s", err)
	}
	log.Info("webhook migration completed")

	log.Info("core domain migration started")
	if err := migrateDomain(
		v091DataDir, v1OceanDataDir, v1TdexdDataDir, v091VaultPassword, esploraUrl,
	); err != nil {
		return err
	}
	log.Info("core domain migration completed")

	return nil
}

func migrateTls(fromDir, toDir string) error {
	tlsLoc := filepath.Join(fromDir, tlsDir)
	if _, err := os.Stat(tlsLoc); os.IsNotExist(err) {
		return nil
	}

	destDir := filepath.Join(toDir, tlsDir)
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		errDir := os.MkdirAll(destDir, 0666)
		if errDir != nil {
			return fmt.Errorf("failed to create directory: %s, error: %w", destDir, errDir)
		}
	}

	files := make([]string, 0)
	if err := filepath.Walk(
		tlsLoc, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				files = append(files, path)
			}
			return nil
		}); err != nil {
		return err
	}

	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("source file does not exist: %s", file)
		}

		if err := copyFile(
			file, filepath.Join(toDir, tlsDir, filepath.Base(file)),
		); err != nil {
			return err
		}
	}

	return nil
}

func migrateMacaroons(fromDir, toDir string) error {
	macaroonDB := filepath.Join(fromDir, dbDir, macaroonsDbFile)
	macaroonPermissions := filepath.Join(fromDir, macaroonsPermissionDir)

	if _, err := os.Stat(macaroonDB); os.IsNotExist(err) {
		return nil
	}

	destDir := filepath.Join(toDir, dbDir)
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		errDir := os.MkdirAll(destDir, 0666)
		if errDir != nil {
			return fmt.Errorf("failed to create directory: %s, error: %w", destDir, errDir)
		}
	}

	destDir = filepath.Join(toDir, macaroonsPermissionDir)
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		errDir := os.MkdirAll(destDir, 0666)
		if errDir != nil {
			return fmt.Errorf("failed to create directory: %s, error: %w", destDir, errDir)
		}
	}

	if err := copyFile(
		macaroonDB, filepath.Join(toDir, dbDir, macaroonsDbFile),
	); err != nil {
		return err
	}

	files := make([]string, 0)
	if err := filepath.Walk(
		macaroonPermissions, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				files = append(files, path)
			}
			return nil
		}); err != nil {
		return err
	}

	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("source file does not exist: %s", file)
		}

		if err := copyFile(
			file, filepath.Join(toDir, macaroonsPermissionDir, filepath.Base(file)),
		); err != nil {
			return err
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Close()
}

func migrateWebhooks(fromDir, toDir, vaultPass, esploraUrl string) error {
	v091WebhookRepoManager, err := v091webhook.NewWebhookRepository(
		filepath.Join(fromDir, dbDir),
	)
	if err != nil {
		return err
	}

	if err := v091WebhookRepoManager.Unlock(vaultPass); err != nil {
		return err
	}

	v091Webhooks, err := v091WebhookRepoManager.GetAllWebhooks()
	if err != nil {
		return err
	}

	mapperSvc := mapper.NewService(nil, esploraUrl)
	v1Webhooks, err := mapperSvc.FromV091WebhooksToV1Subscriptions(v091Webhooks)
	if err != nil {
		return err
	}

	v1WebhookRepoManager, err := v1subscription.NewSubscriptionRepository(
		filepath.Join(toDir, dbDir),
	)
	if err != nil {
		return err
	}

	return v1WebhookRepoManager.InsertSubscriptions(v1Webhooks)
}

func migrateDomain(fromDir, oceanToDir, tdexdToDir, vaultPass, esploraUrl string) error {
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

	oceanToDbDir := filepath.Join(path.Join(oceanToDir, vault.Network.Name), dbDir)
	tdexdToDbDir := filepath.Join(tdexdToDir, dbDir)
	v1RepoManager, err := v1domain.NewRepositoryImpl(oceanToDbDir, tdexdToDbDir, nil)
	if err != nil {
		return err
	}

	mapperSvc := mapper.NewService(v091RepoManager, esploraUrl)

	log.Info("vault to wallet migration started")
	if err := migrateV091VaultToOceanWallet(
		v091RepoManager, v1RepoManager, mapperSvc, vaultPass,
	); err != nil {
		return err
	}
	log.Info("vault to wallet migration completed")

	log.Info("utxo migration started")
	if err := migrateV091UtxosToV1Utxos(
		v091RepoManager, v1RepoManager, mapperSvc,
	); err != nil {
		return err
	}

	log.Info("markets migration started")
	if err := migrateV091MarketsToV1Markets(
		v091RepoManager, v1RepoManager, mapperSvc,
	); err != nil {
		return err
	}

	log.Info("trades migration started")
	if err := migrateV091TradesToV1Trades(
		v091RepoManager, v1RepoManager, mapperSvc,
	); err != nil {
		return err
	}
	log.Info("trades migration completed")

	log.Info("deposits migration started")
	if err := migrateV091DepositsToV1Deposits(
		v091RepoManager, v1RepoManager, mapperSvc,
	); err != nil {
		return err
	}
	log.Info("deposits migration completed")

	log.Info("withdrawals migration started")
	if err := migrateV091WithdrawalsToV1Withdrawals(
		v091RepoManager, v1RepoManager, mapperSvc,
	); err != nil {
		return err
	}
	log.Info("withdrawals migration completed")

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

func migrateV091TradesToV1Trades(
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

func migrateV091DepositsToV1Deposits(
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

func migrateV091WithdrawalsToV1Withdrawals(
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

func migrateV091MarketsToV1Markets(
	v091RepoManager v091domain.Repository,
	v1RepoManager v1domain.Repository,
	mapperSvc mapper.Service,
) error {
	v091Markets, err := v091RepoManager.GetMarketRepository().GetAllMarkets()
	if err != nil {
		return err
	}

	v1Markets, err := mapperSvc.FromV091MarketsToV1Markets(v091Markets)
	if err != nil {
		return err
	}

	return v1RepoManager.GetMarketRepository().InsertMarkets(v1Markets)
}

func migrateV091UtxosToV1Utxos(
	v091RepoManager v091domain.Repository,
	v1RepoManager v1domain.Repository,
	mapperSvc mapper.Service,
) error {
	v091Utxos, err := v091RepoManager.GetUnspentRepository().GetAllUnspents()
	if err != nil {
		return err
	}

	v1Utxos, err := mapperSvc.FromV091UnspentsToV1Utxos(v091Utxos)
	if err != nil {
		return err
	}

	return v1RepoManager.GetUtxoRepository().InsertUtxos(v1Utxos)
}
