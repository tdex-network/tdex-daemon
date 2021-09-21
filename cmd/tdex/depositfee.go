package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var depositfee = cli.Command{
	Name:  "depositfee",
	Usage: "get a deposit address for the fee account used to subsidize liquid network fees",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "num_of_addresses",
			Usage: "the number of addresses to retrieve",
		},
	},
	Action: depositFeeAction,
}

func depositFeeAction(ctx *cli.Context) error {
	return fmt.Errorf(
		"this command is deprecated and will be removed in the next version.\n" +
			"Instead, use the new command 'tdex fee deposit'",
	)
}
