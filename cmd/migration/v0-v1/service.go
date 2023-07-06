package v0migration

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	log "github.com/sirupsen/logrus"
	serviceinterface "github.com/tdex-network/tdex-daemon/cmd/migration/service"
	"github.com/tdex-network/tdex-daemon/cmd/migration/v0-v1/mapper"
	v0webhook "github.com/tdex-network/tdex-daemon/cmd/migration/v0-v1/v0-webhook"
	v1domain "github.com/tdex-network/tdex-daemon/cmd/migration/v0-v1/v1-domain"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/tdex-network/tdex-daemon/internal/infrastructure/pubsub"
	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
	v0domain "github.com/tdex-network/tdex-daemon/old-v0"
	boltsecurestore "github.com/tdex-network/tdex-daemon/pkg/securestore/bolt"
	"github.com/vulpemventures/go-elements/network"
	"go.uber.org/ratelimit"
)

const (
	dbDir                = "db"
	tlsDir               = "tls"
	macaroonsDbFile      = "macaroons.db"
	macaroonsDir         = "macaroons"
	tdexDatadirName      = "tdex-daemon"
	oceanDatadirName     = "oceand"
	tdexTempDatadirName  = ".tdex-daemon-tmp"
	oceanTempDatadirName = ".oceand-tmp"
)

var (
	defaultTdexDatadir  = btcutil.AppDataDir(tdexDatadirName, false)
	defaultOceanDatadir = btcutil.AppDataDir(oceanDatadirName, false)
	explorerByNetwork   = map[string]string{
		network.Liquid.Name:  "https://blockstream.info/liquid/api",
		network.Testnet.Name: "https://blockstream.info/liquidtestnet/api",
		network.Regtest.Name: "http://localhost:3001/",
	}
)

type service struct{}

func NewService() serviceinterface.Service {
	return &service{}
}

func (s *service) Migrate() error {
	passwordFlag := flag.String("password", "", "the password to unlock the daemon")
	v0DatadirFlag := flag.String("v0-datadir", defaultTdexDatadir, "the datadir of the daemon to be migrated from old version to new one")
	v1DatadirFlag := flag.String("v1-datadir", defaultTdexDatadir, "the datadir to be created for the v1 daemon")
	oceanDatadirFlag := flag.String("ocean-datadir", defaultOceanDatadir, "the datadir to be created for the ocean wallet")
	noBackupFlag := flag.Bool("no-backup", false, "do not backup the v0 datadir as compressed archive .tar.gz. Be sure to do it manually or the folder will be forever lost")

	flag.Parse()

	if *passwordFlag == "" {
		return fmt.Errorf("missing password")
	}
	password := *passwordFlag

	v0Datadir := cleanAndExpandPath(*v0DatadirFlag)
	oceanDatadir := cleanAndExpandPath(*oceanDatadirFlag)
	v1Datadir := cleanAndExpandPath(*v1DatadirFlag)
	v1DatadirTemp := filepath.Join(filepath.Dir(v1Datadir), tdexTempDatadirName)
	oceanDatadirTemp := filepath.Join(filepath.Dir(v1Datadir), oceanTempDatadirName)
	defer func() {
		os.RemoveAll(v1DatadirTemp)
		os.RemoveAll(oceanDatadirTemp)
	}()

	// v0 datadir must be there.
	if _, err := os.Stat(v0Datadir); os.IsNotExist(err) {
		return fmt.Errorf("v0 datadir not found: %s", v0Datadir)
	}
	// v1 datadirs must not be there if it's not the same one.
	if v1Datadir != v0Datadir {
		if _, err := os.Stat(v1Datadir); err == nil {
			return fmt.Errorf(
				"v1 datadir already existing, either delete it or change the output " +
					"path with --v1-datadir flag",
			)
		}
	}
	if _, err := os.Stat(oceanDatadir); err == nil {
		return fmt.Errorf(
			"ocean datadir already existing, either delete it or change the output " +
				"path with --ocean-datadir flag",
		)
	}

	if err := migrate(
		v0Datadir, v1DatadirTemp, oceanDatadirTemp, password,
	); err != nil {
		return err
	}

	if !*noBackupFlag {
		v0DatadirDest := filepath.Join(v1DatadirTemp, "tdex-daemon-v0")
		if err := os.Rename(v0Datadir, v0DatadirDest); err != nil {
			return fmt.Errorf("failed to move vo datadir into v1: %s", err)
		}
		if err := archiveAndCompress(v0DatadirDest); err != nil {
			return fmt.Errorf("failed to created compressed archive of v0 datadir: %s", err)
		}
		os.RemoveAll(v0DatadirDest)
	}

	// Let's delete the datadir if it has to be overwritten.
	if v0Datadir == v1Datadir {
		os.RemoveAll(v0Datadir)
	}

	list, _ := os.ReadDir(v1DatadirTemp)
	for _, v := range list {
		source := filepath.Join(v1DatadirTemp, v.Name())
		dest := filepath.Join(v1Datadir, v.Name())
		// if err := os.Rename(source, dest); err != nil {
		// 	if strings.Contains(err.Error(), "cross-device link") {
		if err := copyDir(source, dest); err != nil {
			return err
		}
		// }
		// return err
		// }
	}
	if err := copyDir(oceanDatadirTemp, oceanDatadir); err != nil {
		return err
	}

	return nil
}

func migrate(
	v0Datadir, v1Datadir, oceanDatadir, password string,
) error {
	if err := migrateDomain(
		v0Datadir, v1Datadir, oceanDatadir, password,
	); err != nil {
		return fmt.Errorf("error while migrating domains: %s", err)
	}

	if err := migrateWebhooks(v0Datadir, v1Datadir, password); err != nil {
		return fmt.Errorf("error while migrating webhooks: %s", err)
	}

	if err := migrateTls(v0Datadir, v1Datadir); err != nil {
		return fmt.Errorf("error while migrating tls: %s", err)
	}

	if err := migrateMacaroons(v0Datadir, v1Datadir); err != nil {
		return fmt.Errorf("error while migrating macaroons: %s", err)
	}

	return nil
}

func migrateTls(fromDir, toDir string) error {
	destDir := filepath.Join(toDir, tlsDir)
	sourceDir := filepath.Join(fromDir, tlsDir)
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		log.Info("tls dir not found, skip migrating")
		return nil
	}

	log.Info("migrating tls...")
	start := time.Now()

	if err := copyDir(sourceDir, destDir); err != nil {
		return err
	}

	elapsedTime := time.Since(start).Seconds()
	log.Infof("done in %fs", elapsedTime)
	return nil
}

func migrateMacaroons(fromDir, toDir string) error {
	sourceDB := filepath.Join(fromDir, dbDir, macaroonsDbFile)
	sourceDir := filepath.Join(fromDir, macaroonsDir)
	destDir := filepath.Join(toDir, macaroonsDir)
	destDBDir := filepath.Join(toDir, dbDir)
	destDB := filepath.Join(destDBDir, macaroonsDbFile)

	if _, err := os.Stat(sourceDB); os.IsNotExist(err) {
		log.Info("macaroon dir not found, skip migrating")
		return nil
	}
	if _, err := os.Stat(destDBDir); os.IsNotExist(err) {
		errDir := os.MkdirAll(destDBDir, fs.ModeDir|0755)
		if errDir != nil {
			return fmt.Errorf("failed to create directory: %s, error: %w", destDBDir, errDir)
		}
	}

	log.Info("migrating macaroons...")
	start := time.Now()

	if err := copyFile(sourceDB, destDB); err != nil {
		return err
	}

	if err := copyDir(sourceDir, destDir); err != nil {
		return err
	}

	elapsedTime := time.Since(start).Seconds()
	log.Infof("done in %fs", elapsedTime)
	return nil
}

func migrateWebhooks(fromDir, toDir, password string) error {
	log.Info("migrating webhooks...")
	start := time.Now()

	v0WebhookRepoManager, err := v0webhook.NewWebhookRepository(
		filepath.Join(fromDir, dbDir),
	)
	if err != nil {
		return err
	}

	secureStore, err := boltsecurestore.NewSecureStorage(filepath.Join(toDir, dbDir), "pubsub.db")
	if err != nil {
		return err
	}
	svc, err := pubsub.NewService(secureStore)
	if err != nil {
		return err
	}
	pubsubSvc := application.NewPubSubService(svc)

	if err := v0WebhookRepoManager.Unlock(password); err != nil {
		return err
	}
	if err := pubsubSvc.SecurePubSub().Store().Init(password); err != nil {
		return err
	}
	if err := pubsubSvc.SecurePubSub().Store().Unlock(password); err != nil {
		return err
	}

	v0Webhooks, err := v0WebhookRepoManager.GetAllWebhooks()
	if err != nil {
		return err
	}

	mapperSvc := mapper.NewService(nil)
	v1Webhooks, err := mapperSvc.FromV091WebhooksToV1Subscriptions(v0Webhooks)
	if err != nil {
		return err
	}

	ctx := context.Background()
	for _, v := range v1Webhooks {
		if _, err := pubsubSvc.AddWebhookWithID(ctx, v); err != nil {
			return err
		}
	}

	elapsedTime := time.Since(start).Seconds()
	log.Infof("done in %fs", elapsedTime)
	return nil
}

func migrateDomain(v0Datadir, v1Datadir, oceanDatadir, password string) error {
	log.Info("migrating db...")
	start := time.Now()

	v0RepoManager, err := v0domain.NewTdexRepoManager(
		filepath.Join(v0Datadir, dbDir), nil,
	)
	if err != nil {
		return err
	}

	vault, err := v0RepoManager.GetVaultRepository().GetVault()
	if err != nil {
		return err
	}

	if !vault.IsValidPassword(password) {
		return fmt.Errorf("invalid password")
	}

	oceanDbDir := filepath.Join(path.Join(oceanDatadir, vault.Network.Name), dbDir)
	oceanRepoManager, err := v1domain.NewOceanRepoManager(oceanDbDir)
	if err != nil {
		return err
	}

	v1DbDir := filepath.Join(v1Datadir, dbDir)
	v1RepoManager, err := dbbadger.NewRepoManager(v1DbDir, nil)
	if err != nil {
		return err
	}

	mapperSvc := mapper.NewService(v0RepoManager)

	log.Infof("--> migrating wallet...")
	net, err := migrateV091VaultToOceanWallet(
		v0RepoManager, oceanRepoManager, mapperSvc, password,
	)
	if err != nil {
		return err
	}
	log.Infof("--> done")

	log.Infof("--> migrating utxo set...")
	if err := migrateV091UtxosToV1Utxos(
		v0RepoManager, oceanRepoManager, mapperSvc, net,
	); err != nil {
		return err
	}
	log.Infof("--> done")

	log.Info("--> migrating markets...")
	if err := migrateV091MarketsToV1Markets(
		v0RepoManager, v1RepoManager, mapperSvc,
	); err != nil {
		return err
	}
	log.Infof("--> done")

	log.Info("--> migrating trades...")
	trades, err := migrateV091TradesToV1Trades(
		v0RepoManager, v1RepoManager, mapperSvc,
	)
	if err != nil {
		return err
	}
	log.Infof("--> done")

	log.Info("--> migrating deposits...")
	deposits, err := migrateV091DepositsToV1Deposits(
		v0RepoManager, v1RepoManager, mapperSvc,
	)
	if err != nil {
		return err
	}
	log.Infof("--> done")

	log.Info("--> migrating withdrawals...")
	withdrawals, err := migrateV091WithdrawalsToV1Withdrawals(
		v0RepoManager, v1RepoManager, mapperSvc,
	)
	if err != nil {
		return err
	}
	log.Infof("--> done")

	log.Info("--> migrating transactions...")
	if err := migrateTransactions(
		trades, deposits, withdrawals, mapperSvc, oceanRepoManager,
	); err != nil {
		return err
	}
	log.Infof("--> done")

	elapsedTime := time.Since(start).Seconds()
	log.Infof("done in %fs", elapsedTime)
	return nil
}

func migrateV091VaultToOceanWallet(
	v0RepoManager v0domain.TdexRepoManager,
	v1RepoManager v1domain.OceanRepoManager,
	mapperSvc mapper.Service,
	vaultPass string,
) (string, error) {
	v0Vault, err := v0RepoManager.GetVaultRepository().GetVault()
	if err != nil {
		return "", err
	}

	wallet, err := mapperSvc.FromV091VaultToV1Wallet(*v0Vault, vaultPass)
	if err != nil {
		return "", err
	}

	if err := v1RepoManager.WalletRepository().InsertWallet(wallet); err != nil {
		return "", err
	}
	return v0Vault.Network.Name, nil
}

func migrateV091TradesToV1Trades(
	v0RepoManager v0domain.TdexRepoManager,
	v1RepoManager ports.RepoManager,
	mapperSvc mapper.Service,
) ([]*domain.Trade, error) {
	v0Trades, err := v0RepoManager.GetTradeRepository().GetAllTrades()
	if err != nil {
		return nil, err
	}

	v1Trades, err := mapperSvc.FromV091TradesToV1Trades(v0Trades)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	for _, trade := range v1Trades {
		if err := v1RepoManager.TradeRepository().AddTrade(ctx, trade); err != nil {
			return nil, err
		}
	}
	return v1Trades, nil
}

func migrateV091DepositsToV1Deposits(
	v0RepoManager v0domain.TdexRepoManager,
	v1RepoManager ports.RepoManager,
	mapperSvc mapper.Service,
) ([]*domain.Deposit, error) {
	v0Deposits, err := v0RepoManager.GetDepositRepository().GetAllDeposits()
	if err != nil {
		return nil, err
	}

	v1Deposits, err := mapperSvc.FromV091DepositsToV1Deposits(v0Deposits)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	if _, err := v1RepoManager.DepositRepository().AddDeposits(
		ctx, v1Deposits,
	); err != nil {
		return nil, err
	}

	deposits := make([]*domain.Deposit, 0, len(v1Deposits))
	for i := range v1Deposits {
		d := v1Deposits[i]
		deposits = append(deposits, &d)
	}
	return deposits, nil
}

func migrateV091WithdrawalsToV1Withdrawals(
	v0RepoManager v0domain.TdexRepoManager,
	v1RepoManager ports.RepoManager,
	mapperSvc mapper.Service,
) ([]*domain.Withdrawal, error) {
	v0Withdrawals, err := v0RepoManager.GetWithdrawalRepository().GetAllWithdrawals()
	if err != nil {
		return nil, err
	}

	v1Withdrawals, err := mapperSvc.FromV091WithdrawalsToV1Withdrawals(v0Withdrawals)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	if _, err := v1RepoManager.WithdrawalRepository().AddWithdrawals(
		ctx, v1Withdrawals,
	); err != nil {
		return nil, err
	}

	withdrawals := make([]*domain.Withdrawal, 0, len(v1Withdrawals))
	for i := range v1Withdrawals {
		w := v1Withdrawals[i]
		withdrawals = append(withdrawals, &w)
	}
	return withdrawals, nil
}

func migrateV091MarketsToV1Markets(
	v0RepoManager v0domain.TdexRepoManager,
	v1RepoManager ports.RepoManager,
	mapperSvc mapper.Service,
) error {
	v0Markets, err := v0RepoManager.GetMarketRepository().GetAllMarkets()
	if err != nil {
		return err
	}

	v1Markets, err := mapperSvc.FromV091MarketsToV1Markets(v0Markets)
	if err != nil {
		return err
	}

	ctx := context.Background()
	for _, m := range v1Markets {
		if err := v1RepoManager.MarketRepository().AddMarket(ctx, m); err != nil {
			return err
		}
	}
	return nil
}

func migrateV091UtxosToV1Utxos(
	v0RepoManager v0domain.TdexRepoManager,
	v1RepoManager v1domain.OceanRepoManager,
	mapperSvc mapper.Service,
	net string,
) error {
	v0Utxos, err := v0RepoManager.GetUnspentRepository().GetAllUnspents()
	if err != nil {
		return err
	}

	v1Utxos, err := mapperSvc.FromV091UnspentsToV1Utxos(v0Utxos)
	if err != nil {
		return err
	}

	confirmedUtxos := make(map[string][]*v1domain.Utxo)
	spentUtxos := make(map[string]map[int]*v1domain.Utxo)
	indexedUtxos := make(map[string][]int)
	txids := make(map[string]struct{})
	empty := v1domain.UtxoStatus{}
	for i := range v1Utxos {
		u := v1Utxos[i]
		if u.SpentStatus != empty {
			if len(indexedUtxos[u.TxID]) <= 0 {
				indexedUtxos[u.TxID] = make([]int, 0)
			}
			indexedUtxos[u.TxID] = append(indexedUtxos[u.TxID], int(u.VOut))
			if spentUtxos[u.TxID] == nil {
				spentUtxos[u.TxID] = make(map[int]*v1domain.Utxo)
			}
			spentUtxos[u.TxID][int(u.VOut)] = u
		}
		if u.ConfirmedStatus != empty {
			txids[u.TxID] = struct{}{}
			if len(confirmedUtxos[u.TxID]) <= 0 {
				confirmedUtxos[u.TxID] = make([]*v1domain.Utxo, 0)
			}
			confirmedUtxos[u.TxID] = append(confirmedUtxos[u.TxID], u)
		}
	}

	limiter := ratelimit.New(300)
	confirmedStatuses := make(map[string]v1domain.UtxoStatus)
	for txid := range txids {
		limiter.Take()
		status, err := getConfirmationStatus(net, txid)
		if err != nil {
			return err
		}
		confirmedStatuses[txid] = *status
	}

	for txid, status := range confirmedStatuses {
		for i := range confirmedUtxos[txid] {
			confirmedUtxos[txid][i].ConfirmedStatus = status
		}
	}

	limiter = ratelimit.New(5)
	spentStatuses := make(map[string][]v1domain.UtxoStatus)
	for txid := range indexedUtxos {
		limiter.Take()
		statuses, err := getSpentStatus(net, txid)
		if err != nil {
			return err
		}
		spentStatuses[txid] = statuses
	}

	for txid, vouts := range indexedUtxos {
		for _, vout := range vouts {
			spentUtxos[txid][vout].SpentStatus = spentStatuses[txid][vout]
		}
	}

	return v1RepoManager.UtxoRepository().AddUtxos(v1Utxos)
}

func migrateTransactions(
	trades []*domain.Trade,
	deposits []*domain.Deposit, withdrawals []*domain.Withdrawal,
	mapperSvc mapper.Service, v1RepoManager v1domain.OceanRepoManager,
) error {
	wallet, err := v1RepoManager.WalletRepository().GetWallet()
	if err != nil {
		return err
	}

	indexedTxs := mapperSvc.FromV091TransactionsToV1Transactions(
		trades, deposits, withdrawals, wallet.AccountsByLabel,
	)

	txs := make([]*v1domain.Transaction, 0, len(indexedTxs))
	limiter := ratelimit.New(300)
	for txid, tx := range indexedTxs {
		limiter.Take()
		status, err := getConfirmationStatus(wallet.NetworkName, txid)
		if err != nil {
			return err
		}
		tx.Confirm(status.BlockHash, status.BlockHeight)
		if len(tx.TxHex) <= 0 {
			limiter.Take()
			txHex, err := getTxHex(wallet.NetworkName, txid)
			if err != nil {
				return err
			}
			tx.TxHex = txHex
		}
		txs = append(txs, tx)
	}

	return v1RepoManager.TransactionRepository().AddTransactions(txs)
}

func getConfirmationStatus(
	net, txid string,
) (*v1domain.UtxoStatus, error) {
	resp, err := http.Get(fmt.Sprintf("%s/tx/%s/status", explorerByNetwork[net], txid))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}

	m := make(map[string]interface{})
	json.Unmarshal(body, &m)
	return &v1domain.UtxoStatus{
		BlockHash:   m["block_hash"].(string),
		BlockHeight: uint64(m["block_height"].(float64)),
		BlockTime:   int64(m["block_time"].(float64)),
	}, nil
}

func getSpentStatus(
	net, txid string,
) ([]v1domain.UtxoStatus, error) {
	resp, err := http.Get(fmt.Sprintf("%s/tx/%s/outspends", explorerByNetwork[net], txid))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}

	m := make([]interface{}, 0)
	json.Unmarshal(body, &m)
	statuses := make([]v1domain.UtxoStatus, 0)
	for i := range m {
		mm := m[i].(map[string]interface{})
		s := v1domain.UtxoStatus{}
		if mm["spent"].(bool) {
			mmm := mm["status"].(map[string]interface{})
			s = v1domain.UtxoStatus{
				Txid:        mm["txid"].(string),
				BlockHash:   mmm["block_hash"].(string),
				BlockHeight: uint64(mmm["block_height"].(float64)),
				BlockTime:   int64(mmm["block_time"].(float64)),
			}
		}
		statuses = append(statuses, s)
	}
	return statuses, nil
}

func getTxHex(net, txid string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/tx/%s/hex", explorerByNetwork[net], txid))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(string(body))
	}

	return string(body), nil
}

func archiveAndCompress(dir string) error {
	start := time.Now()
	log.Info("making compressed archive out of the v0 datadir...")

	tar := func(source, target string) error {
		filename := filepath.Base(source)
		target = filepath.Join(target, fmt.Sprintf("%s.tar", filename))
		tarfile, err := os.Create(target)
		if err != nil {
			return err
		}
		defer tarfile.Close()

		tarball := tar.NewWriter(tarfile)
		defer tarball.Close()

		info, err := os.Stat(source)
		if err != nil {
			return nil
		}

		var baseDir string
		if info.IsDir() {
			baseDir = filepath.Base(source)
		}

		return filepath.Walk(
			source, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				header, err := tar.FileInfoHeader(info, info.Name())
				if err != nil {
					return err
				}

				if baseDir != "" {
					header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
				}

				if err := tarball.WriteHeader(header); err != nil {
					return err
				}

				if info.IsDir() {
					return nil
				}

				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()
				_, err = io.Copy(tarball, file)
				return err
			},
		)
	}

	gzip := func(source, target string) error {
		reader, err := os.Open(source)
		if err != nil {
			return err
		}

		filename := filepath.Base(source)
		target = filepath.Join(target, fmt.Sprintf("%s.gz", filename))
		writer, err := os.Create(target)
		if err != nil {
			return err
		}
		defer writer.Close()

		archiver := gzip.NewWriter(writer)
		archiver.Name = filename
		defer archiver.Close()

		_, err = io.Copy(archiver, reader)
		return err
	}

	baseDir := filepath.Dir(dir)
	if err := tar(dir, baseDir); err != nil {
		return err
	}
	if err := gzip(fmt.Sprintf("%s.tar", dir), baseDir); err != nil {
		return err
	}
	if err := os.RemoveAll(fmt.Sprintf("%s.tar", dir)); err != nil {
		return err
	}

	elapsedTime := time.Since(start).Seconds()
	log.Infof("done in %fs", elapsedTime)
	return nil
}

// cleanAndExpandPath expands environment variables and leading ~ in the
// passed path, cleans the result, and returns it.
// This function is taken from https://github.com/btcsuite/btcd
func cleanAndExpandPath(path string) string {
	if path == "" {
		return ""
	}

	// Expand initial ~ to OS specific home directory.
	if strings.HasPrefix(path, "~") {
		var homeDir string
		u, err := user.Current()
		if err == nil {
			homeDir = u.HomeDir
		} else {
			homeDir = os.Getenv("HOME")
		}

		path = strings.Replace(path, "~", homeDir, 1)
	}

	// NOTE: The os.ExpandEnv doesn't work with Windows-style %VARIABLE%,
	// but the variables can still be expanded via POSIX-style $VARIABLE.
	return filepath.Clean(os.ExpandEnv(path))
}

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
// Symlinks are ignored and skipped.
func copyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if err == nil {
		return fmt.Errorf("destination already exists")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Type()&os.ModeSymlink != 0 {
				continue
			}

			err = copyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}
