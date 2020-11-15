package main

import (
	"context"

	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"

	"github.com/urfave/cli/v2"
)

var depositfee = cli.Command{
	Name:   "depositfee",
	Usage:  "get a deposit address for the fee account used to subsidize liquid network fees",
	Action: depositFeeAction,
}

func depositFeeAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	resp, err := client.DepositFeeAccount(
		context.Background(), &pboperator.DepositFeeAccountRequest{},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}
