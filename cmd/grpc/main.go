package main

import (
	"context"
	"fmt"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/soheilhy/cmux"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/domain"
	"github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/inmemory"
	grpchandler "github.com/tdex-network/tdex-daemon/internal/interfaces/grpc/handler"
	"github.com/tdex-network/tdex-daemon/internal/interfaces/grpc/interceptor"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/storage"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"google.golang.org/grpc"

	tradeservice "github.com/tdex-network/tdex-daemon/internal/service/trader"
	pbhandshake "github.com/tdex-network/tdex-protobuf/generated/go/handshake"
	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbtrader "github.com/tdex-network/tdex-protobuf/generated/go/trade"
	pbwallet "github.com/tdex-network/tdex-protobuf/generated/go/wallet"
)

func main() {
	log.SetLevel(log.Level(config.GetInt(config.LogLevelKey)))
	config.Validate()

	// Ports
	var traderAddress = fmt.Sprintf(":%+v", config.GetInt(config.TraderListeningPortKey))
	var operatorAddress = fmt.Sprintf(":%+v", config.GetInt(config.OperatorListeningPortKey))
	// Grpc Server
	traderGrpcServer := grpc.NewServer(
		interceptor.UnaryLoggerInterceptor(),
		interceptor.StreamLoggerInterceptor(),
	)
	operatorGrpcServer := grpc.NewServer(
		interceptor.UnaryLoggerInterceptor(),
		interceptor.StreamLoggerInterceptor(),
	)

	// Init market in-memory storage
	marketRepository := storage.NewInMemoryMarketRepository()
	unspentRepository := storage.NewInMemoryUnspentRepository()
	vaultRepository := storage.NewInMemoryVaultRepository()
	tradeRepository := storage.NewInMemoryTradeRepository()

	explorerSvc := explorer.NewService(config.GetString(config.ExplorerEndpointKey))

	unspentRepo := inmemory.NewUnspentRepository()
	vaultRepo := inmemory.NewVaultRepositoryImpl()
	marketRepo := inmemory.NewMarketRepositoryImpl()
	tradeRepo := inmemory.NewTradeRepositoryImpl()
	observables, err := getObjectsToObserve(marketRepo, vaultRepo)
	if err != nil {
		log.WithError(err).Panic(err)
	}

	errorHandler := func(err error) {
		log.Warn(err)
	}
	crawlerSvc := crawler.NewService(explorerSvc, observables, errorHandler)

	// Init services
	tradeSvc := tradeservice.NewService(
		marketRepository,
		unspentRepository,
		vaultRepository,
		tradeRepository,
		explorerSvc,
	)

	walletSvc := application.NewWalletService(vaultRepo, unspentRepo, explorerSvc)
	walletHandler := grpchandler.NewWalletHandler(walletSvc)

	operatorSvc := application.NewOperatorService(
		marketRepo,
		vaultRepo,
		tradeRepo,
		unspentRepo,
		explorerSvc,
		crawlerSvc,
	)

	operatorSvc.ObserveBlockchain()
	operatorHandler := grpchandler.NewOperatorHandler(operatorSvc)

	// Register proto implementations on Trader interface
	pbtrader.RegisterTradeServer(traderGrpcServer, tradeSvc)
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

func getObjectsToObserve(
	marketRepository domain.MarketRepository,
	vaultRepository domain.VaultRepository,
) ([]crawler.Observable, error) {

	//get all market addresses to observe
	markets, err := marketRepository.GetAllMarkets(context.Background())
	if err != nil {
		return nil, err
	}

	observables := make([]crawler.Observable, 0)
	for _, m := range markets {
		addresses, blindingKeys, err := vaultRepository.GetAllDerivedAddressesAndBlindingKeysForAccount(
			context.Background(),
			m.AccountIndex(),
		)
		if err != nil {
			return nil, err
		}

		for i, a := range addresses {
			observables = append(observables, &crawler.AddressObservable{
				AccountIndex: m.AccountIndex(),
				Address:      a,
				BlindingKey:  blindingKeys[i],
			})
		}
	}

	//get fee account addresses to observe
	addresses, blindingKeys, err := vaultRepository.GetAllDerivedAddressesAndBlindingKeysForAccount(
		context.Background(),
		domain.FeeAccount,
	)
	if err != nil {
		return nil, err
	}
	for i, a := range addresses {
		observables = append(observables, &crawler.AddressObservable{
			AccountIndex: domain.FeeAccount,
			Address:      a,
			BlindingKey:  blindingKeys[i],
		})
	}

	return observables, nil
}
