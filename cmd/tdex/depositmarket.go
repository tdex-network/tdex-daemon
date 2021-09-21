package main

import (
	"fmt"

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
	return fmt.Errorf(
		"this command is deprecated and will be removed in the next version.\n" +
			"Instead, use the new command 'tdex market deposit'",
	)
}
