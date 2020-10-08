package main

import (
	"fmt"
	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/soheilhy/cmux"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/inmemory"
	grpchandler "github.com/tdex-network/tdex-daemon/internal/interfaces/grpc/handler"
	"github.com/tdex-network/tdex-daemon/internal/interfaces/grpc/interceptor"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"google.golang.org/grpc"

	pbhandshake "github.com/tdex-network/tdex-protobuf/generated/go/handshake"
	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbtrader "github.com/tdex-network/tdex-protobuf/generated/go/trade"
	pbwallet "github.com/tdex-network/tdex-protobuf/generated/go/wallet"
)

func main() {
	log.SetLevel(log.Level(config.GetInt(config.LogLevelKey)))
	config.Validate()
	if err := config.InitDataDir(); err != nil {
		log.WithError(err).Panic("error while init data dir")
	}

	dbDir := filepath.Join(config.GetString(config.DataDirPathKey), "db")
	dbManager, err := dbbadger.NewDbManager(dbDir)
	if err != nil {
		log.WithError(err).Panic("error while opening db")
	}
	defer dbManager.Store.Close()

	unspentRepository := inmemory.NewUnspentRepositoryImpl()
	vaultRepository := inmemory.NewVaultRepositoryImpl()
	marketRepository := dbbadger.NewMarketRepositoryImpl(dbManager)
	tradeRepository := inmemory.NewTradeRepositoryImpl()

	errorHandler := func(err error) { log.Warn(err) }
	explorerSvc := explorer.NewService(config.GetString(config.ExplorerEndpointKey))
	crawlerSvc := crawler.NewService(explorerSvc, []crawler.Observable{}, errorHandler)
	traderSvc := application.NewTraderService(
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
	operatorSvc := application.NewOperatorService(
		marketRepository,
		vaultRepository,
		tradeRepository,
		unspentRepository,
		explorerSvc,
		crawlerSvc,
	)
	operatorSvc.ObserveBlockchain()

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
	pbhandshake.RegisterHandshakeServer(traderGrpcServer, &pbhandshake.UnimplementedHandshakeServer{})
	// Register proto implementations on Operator interface
	pboperator.RegisterOperatorServer(operatorGrpcServer, operatorHandler)
	pbwallet.RegisterWalletServer(operatorGrpcServer, walletHandler)

	log.Debug("starting daemon")

	// Serve grpc and grpc-web multiplexed on the same port
	if err := serveMux(traderAddress, traderGrpcServer); err != nil {
		log.WithError(err).Panic("error listening on trader interface")
	}
	if err := serveMux(operatorAddress, operatorGrpcServer); err != nil {
		log.WithError(err).Panic("error listening on operator interface")
	}

	log.Debug("trader interface is listening on " + traderAddress)
	log.Debug("operator interface is listening on " + operatorAddress)

	defer traderGrpcServer.Stop()
	defer operatorGrpcServer.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

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
