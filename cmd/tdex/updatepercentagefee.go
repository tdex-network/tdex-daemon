package main

import (
	"github.com/urfave/cli/v2"
)

var updatePercentagefee = cli.Command{
	Name:  "percentagefee",
	Usage: "updates the current market percentage fee",
	Flags: []cli.Flag{
		&cli.Int64Flag{
			Name:  "basis_point",
			Usage: "set the fee basis point",
		},
	},
	Action: updatePercentageFeeAction,
}

func updatePercentageFeeAction(ctx *cli.Context) error {
	printDeprecatedWarn("tdex market percentagefee")
	return nil
}
