package main

import (
	"github.com/urfave/cli/v2"
)

var claimfee = cli.Command{
	Name:  "claimfee",
	Usage: "claim deposits for the fee account",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "outpoints",
			Usage: "list of outpoints referring to utxos [{\"hash\": <string>, \"index\": <number>}]",
		},
	},
	Action: claimFeeAction,
}

func claimFeeAction(ctx *cli.Context) error {
	printDeprecatedWarn("tdex fee claim")
	return nil
}
