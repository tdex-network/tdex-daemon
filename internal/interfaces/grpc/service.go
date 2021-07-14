package grpcinterface

import (
	"context"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	interfaces "github.com/tdex-network/tdex-daemon/internal/interfaces"
	grpchandler "github.com/tdex-network/tdex-daemon/internal/interfaces/grpc/handler"
	"github.com/tdex-network/tdex-daemon/internal/interfaces/grpc/interceptor"
	"github.com/tdex-network/tdex-daemon/internal/interfaces/grpc/permissions"
	"github.com/tdex-network/tdex-daemon/pkg/macaroons"
	"google.golang.org/grpc"
	"gopkg.in/macaroon-bakery.v2/bakery"

	pboperator "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/operator"
	pbwallet "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/wallet"
	pbwalletunlocker "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/walletunlocker"
	pbtrade "github.com/tdex-network/tdex-protobuf/generated/go/trade"
)

const (
	// OperatorTLSKeyFile is the name of the TLS key file for the Operator
	// interface.
	OperatorTLSKeyFile = "key.pem"
	// OperatorTLSCertFile is the name of the TLS certificate file for the
	// Operator interface.
	OperatorTLSCertFile = "cert.pem"
	// Location is used as the macaroon's location hint. This is not verified as
	// part of the macaroons itself. Check the doc for more info:
	// https://github.com/go-macaroon/macaroon#func-macaroon-location.
	Location = "tdexd"
	// DbFile is the name of the macaroon database file.
	DBFile = "macaroons.db"
	// AdminMacaroonFile is the name of the admin macaroon.
	AdminMacaroonFile = "admin.macaroon"
	// ReadOnlyMacaroonFile is the name of the read-only macaroon.
	ReadOnlyMacaroonFile = "readonly.macaroon"
	// MarketMacaroonFile is the name of the macaroon allowing to open, close and
	// update the strategy of a market.
	MarketMacaroonFile = "market.macaroon"
	// PriceMacaroonFile is the name of the macaroon allowing to update only the
	// prices of markets.
	PriceMacaroonFile = "price.macaroon"
	// WalletMacaroonFile is the name of the macaroon allowing to manage the
	// so called "Wallet" subaccount of the daemon's wallet.
	WalletMacaroonFile = "wallet.macaroon"
	// WebhookMacaroonFile is the name of the macaroon allowing to add, remove or
	// list webhooks.
	WebhookMacaroonFile = "webhook.macaroon"
)

var (
	serialNumberLimit = new(big.Int).Lsh(big.NewInt(1), 128)

	Macaroons = map[string][]bakery.Op{
		AdminMacaroonFile:    permissions.AdminPermissions(),
		ReadOnlyMacaroonFile: permissions.ReadOnlyPermissions(),
		MarketMacaroonFile:   permissions.MarketPermissions(),
		PriceMacaroonFile:    permissions.PricePermissions(),
		WebhookMacaroonFile:  permissions.WebhookPermissions(),
		WalletMacaroonFile:   permissions.WalletPermissions(),
	}
)

type service struct {
	opts        ServiceOpts
	macaroonSvc *macaroons.Service

	grpcOperatorServer *grpc.Server
	grpcTradeServer    *grpc.Server
	httpOperatorServer *http.Server
	httpTradeServer    *http.Server
	muxOperator        cmux.CMux
	muxTrade           cmux.CMux

	passphraseChan chan application.PassphraseMsg
	readyChan      chan bool
}

type ServiceOpts struct {
	NoMacaroons bool

	Datadir              string
	DBLocation           string
	TLSLocation          string
	MacaroonsLocation    string
	OperatorExtraIPs     []string
	OperatorExtraDomains []string

	OperatorAddress string
	TradeAddress    string
	TradeTLSKey     string
	TradeTLSCert    string

	WalletUnlockerSvc application.WalletUnlockerService
	WalletSvc         application.WalletService
	OperatorSvc       application.OperatorService
	TradeSvc          application.TradeService
}

func (o ServiceOpts) validate() error {
	if !pathExists(o.Datadir) {
		return fmt.Errorf("%s: datadir must be an existing directory", o.Datadir)
	}

	if !o.NoMacaroons {
		macDir := o.macaroonsDatadir()
		adminMacExists := pathExists(filepath.Join(macDir, AdminMacaroonFile))
		roMacExists := pathExists(filepath.Join(macDir, ReadOnlyMacaroonFile))
		marketMacExists := pathExists(filepath.Join(macDir, MarketMacaroonFile))
		priceMacExists := pathExists(filepath.Join(macDir, PriceMacaroonFile))

		if adminMacExists != roMacExists ||
			adminMacExists != marketMacExists ||
			adminMacExists != priceMacExists {
			return fmt.Errorf(
				"all macaroons must be either existing or not in path %s", macDir,
			)
		}

		// TLS over operator interface is automatically enabled if macaroons auth
		// is active.
		tlsDir := o.tlsDatadir()
		tlsKeyExists := pathExists(filepath.Join(tlsDir, OperatorTLSKeyFile))
		tlsCertExists := pathExists(filepath.Join(tlsDir, OperatorTLSCertFile))
		if !tlsKeyExists && tlsCertExists {
			return fmt.Errorf(
				"found %s file but %s is missing. Please delete %s to have the daemon recreate both in path %s",
				OperatorTLSCertFile, OperatorTLSKeyFile, OperatorTLSCertFile, tlsDir,
			)
		}

		if len(o.OperatorExtraIPs) > 0 {
			for _, ip := range o.OperatorExtraIPs {
				if net.ParseIP(ip) == nil {
					return fmt.Errorf("invalid operator extra ip %s", ip)
				}
			}
		}
	}

	if ok := isValidAddress(o.OperatorAddress); !ok {
		return fmt.Errorf("operator address is not valid: %s", o.OperatorAddress)
	}
	if ok := isValidAddress(o.TradeAddress); !ok {
		return fmt.Errorf("trade address is not valid: %s", o.OperatorAddress)
	}

	tradeTLSKeyExist := pathExists(o.TradeTLSKey)
	tradeTLSCertExist := pathExists(o.TradeTLSCert)
	if tradeTLSKeyExist != tradeTLSCertExist {
		return fmt.Errorf(
			"TLS key and certificate for Trade interface must be either existing " +
				"or not",
		)
	}

	if o.WalletUnlockerSvc == nil {
		return fmt.Errorf("wallet unlocker app service must not be null")
	}
	if o.WalletSvc == nil {
		return fmt.Errorf("wallet app service must not be null")
	}
	if o.OperatorSvc == nil {
		return fmt.Errorf("operator app service must not be null")
	}
	if o.TradeSvc == nil {
		return fmt.Errorf("trade app service must not be null")
	}
	return nil
}

func (o ServiceOpts) dbDatadir() string {
	return filepath.Join(o.Datadir, o.DBLocation)
}

func (o ServiceOpts) macaroonsDatadir() string {
	return filepath.Join(o.Datadir, o.MacaroonsLocation)
}

func (o ServiceOpts) tlsDatadir() string {
	return filepath.Join(o.Datadir, o.TLSLocation)
}

func (o ServiceOpts) operatorTLSKey() string {
	if o.NoMacaroons {
		return ""
	}
	return filepath.Join(o.tlsDatadir(), OperatorTLSKeyFile)
}

func (o ServiceOpts) operatorTLSCert() string {
	if o.NoMacaroons {
		return ""
	}
	return filepath.Join(o.tlsDatadir(), OperatorTLSCertFile)
}

func NewService(opts ServiceOpts) (interfaces.Service, error) {
	if err := opts.validate(); err != nil {
		return nil, fmt.Errorf("invalid opts: %s", err)
	}

	var macaroonSvc *macaroons.Service
	if !opts.NoMacaroons {
		macaroonSvc, _ = macaroons.NewService(
			opts.dbDatadir(), Location, DBFile, false, macaroons.IPLockChecker,
		)
		if err := generateOperatorTLSKeyCert(
			opts.tlsDatadir(), opts.OperatorExtraIPs, opts.OperatorExtraDomains,
		); err != nil {
			return nil, err
		}
	}

	return &service{
		opts:           opts,
		macaroonSvc:    macaroonSvc,
		passphraseChan: opts.WalletUnlockerSvc.PassphraseChan(),
		readyChan:      opts.WalletUnlockerSvc.ReadyChan(),
	}, nil
}

func (s *service) Start() error {
	walletUnlockerOnly := true
	services, err := s.start(walletUnlockerOnly)
	if err != nil {
		return err
	}

	log.Infof("wallet unlocker interface is listening on %s", s.opts.OperatorAddress)

	go s.startListeningToPassphraseChan()
	go s.startListeningToReadyChan()

	s.grpcOperatorServer = services.grpcOperator
	s.grpcTradeServer = services.grpcTrade
	s.httpOperatorServer = services.httpOperator
	s.httpTradeServer = services.httpTrade
	s.muxOperator = services.muxOperator
	s.muxTrade = services.muxTrade

	return nil
}

func (s *service) Stop() {
	stopMacaroonSvc := true
	s.stop(stopMacaroonSvc)
}

func (s *service) withMacaroons() bool {
	return s.macaroonSvc != nil
}

type services struct {
	grpcOperator *grpc.Server
	grpcTrade    *grpc.Server
	httpOperator *http.Server
	httpTrade    *http.Server
	muxOperator  cmux.CMux
	muxTrade     cmux.CMux
}

func (s *service) start(withUnlockerOnly bool) (*services, error) {
	unaryInterceptor := interceptor.UnaryInterceptor(s.macaroonSvc)
	streamInterceptor := interceptor.StreamInterceptor(s.macaroonSvc)

	// gRPC Operator server
	operatorAddr := s.opts.OperatorAddress
	grpcOperatorServer := grpc.NewServer(
		unaryInterceptor,
		streamInterceptor,
	)
	walletUnlockerHandler := grpchandler.NewWalletUnlockerHandler(
		s.opts.WalletUnlockerSvc,
	)
	pbwalletunlocker.RegisterWalletUnlockerServer(
		grpcOperatorServer, walletUnlockerHandler,
	)

	if !withUnlockerOnly {
		walletHandler := grpchandler.NewWalletHandler(s.opts.WalletSvc)
		operatorHandler := grpchandler.NewOperatorHandler(s.opts.OperatorSvc)
		pboperator.RegisterOperatorServer(grpcOperatorServer, operatorHandler)
		pbwallet.RegisterWalletServer(grpcOperatorServer, walletHandler)
	}

	// http Operator server for grpc-web
	httpOperatorServer := newGRPCWrappedServer(operatorAddr, grpcOperatorServer)

	// Serve grpc and grpc-web multiplexed on the same port
	muxOperator, err := serveMux(
		operatorAddr, s.opts.operatorTLSKey(), s.opts.operatorTLSCert(),
		grpcOperatorServer, httpOperatorServer,
	)
	if err != nil {
		return nil, err
	}

	var muxTrade cmux.CMux
	var grpcTradeServer *grpc.Server
	var httpTradeServer *http.Server
	if !withUnlockerOnly {
		// gRPC Trade server
		tradeAddr := s.opts.TradeAddress
		tradeTLSKey := s.opts.TradeTLSKey
		tradeTLSCert := s.opts.TradeTLSCert
		grpcTradeServer = grpc.NewServer(
			unaryInterceptor,
			streamInterceptor,
		)
		tradeHandler := grpchandler.NewTraderHandler(s.opts.TradeSvc)
		pbtrade.RegisterTradeServer(grpcTradeServer, tradeHandler)
		httpTradeServer = newGRPCWrappedServer(tradeAddr, grpcTradeServer)
		muxTrade, err = serveMux(
			tradeAddr, tradeTLSKey, tradeTLSCert,
			grpcTradeServer, httpTradeServer,
		)
		if err != nil {
			return nil, err
		}
	}

	return &services{
		grpcOperator: grpcOperatorServer,
		grpcTrade:    grpcTradeServer,
		httpOperator: httpOperatorServer,
		httpTrade:    httpTradeServer,
		muxOperator:  muxOperator,
		muxTrade:     muxTrade,
	}, nil
}

func (s *service) stop(stopMacaroonSvc bool) {
	if s.withMacaroons() && stopMacaroonSvc {
		s.macaroonSvc.Close()
		log.Debug("stopped macaroon service")
	}

	log.Debug("stop grpc-web Operator server")
	s.httpOperatorServer.Shutdown(context.Background())

	log.Debug("stop grpc Operator server")
	s.grpcOperatorServer.GracefulStop()

	log.Debug("stop mux Operator")
	s.muxOperator.Close()

	if s.grpcTradeServer != nil {
		log.Debug("stop grpc-web Trade server")
		s.httpTradeServer.Shutdown(context.Background())

		log.Debug("stop grpc Trade server")
		s.grpcTradeServer.GracefulStop()

		log.Debug("stop mux Trade")
		s.muxTrade.Close()
	}
}

func (s *service) startListeningToPassphraseChan() {
	for msg := range s.passphraseChan {
		if s.withMacaroons() {
			switch msg.Method {
			case application.UnlockWallet:
				pwd := []byte(msg.CurrentPwd)
				if err := s.macaroonSvc.CreateUnlock(&pwd); err != nil {
					if err != macaroons.ErrAlreadyUnlocked {
						log.WithError(err).Warn(
							"an error occured while unlocking macaroon service",
						)
					}
				}
				ctx := context.Background()
				if err := genMacaroons(
					ctx, s.macaroonSvc, s.opts.macaroonsDatadir(),
				); err != nil {
					log.WithError(err).Warn("an error occured while creating macaroons")
				}
				break
			case application.ChangePassphrase:
				currentPwd := []byte(msg.CurrentPwd)
				newPwd := []byte(msg.NewPwd)
				if err := s.macaroonSvc.ChangePassword(currentPwd, newPwd); err != nil {
					log.WithError(err).Warn(
						"an error occured while changing password of macaroon service",
					)
				}
			default:
				pwd := []byte(msg.CurrentPwd)
				if err := s.macaroonSvc.CreateUnlock(&pwd); err != nil {
					log.WithError(err).Warn(
						"an error occured while creating macaroon service",
					)
				}
				ctx := context.Background()
				if err := genMacaroons(
					ctx, s.macaroonSvc, s.opts.macaroonsDatadir(),
				); err != nil {
					log.WithError(err).Warn("an error occured while creating macaroons")
				}
				break
			}
		}
	}
}

func (s *service) startListeningToReadyChan() {
	isReady := <-s.readyChan

	dontStopMacaroonSvc := false
	s.stop(dontStopMacaroonSvc)

	if !isReady {
		panic("failed to initialize wallet")
	}

	withoutUnlockerOnly := false
	services, err := s.start(withoutUnlockerOnly)
	if err != nil {
		log.WithError(err).Warn(
			"an error occured while enabling operator and trade interfaces. Shutting down",
		)
		panic(nil)
	}

	log.Infof("operator interface is listening on %s", s.opts.OperatorAddress)
	log.Infof("trade interface is listening on %s", s.opts.TradeAddress)

	s.grpcOperatorServer = services.grpcOperator
	s.grpcTradeServer = services.grpcTrade
	s.httpOperatorServer = services.httpOperator
	s.httpTradeServer = services.httpTrade
	s.muxOperator = services.muxOperator
	s.muxTrade = services.muxTrade
}
