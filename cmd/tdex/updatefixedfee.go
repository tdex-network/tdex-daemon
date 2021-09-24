package main

import (
	"github.com/urfave/cli/v2"
)

var updateFixedfee = cli.Command{
	Name:  "fixedfee",
	Usage: "updates the current market fixed fee",
	Flags: []cli.Flag{
		&cli.Int64Flag{
			Name:  "base_fee",
			Usage: "set the fixed fee for base asset",
		},
		&cli.Int64Flag{
			Name:  "quote_fee",
			Usage: "set the fixed fee for quote asset",
		},
	},
	Action: updateFixedFeeAction,
}

func updateFixedFeeAction(ctx *cli.Context) error {
	printDeprecatedWarn("tdex market fixedfee")
	return nil
}
