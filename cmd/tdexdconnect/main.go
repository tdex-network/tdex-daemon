package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/btcsuite/btcutil"
	"github.com/mdp/qrterminal/v3"
	"github.com/skip2/go-qrcode"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tdex-network/tdex-daemon/pkg/tdexdconnect"
)

const (
	rpcServerKey = "rpcserver"
	tlsCertKey   = "tls_cert_path"
	macaroonsKey = "macaroons_path"
	outputKey    = "out"

	qrFilename = "tdexdconnect-qr.png"
)

var (
	defaultRPCServer     = "localhost:9000"
	defaultDatadir       = btcutil.AppDataDir("tdex-daemon", false)
	defaultTLSCertPath   = filepath.Join(defaultDatadir, "tls", "cert.pem")
	defaultMacaroonsPath = filepath.Join(defaultDatadir, "macaroons", "admin.macaroon")
	defaultOutputKey     = "qr"

	supportedOutputs = map[string]struct{}{
		"qr":    {},
		"url":   {},
		"image": {},
	}

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
		outputKey, defaultOutputKey,
		"whether 'qr' to display QRCode, 'url' to display string URL, "+
			"or 'image' to save QRCode to file",
	)
)

func validateFlags(
	rpcServerAddr, tlsCertPath, macaroonsPath, out string,
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
			if net.ParseIP(parts[0]) == nil {
				return fmt.Errorf("%s host is invalid", rpcServerKey)
			}
		}
		port, err := strconv.Atoi(parts[1])
		if err != nil || port <= 1024 {
			return fmt.Errorf("%s port is invalid", rpcServerKey)
		}
	}

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

func main() {
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	rpcServerAddr := viper.GetString(rpcServerKey)
	tlsCertPath := viper.GetString(tlsCertKey)
	macaroonsPath := viper.GetString(macaroonsKey)
	out := strings.ToLower(viper.GetString(outputKey))

	if err := validateFlags(
		rpcServerAddr, tlsCertPath, macaroonsPath, out,
	); err != nil {
		log.Fatalf("invalid flag: %s", err)
	}

	macBytes, _ := ioutil.ReadFile(macaroonsPath)
	certBytes, err := ioutil.ReadFile(tlsCertPath)
	if err != nil {
		log.Fatalf("failed to read TLS certificate file: %s", err)
	}

	connectUrl, err := tdexdconnect.EncodeToString(
		rpcServerAddr, certBytes, macBytes,
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
