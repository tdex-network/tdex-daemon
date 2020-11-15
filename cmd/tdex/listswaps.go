package main

import (
	"context"

	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"

	"github.com/urfave/cli/v2"
)

var listswaps = cli.Command{
	Name:   "listswaps",
	Usage:  "list all completed swaps",
	Action: listSwapsAction,
}

func listSwapsAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	resp, err := client.ListSwaps(
		context.Background(), &pboperator.ListSwapsRequest{},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}
