package grpcinterface

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"path/filepath"

	"github.com/tdex-network/reflection"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	httpinterface "github.com/tdex-network/tdex-daemon/internal/interfaces/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	log "github.com/sirupsen/logrus"
	daemonv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v1"
	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v1"
	"github.com/tdex-network/tdex-daemon/internal/interfaces"
	grpchandler "github.com/tdex-network/tdex-daemon/internal/interfaces/grpc/handler"
	"github.com/tdex-network/tdex-daemon/internal/interfaces/grpc/interceptor"
	"github.com/tdex-network/tdex-daemon/internal/interfaces/grpc/permissions"
	tdexold "github.com/tdex-network/tdex-protobuf/generated/go/trade"

	"github.com/soheilhy/cmux"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/pkg/macaroons"
	"google.golang.org/grpc"
)

type ServiceOptsOnePort struct {
	NoMacaroons bool

	Datadir                  string
	DBLocation               string
	MacaroonsLocation        string
	WalletUnlockPasswordFile string

	Address string

	WalletUnlockerSvc application.WalletUnlockerService
	WalletSvc         application.WalletService
	OperatorSvc       application.OperatorService
	TradeSvc          application.TradeService

	RepoManager ports.RepoManager

	TLSLocation  string
	NoTls        bool
	ExtraIPs     []string
	ExtraDomains []string
	ConnectAddr  string
	ConnectProto string
}

type serviceOnePort struct {
	opts        ServiceOptsOnePort
	macaroonSvc *macaroons.Service

	grpcServer  *grpc.Server
	http1Server *http.Server
	http2Server *http.Server
	mux         cmux.CMux

	walletPassword string

	passphraseChan chan application.PassphraseMsg
	readyChan      chan bool
}

func NewServiceOnePort(opts ServiceOptsOnePort) (interfaces.Service, error) {
	if err := opts.validate(); err != nil {
		return nil, fmt.Errorf("invalid opts: %s", err)
	}

	var macaroonSvc *macaroons.Service
	if !opts.NoMacaroons {
		macaroonSvc, _ = macaroons.NewService(
			opts.dbDatadir(), Location, DBFile, false, macaroons.IPLockChecker,
		)
		if err := permissions.Validate(); err != nil {
			return nil, err
		}
	}

	if !opts.NoTls {
		if err := generateOperatorTLSKeyCert(
			opts.tlsDatadir(), opts.ExtraIPs, opts.ExtraDomains,
		); err != nil {
			return nil, err
		}
	}

	var walletPassword string
	if opts.WalletUnlockPasswordFile != "" {
		walletPasswordBytes, err := ioutil.ReadFile(opts.WalletUnlockPasswordFile)
		if err != nil {
			return nil, err
		}

		trimmedPass := bytes.TrimFunc(walletPasswordBytes, func(r rune) bool {
			return r == 10 || r == 32
		})

		walletPassword = string(trimmedPass)
	}

	return &serviceOnePort{
		opts:           opts,
		macaroonSvc:    macaroonSvc,
		passphraseChan: opts.WalletUnlockerSvc.PassphraseChan(),
		readyChan:      opts.WalletUnlockerSvc.ReadyChan(),
		walletPassword: walletPassword,
	}, nil
}

func (s *serviceOnePort) Start() error {
	walletUnlockerOnly := true
	services, err := s.start(walletUnlockerOnly)
	if err != nil {
		return err
	}

	log.Infof("wallet unlocker interface is listening on %s", s.opts.Address)

	go s.startListeningToPassphraseChan()
	go s.startListeningToReadyChan()

	s.grpcServer = services.grpcServer
	s.http1Server = services.http1Server
	s.http2Server = services.http2Server
	s.mux = services.mux

	return nil
}

func (s *serviceOnePort) startListeningToPassphraseChan() {
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
			}
		}
	}
}

func (s *serviceOnePort) startListeningToReadyChan() {
	isReady := <-s.readyChan

	dontStopMacaroonSvc := false
	s.stop(dontStopMacaroonSvc)

	if !isReady {
		panic("failed to initialize wallet")
	}

	if s.walletPassword != "" {
		if err := s.opts.WalletUnlockerSvc.UnlockWallet(context.Background(), s.walletPassword); err != nil {
			panic(err)
		}

		s.walletPassword = ""
		log.Infoln("wallet unlocked")
	}

	withoutUnlockerOnly := false
	services, err := s.start(withoutUnlockerOnly)
	if err != nil {
		log.WithError(err).Warn(
			"an error occured while enabling operator and trade interfaces. Shutting down",
		)
		panic(nil)
	}

	log.Infof("operator, trade interfaces is listening on %s", s.opts.Address)

	s.grpcServer = services.grpcServer
	s.http1Server = services.http1Server
	s.http2Server = services.http2Server
	s.mux = services.mux
}

func (s *serviceOnePort) stop(stopMacaroonSvc bool) {
	if s.withMacaroons() && stopMacaroonSvc {
		s.macaroonSvc.Close()
		log.Debug("stopped macaroon service")
	}

	log.Debug("stop grpc-web server")
	s.http1Server.Shutdown(context.Background())
	s.http2Server.Shutdown(context.Background())

	log.Debug("stop grpc server")
	s.grpcServer.GracefulStop()

	log.Debug("stop mux")
	s.mux.Close()
}

func (s *serviceOnePort) Stop() {
	stopMacaroonSvc := true
	s.stop(stopMacaroonSvc)
}

func (s *serviceOnePort) start(withUnlockerOnly bool) (*serviceOnePort, error) {
	unaryInterceptor := interceptor.UnaryInterceptor(s.macaroonSvc)
	streamInterceptor := interceptor.StreamInterceptor(s.macaroonSvc)

	var adminMacaroonPath string
	if s.withMacaroons() {
		adminMacaroonPath = filepath.Join(s.opts.macaroonsDatadir(), AdminMacaroonFile)
	}

	address := s.opts.Address
	grpcServer := grpc.NewServer(
		unaryInterceptor, streamInterceptor,
	)
	walletUnlockerHandler := grpchandler.NewWalletUnlockerHandler(
		s.opts.WalletUnlockerSvc, adminMacaroonPath,
	)
	daemonv1.RegisterWalletUnlockerServiceServer(
		grpcServer, walletUnlockerHandler,
	)
	reflection.Register(grpcServer)

	var grpcGateway http.Handler
	if !withUnlockerOnly {
		walletHandler := grpchandler.NewWalletHandler(s.opts.WalletSvc)
		operatorHandler := grpchandler.NewOperatorHandler(s.opts.OperatorSvc)
		daemonv1.RegisterOperatorServiceServer(grpcServer, operatorHandler)
		daemonv1.RegisterWalletServiceServer(grpcServer, walletHandler)
		tradeHandler := grpchandler.NewTradeHandler(s.opts.TradeSvc)
		tradeOldHandler := grpchandler.NewTradeOldHandler(s.opts.TradeSvc)
		tdexv1.RegisterTradeServiceServer(grpcServer, tradeHandler)
		tdexold.RegisterTradeServer(grpcServer, tradeOldHandler)
		tradeGrpcGateway, err := s.tradeGrpcGateway(context.Background(), true)
		if err != nil {
			return nil, err
		}
		transportHandler := grpchandler.NewTransportHandler()
		tdexv1.RegisterTransportServiceServer(grpcServer, transportHandler)
		grpcGateway = tradeGrpcGateway
	}

	operatorTlsCert := s.opts.tlsCert()
	operatorTlsKey := s.opts.tlsKey()

	tdexConnectSvc, err := httpinterface.NewTdexConnectService(
		s.opts.RepoManager,
		s.opts.WalletUnlockerSvc,
		adminMacaroonPath,
		operatorTlsCert,
		s.opts.ConnectAddr,
		s.opts.ConnectProto,
	)
	if err != nil {
		return nil, err
	}

	http1Server, http2Server := newGRPCWrappedServer(
		address,
		grpcServer,
		grpcGateway,
		map[string]func(w http.ResponseWriter, req *http.Request){
			"/":             tdexConnectSvc.RootHandler,
			"/tdexdconnect": tdexConnectSvc.AuthHandler,
		},
	)
	mux, err := serveMux(
		s.opts.Address, operatorTlsKey, operatorTlsCert,
		grpcServer, http1Server, http2Server,
	)
	if err != nil {
		return nil, err
	}

	return &serviceOnePort{
		opts:           s.opts,
		macaroonSvc:    s.macaroonSvc,
		grpcServer:     grpcServer,
		http1Server:    http1Server,
		http2Server:    http2Server,
		mux:            mux,
		walletPassword: s.walletPassword,
		passphraseChan: s.passphraseChan,
		readyChan:      s.readyChan,
	}, nil
}

func (s *serviceOnePort) withMacaroons() bool {
	return s.macaroonSvc != nil
}

func (o ServiceOptsOnePort) macaroonsDatadir() string {
	return filepath.Join(o.Datadir, o.MacaroonsLocation)
}

func (s *serviceOnePort) tradeGrpcGateway(
	ctx context.Context, insecureConn bool,
) (http.Handler, error) {
	creds := make([]grpc.DialOption, 0)
	if insecureConn {
		creds = append(creds, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		// #nosec
		creds = append(creds, grpc.WithTransportCredentials(
			/* #nosec */
			credentials.NewTLS(&tls.Config{
				InsecureSkipVerify: true,
			}),
		))
	}

	conn, err := grpc.Dial(s.opts.Address, creds...)
	if err != nil {
		return nil, err
	}

	grpcGatewayMux := runtime.NewServeMux()
	if err := tdexv1.RegisterTradeServiceHandler(ctx, grpcGatewayMux, conn); err != nil {
		return nil, err
	}
	if err := tdexv1.RegisterTransportServiceHandler(ctx, grpcGatewayMux, conn); err != nil {
		return nil, err
	}

	grpcGatewayHandler := http.Handler(grpcGatewayMux)

	return grpcGatewayHandler, nil
}

func (o ServiceOptsOnePort) validate() error {
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
	}

	if !o.NoTls {
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

		if len(o.ExtraIPs) > 0 {
			for _, ip := range o.ExtraIPs {
				if net.ParseIP(ip) == nil {
					return fmt.Errorf("invalid operator extra ip %s", ip)
				}
			}
		}
	}

	if ok := isValidAddress(o.Address); !ok {
		return fmt.Errorf("address is not valid: %s", o.Address)
	}

	if o.WalletUnlockPasswordFile != "" {
		if !pathExists(o.WalletUnlockPasswordFile) {
			return fmt.Errorf("wallet unlock password file not found")
		}
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

func (o ServiceOptsOnePort) dbDatadir() string {
	return filepath.Join(o.Datadir, o.DBLocation)
}

func (o ServiceOptsOnePort) tlsDatadir() string {
	return filepath.Join(o.Datadir, o.TLSLocation)
}

func (o ServiceOptsOnePort) tlsKey() string {
	if o.NoTls {
		return ""
	}
	return filepath.Join(o.tlsDatadir(), OperatorTLSKeyFile)
}

func (o ServiceOptsOnePort) tlsCert() string {
	if o.NoTls {
		return ""
	}
	return filepath.Join(o.tlsDatadir(), OperatorTLSCertFile)
}
