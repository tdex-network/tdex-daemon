package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/domain/market"
	"github.com/tdex-network/tdex-daemon/internal/domain/vault"
	"github.com/tdex-network/tdex-daemon/internal/grpcutil"
	"github.com/tdex-network/tdex-daemon/internal/storage"
	"github.com/tdex-network/tdex-daemon/pkg/crawler"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"google.golang.org/grpc"

	operatorservice "github.com/tdex-network/tdex-daemon/internal/service/operator"
	tradeservice "github.com/tdex-network/tdex-daemon/internal/service/trader"
	walletservice "github.com/tdex-network/tdex-daemon/internal/service/wallet"

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
	traderGrpcServer := grpc.NewServer(grpcutil.UnaryLoggerInterceptor(), grpcutil.StreamLoggerInterceptor())
	operatorGrpcServer := grpc.NewServer(grpcutil.UnaryLoggerInterceptor(), grpcutil.StreamLoggerInterceptor())

	// Init market in-memory storage
	marketRepository := storage.NewInMemoryMarketRepository()
	unspentRepository := storage.NewInMemoryUnspentRepository()
	vaultRepository := storage.NewInMemoryVaultRepository()

	explorerSvc := explorer.NewService()
	observables, err := getObjectsToObserve(marketRepository, vaultRepository)
	crawlerSvc := crawler.NewService(explorerSvc, observables, nil)

	// Init services
	tradeSvc := tradeservice.NewService(marketRepository)
	walletSvc := walletservice.NewService(
		vaultRepository,
		explorerSvc,
	)
	operatorSvc, err := operatorservice.NewService(
		marketRepository,
		unspentRepository,
		vaultRepository,
		crawlerSvc,
		explorerSvc,
	)
	if err != nil {
		log.WithError(err).Panic(err)
	}
	operatorSvc.ObserveBlockchain()

	// Register proto implementations on Trader interface
	pbtrader.RegisterTradeServer(traderGrpcServer, tradeSvc)
	pbhandshake.RegisterHandshakeServer(traderGrpcServer, &pbhandshake.UnimplementedHandshakeServer{})
	// Register proto implementations on Operator interface
	pboperator.RegisterOperatorServer(operatorGrpcServer, operatorSvc)
	pbwallet.RegisterWalletServer(operatorGrpcServer, walletSvc)

	log.Debug("starting daemon")

	// Serve grpc and grpc-web multiplexed on the same port
	if err := grpcutil.ServeMux(traderAddress, traderGrpcServer); err != nil {
		log.WithError(err).Panic("error listening on trader interface")
	}
	if err := grpcutil.ServeMux(operatorAddress, operatorGrpcServer); err != nil {
		log.WithError(err).Panic("error listening on operator interface")
	}

	log.Debug("trader interface is listening on " + traderAddress)
	log.Debug("operator interface is listening on " + operatorAddress)

	// TODO: to be removed.
	// Add a sample market
	tradeSvc.AddTestMarket(true)
	// Add anothet right away
	tradeSvc.AddTestMarket(false)

	defer traderGrpcServer.Stop()
	defer operatorGrpcServer.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	log.Debug("exiting")
}

func getObjectsToObserve(
	marketRepository market.Repository,
	vaultRepository vault.Repository,
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
				AccountType: m.AccountIndex(),
				Address:     a,
				BlindingKey: blindingKeys[i],
			})
		}
	}

	//get fee account addresses to observe
	addresses, blindingKeys, err := vaultRepository.GetAllDerivedAddressesAndBlindingKeysForAccount(
		context.Background(),
		vault.FeeAccount,
	)
	if err != nil {
		return nil, err
	}
	for i, a := range addresses {
		observables = append(observables, &crawler.AddressObservable{
			AccountType: vault.FeeAccount,
			Address:     a,
			BlindingKey: blindingKeys[i],
		})
	}

	return observables, nil
}
