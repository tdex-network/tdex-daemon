package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/btcsuite/btcutil"
	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/explorer/esplora"
	"github.com/urfave/cli/v2"
	"github.com/vulpemventures/go-elements/network"
)

const (
	noMacaroonsKey   = "no_macaroons"
	macaroonsPathKey = "macaroons_path"
	tlsCertPathKey   = "tls_cert_path"
)

var (
	daemonDatadir = btcutil.AppDataDir("tdex-daemon", false)

	defaultNetwork       = network.Liquid.Name
	defaultExplorer      = "https://blockstream.info/liquid/api"
	defaultRPCServer     = "localhost:9000"
	defaultMacaroonsAuth = false
	defaultTLSCertPath   = filepath.Join(daemonDatadir, "tls", "cert.pem")
	defaultMacaroonsPath = filepath.Join(daemonDatadir, "macaroons", "admin.macaroon")

	networkFlag = cli.StringFlag{
		Name:  "network, n",
		Usage: "the network tdexd is running on: liquid or regtest",
		Value: defaultNetwork,
	}

	explorerUrlFlag = cli.StringFlag{
		Name:  "explorer_url",
		Usage: "explorer url for the current network",
		Value: defaultExplorer,
	}

	rpcFlag = cli.StringFlag{
		Name:  "rpcserver",
		Usage: "tdexd daemon address host:port",
		Value: defaultRPCServer,
	}

	tlsCertFlag = cli.StringFlag{
		Name:  tlsCertPathKey,
		Usage: "the path of the TLS certificate file to use",
		Value: defaultTLSCertPath,
	}

	noMacaroonsFlag = cli.BoolFlag{
		Name:  noMacaroonsKey,
		Usage: "used to start the daemon without macaroon auth",
		Value: defaultMacaroonsAuth,
	}

	macaroonsFlag = cli.StringFlag{
		Name:  macaroonsPathKey,
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
				&explorerUrlFlag,
				&rpcFlag,
				&tlsCertFlag,
				&noMacaroonsFlag,
				&macaroonsFlag,
			},
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
		"explorer_url":   c.String("explorer_url"),
		"rpcserver":      c.String("rpcserver"),
		"no_macaroons":   c.String(noMacaroonsKey),
		"tls_cert_path":  cleanAndExpandPath(c.String(tlsCertPathKey)),
		"macaroons_path": cleanAndExpandPath(c.String(macaroonsPathKey)),
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

func getNetworkFromState() (*network.Network, error) {
	state, err := getState()
	if err != nil {
		return nil, err
	}

	net, ok := state["network"]
	if !ok {
		return &network.Liquid, nil
	}
	if net == "regtest" {
		return &network.Regtest, nil
	}
	return &network.Liquid, nil
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

func getExplorerFromState() (explorer.Service, error) {
	state, err := getState()
	if err != nil {
		return nil, err
	}

	reqTimeout := 15000
	url, ok := state["explorer_url"]
	if !ok {
		url = "https://blockstream.info/liquid/api"
	}

	return esplora.NewService(url, reqTimeout)
}

func getWalletFromState(walletType string) (map[string]string, error) {
	state, err := getState()
	if err != nil {
		return nil, err
	}

	walletKey := fmt.Sprintf("%s_wallet", walletType)
	walletJSON, ok := state[walletKey]
	if !ok || walletJSON == "" {
		return nil, nil
	}

	wallet := map[string]string{}
	if err := json.Unmarshal([]byte(walletJSON), &wallet); err != nil {
		return nil, err
	}
	return wallet, nil
}

func flushWallet(walletType string) {
	state, _ := getState()
	walletKey := fmt.Sprintf("%s_wallet", walletType)
	if _, ok := state[walletKey]; ok {
		setState(map[string]string{walletKey: ""})
	}
}
