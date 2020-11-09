package main

import (
	"errors"
	"fmt"

	"github.com/urfave/cli/v2"
)

var market = cli.Command{
	Name:  "market",
	Usage: "select a market to run commands against",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "base_asset",
			Usage:    "the base asset hash of an existent market",
			Value:    "",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "quote_asset",
			Usage:    "the base asset hash of an existent market",
			Value:    "",
			Required: true,
		},
	},
	Action: marketAction,
}

func marketAction(ctx *cli.Context) error {

	base := ctx.String("base_asset")
	quote := ctx.String("quote_asset")

	if len(base) == 0 || len(quote) == 0 {
		return &invalidUsageError{ctx, ctx.Command.Name}
	}

	err := setMarketIntoState(base, quote)
	if err != nil {
		return err
	}

	fmt.Println("market has been selected")
	return nil
}

func getMarketFromState() (string, string, error) {
	state, err := getState()
	if err != nil {
		return "", "", errors.New("a market must be selected")
	}
	baseAsset := state["base_asset"]
	quoteAsset := state["quote_asset"]

	return baseAsset, quoteAsset, nil
}

func setMarketIntoState(baseAsset, quoteAsset string) error {
	return setState(map[string]string{
		"base_asset":  baseAsset,
		"quote_asset": quoteAsset,
	})
}
