package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	webhookpubsub "github.com/tdex-network/tdex-daemon/internal/infrastructure/pubsub/webhook"
	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
	"github.com/tdex-network/tdex-daemon/internal/interfaces"
	grpcinterface "github.com/tdex-network/tdex-daemon/internal/interfaces/grpc"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	"github.com/tdex-network/tdex-daemon/pkg/explorer/esplora"
	boltsecurestore "github.com/tdex-network/tdex-daemon/pkg/securestore/bolt"
	"github.com/tdex-network/tdex-daemon/pkg/stats"

	_ "net/http/pprof"
)

var (
	// General config
	logLevel               = config.GetInt(config.LogLevelKey)
	network                = config.GetNetwork()
	profilerEnabled        = config.GetBool(config.EnableProfilerKey)
	datadir                = config.GetDatadir()
	dbDir                  = filepath.Join(datadir, config.DbLocation)
	profilerDir            = filepath.Join(datadir, config.ProfilerLocation)
	noMacaroons            = config.GetBool(config.NoMacaroonsKey)
	statsIntervalInSeconds = config.GetDuration(config.StatsIntervalKey) * time.Second
	tradeTLSKey            = config.GetString(config.TradeTLSKeyKey)
	tradeTLSCert           = config.GetString(config.TradeTLSCertKey)
	// App services config
	marketsFee                    = int64(config.GetFloat(config.DefaultFeeKey) * 100)
	marketsBaseAsset              = config.GetString(config.BaseAssetKey)
	tradesExpiryDurationInSeconds = config.GetDuration(config.TradeExpiryTimeKey) * time.Second
	pricesSlippagePercentage      = decimal.NewFromFloat(config.GetFloat(config.PriceSlippageKey))
	feeThreshold                  = uint64(config.GetInt(config.FeeAccountBalanceThresholdKey))
	tradeSvcPort                  = config.GetInt(config.TradeListeningPortKey)
	operatorSvcPort               = config.GetInt(config.OperatorListeningPortKey)
	crawlerIntervalInMilliseconds = time.Duration(config.GetInt(config.CrawlIntervalKey)) * time.Millisecond
)

func main() {
	log.SetLevel(log.Level(logLevel))

	// Profiler is enabled at url http://localhost:8024/debug/pprof/
	if profilerEnabled {
		runtime.SetBlockProfileRate(1)
		go func() {
			http.ListenAndServe(":8024", nil)
		}()
	}

	// Init services to be used by those of the application layer.
	repoManager, err := dbbadger.NewRepoManager(dbDir, log.New())
	if err != nil {
		log.WithError(err).Panic("error while opening db")
	}
	explorerSvc, err := config.GetExplorer()
	if err != nil {
		log.WithError(err).Panic("error while setting up explorer service")
	}
	crawlerSvc := crawler.NewService(crawler.Opts{
		ExplorerSvc:        explorerSvc,
		ErrorHandler:       func(err error) { log.Warn(err) },
		CrawlerInterval:    crawlerIntervalInMilliseconds,
		ExplorerLimit:      config.GetInt(config.CrawlLimitKey),
		ExplorerTokenBurst: config.GetInt(config.CrawlTokenBurst),
	})
	webhookPubSub, err := newWebhookPubSubService(
		dbDir, config.GetDuration(config.ExplorerRequestTimeoutKey),
	)
	if err != nil {
		log.WithError(err).Panic(
			"an error occured while setting up webhook pubsub service",
		)
	}
	blockchainListener := application.NewBlockchainListener(
		crawlerSvc,
		repoManager,
		webhookPubSub,
		marketsBaseAsset,
		network,
	)

	// Init application services
	tradeSvc := application.NewTradeService(
		repoManager,
		explorerSvc,
		blockchainListener,
		marketsBaseAsset,
		tradesExpiryDurationInSeconds,
		pricesSlippagePercentage,
		network,
	)
	operatorSvc := application.NewOperatorService(
		repoManager,
		explorerSvc,
		blockchainListener,
		marketsBaseAsset,
		marketsFee,
		network,
		feeThreshold,
	)
	walletSvc, err := application.NewWalletService(
		repoManager,
		explorerSvc,
		blockchainListener,
		network,
		marketsFee,
		marketsBaseAsset,
	)
	if err != nil {
		log.WithError(err).Panic("error while setting up wallet service")
	}

	// Init gRPC interfaces.
	opts := grpcinterface.ServiceOpts{
		NoMacaroons:       noMacaroons,
		Datadir:           datadir,
		DBLocation:        config.DbLocation,
		TLSLocation:       config.TLSLocation,
		MacaroonsLocation: config.MacaroonsLocation,
		WalletSvc:         walletSvc,
		OperatorSvc:       operatorSvc,
		TradeSvc:          tradeSvc,
	}
	svc, err := grpcinterface.NewService(opts)
	if err != nil {
		log.WithError(err).Panic("an error occured while setting up gRPC service")
	}

	log.Info("starting daemon")

	var cancelStats context.CancelFunc
	if log.GetLevel() >= log.DebugLevel {
		var ctx context.Context
		ctx, cancelStats = context.WithCancel(context.Background())
		stats.EnableMemoryStatistics(ctx, statsIntervalInSeconds, profilerDir)
	}

	defer stop(repoManager, webhookPubSub, blockchainListener, svc, cancelStats)

	// Start gRPC service interfaces.
	tradeAddress := fmt.Sprintf(":%+v", tradeSvcPort)
	operatorAddress := fmt.Sprintf(":%+v", operatorSvcPort)

	if err := svc.Start(operatorAddress, tradeAddress, tradeTLSKey, tradeTLSCert); err != nil {
		log.WithError(err).Panic("an error occured while starting daemon")
	}

	log.Info("trade interface is listening on " + tradeAddress)
	log.Info("operator interface is listening on " + operatorAddress)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	log.Info("shutting down daemon")
}

func stop(
	repoManager ports.RepoManager,
	pubsubSvc application.SecurePubSub,
	blockchainListener application.BlockchainListener,
	svc interfaces.Service,
	cancelStats context.CancelFunc,
) {
	if profilerEnabled && log.GetLevel() >= log.DebugLevel {
		cancelStats()
		time.Sleep(1 * time.Second)
		log.Debug("stopped profiler")
	}

	svc.Stop()

	blockchainListener.StopObservation()

	pubsubSvc.Store().Close()
	log.Debug("stopped pubsub service")

	// give the crawler the time to terminate
	time.Sleep(
		time.Duration(config.GetInt(config.CrawlIntervalKey)) * time.Millisecond,
	)

	repoManager.Close()
	log.Debug("closed connection with database")

	log.Debug("exiting")
}

func newWebhookPubSubService(
	datadir string, reqTimeout time.Duration,
) (application.SecurePubSub, error) {
	secureStore, err := boltsecurestore.NewSecureStorage(datadir, "pubsub.db")
	if err != nil {
		return nil, err
	}
	httpClient := esplora.NewHTTPClient(time.Duration(reqTimeout) * time.Second)
	return webhookpubsub.NewWebhookPubSubService(secureStore, httpClient)
}
