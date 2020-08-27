package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/internal/grpcutil"
	"github.com/tdex-network/tdex-daemon/internal/storage"
	"google.golang.org/grpc"

	operatorservice "github.com/tdex-network/tdex-daemon/internal/service/operator"
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
	traderGrpcServer := grpc.NewServer(grpcutil.UnaryLoggerInterceptor(), grpcutil.StreamLoggerInterceptor())
	operatorGrpcServer := grpc.NewServer(grpcutil.UnaryLoggerInterceptor(), grpcutil.StreamLoggerInterceptor())

	// Init market in-memory storage
	mktStore := storage.NewInMemoryMarketRepository()
	// Init services
	tradeSvc := tradeservice.NewService(mktStore)
	operatorSvc := operatorservice.NewService(mktStore)
	// Register proto implementations on Trader interface
	pbtrader.RegisterTradeServer(traderGrpcServer, tradeSvc)
	pbhandshake.RegisterHandshakeServer(traderGrpcServer, &pbhandshake.UnimplementedHandshakeServer{})
	// Register proto implementations on Operator interface
	pboperator.RegisterOperatorServer(operatorGrpcServer, operatorSvc)
	pbwallet.RegisterWalletServer(operatorGrpcServer, &pbwallet.UnimplementedWalletServer{})

	log.Debug("starting daemon")

	// Serve grpc and grpc-web multiplexed on the same port
	err := grpcutil.ServeMux(traderAddress, traderGrpcServer)
	if err != nil {
		log.WithError(err).Panic("error listening on trader interface")
	}
	err = grpcutil.ServeMux(operatorAddress, operatorGrpcServer)
	if err != nil {
		log.WithError(err).Panic("error listening on operator interface")
	}

	log.Debug("trader interface is listening on " + traderAddress)
	log.Debug("operator interface is listening on " + operatorAddress)

	// TODO: to be removed.
	// Add a sample market
	tradeSvc.AddTestMarket()
	// Add anothet right away
	tradeSvc.AddTestMarket()

	defer traderGrpcServer.Stop()
	defer operatorGrpcServer.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	log.Debug("exiting")
}
