package main

import (
	"github.com/urfave/cli/v2"
)

var depositmarket = cli.Command{
	Name:  "depositmarket",
	Usage: "get a deposit address for a given market or create a new one",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "base_asset",
			Usage: "the base asset hash of an existent market",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "quote_asset",
			Usage: "the base asset hash of an existent market",
			Value: "",
		},
		&cli.IntFlag{
			Name:  "num_of_addresses",
			Usage: "the number of addresses to generate for the market",
		},
	},
	Action: depositMarketAction,
}

func depositMarketAction(ctx *cli.Context) error {
	printDeprecatedWarn("tdex market deposit")
	return nil
}
