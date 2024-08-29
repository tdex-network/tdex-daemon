package grpcinterface

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	grpchealth "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	reflectionv1 "github.com/tdex-network/reflection/api-spec/protobuf/gen/reflection/v1"
	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"
	tdexv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v2"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	log "github.com/sirupsen/logrus"
	"github.com/tdex-network/reflection"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"
	"github.com/tdex-network/tdex-daemon/internal/interfaces"
	grpchandler "github.com/tdex-network/tdex-daemon/internal/interfaces/grpc/handler"
	"github.com/tdex-network/tdex-daemon/internal/interfaces/grpc/interceptor"
	httpinterface "github.com/tdex-network/tdex-daemon/internal/interfaces/http"
	"github.com/tdex-network/tdex-daemon/pkg/macaroons"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type serviceOnePort struct {
	opts        ServiceOptsOnePort
	macaroonSvc *macaroons.Service
	server      *http.Server
	password    string
}

type ServiceOptsOnePort struct {
	NoMacaroons bool

	Datadir                  string
	DBLocation               string
	MacaroonsLocation        string
	WalletUnlockPasswordFile string

	Port    int
	TLSKey  string
	TLSCert string

	AppConfig *application.Config
	BuildData ports.BuildData

	ConnectAddr  string
	ConnectProto string
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

	if o.withTls() {
		tlsKeyExists := pathExists(o.TLSKey)
		tlsCertExists := pathExists(o.TLSCert)
		if !tlsKeyExists && tlsCertExists {
			return fmt.Errorf(
				"TLS key and certificate must be either existing or undefined",
			)
		}
	}

	if ok := isValidPort(o.Port); !ok {
		return fmt.Errorf("port must be in range [%d, %d]", minPort, maxPort)
	}

	if o.WalletUnlockPasswordFile != "" {
		if !pathExists(o.WalletUnlockPasswordFile) {
			return fmt.Errorf("wallet unlock password file not found")
		}
	}

	if o.AppConfig == nil {
		return fmt.Errorf("missing app config")
	}
	if err := o.AppConfig.Validate(); err != nil {
		return fmt.Errorf("invalid app config: %s", err)
	}

	if o.BuildData == nil {
		return fmt.Errorf("missing build data")
	}

	return nil
}

func (o ServiceOptsOnePort) dbDatadir() string {
	return filepath.Join(o.Datadir, o.DBLocation)
}

func (o ServiceOptsOnePort) macaroonsDatadir() string {
	return filepath.Join(o.Datadir, o.MacaroonsLocation)
}

func (o ServiceOptsOnePort) withTls() bool {
	return len(o.TLSCert) > 0
}

func (o ServiceOptsOnePort) tlsConfig() (*tls.Config, error) {
	if !o.withTls() {
		return nil, nil
	}
	return getTlsConfig(o.TLSKey, o.TLSCert)
}

func (o ServiceOptsOnePort) serverAddr() string {
	return fmt.Sprintf(":%d", o.Port)
}

func (o ServiceOptsOnePort) clientAddr() string {
	return fmt.Sprintf("localhost:%d", o.Port)
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
	}

	var password string
	if opts.WalletUnlockPasswordFile != "" {
		passwordBytes, err := os.ReadFile(opts.WalletUnlockPasswordFile)
		if err != nil {
			return nil, err
		}

		trimmedPass := bytes.TrimFunc(passwordBytes, func(r rune) bool {
			return r == 10 || r == 13 || r == 32
		})

		password = string(trimmedPass)
	}

	return &serviceOnePort{
		opts:        opts,
		macaroonSvc: macaroonSvc,
		password:    password,
	}, nil
}

func (s *serviceOnePort) Start() error {
	withWalletOnly := true
	if err := s.start(withWalletOnly); err != nil {
		return err
	}

	if s.opts.WalletUnlockPasswordFile != "" {
		if err := s.opts.AppConfig.UnlockerService().UnlockWallet(
			context.Background(), s.password,
		); err != nil {
			return fmt.Errorf("failed to auto unlock wallet: %s", err)
		}

		s.onUnlock(s.password)
	}

	return nil
}

func (s *serviceOnePort) Stop() {
	if s.password != "" {
		// nolint
		s.opts.AppConfig.UnlockerService().LockWallet(
			context.Background(), s.password,
		)
	}

	stopMacaroonSvc := true
	s.stop(stopMacaroonSvc)

	s.opts.AppConfig.FeederService().Close()
	log.Debug("closed connection with feeder")

	s.opts.AppConfig.PubSubService().Close()
	log.Debug("closed connection with pubsub")

	s.opts.AppConfig.RepoManager().Close()
	log.Debug("closed connection with database")

	s.opts.AppConfig.WalletService().Close()
	log.Debug("closed connection with ocean wallet")
}

func (s *serviceOnePort) withMacaroons() bool {
	return !s.opts.NoMacaroons
}

func (s *serviceOnePort) start(withWalletOnly bool) error {
	tlsConfig, err := s.opts.tlsConfig()
	if err != nil {
		return err
	}
	server, err := s.newServer(tlsConfig, withWalletOnly)
	if err != nil {
		return err
	}

	s.server = server

	if s.opts.withTls() {
		//nolint
		go s.server.ListenAndServeTLS("", "")
	} else {
		//nolint
		go s.server.ListenAndServe()
	}

	log.Infof("wallet interface is listening on %s", s.opts.serverAddr())
	if !withWalletOnly {
		log.Infof("operator interface is listening on %s", s.opts.serverAddr())
		log.Infof("trade interface is listening on %s", s.opts.serverAddr())
	}

	return nil
}

func (s *serviceOnePort) stop(stopMacaroonSvc bool) {
	if s.withMacaroons() && stopMacaroonSvc {
		//nolint
		s.macaroonSvc.Close()
		log.Debug("closed connection with macaroon db")
	}

	//nolint
	s.server.Shutdown(context.Background())
	log.Debug("stopped server")
}

func (s *serviceOnePort) newServer(
	tlsConfig *tls.Config, withWalletOnly bool,
) (*http.Server, error) {
	serverOpts := []grpc.ServerOption{
		interceptor.UnaryInterceptor(s.macaroonSvc),
		interceptor.StreamInterceptor(s.macaroonSvc),
	}

	creds := insecure.NewCredentials()
	if tlsConfig != nil {
		creds = credentials.NewTLS(tlsConfig)
	}
	serverOpts = append(serverOpts, grpc.Creds(creds))

	var adminMacaroonPath string
	if s.withMacaroons() {
		adminMacaroonPath = filepath.Join(
			s.opts.macaroonsDatadir(), AdminMacaroonFile,
		)
	}

	// Server grpc.
	grpcServer := grpc.NewServer(serverOpts...)

	// Creds for grpc gateway reverse proxy.
	gatewayCreds := insecure.NewCredentials()
	if s.opts.withTls() {
		gatewayCreds = credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true, // #nosec
		})
	}
	ctx := context.Background()
	gatewayOpts := grpc.WithTransportCredentials(gatewayCreds)
	conn, err := grpc.DialContext(
		ctx, s.opts.clientAddr(), gatewayOpts,
	)
	if err != nil {
		return nil, err
	}
	// Reverse proxy grpc-gateway.
	gwmux := runtime.NewServeMux(
		runtime.WithHealthzEndpoint(grpchealth.NewHealthClient(conn)),
		runtime.WithMarshalerOption("application/json+pretty", &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				Indent:    "  ",
				Multiline: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)

	// Register wallet interface.
	walletHandler := grpchandler.NewWalletHandler(
		s.opts.AppConfig.UnlockerService(), s.opts.BuildData, adminMacaroonPath,
		s.onInit, s.onUnlock, s.onLock, s.onChangePwd,
	)
	daemonv2.RegisterWalletServiceServer(
		grpcServer, walletHandler,
	)
	if err := daemonv2.RegisterWalletServiceHandler(
		ctx, gwmux, conn,
	); err != nil {
		return nil, err
	}

	// Register healthz and reflection.
	healthHandler := grpchandler.NewHealthHandler()
	grpchealth.RegisterHealthServer(grpcServer, healthHandler)
	reflection.Register(grpcServer)
	if err := reflectionv1.RegisterReflectionServiceHandler(
		ctx, gwmux, conn,
	); err != nil {
		return nil, err
	}

	// Register operator, feeder, webhook, transport and trade interfaces if
	// needed.
	if !withWalletOnly {
		operatorHandler := grpchandler.NewOperatorHandler(
			s.opts.AppConfig.OperatorService(),
		)
		daemonv2.RegisterOperatorServiceServer(grpcServer, operatorHandler)
		if err := daemonv2.RegisterOperatorServiceHandler(
			ctx, gwmux, conn,
		); err != nil {
			return nil, err
		}

		feederHandler := grpchandler.NewFeederHandler(
			s.opts.AppConfig.FeederService(),
		)
		daemonv2.RegisterFeederServiceServer(grpcServer, feederHandler)
		if err := daemonv2.RegisterFeederServiceHandler(
			ctx, gwmux, conn,
		); err != nil {
			return nil, err
		}

		webhookHandler := grpchandler.NewWebhookHandler(s.opts.AppConfig.OperatorService())
		daemonv2.RegisterWebhookServiceServer(grpcServer, webhookHandler)
		if err := daemonv2.RegisterWebhookServiceHandler(
			ctx, gwmux, conn,
		); err != nil {
			return nil, err
		}

		transportHandler := grpchandler.NewTransportHandler()
		tdexv2.RegisterTransportServiceServer(grpcServer, transportHandler)
		if err := tdexv2.RegisterTransportServiceHandler(
			ctx, gwmux, conn,
		); err != nil {
			return nil, err
		}

		tradeHandler := grpchandler.NewTradeHandler(s.opts.AppConfig.TradeService())
		tdexv2.RegisterTradeServiceServer(grpcServer, tradeHandler)
		if err := tdexv2.RegisterTradeServiceHandler(
			ctx, gwmux, conn,
		); err != nil {
			return nil, err
		}
	}
	grpcGateway := http.Handler(gwmux)

	// Wrapped server grpc-web.
	grpcWebServer := grpcweb.WrapServer(
		grpcServer,
		grpcweb.WithCorsForRegisteredEndpointsOnly(false),
		grpcweb.WithOriginFunc(func(origin string) bool { return true }),
	)

	// Custom tdexconnect endpoints.
	tdexConnectSvc, err := httpinterface.NewTdexConnectService(
		s.opts.AppConfig.WalletService().Wallet(),
		adminMacaroonPath,
		s.opts.TLSCert,
		s.opts.ConnectAddr,
		s.opts.ConnectProto,
	)
	if err != nil {
		return nil, err
	}
	httpHandlers := map[string]http.HandlerFunc{
		"/":             tdexConnectSvc.RootHandler,
		"/tdexdconnect": tdexConnectSvc.AuthHandler,
	}

	// Server mux.
	handler := router(grpcServer, grpcWebServer, grpcGateway, httpHandlers)
	mux := http.NewServeMux()
	mux.Handle("/", handler)

	httpServerHandler := http.Handler(mux)
	if !s.opts.withTls() {
		httpServerHandler = h2c.NewHandler(httpServerHandler, &http2.Server{})
	}

	return &http.Server{
		Addr:      s.opts.serverAddr(),
		Handler:   httpServerHandler,
		TLSConfig: tlsConfig,
	}, nil
}

func (s *serviceOnePort) onInit(password string) {
	s.password = password

	if !s.withMacaroons() {
		return
	}

	pwd := []byte(password)
	if err := s.macaroonSvc.CreateUnlock(&pwd); err != nil {
		log.WithError(err).Warn("failed to unlock macaroon store")
	}
	if err := genMacaroons(
		context.Background(), s.macaroonSvc, s.opts.macaroonsDatadir(),
	); err != nil {
		log.WithError(err).Warn("failed to create macaroons")
	}
}

func (s *serviceOnePort) onUnlock(password string) {
	if s.password == "" {
		s.password = password
	}

	if s.withMacaroons() {
		pwd := []byte(password)
		if err := s.macaroonSvc.CreateUnlock(&pwd); err != nil {
			if err != macaroons.ErrAlreadyUnlocked {
				log.WithError(err).Warn("failed to unlock macaroon store")
			}
		}
		if err := genMacaroons(
			context.Background(), s.macaroonSvc, s.opts.macaroonsDatadir(),
		); err != nil {
			log.WithError(err).Warn("failed to create macaroons")
		}
	}

	stopMacaroonSvc := true
	s.stop(!stopMacaroonSvc)

	withWalletOnly := true
	if err := s.start(!withWalletOnly); err != nil {
		log.WithError(err).Warn("failed to load handlers to interface after unlock")
	}
}

func (s *serviceOnePort) onLock(_ string) {
	stopMacaroonSvc := false
	s.stop(stopMacaroonSvc)
	withWalletOnly := true
	//nolint
	s.start(withWalletOnly)
}

func (s *serviceOnePort) onChangePwd(oldPassword, newPassword string) {
	if !s.withMacaroons() {
		return
	}
	oldPwd, newPwd := []byte(oldPassword), []byte(newPassword)
	if err := s.macaroonSvc.ChangePassword(oldPwd, newPwd); err != nil {
		log.WithError(err).Warn("failed to change password of macaroon store")
	}
}
