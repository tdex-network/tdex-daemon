package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/tdex-network/tdex-daemon/pkg/stats"

	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
	"golang.org/x/net/http2"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/soheilhy/cmux"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	grpchandler "github.com/tdex-network/tdex-daemon/internal/interfaces/grpc/handler"
	"github.com/tdex-network/tdex-daemon/internal/interfaces/grpc/interceptor"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	"google.golang.org/grpc"

	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbtrader "github.com/tdex-network/tdex-protobuf/generated/go/trade"
	pbwallet "github.com/tdex-network/tdex-protobuf/generated/go/wallet"

	_ "net/http/pprof"
)

func main() {
	log.SetLevel(log.Level(config.GetInt(config.LogLevelKey)))

	//http://localhost:8024/debug/pprof/
	if config.GetBool(config.EnableProfilerKey) {
		runtime.SetBlockProfileRate(1)
		go func() {
			http.ListenAndServe(
				":8024",
				nil,
			)
		}()
	}

	dbDir := filepath.Join(config.GetString(config.DataDirPathKey), "db")
	dbManager, err := dbbadger.NewDbManager(dbDir, log.New())
	if err != nil {
		log.WithError(err).Panic("error while opening db")
	}

	unspentRepository := dbbadger.NewUnspentRepositoryImpl(dbManager)
	vaultRepository := dbbadger.NewVaultRepositoryImpl(dbManager)
	marketRepository := dbbadger.NewMarketRepositoryImpl(dbManager)
	tradeRepository := dbbadger.NewTradeRepositoryImpl(dbManager)

	explorerSvc, err := config.GetExplorer()
	if err != nil {
		log.WithError(err).Panic("error while setting up explorer service")
	}

	crawlerSvc := crawler.NewService(crawler.Opts{
		ExplorerSvc:            explorerSvc,
		Observables:            []crawler.Observable{},
		ErrorHandler:           func(err error) { log.Warn(err) },
		IntervalInMilliseconds: config.GetInt(config.CrawlIntervalKey),
	})
	blockchainListener := application.NewBlockchainListener(
		unspentRepository,
		marketRepository,
		vaultRepository,
		tradeRepository,
		crawlerSvc,
		explorerSvc,
		dbManager,
	)

	traderSvc := application.NewTradeService(
		marketRepository,
		tradeRepository,
		vaultRepository,
		unspentRepository,
		explorerSvc,
		crawlerSvc,
	)
	walletSvc := application.NewWalletService(
		vaultRepository,
		unspentRepository,
		crawlerSvc,
		explorerSvc,
		blockchainListener,
	)

	operatorSvc := application.NewOperatorService(
		marketRepository,
		vaultRepository,
		tradeRepository,
		unspentRepository,
		explorerSvc,
		crawlerSvc,
	)

	// Ports
	traderAddress := fmt.Sprintf(":%+v", config.GetInt(config.TraderListeningPortKey))
	operatorAddress := fmt.Sprintf(":%+v", config.GetInt(config.OperatorListeningPortKey))
	// Grpc Server
	traderGrpcServer := grpc.NewServer(
		interceptor.UnaryInterceptor(dbManager),
		interceptor.StreamInterceptor(dbManager),
	)
	operatorGrpcServer := grpc.NewServer(
		interceptor.UnaryInterceptor(dbManager),
		interceptor.StreamInterceptor(dbManager),
	)

	traderHandler := grpchandler.NewTraderHandler(traderSvc, dbManager)
	walletHandler := grpchandler.NewWalletHandler(walletSvc, dbManager)
	operatorHandler := grpchandler.NewOperatorHandler(operatorSvc, dbManager)

	// Register proto implementations on Trader interface
	pbtrader.RegisterTradeServer(traderGrpcServer, traderHandler)
	// Register proto implementations on Operator interface
	pboperator.RegisterOperatorServer(operatorGrpcServer, operatorHandler)
	pbwallet.RegisterWalletServer(operatorGrpcServer, walletHandler)

	log.Info("starting daemon")

	var cancelStats context.CancelFunc
	if log.GetLevel() >= log.DebugLevel {
		statsDir := filepath.Join(config.GetString(config.DataDirPathKey), "stats")
		var ctx context.Context
		ctx, cancelStats = context.WithCancel(context.Background())
		stats.EnableMemoryStatistics(
			ctx,
			config.GetDuration(config.StatsIntervalKey)*time.Second,
			statsDir,
		)
	}

	defer stop(
		dbManager,
		blockchainListener,
		traderGrpcServer,
		operatorGrpcServer,
		cancelStats,
	)

	// Serve grpc and grpc-web multiplexed on the same port
	if err := serveMux(traderAddress, true, traderGrpcServer); err != nil {
		log.WithError(err).Panic("error listening on trader interface")
	}
	if err := serveMux(operatorAddress, false, operatorGrpcServer); err != nil {
		log.WithError(err).Panic("error listening on operator interface")
	}

	log.Info("trader interface is listening on " + traderAddress)
	log.Info("operator interface is listening on " + operatorAddress)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	log.Info("shutting down daemon")
}

func stop(
	dbManager *dbbadger.DbManager,
	blockchainListener application.BlockchainListener,
	traderServer *grpc.Server,
	operatorServer *grpc.Server,
	cancelStats context.CancelFunc,
) {
	if log.GetLevel() >= log.DebugLevel {
		cancelStats()
		time.Sleep(1 * time.Second)
		log.Debug("cancel printing statistics")
	}

	operatorServer.Stop()
	log.Debug("disabled operator interface")

	traderServer.Stop()
	log.Debug("disabled trader interface")

	blockchainListener.StopObserveBlockchain()
	// give the crawler the time to terminate
	time.Sleep(
		time.Duration(config.GetInt(config.CrawlIntervalKey)) * time.Millisecond,
	)
	log.Debug("stopped observing blockchain")

	dbManager.Store.Close()
	dbManager.UnspentStore.Close()
	dbManager.PriceStore.Close()
	log.Debug("closed connection with database")

	log.Debug("exiting")
}

func serveMux(address string, withSsl bool, grpcServer *grpc.Server) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	if sslKey := config.GetString(config.SSLKeyPathKey); sslKey != "" && withSsl {
		certificate, err := tls.LoadX509KeyPair(config.GetString(config.SSLCertPathKey), sslKey)
		if err != nil {
			return err
		}

		const requiredCipher = tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
		config := &tls.Config{
			CipherSuites: []uint16{requiredCipher},
			NextProtos:   []string{"http/1.1", http2.NextProtoTLS, "h2-14"}, // h2-14 is just for compatibility. will be eventually removed.
			Certificates: []tls.Certificate{certificate},
		}
		config.Rand = rand.Reader

		lis = tls.NewListener(lis, config)
	}

	mux := cmux.New(lis)
	grpcL := mux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldPrefixSendSettings("content-type", "application/grpc"))
	httpL := mux.Match(cmux.HTTP1Fast())

	grpcWebServer := grpcweb.WrapServer(
		grpcServer,
		grpcweb.WithCorsForRegisteredEndpointsOnly(false),
		grpcweb.WithOriginFunc(func(origin string) bool { return true }),
	)

	go grpcServer.Serve(grpcL)
	go http.Serve(httpL, http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if isValidRequest(req) {
			grpcWebServer.ServeHTTP(resp, req)
		}
	}))

	go mux.Serve()
	return nil
}

func isValidRequest(req *http.Request) bool {
	return isValidGrpcWebOptionRequest(req) || isValidGrpcWebRequest(req)
}

func isValidGrpcWebRequest(req *http.Request) bool {
	return req.Method == http.MethodPost && isValidGrpcContentTypeHeader(req.Header.Get("content-type"))
}

func isValidGrpcContentTypeHeader(contentType string) bool {
	return strings.HasPrefix(contentType, "application/grpc-web-text") ||
		strings.HasPrefix(contentType, "application/grpc-web")
}

func isValidGrpcWebOptionRequest(req *http.Request) bool {
	accessControlHeader := req.Header.Get("Access-Control-Request-Headers")
	return req.Method == http.MethodOptions &&
		strings.Contains(accessControlHeader, "x-grpc-web") &&
		strings.Contains(accessControlHeader, "content-type")
}
