package main

import (
	"context"
	"encoding/hex"
	"fmt"
	dbbadger "github.com/tdex-network/tdex-daemon/internal/infrastructure/storage/db/badger"
	"github.com/tdex-network/tdex-daemon/pkg/macaroons"
	"gopkg.in/macaroon-bakery.v2/bakery"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

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

	pbhandshake "github.com/tdex-network/tdex-protobuf/generated/go/handshake"
	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbtrader "github.com/tdex-network/tdex-protobuf/generated/go/trade"
	pbwallet "github.com/tdex-network/tdex-protobuf/generated/go/wallet"
)

var (
	pricePermissions = []bakery.Op{
		{
			Entity: "market",
			Action: "updatemarketprice",
		},
	}
	marketPermissions = []bakery.Op{
		{
			Entity: "market",
			Action: "openmarket",
		},
		{
			Entity: "market",
			Action: "closemarket",
		},
		{
			Entity: "market",
			Action: "updatemarketstrategy",
		},
	}
	//TODO check permissions
	readonlyPermissions = []bakery.Op{}

	adminPermissions = []bakery.Op{
		{
			Entity: "wallet",
			Action: "genseed",
		},
		{
			Entity: "wallet",
			Action: "initwallet",
		},
		{
			Entity: "wallet",
			Action: "unlockwallet",
		},
		{
			Entity: "wallet",
			Action: "changepassword",
		},
		{
			Entity: "wallet",
			Action: "walletaddress",
		},
		{
			Entity: "wallet",
			Action: "walletbalance",
		},
		{
			Entity: "wallet",
			Action: "sendtomany",
		},
		{
			Entity: "operator",
			Action: "depositmarket",
		},
		{
			Entity: "operator",
			Action: "listdepositmarket",
		},
		{
			Entity: "operator",
			Action: "depositfeeaccount",
		},
		{
			Entity: "operator",
			Action: "balancefeeaccount",
		},
		{
			Entity: "operator",
			Action: "openmarket",
		},
		{
			Entity: "operator",
			Action: "closemarket",
		},
		{
			Entity: "operator",
			Action: "listmarket",
		},
		{
			Entity: "operator",
			Action: "updatemarketfee",
		},
		{
			Entity: "operator",
			Action: "updatemarketprice",
		},
		{
			Entity: "operator",
			Action: "updatemarketstrategy",
		},
		{
			Entity: "operator",
			Action: "withdrawmarket",
		},
		{
			Entity: "operator",
			Action: "listswaps",
		},
		{
			Entity: "operator",
			Action: "reportmarketfee",
		},
		{
			Entity: "trade",
			Action: "markets",
		},
		{
			Entity: "trade",
			Action: "balances",
		},
		{
			Entity: "trade",
			Action: "marketprice",
		},
		{
			Entity: "trade",
			Action: "tradepropose",
		},
		{
			Entity: "trade",
			Action: "tradecomplete",
		},
	}
)

func main() {
	log.SetLevel(log.Level(config.GetInt(config.LogLevelKey)))

	dbDir := filepath.Join(config.GetString(config.DataDirPathKey), "db")
	dbManager, err := dbbadger.NewDbManager(dbDir)
	if err != nil {
		log.WithError(err).Panic("error while opening db")
	}
	defer dbManager.Store.Close()

	// ****
	var macaroonService *macaroons.Service
	macaroonService, err = macaroons.NewService(
		config.GetString(config.DataDirPathKey),
		"tdex-daemon",
		macaroons.IPLockChecker,
	)
	if err != nil {
		log.WithError(err).Panic("unable to set up macaroon authentication")
	}
	defer macaroonService.Close()

	var pass = []byte("hello")
	// Try to unlock the macaroon store with the private password.
	err = macaroonService.CreateUnlock(&pass)
	if err != nil {
		log.WithError(err).Panic("unable to unlock macaroons")
	}

	adminMacPath := filepath.Join(
		config.GetString(config.DataDirPathKey), "admin.macaroon",
	)

	readMacPath := filepath.Join(
		config.GetString(config.DataDirPathKey), "readonly.macaroon",
	)

	priceMacPath := filepath.Join(
		config.GetString(config.DataDirPathKey), "price.macaroon",
	)

	marketMacPath := filepath.Join(
		config.GetString(config.DataDirPathKey), "market.macaroon",
	)

	// Create macaroon files for lncli to use if they don't exist.
	if !fileExists(adminMacPath) && !fileExists(readMacPath) &&
		!fileExists(priceMacPath) && !fileExists(marketMacPath) {

		err = genMacaroons(
			context.Background(),
			macaroonService,
			adminMacPath,
			readMacPath,
			priceMacPath,
			marketMacPath,
		)
		if err != nil {
			log.WithError(err).Panic("unable to create macaroons")
		}
	}
	// ****

	unspentRepository := dbbadger.NewUnspentRepositoryImpl(dbManager)
	vaultRepository := dbbadger.NewVaultRepositoryImpl(dbManager)
	marketRepository := dbbadger.NewMarketRepositoryImpl(dbManager)
	tradeRepository := dbbadger.NewTradeRepositoryImpl(dbManager)

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
		interceptor.UnaryInterceptor(dbManager, macaroonService),
		interceptor.StreamInterceptor(dbManager, macaroonService),
	)
	operatorGrpcServer := grpc.NewServer(
		interceptor.UnaryInterceptor(dbManager, macaroonService),
		interceptor.StreamInterceptor(dbManager, macaroonService),
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

// ****

// genMacaroons generates three macaroon files; one admin-level, one for
// invoice access and one read-only. These can also be used to generate more
// granular macaroons.
func genMacaroons(ctx context.Context, svc *macaroons.Service,
	admFile, roFile, priceFile, marketFile string) error {

	priceMac, err := svc.NewMacaroon(
		ctx, macaroons.DefaultRootKeyID, pricePermissions...,
	)
	if err != nil {
		return err
	}
	priceUpdateMacBytes, err := priceMac.M().MarshalBinary()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(priceFile, priceUpdateMacBytes, 0644)
	if err != nil {
		os.Remove(priceFile)
		return err
	}

	marketMac, err := svc.NewMacaroon(
		ctx, macaroons.DefaultRootKeyID, marketPermissions...,
	)
	if err != nil {
		return err
	}
	marketMacBytes, err := marketMac.M().MarshalBinary()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(marketFile, marketMacBytes, 0644)
	if err != nil {
		os.Remove(marketFile)
		return err
	}

	//TODO uncomment once read only permissions are defined
	//// Generate the read-only macaroon and write it to a file.
	//roMacaroon, err := svc.NewMacaroon(
	//	ctx, macaroons.DefaultRootKeyID, readonlyPermissions...,
	//)
	//if err != nil {
	//	return err
	//}
	//roBytes, err := roMacaroon.M().MarshalBinary()
	//if err != nil {
	//	return err
	//}
	//if err = ioutil.WriteFile(roFile, roBytes, 0644); err != nil {
	//	os.Remove(admFile)
	//	return err
	//}

	// Generate the admin macaroon and write it to a file.
	admMacaroon, err := svc.NewMacaroon(
		ctx, macaroons.DefaultRootKeyID, adminPermissions...,
	)
	if err != nil {
		return err
	}
	admBytes, err := admMacaroon.M().MarshalBinary()
	if err != nil {
		return err
	}
	println(hex.EncodeToString(admBytes))

	if err = ioutil.WriteFile(admFile, admBytes, 0600); err != nil {
		return err
	}

	return nil
}

// fileExists reports whether the named file or directory exists.
// This function is taken from https://github.com/btcsuite/btcd
func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// ****
