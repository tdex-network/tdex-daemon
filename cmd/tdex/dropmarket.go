package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var dropmarket = cli.Command{
	Name:  "dropmarket",
	Usage: "drop a market",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "account_index",
			Usage: "the account index of the market to drop",
		},
	},
	Action: dropMarketAction,
}

func dropMarketAction(ctx *cli.Context) error {
	return fmt.Errorf(
		"this command is deprecated and will be removed in the next version.\n" +
			"Instead, use the new command 'tdex market drop'",
	)
}
