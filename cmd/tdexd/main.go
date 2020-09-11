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
	"github.com/tdex-network/tdex-daemon/pkg/wallet"
	"github.com/vulpemventures/go-elements/network"
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
	observables, err := getObjectsToObserv(marketRepository)
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
	log.Debug("wallet interface is listening on " + operatorAddress)

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

//TODO fetch all addresses to be observed - dummy implementation below
func getObjectsToObserv(marketRepo market.Repository) (
	[]crawler.Observable, error) {
	markets, err := marketRepo.GetAllMarkets(context.Background())
	if err != nil {
		return nil, err
	}

	w, err := wallet.NewWallet(wallet.NewWalletOpts{
		ExtraMnemonic: false,
	})

	observables := make([]crawler.Observable, 0)
	for _, m := range markets {
		opts := wallet.DeriveConfidentialAddressOpts{
			DerivationPath: fmt.Sprintf("%v'/0/0", m.AccountIndex()),
			Network:        &network.Liquid,
		}
		ctAddress, _, err := w.DeriveConfidentialAddress(opts)
		if err != nil {
			return nil, err
		}
		observables = append(observables, crawler.Observable{
			AccountType: vault.FeeAccount,
			AssetHash:   config.GetString(config.BaseAssetKey),
			Address:     ctAddress,
			BlindingKey: nil,
		})
	}

	return observables, nil
}
