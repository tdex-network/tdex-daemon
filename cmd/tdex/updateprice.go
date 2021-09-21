package main

import (
	"fmt"

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
	return fmt.Errorf(
		"this command is deprecated and will be removed in the next version.\n" +
			"Instead, use the new command 'tdex market price'",
	)
}
