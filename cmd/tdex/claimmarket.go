package main

import (
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
	printDeprecatedWarn("tdex market claim")
	return nil
}
