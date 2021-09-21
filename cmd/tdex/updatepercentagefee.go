package main

import (
	"fmt"

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
	return fmt.Errorf(
		"this command is deprecated and will be removed in the next version.\n" +
			"Instead, use the new command 'tdex market percentagefee'",
	)
}
