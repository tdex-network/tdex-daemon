package main

import (
	"github.com/urfave/cli/v2"
)

var withdrawmarket = cli.Command{
	Name:  "withdrawmarket",
	Usage: "withdraw funds from some market.",
	Flags: []cli.Flag{
		&cli.Uint64Flag{
			Name:  "base_amount",
			Usage: "the amount in Satoshi of base asset to withdraw from the market.",
		},
		&cli.Uint64Flag{
			Name:  "quote_amount",
			Usage: "the amount in Satoshi of quote asset to withdraw from the market.",
		},
		&cli.StringFlag{
			Name:  "address",
			Usage: "the address where to send the withdrew amount(s).",
		},
		&cli.Int64Flag{
			Name:  "millisatsperbyte",
			Usage: "the mSat/byte to pay for the transaction",
			Value: 100,
		},
	},
	Action: withdrawMarketAction,
}

func withdrawMarketAction(ctx *cli.Context) error {
	printDeprecatedWarn("tdex market withdraw")
	return nil
}
