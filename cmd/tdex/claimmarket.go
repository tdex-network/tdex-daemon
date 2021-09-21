package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var claimmarket = cli.Command{
	Name:  "claimmarket",
	Usage: "claim deposits for a market",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "outpoints",
			Usage: "list of outpoints referring to utxos [{\"hash\": <string>, \"index\": <number>}]",
		},
	},
	Action: claimMarketAction,
}

func claimMarketAction(ctx *cli.Context) error {
	return fmt.Errorf(
		"this command is deprecated and will be removed in the next version.\n" +
			"Instead, use the new command 'tdex market claim'",
	)
}
