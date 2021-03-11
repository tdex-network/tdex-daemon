package main

import (
	"context"

	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"

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
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	numOfAddresses := ctx.Int64("num_of_addresses")
	resp, err := client.DepositFeeAccount(
		context.Background(), &pboperator.DepositFeeAccountRequest{
			NumOfAddresses: numOfAddresses,
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}
