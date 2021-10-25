package main

import (
	"github.com/urfave/cli/v2"
)

var fragmentmarket = cli.Command{
	Name: "fragmentmarket",
	Usage: "deposit funds for a market (either existing or to be created) " +
		"into an ephemeral wallet, then split the amount into multiple " +
		"fragments and deposit into the daemon",
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Name:  "txid",
			Usage: "txid of the funds to resume a fragmentmarket",
		},
		&cli.StringFlag{
			Name:  "recover_funds_to_address",
			Usage: "specify an address where to send funds stuck into the fragmenter to",
		},
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "print tx hex in case the transaction fails to be broadcasted",
		},
	},
	Action: fragmentMarketAction,
}

func fragmentMarketAction(ctx *cli.Context) error {
	printDeprecatedWarn("tdex market deposit --fragment")
	return nil
}
