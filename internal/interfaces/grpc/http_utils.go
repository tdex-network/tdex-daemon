package grpcinterface

import (
	"bytes"
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
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	log "github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
)

func isValidAddress(addr string) bool {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return false
	}
	if parts[0] != "" {
		if ip := net.ParseIP(parts[0]); ip == nil {
			return false
		}
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}
	if port <= 1024 {
		return false
	}
	return true
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

func serveMux(
	address, tlsKey, tlsCert string,
	grpcServer *grpc.Server, http1Server, http2Server *http.Server,
) (cmux.CMux, error) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	if tlsKey != "" {
		tlsConfig, err := getTlsConfig(tlsKey, tlsCert)
		if err != nil {
			return nil, err
		}

		lis = tls.NewListener(lis, tlsConfig)
	}

	mux := cmux.New(lis)
	grpcL := mux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	http2L := mux.Match(cmux.HTTP2())
	http1L := mux.Match(cmux.HTTP1())

	go grpcServer.Serve(grpcL)
	go http2Server.Serve(http2L)
	go http1Server.Serve(http1L)
	go mux.Serve()

	return mux, nil
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

/*
	gRPC web wrapper
*/

func newGRPCWrappedServer(
	addr string,
	grpcServer *grpc.Server,
	grpcGateway http.Handler,
	httpHandlers map[string]func(w http.ResponseWriter, req *http.Request),
) (httpSrv *http.Server, http2Srv *http.Server) {
	grpcWebServer := grpcweb.WrapServer(
		grpcServer,
		grpcweb.WithCorsForRegisteredEndpointsOnly(false),
		grpcweb.WithOriginFunc(func(origin string) bool { return true }),
	)

	handler := func(w http.ResponseWriter, req *http.Request) {
		if isOptionRequest(req) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			return
		}
		if isGetRequest(req) {
			if handler, ok := httpHandlers[req.URL.Path]; ok {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Headers", "*")
				w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
				handler(w, req)
				return
			}
		}

		if isValidRequest(req) {
			grpcWebServer.ServeHTTP(w, req)
			return
		}

		if grpcGateway != nil {
			if isHttpRequest(req) {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Headers", "*")
				w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
				grpcGateway.ServeHTTP(w, req)
				return
			}
		}

		msg := "received a request that could not be matched to grpc or grpc-web"
		log.Warn(msg)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(msg))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	httpSrv = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	http2Srv = &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(http.HandlerFunc(handler), &http2.Server{}),
	}

	return
}

func isGetRequest(req *http.Request) bool {
	return req.Method == http.MethodGet
}

func isOptionRequest(req *http.Request) bool {
	return req.Method == http.MethodOptions
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

func isHttpRequest(req *http.Request) bool {
	return strings.ToLower(req.Method) == "get" ||
		strings.Contains(req.Header.Get("Content-Type"), "application/json")
}

func getTlsConfig(tlsKey, tlsCert string) (*tls.Config, error) {
	if tlsKey == "" || tlsCert == "" {
		return nil, errors.New("tls_key and tls_cert both needs to be provided")
	}

	certificate, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		NextProtos:   []string{"http/1.1", http2.NextProtoTLS, "h2-14"}, // h2-14 is just for compatibility. will be eventually removed.
		Certificates: []tls.Certificate{certificate},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
	}
	config.Rand = rand.Reader

	return config, nil
}
