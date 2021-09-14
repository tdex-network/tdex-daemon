package main

import (
	"context"

	pb "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/operator"
	"github.com/urfave/cli/v2"
)

var balancefee = cli.Command{
	Name:   "balancefee",
	Usage:  "check the balance of the fee account.",
	Action: balanceFeeAccountAction,
}

func balanceFeeAccountAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.BalanceFeeAccount(context.Background(), &pb.BalanceFeeAccountRequest{})
	if err != nil {
		return err
	}

	printRespJSON(reply)

	return nil
}
