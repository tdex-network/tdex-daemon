package main

import (
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
	printDeprecatedWarn("tdex fee deposit")
	return nil
}
