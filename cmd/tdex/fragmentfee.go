package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

const (
	MinFee          = 5000
	MaxNumOfOutputs = 50
)

var fragmentfee = cli.Command{
	Name: "fragmentfee",
	Usage: "deposit funds for fee account into an ephemeral wallet, then " +
		"split the amount into multiple fragments and deposit into the daemon",
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Name:  "txid",
			Usage: "txid of the funds to resume a previous fragmentfee",
		},
		&cli.Uint64Flag{
			Name: "max_fragments",
			Usage: fmt.Sprintf(
				"specify the max number of fragments created. "+
					"Values over %d will be overridden to %d",
				MaxNumOfOutputs, MaxNumOfOutputs,
			),
			Value: MaxNumOfOutputs,
		},
		&cli.StringFlag{
			Name:  "recover_funds_to_address",
			Usage: "specify an address where to send funds stuck into the fragmenter to",
		},
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "print tx hex in case the transaction fails to be broadcasted",
		},
	},
	Action: fragmentFeeAction,
}

func fragmentFeeAction(ctx *cli.Context) error {
	printDeprecatedWarn("tdex fragmenter fee")
	return nil
}
