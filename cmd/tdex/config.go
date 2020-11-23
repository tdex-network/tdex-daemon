package main

import (
	"errors"
	"fmt"

	"github.com/urfave/cli/v2"
	"github.com/vulpemventures/go-elements/network"
)

var (
	networkFlag = cli.StringFlag{
		Name:  "network, n",
		Usage: "the network tdexd is running on: liquid or regtest",
		Value: network.Liquid.Name,
	}

	rpcFlag = cli.StringFlag{
		Name:  "rpcserver",
		Usage: "tdexd daemon address host:port",
		Value: "localhost:9000",
	}

	macaroonFlag = cli.StringFlag{
		Name:  "macaroon",
		Usage: "hex encoded admin macaroon",
		Value: "",
	}
)

var config = cli.Command{
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
				&macaroonFlag,
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
		"network":   c.String("network"),
		"rpcserver": c.String("rpcserver"),
		"macaroon":  c.String("macaroon"),
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
		return "", "", errors.New("set base asset with `config set quote_asset`")
	}

	return baseAsset, quoteAsset, nil
}

func setMarketIntoState(baseAsset, quoteAsset string) error {
	return setState(map[string]string{
		"base_asset":  baseAsset,
		"quote_asset": quoteAsset,
	})
}
