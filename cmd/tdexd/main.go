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
	"github.com/tdex-network/tdex-daemon/pkg/circuitbreaker"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	"github.com/tdex-network/tdex-daemon/pkg/explorer/esplora"
	boltsecurestore "github.com/tdex-network/tdex-daemon/pkg/securestore/bolt"
	"github.com/tdex-network/tdex-daemon/pkg/stats"

	_ "net/http/pprof" // #nosec
)

var (
	// General config
	logLevel                = config.GetInt(config.LogLevelKey)
	profilerEnabled         = config.GetBool(config.EnableProfilerKey)
	datadir                 = config.GetDatadir()
	dbDir                   = filepath.Join(datadir, config.DbLocation)
	profilerDir             = filepath.Join(datadir, config.ProfilerLocation)
	noMacaroons             = config.GetBool(config.NoMacaroonsKey)
	statsIntervalInSeconds  = config.GetDuration(config.StatsIntervalKey) * time.Second
	tradeTLSKey             = config.GetString(config.TradeTLSKeyKey)
	tradeTLSCert            = config.GetString(config.TradeTLSCertKey)
	operatorTLSExtraIPs     = config.GetStringSlice(config.OperatorExtraIPKey)
	operatorTLSExtraDomains = config.GetStringSlice(config.OperatorExtraDomainKey)
	// App services config
	marketsFee                    = int64(config.GetFloat(config.DefaultFeeKey) * 100)
	marketsBaseAsset              = config.GetString(config.BaseAssetKey)
	marketsQuoteAsset             = config.GetString(config.QuoteAssetKey)
	tradesExpiryDurationInSeconds = config.GetDuration(config.TradeExpiryTimeKey) * time.Second
	tradesSatsPerByte             = config.GetFloat(config.TradeSatsPerByte)
	pricesSlippagePercentage      = decimal.NewFromFloat(config.GetFloat(config.PriceSlippageKey))
	feeThreshold                  = uint64(config.GetInt(config.FeeAccountBalanceThresholdKey))
	tradeSvcPort                  = config.GetInt(config.TradeListeningPortKey)
	operatorSvcPort               = config.GetInt(config.OperatorListeningPortKey)
	crawlerIntervalInMilliseconds = time.Duration(config.GetInt(config.CrawlIntervalKey)) * time.Millisecond
	explorerTimoutRequest         = config.GetDuration(config.ExplorerRequestTimeoutKey)
	cbMaxFailingRequest           = config.GetInt(config.CBMaxFailingRequestsKey)
	cbFailingRatio                = config.GetFloat(config.CBFailingRatioKey)
	rescanRangeStart              = config.GetInt(config.RescanRangeStartKey)
	rescanGapLimit                = config.GetInt(config.RescanGapLimitKey)
	walletUnlockPasswordFile      = config.GetString(config.WalletUnlockPasswordFile)
	noOperatorTls                 = config.GetBool(config.NoOperatorTlsKey)
	protocol                      = config.GetString(config.ConnectUrlProto)

	version = "dev"
	commit  = "none"
	date    = "unknown"
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

	// Set default params for circuitbreaker pkg
	circuitbreaker.MaxNumOfFailingRequests = cbMaxFailingRequest
	circuitbreaker.FailingRatio = cbFailingRatio

	// Init services to be used by those of the application layer.
	repoManager, err := dbbadger.NewRepoManager(dbDir, log.New())
	if err != nil {
		log.Errorf("error while opening db: %s", err)
		return
	}

	explorerSvc, err := config.GetExplorer()
	if err != nil {
		repoManager.Close()

		log.Errorf("error while setting up explorer service: %s", err)
		return
	}
	crawlerSvc := crawler.NewService(crawler.Opts{
		ExplorerSvc:     explorerSvc,
		ErrorHandler:    func(err error) { log.Warn(err) },
		CrawlerInterval: crawlerIntervalInMilliseconds,
	})
	webhookPubSub, err := newWebhookPubSubService(dbDir, explorerTimoutRequest)
	if err != nil {
		crawlerSvc.Stop()
		repoManager.Close()

		log.Errorf("error while setting up webhook pubsub service: %s", err)
		return
	}

	network, err := config.GetNetwork()
	if err != nil {
		crawlerSvc.Stop()
		repoManager.Close()

		log.Errorf("error while setting up network: %s", err)
		return
	}

	blockchainListener := application.NewBlockchainListener(
		crawlerSvc,
		repoManager,
		webhookPubSub,
		network,
	)

	// Init application services
	skipCheckTrades := true
	tradeSvc := application.NewTradeService(
		repoManager,
		explorerSvc,
		blockchainListener,
		tradesExpiryDurationInSeconds,
		tradesSatsPerByte,
		pricesSlippagePercentage,
		network,
		feeThreshold,
		!skipCheckTrades,
	)
	operatorSvc := application.NewOperatorService(
		repoManager,
		explorerSvc,
		blockchainListener,
		marketsBaseAsset,
		marketsQuoteAsset,
		marketsFee,
		network,
		feeThreshold,
		application.BuildInfo{
			Version: version,
			Commit:  commit,
			Date:    date,
		},
	)
	walletSvc := application.NewWalletService(
		repoManager,
		explorerSvc,
		blockchainListener,
		network,
		marketsFee,
	)
	walletUnlockerSvc := application.NewWalletUnlockerService(
		repoManager,
		explorerSvc,
		blockchainListener,
		network,
		marketsFee,
		marketsBaseAsset,
		marketsQuoteAsset,
		rescanRangeStart,
		rescanGapLimit,
	)

	runOnOnePort := operatorSvcPort == tradeSvcPort
	svc, err := NewGrpcService(
		runOnOnePort,
		walletUnlockerSvc,
		walletSvc,
		operatorSvc,
		tradeSvc,
		repoManager,
	)
	if err != nil {
		crawlerSvc.Stop()
		repoManager.Close()

		log.Errorf("error while setting up gRPC service: %s", err)
		return
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
	if err := svc.Start(); err != nil {
		log.Errorf("error while starting daemon: %s", err)
		return
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	log.Info("shutting down daemon")
}

func stop(
	repoManager ports.RepoManager,
	pubsubSvc ports.SecurePubSub,
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
	time.Sleep(crawlerIntervalInMilliseconds)

	repoManager.Close()
	log.Debug("closed connection with database")

	log.Info("disabled all active interfaces. Exiting")
}

func newWebhookPubSubService(
	datadir string, reqTimeout time.Duration,
) (ports.SecurePubSub, error) {
	secureStore, err := boltsecurestore.NewSecureStorage(datadir, "pubsub.db")
	if err != nil {
		return nil, err
	}
	httpClient := esplora.NewHTTPClient(time.Duration(reqTimeout) * time.Second)
	return webhookpubsub.NewWebhookPubSubService(secureStore, httpClient)
}

func NewGrpcService(
	runOnOnePort bool,
	walletUnlockerSvc application.WalletUnlockerService,
	walletSvc application.WalletService,
	operatorSvc application.OperatorService,
	tradeSvc application.TradeService,
	repoManager ports.RepoManager,
) (interfaces.Service, error) {
	if runOnOnePort {
		opts := grpcinterface.ServiceOptsOnePort{
			NoMacaroons:              noMacaroons,
			Datadir:                  datadir,
			DBLocation:               config.DbLocation,
			MacaroonsLocation:        config.MacaroonsLocation,
			WalletUnlockPasswordFile: walletUnlockPasswordFile,
			Address:                  fmt.Sprintf(":%d", operatorSvcPort),
			WalletUnlockerSvc:        walletUnlockerSvc,
			WalletSvc:                walletSvc,
			OperatorSvc:              operatorSvc,
			TradeSvc:                 tradeSvc,
			RepoManager:              repoManager,
			TLSLocation:              config.TLSLocation,
			NoTls:                    noOperatorTls,
			ExtraIPs:                 operatorTLSExtraIPs,
			ExtraDomains:             operatorTLSExtraDomains,
			Host:                     config.GetHost(),
			Protocol:                 protocol,
		}

		return grpcinterface.NewServiceOnePort(opts)
	}

	opts := grpcinterface.ServiceOpts{
		NoMacaroons:              noMacaroons,
		Datadir:                  datadir,
		DBLocation:               config.DbLocation,
		TLSLocation:              config.TLSLocation,
		MacaroonsLocation:        config.MacaroonsLocation,
		OperatorExtraIPs:         operatorTLSExtraIPs,
		OperatorExtraDomains:     operatorTLSExtraDomains,
		OperatorAddress:          fmt.Sprintf(":%d", operatorSvcPort),
		TradeAddress:             fmt.Sprintf(":%d", tradeSvcPort),
		TradeTLSKey:              tradeTLSKey,
		TradeTLSCert:             tradeTLSCert,
		WalletSvc:                walletSvc,
		WalletUnlockerSvc:        walletUnlockerSvc,
		OperatorSvc:              operatorSvc,
		TradeSvc:                 tradeSvc,
		WalletUnlockPasswordFile: walletUnlockPasswordFile,
		RepoManager:              repoManager,
		NoOperatorTls:            noOperatorTls,
		Host:                     config.GetHost(),
		Protocol:                 protocol,
	}

	return grpcinterface.NewService(opts)
}
