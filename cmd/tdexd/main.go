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

	pricefeederinfra "github.com/tdex-network/tdex-daemon/internal/infrastructure/price-feeder"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/internal/config"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	oceanwallet "github.com/tdex-network/tdex-daemon/internal/infrastructure/ocean-wallet"
	webhookpubsub "github.com/tdex-network/tdex-daemon/internal/infrastructure/pubsub/webhook"
	swap_parser "github.com/tdex-network/tdex-daemon/internal/infrastructure/swap-parser"
	"github.com/tdex-network/tdex-daemon/internal/interfaces"
	grpcinterface "github.com/tdex-network/tdex-daemon/internal/interfaces/grpc"
	boltsecurestore "github.com/tdex-network/tdex-daemon/pkg/securestore/bolt"
	"github.com/tdex-network/tdex-daemon/pkg/stats"

	_ "net/http/pprof" // #nosec
)

var (
	// General config
	logLevel, tradeSvcPort, operatorSvcPort, statsInterval int
	noMacaroons, noOperatorTls, profilerEnabled            bool
	datadir, dbDir, profilerDir, tradeTLSKey, tradeTLSCert string
	walletUnlockPasswordFile, dbType, oceanWalletAddr      string
	connectAddr, connectProto                              string
	operatorTLSExtraIPs, operatorTLSExtraDomains           []string
	// App services config
	feeBalanceThreshold                   uint64
	pricesSlippagePercentage, satsPerByte decimal.Decimal

	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := loadConfig(); err != nil {
		log.WithError(err).Fatal("failed to init config")
	}

	log.SetLevel(log.Level(logLevel))
	domain.SwapParserManager = swap_parser.NewService()

	// Profiler is enabled at url http://localhost:8024/debug/pprof/
	if profilerEnabled {
		runtime.SetBlockProfileRate(1)
		//nolint
		go http.ListenAndServe(":8024", nil)
	}

	// Init services to be used by those of the application layer.
	wallet, err := oceanwallet.NewService(oceanWalletAddr)
	if err != nil {
		log.WithError(err).Fatal("failed to connect to ocean wallet")
	}

	pubsub, err := newWebhookPubSubService(dbDir)
	if err != nil {
		log.WithError(err).Fatal("failed to initialize pubsub service")
	}

	priceFeederSvc, err := newPriceFeederService()
	if err != nil {
		log.WithError(err).Fatal("failed to initialize price feeder service")
	}

	appConfig := &application.Config{
		OceanWallet:         wallet,
		SecurePubSub:        pubsub,
		PriceFeederSvc:      priceFeederSvc,
		FeeBalanceThreshold: feeBalanceThreshold,
		TradePriceSlippage:  pricesSlippagePercentage,
		TradeSatsPerByte:    satsPerByte,
		DBType:              dbType,
		DBConfig:            dbDir,
	}

	runOnOnePort := operatorSvcPort == tradeSvcPort
	svc, err := NewGrpcService(runOnOnePort, appConfig)
	if err != nil {
		log.WithError(err).Fatal("failed to initialize grpc service")
	}
	log.RegisterExitHandler(svc.Stop)

	log.Info("starting daemon")

	if log.GetLevel() >= log.DebugLevel {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		interval := time.Duration(statsInterval) * time.Second
		stats.EnableMemoryStatistics(ctx, interval, profilerDir)
	}

	// Start gRPC service interfaces.
	if err := svc.Start(); err != nil {
		log.WithError(err).Error("failed to start daemon")
		return
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	log.Info("shutting down daemon")
	log.Exit(0)
}

func loadConfig() error {
	if err := config.InitConfig(); err != nil {
		return err
	}
	logLevel = config.GetInt(config.LogLevelKey)
	profilerEnabled = config.GetBool(config.EnableProfilerKey)
	datadir = config.GetDatadir()
	dbDir = filepath.Join(datadir, config.DbLocation)
	profilerDir = filepath.Join(datadir, config.ProfilerLocation)
	noMacaroons = config.GetBool(config.NoMacaroonsKey)
	noOperatorTls = config.GetBool(config.NoOperatorTlsKey)
	statsInterval = config.GetInt(config.StatsIntervalKey)
	tradeTLSKey = config.GetString(config.TradeTLSKeyKey)
	tradeTLSCert = config.GetString(config.TradeTLSCertKey)
	operatorTLSExtraIPs = config.GetStringSlice(config.OperatorExtraIPKey)
	operatorTLSExtraDomains = config.GetStringSlice(config.OperatorExtraDomainKey)
	walletUnlockPasswordFile = config.GetString(config.WalletUnlockPasswordFile)
	connectAddr = config.GetString(config.ConnectAddrKey)
	connectProto = config.GetString(config.ConnectProtoKey)
	dbType = config.GetString(config.DBTypeKey)
	// App services config
	pricesSlippagePercentage = decimal.NewFromFloat(config.GetFloat(config.PriceSlippageKey))
	satsPerByte = decimal.NewFromFloat(config.GetFloat(config.TradeSatsPerByte))
	feeBalanceThreshold = uint64(config.GetInt(config.FeeAccountBalanceThresholdKey))
	tradeSvcPort = config.GetInt(config.TradeListeningPortKey)
	operatorSvcPort = config.GetInt(config.OperatorListeningPortKey)
	oceanWalletAddr = config.GetString(config.OceanWalletAddrKey)

	return nil
}

type buildData struct{}

func (bd buildData) GetVersion() string {
	return version
}
func (bd buildData) GetCommit() string {
	return commit
}
func (bd buildData) GetDate() string {
	return date
}

func newWebhookPubSubService(datadir string) (ports.SecurePubSub, error) {
	secureStore, err := boltsecurestore.NewSecureStorage(datadir, "pubsub.db")
	if err != nil {
		return nil, err
	}
	return webhookpubsub.NewWebhookPubSubService(secureStore)
}

func newPriceFeederService() (ports.PriceFeeder, error) {
	dbDir := filepath.Join(datadir, "db")
	store, err := pricefeederinfra.NewPriceFeedStoreImpl(dbDir, log.New())
	if err != nil {
		return nil, err
	}

	priceFeedSvc := pricefeederinfra.NewService(store)

	return priceFeedSvc, nil
}

func NewGrpcService(
	runOnOnePort bool, appConfig *application.Config,
) (interfaces.Service, error) {
	addr := fmt.Sprintf("localhost:%d", operatorSvcPort)
	if connectAddr != "" {
		addr = connectAddr
	}

	if runOnOnePort {
		opts := grpcinterface.ServiceOptsOnePort{
			NoMacaroons:              noMacaroons,
			Datadir:                  datadir,
			DBLocation:               config.DbLocation,
			MacaroonsLocation:        config.MacaroonsLocation,
			WalletUnlockPasswordFile: walletUnlockPasswordFile,
			Port:                     tradeSvcPort,
			TLSKey:                   tradeTLSKey,
			TLSCert:                  tradeTLSCert,
			ConnectAddr:              addr,
			ConnectProto:             connectProto,
			BuildData:                buildData{},
			AppConfig:                appConfig,
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
		OperatorPort:             operatorSvcPort,
		TradePort:                tradeSvcPort,
		TradeTLSKey:              tradeTLSKey,
		TradeTLSCert:             tradeTLSCert,
		WalletUnlockPasswordFile: walletUnlockPasswordFile,
		NoOperatorTls:            noOperatorTls,
		ConnectAddr:              addr,
		ConnectProto:             connectProto,
		BuildData:                buildData{},
		AppConfig:                appConfig,
	}

	return grpcinterface.NewService(opts)
}
