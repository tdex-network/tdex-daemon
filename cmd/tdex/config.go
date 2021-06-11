package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tdex-network/tdex-daemon/pkg/explorer"
	"github.com/tdex-network/tdex-daemon/pkg/explorer/esplora"
	"github.com/urfave/cli/v2"
	"github.com/vulpemventures/go-elements/network"
)

var (
	networkFlag = cli.StringFlag{
		Name:  "network, n",
		Usage: "the network tdexd is running on: liquid or regtest",
		Value: network.Liquid.Name,
	}

	explorerUrlFlag = cli.StringFlag{
		Name:  "explorer_url",
		Usage: "explorer url for the current network",
		Value: "https://blockstream.info/liquid/api",
	}

	rpcFlag = cli.StringFlag{
		Name:  "rpcserver",
		Usage: "tdexd daemon address host:port",
		Value: "localhost:9000",
	}

	tlsCertFlag = cli.StringFlag{
		Name:  "tls_cert_path",
		Usage: fmt.Sprintf("the directory where to fing the %s file", tlsCertFile),
		Value: "",
	}

	noMacaroonsFlag = cli.BoolFlag{
		Name:  "no_macaroons",
		Usage: "used to start the daemon without macaroon auth",
		Value: false,
	}

	macaroonsFlag = cli.StringFlag{
		Name:  "macaroons_path",
		Usage: fmt.Sprintf("the directory where to find the %s file", adminMacaroonFile),
		Value: "",
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
	err := setState(map[string]string{
		"network":        c.String("network"),
		"explorer_url":   c.String("explorer_url"),
		"rpcserver":      c.String("rpcserver"),
		"tls_cert_path":  cleanAndExpandPath(c.String("tls_cert_path")),
		"no_macaroons":   c.String("no_macaroons"),
		"macaroons_path": cleanAndExpandPath(c.String("macaroons_path")),
	})

	if err != nil {
		return err
	}

	return nil
}

func configSetAction(c *cli.Context) error {
	if c.NArg() < 2 {
		return errors.New("key and value are missing")
	}

	key := c.Args().Get(0)
	value := c.Args().Get(1)

	err := setState(map[string]string{key: value})
	if err != nil {
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
		return "", "", errors.New("set base asset with `config set base_asset`")
	}
	quoteAsset, ok := state["quote_asset"]
	if !ok {
		return "", "", errors.New("set quote asset with `config set quote_asset`")
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
