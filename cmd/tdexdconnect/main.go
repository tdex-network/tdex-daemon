package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/mdp/qrterminal/v3"
	"github.com/skip2/go-qrcode"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tdex-network/tdex-daemon/pkg/tdexdconnect"
)

const (
	rpcServerKey   = "rpcserver"
	tlsCertKey     = "tls_cert_path"
	macaroonsKey   = "macaroons_path"
	outputKey      = "out"
	noTlsKey       = "no_tls"
	noMacaroonsKey = "no_macaroons"
	protoKey       = "proto"

	qrFilename = "tdexdconnect-qr.png"
)

var (
	defaultProto         = "https"
	defaultRPCServer     = "localhost:9000"
	defaultDatadir       = btcutil.AppDataDir("tdex-daemon", false)
	defaultTLSCertPath   = filepath.Join(defaultDatadir, "tls", "cert.pem")
	defaultMacaroonsPath = filepath.Join(defaultDatadir, "macaroons", "admin.macaroon")
	defaultOutput        = "qr"
	defaultNoTls         = false
	defaultNoMacaroon    = false

	supportedOutputs = map[string]struct{}{
		"qr":    {},
		"url":   {},
		"image": {},
	}

	rxDNSName = regexp.MustCompile(
		`^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])(\.([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9]))*$`,
	)
	rxIP = regexp.MustCompile(
		`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`,
	)

	protoFlag = pflag.String(
		protoKey, defaultProto, "http or https protocol",
	)
	rpcServerFlag = pflag.String(
		rpcServerKey, defaultRPCServer, "the rpc address and port of tdexd",
	)
	tlsCertFlag = pflag.String(
		tlsCertKey, defaultTLSCertPath, "the path of the TLS certificate file",
	)
	macaroonsFlag = pflag.String(
		macaroonsKey, defaultMacaroonsPath, "the path of the macaroon file",
	)
	outputFlag = pflag.String(
		outputKey, defaultOutput,
		"whether 'qr' to display QRCode, 'url' to display string URL, "+
			"or 'image' to save QRCode to file",
	)
	noTlsFlag = pflag.Bool(
		noTlsKey, defaultNoTls, "to be used in case the daemon has TLS disabled",
	)
	noMacaroonFlag = pflag.Bool(
		noMacaroonsKey, defaultNoMacaroon, "to be used in case the daemon has macaroon auth disabled",
	)
)

func validateFlags(
	rpcServerAddr, tlsCertPath, macaroonsPath, out string, noTls, noMacaroon bool,
) error {
	// Validate rpc server address
	if rpcServerAddr == "" {
		return fmt.Errorf("%s must not be null", rpcServerKey)
	}
	if rpcServerAddr != defaultRPCServer {
		parts := strings.Split(rpcServerAddr, ":")
		if len(parts) != 2 {
			return fmt.Errorf("%s must be a valid address in the form host:port", rpcServerKey)
		}
		if parts[0] != "" && parts[0] != "localhost" {
			if !validateIp(parts[0]) && !validateDomain(parts[0]) {
				return fmt.Errorf("%s host is invalid", rpcServerKey)
			}
		}
		port, err := strconv.Atoi(parts[1])
		if err != nil || port <= 1024 {
			return fmt.Errorf("%s port is invalid", rpcServerKey)
		}
	}

	if !noTls {
		// Make sure the TLS cert filepath is defined and that the file exists.
		if tlsCertPath == "" {
			return fmt.Errorf("%s must not be null", tlsCertKey)
		}
		if _, err := os.Stat(tlsCertPath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("TLS certificate not found at path %s", macaroonsPath)
			}
			return fmt.Errorf("%s is not a valid path", tlsCertPath)
		}
	}

	if !noMacaroon {
		if macaroonsPath == "" {
			return fmt.Errorf("%s must not be null", macaroonsPath)
		}
		// In case the macaroon path is customized, make sure the file exists.
		if macaroonsPath != defaultMacaroonsPath {
			if _, err := os.Stat(macaroonsPath); err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("macaroon not found at path %s", macaroonsPath)
				}
				return fmt.Errorf("%s is not a valid path", macaroonsPath)
			}
		}
	}

	if _, ok := supportedOutputs[out]; !ok {
		outs := make([]string, 0, len(supportedOutputs))
		for o := range supportedOutputs {
			outs = append(outs, o)
		}
		return fmt.Errorf(
			"output format is invalid. It must be one of %s", strings.Join(outs, " | "),
		)
	}

	return nil
}

func validateIp(str string) bool {
	return net.ParseIP(str) != nil && rxIP.MatchString(str)
}

func validateDomain(str string) bool {
	if str == "" || len(strings.Replace(str, ".", "", -1)) > 255 {
		// Domain must not be longer than 255 chars.
		return false
	}
	return rxDNSName.MatchString(str)
}

func main() {
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	proto := viper.GetString(protoKey)
	rpcServerAddr := viper.GetString(rpcServerKey)
	tlsCertPath := viper.GetString(tlsCertKey)
	macaroonsPath := viper.GetString(macaroonsKey)
	out := strings.ToLower(viper.GetString(outputKey))
	noTls := viper.GetBool(noTlsKey)
	noMacaroons := viper.GetBool(noMacaroonsKey)

	if err := validateFlags(
		rpcServerAddr, tlsCertPath, macaroonsPath, out, noTls, noMacaroons,
	); err != nil {
		log.Fatal(err)
	}

	var macBytes, certBytes []byte
	if !noTls {
		var err error
		certBytes, err = ioutil.ReadFile(tlsCertPath)
		if err != nil {
			log.Fatalf("failed to read TLS certificate file: %s", err)
		}
	}

	if !noMacaroons {
		macBytes, _ = ioutil.ReadFile(macaroonsPath)
	}

	connectUrl, err := tdexdconnect.EncodeToString(
		rpcServerAddr, proto, certBytes, macBytes,
	)
	if err != nil {
		log.Fatalf("failed to encode url string: %s", err)
	}

	if out == "qr" {
		qrterminal.Generate(connectUrl, qrterminal.L, os.Stdout)
		return
	}

	if out == "url" {
		fmt.Println(connectUrl)
		return
	}

	qrcode.WriteFile(connectUrl, qrcode.Medium, 512, qrFilename)
	fmt.Println("QRCode written to file", qrFilename)
}
