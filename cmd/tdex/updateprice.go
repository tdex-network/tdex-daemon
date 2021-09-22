package main

import (
	"github.com/urfave/cli/v2"
)

var updateprice = cli.Command{
	Name:  "price",
	Usage: "updates the current market price to be used for future trades",
	Flags: []cli.Flag{
		&cli.Float64Flag{
			Name:     "base_price",
			Usage:    "set the base price",
			Required: true,
		},
		&cli.Float64Flag{
			Name:     "quote_price",
			Usage:    "set the quote price",
			Required: true,
		},
	},
	Action: updatePriceAction,
}

func updatePriceAction(ctx *cli.Context) error {
	printDeprecatedWarn("tdex market price")
	return nil
}
