package main

import (
	"bytes"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/tdex-network/tdex-daemon/pkg/tdexdconnect"
	"github.com/urfave/cli/v2"
	"github.com/vulpemventures/go-elements/network"
)

const (
	noMacaroonsKey   = "no_macaroons"
	macaroonsPathKey = "macaroons_path"
	tlsCertPathKey   = "tls_cert_path"
	noTlsKey         = "no_tls"
)

var (
	daemonDatadir = btcutil.AppDataDir("tdex-daemon", false)

	defaultNetwork         = network.Liquid.Name
	defaultRPCServer       = "localhost:9000"
	defaultNoMacaroonsAuth = false
	defaultTLSCertPath     = filepath.Join(daemonDatadir, "tls", "cert.pem")
	defaultMacaroonsPath   = filepath.Join(daemonDatadir, "macaroons", "admin.macaroon")
	defaultNoTLS           = false

	noMacaroonsFlagName   = strings.Replace(noMacaroonsKey, "_", "-", -1)
	macaroonsPathFlagName = strings.Replace(macaroonsPathKey, "_", "-", -1)
	tlsCertPathFlagName   = strings.Replace(tlsCertPathKey, "_", "-", -1)
	noTlsFlagName         = strings.Replace(noTlsKey, "_", "-", -1)

	networkFlag = cli.StringFlag{
		Name:  "network, n",
		Usage: "the network tdexd is running on: liquid or regtest",
		Value: defaultNetwork,
	}

	rpcFlag = cli.StringFlag{
		Name:  "rpcserver",
		Usage: "tdexd daemon address host:port",
		Value: defaultRPCServer,
	}

	tlsCertFlag = cli.StringFlag{
		Name:  tlsCertPathFlagName,
		Usage: "the path of the TLS certificate file to use",
		Value: defaultTLSCertPath,
	}

	noTlsFlag = cli.BoolFlag{
		Name:  noTlsFlagName,
		Usage: "used to disable operator TLS certificate",
		Value: defaultNoTLS,
	}

	noMacaroonsFlag = cli.BoolFlag{
		Name:  noMacaroonsFlagName,
		Usage: "used to start the daemon without macaroon auth",
		Value: defaultNoMacaroonsAuth,
	}

	macaroonsFlag = cli.StringFlag{
		Name:  macaroonsPathFlagName,
		Usage: "the path of the macaroons file to use",
		Value: defaultMacaroonsPath,
	}
)

var cliConfig = cli.Command{
	Name:   "config",
	Usage:  "Print local configuration of the tdex CLI",
	Action: configAction,
	Subcommands: []*cli.Command{
		{
			Name:   "set",
			Usage:  "set a <key> <value> in the local state",
			Action: configSetAction,
		},
		{
			Name:   "init",
			Usage:  "initialize the local state with flags",
			Action: configInitAction,
			Flags: []cli.Flag{
				&networkFlag,
				&rpcFlag,
				&tlsCertFlag,
				&noMacaroonsFlag,
				&macaroonsFlag,
				&noTlsFlag,
			},
		},
		{
			Name:   "connect",
			Usage:  "configure the CLI with a tdexdconnect URL",
			Action: configConnectAction,
		},
	},
}

func configAction(ctx *cli.Context) error {
	state, err := getState()
	if err != nil {
		return err
	}

	for key, value := range state {
		fmt.Println(key + ": " + value)
	}

	return nil
}

func configInitAction(c *cli.Context) error {
	return setState(map[string]string{
		"network":        c.String("network"),
		"rpcserver":      c.String("rpcserver"),
		noMacaroonsKey:   c.String(noMacaroonsFlagName),
		noTlsKey:         c.String(noTlsFlagName),
		tlsCertPathKey:   cleanAndExpandPath(c.String(tlsCertPathFlagName)),
		macaroonsPathKey: cleanAndExpandPath(c.String(macaroonsPathFlagName)),
	})
}

func configSetAction(c *cli.Context) error {
	if c.NArg() < 2 {
		return fmt.Errorf("key and value are missing")
	}

	key := c.Args().Get(0)
	value := c.Args().Get(1)

	if value == "" {
		return fmt.Errorf("value must not be an empty string")
	}

	if err := setState(map[string]string{key: value}); err != nil {
		return err
	}

	fmt.Printf("%s %s has been set\n", key, value)

	return nil
}

func configConnectAction(c *cli.Context) (err error) {
	connectUrl := c.Args().Get(0)
	if connectUrl == "" {
		err = fmt.Errorf("tdexdconnect URI is missing")
		return
	}

	rpcServerAddr, proto, certificate, macaroon, err :=
		tdexdconnect.Decode(connectUrl)
	if err != nil {
		return
	}

	var tlsCertPath string
	if len(certificate) > 0 {
		tlsCertPath = filepath.Join(tdexDataDir, "cert.pem")
		buf := &bytes.Buffer{}
		if err = pem.Encode(
			buf, &pem.Block{Type: "CERTIFICATE", Bytes: certificate},
		); err != nil {
			err = fmt.Errorf("failed to encode certificate: %v", err)
			return
		}

		if err = os.WriteFile(tlsCertPath, buf.Bytes(), 0644); err != nil {
			err = fmt.Errorf("failed to write certificate to file: %s", err)
			return
		}
	}
	defer func() {
		if err != nil && tlsCertPath != "" {
			os.Remove(tlsCertPath)
		}
	}()

	var macaroonsPath string
	if len(macaroon) > 0 {
		macaroonsPath = filepath.Join(tdexDataDir, "admin.macaroon")
		if err = os.WriteFile(macaroonsPath, macaroon, 0644); err != nil {
			err = fmt.Errorf("failed to write macaroon to file: %s", err)
			return
		}
	}
	defer func() {
		if err != nil && macaroonsPath != "" {
			os.Remove(macaroonsPath)
		}
	}()

	noMacaroons := strconv.FormatBool(len(certificate) == 0 && len(macaroon) == 0)

	if err = setState(map[string]string{
		"proto":          proto,
		"no_tls":         strconv.FormatBool(proto == "http"),
		"rpcserver":      rpcServerAddr,
		"no_macaroons":   noMacaroons,
		"tls_cert_path":  tlsCertPath,
		"macaroons_path": macaroonsPath,
	}); err != nil {
		return
	}

	fmt.Println()
	fmt.Println("CLI configured via tdexdconnect URL.")
	fmt.Println("Check configuration with `tdex config`")
	return nil
}

func getMarketFromState() (string, string, error) {
	state, err := getState()
	if err != nil {
		return "", "", err
	}
	baseAsset, ok := state["base_asset"]
	if !ok {
		return "", "", fmt.Errorf("set base asset with `config set base_asset`")
	}
	quoteAsset, ok := state["quote_asset"]
	if !ok {
		return "", "", fmt.Errorf("set quote asset with `config set quote_asset`")
	}

	return baseAsset, quoteAsset, nil
}
