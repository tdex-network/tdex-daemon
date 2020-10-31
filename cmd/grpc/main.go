package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/soheilhy/cmux"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	grpchandler "github.com/tdex-network/tdex-daemon/internal/interfaces/grpc/handler"
	"github.com/tdex-network/tdex-daemon/internal/interfaces/grpc/interceptor"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"google.golang.org/grpc"

	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbtrader "github.com/tdex-network/tdex-protobuf/generated/go/trade"
	pbwallet "github.com/tdex-network/tdex-protobuf/generated/go/wallet"
)

func main() {
	log.SetLevel(log.Level(config.GetInt(config.LogLevelKey)))

	dbDir := filepath.Join(config.GetString(config.DataDirPathKey), "db")
	dbManager, err := dbbadger.NewDbManager(dbDir, log.New())
	if err != nil {
		log.WithError(err).Panic("error while opening db")
	}

	unspentRepository := dbbadger.NewUnspentRepositoryImpl(dbManager)
	vaultRepository := dbbadger.NewVaultRepositoryImpl(dbManager)
	marketRepository := dbbadger.NewMarketRepositoryImpl(dbManager)
	tradeRepository := dbbadger.NewTradeRepositoryImpl(dbManager)

	errorHandler := func(err error) { log.Warn(err) }
	explorerSvc := explorer.NewService(config.GetString(config.ExplorerEndpointKey))
	crawlerSvc := crawler.NewService(explorerSvc, []crawler.Observable{}, errorHandler)
	traderSvc := application.NewTradeService(
		marketRepository,
		tradeRepository,
		vaultRepository,
		unspentRepository,
		explorerSvc,
	)
	walletSvc := application.NewWalletService(
		vaultRepository,
		unspentRepository,
		crawlerSvc,
		explorerSvc,
	)

	blockchainListener := application.NewBlockchainListener(
		unspentRepository,
		marketRepository,
		vaultRepository,
		crawlerSvc,
		explorerSvc,
		dbManager,
	)
	blockchainListener.ObserveBlockchain()

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

	traderHandler := grpchandler.NewTraderHandler(traderSvc)
	walletHandler := grpchandler.NewWalletHandler(walletSvc)
	operatorHandler := grpchandler.NewOperatorHandler(operatorSvc)

	// Register proto implementations on Trader interface
	pbtrader.RegisterTradeServer(traderGrpcServer, traderHandler)
	// Register proto implementations on Operator interface
	pboperator.RegisterOperatorServer(operatorGrpcServer, operatorHandler)
	pbwallet.RegisterWalletServer(operatorGrpcServer, walletHandler)

	log.Debug("starting daemon")

	defer stop(
		dbManager,
		crawlerSvc,
		traderGrpcServer,
		operatorGrpcServer,
	)

	// Serve grpc and grpc-web multiplexed on the same port
	if err := serveMux(traderAddress, traderGrpcServer); err != nil {
		log.WithError(err).Panic("error listening on trader interface")
	}
	if err := serveMux(operatorAddress, operatorGrpcServer); err != nil {
		log.WithError(err).Panic("error listening on operator interface")
	}

	log.Debug("trader interface is listening on " + traderAddress)
	log.Debug("operator interface is listening on " + operatorAddress)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	log.Debug("shutting down daemon")
}

func stop(
	dbManager *dbbadger.DbManager,
	crawlerSvc crawler.Service,
	traderServer *grpc.Server,
	operatorServer *grpc.Server,
) {
	operatorServer.Stop()
	log.Debug("disabled operator interface")
	traderServer.Stop()
	log.Debug("disabled trader interface")
	crawlerSvc.Stop()
	log.Debug("stop observing blockchain")
	dbManager.Store.Close()
	dbManager.UnspentStore.Close()
	dbManager.PriceStore.Close()
	log.Debug("closed connection with database")
	log.Debug("exiting")
}

func serveMux(address string, grpcServer *grpc.Server) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	mux := cmux.New(lis)
	grpcL := mux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldPrefixSendSettings("content-type", "application/grpc"))
	httpL := mux.Match(cmux.HTTP1Fast())

	grpcWebServer := grpcweb.WrapServer(grpcServer)

	go grpcServer.Serve(grpcL)
	go http.Serve(httpL, http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if grpcWebServer.IsGrpcWebRequest(req) {
			grpcWebServer.ServeHTTP(resp, req)
		}
	}))

	go mux.Serve()
	return nil
}
