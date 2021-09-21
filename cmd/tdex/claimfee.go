package main

import (
	"fmt"

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
	return fmt.Errorf(
		"this command is deprecated and will be removed in the next version.\n" +
			"Instead, use the new command 'tdex fee claim'",
	)
}
