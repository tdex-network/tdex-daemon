package main

import (
	"fmt"

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
	return fmt.Errorf(
		"this command is deprecated and will be removed in the next version.\n" +
			"Instead, use the new command 'tdex market fixedfee'",
	)
}
