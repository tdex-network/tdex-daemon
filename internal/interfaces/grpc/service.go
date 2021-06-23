package grpcinterface

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	log "github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
	"github.com/tdex-network/tdex-daemon/internal/core/application"
	interfaces "github.com/tdex-network/tdex-daemon/internal/interfaces"
	grpchandler "github.com/tdex-network/tdex-daemon/internal/interfaces/grpc/handler"
	"github.com/tdex-network/tdex-daemon/internal/interfaces/grpc/interceptor"
	"github.com/tdex-network/tdex-daemon/internal/interfaces/grpc/permissions"
	"github.com/tdex-network/tdex-daemon/pkg/macaroons"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"gopkg.in/macaroon-bakery.v2/bakery"

	pboperator "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/operator"
	pbwallet "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/wallet"
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
	opts           ServiceOpts
	macaroonSvc    *macaroons.Service
	operatorServer *grpc.Server
	tradeServer    *grpc.Server
	passphraseChan chan application.PassphraseMsg
}

type ServiceOpts struct {
	NoMacaroons bool

	Datadir              string
	DBLocation           string
	TLSLocation          string
	MacaroonsLocation    string
	OperatorExtraIPs     []string
	OperatorExtraDomains []string

	WalletSvc   application.WalletService
	OperatorSvc application.OperatorService
	TradeSvc    application.TradeService
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
		passphraseChan: opts.WalletSvc.PassphraseChan(),
	}, nil
}

func (s *service) Start(
	operatorAddress, tradeAddress,
	tradeTLSKey, tradeTLSCert string,
) error {
	unaryInterceptor := interceptor.UnaryInterceptor(s.macaroonSvc)
	streamInterceptor := interceptor.StreamInterceptor(s.macaroonSvc)

	walletHandler := grpchandler.NewWalletHandler(s.opts.WalletSvc)
	operatorHandler := grpchandler.NewOperatorHandler(s.opts.OperatorSvc)
	tradeHandler := grpchandler.NewTraderHandler(s.opts.TradeSvc)

	// Server
	operatorServer := grpc.NewServer(
		unaryInterceptor,
		streamInterceptor,
	)
	tradeServer := grpc.NewServer(
		unaryInterceptor,
		streamInterceptor,
	)
	// Register proto implementations on Trade interface
	pbtrade.RegisterTradeServer(tradeServer, tradeHandler)
	// Register proto implementations on Operator interface
	pboperator.RegisterOperatorServer(operatorServer, operatorHandler)
	pbwallet.RegisterWalletServer(operatorServer, walletHandler)

	// Serve grpc and grpc-web multiplexed on the same port
	if err := serveMux(
		operatorAddress, s.operatorTLSKey(), s.operatorTLSCert(), operatorServer,
	); err != nil {
		return err
	}
	if err := serveMux(
		tradeAddress, tradeTLSKey, tradeTLSCert, tradeServer,
	); err != nil {
		return err
	}

	go s.startListeningToPassphraseChan()

	s.operatorServer = operatorServer
	s.tradeServer = tradeServer

	return nil
}

func (s *service) Stop() {
	if s.withMacaroons() {
		s.macaroonSvc.Close()
		log.Debug("stopped macaroon service")
	}

	s.operatorServer.GracefulStop()
	log.Debug("disabled operator interface")

	s.tradeServer.GracefulStop()
	log.Debug("disabled trader interface")
}

func (s *service) operatorTLSKey() string {
	if s.opts.NoMacaroons {
		return ""
	}
	return filepath.Join(s.opts.tlsDatadir(), OperatorTLSKeyFile)
}

func (s *service) operatorTLSCert() string {
	if s.opts.NoMacaroons {
		return ""
	}
	return filepath.Join(s.opts.tlsDatadir(), OperatorTLSCertFile)
}

func (s *service) withMacaroons() bool {
	return s.macaroonSvc != nil
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

func generateOperatorTLSKeyCert(
	datadir string, extraIPs, extraDomains []string,
) error {
	if err := makeDirectoryIfNotExists(datadir); err != nil {
		return err
	}
	keyPath := filepath.Join(datadir, OperatorTLSKeyFile)
	certPath := filepath.Join(datadir, OperatorTLSCertFile)

	// if key and cert files already exist nothing to do here.
	if pathExists(keyPath) && pathExists(certPath) {
		return nil
	}

	organization := "tdex"
	now := time.Now()
	validUntil := now.AddDate(1, 0, 0)

	// Generate a serial number that's below the serialNumberLimit.
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %s", err)
	}

	// Collect the host's IP addresses, including loopback, in a slice.
	ipAddresses := []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}

	if len(extraIPs) > 0 {
		for _, ip := range extraIPs {
			ipAddresses = append(ipAddresses, net.ParseIP(ip))
		}
	}

	// addIP appends an IP address only if it isn't already in the slice.
	addIP := func(ipAddr net.IP) {
		for _, ip := range ipAddresses {
			if bytes.Equal(ip, ipAddr) {
				return
			}
		}
		ipAddresses = append(ipAddresses, ipAddr)
	}

	// Add all the interface IPs that aren't already in the slice.
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}
	for _, a := range addrs {
		ipAddr, _, err := net.ParseCIDR(a.String())
		if err == nil {
			addIP(ipAddr)
		}
	}

	host, err := os.Hostname()
	if err != nil {
		return err
	}

	dnsNames := []string{host}
	if host != "localhost" {
		dnsNames = append(dnsNames, "localhost")
	}

	if len(extraDomains) > 0 {
		for _, domain := range extraDomains {
			dnsNames = append(dnsNames, domain)
		}
	}

	dnsNames = append(dnsNames, "unix", "unixpacket")

	priv, err := createOrLoadTLSKey(keyPath)
	if err != nil {
		return err
	}

	keybytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return err
	}

	// construct certificate template
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{organization},
			CommonName:   host,
		},
		NotBefore: now.Add(-time.Hour * 24),
		NotAfter:  validUntil,

		KeyUsage: x509.KeyUsageKeyEncipherment |
			x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		IsCA:                  true,
		BasicConstraintsValid: true,

		DNSNames:    dnsNames,
		IPAddresses: ipAddresses,
	}

	derBytes, err := x509.CreateCertificate(
		rand.Reader, &template, &template, &priv.PublicKey, priv,
	)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %v", err)
	}

	certBuf := &bytes.Buffer{}
	if err := pem.Encode(
		certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes},
	); err != nil {
		return fmt.Errorf("failed to encode certificate: %v", err)
	}

	keyBuf := &bytes.Buffer{}
	if err := pem.Encode(
		keyBuf, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keybytes},
	); err != nil {
		return fmt.Errorf("failed to encode private key: %v", err)
	}

	if err := ioutil.WriteFile(certPath, certBuf.Bytes(), 0644); err != nil {
		return err
	}
	if err := ioutil.WriteFile(keyPath, keyBuf.Bytes(), 0600); err != nil {
		os.Remove(certPath)
		return err
	}

	return nil
}

func serveMux(address, tlsKey, tlsCert string, grpcServer *grpc.Server) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	if tlsKey != "" {
		certificate, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
		if err != nil {
			return err
		}

		const requiredCipher = tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
		config := &tls.Config{
			CipherSuites: []uint16{requiredCipher},
			NextProtos:   []string{"http/1.1", http2.NextProtoTLS, "h2-14"}, // h2-14 is just for compatibility. will be eventually removed.
			Certificates: []tls.Certificate{certificate},
		}
		config.Rand = rand.Reader

		lis = tls.NewListener(lis, config)
	}

	mux := cmux.New(lis)
	grpcL := mux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldPrefixSendSettings("content-type", "application/grpc"))
	httpL := mux.Match(cmux.HTTP1Fast())

	grpcWebServer := grpcweb.WrapServer(
		grpcServer,
		grpcweb.WithCorsForRegisteredEndpointsOnly(false),
		grpcweb.WithOriginFunc(func(origin string) bool { return true }),
	)

	go grpcServer.Serve(grpcL)
	go http.Serve(httpL, http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if isValidRequest(req) {
			grpcWebServer.ServeHTTP(resp, req)
		}
	}))

	go mux.Serve()
	return nil
}

func isValidRequest(req *http.Request) bool {
	return isValidGrpcWebOptionRequest(req) || isValidGrpcWebRequest(req)
}

func isValidGrpcWebRequest(req *http.Request) bool {
	return req.Method == http.MethodPost && isValidGrpcContentTypeHeader(req.Header.Get("content-type"))
}

func isValidGrpcContentTypeHeader(contentType string) bool {
	return strings.HasPrefix(contentType, "application/grpc-web-text") ||
		strings.HasPrefix(contentType, "application/grpc-web")
}

func isValidGrpcWebOptionRequest(req *http.Request) bool {
	accessControlHeader := req.Header.Get("Access-Control-Request-Headers")
	return req.Method == http.MethodOptions &&
		strings.Contains(accessControlHeader, "x-grpc-web") &&
		strings.Contains(accessControlHeader, "content-type")
}

func createOrLoadTLSKey(keyPath string) (*ecdsa.PrivateKey, error) {
	if !pathExists(keyPath) {
		return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	}

	b, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	key, err := privateKeyFromPEM(b)
	if err != nil {
		return nil, err
	}
	return key.(*ecdsa.PrivateKey), nil
}

func privateKeyFromPEM(pemBlock []byte) (crypto.PrivateKey, error) {
	var derBlock *pem.Block
	for {
		derBlock, pemBlock = pem.Decode(pemBlock)
		if derBlock == nil {
			return nil, fmt.Errorf("tls: failed to find any PEM data in key input")
		}
		if derBlock.Type == "PRIVATE KEY" || strings.HasSuffix(derBlock.Type, " PRIVATE KEY") {
			break
		}
	}
	return parsePrivateKey(derBlock.Bytes)
}

func parsePrivateKey(der []byte) (crypto.PrivateKey, error) {
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey, ed25519.PrivateKey:
			return key, nil
		default:
			return nil, fmt.Errorf("tls: found unknown private key type in PKCS#8 wrapping")
		}
	}
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}

	return nil, fmt.Errorf("tls: failed to parse private key")
}
