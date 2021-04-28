package main

import (
	"context"

	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"

	"github.com/urfave/cli/v2"
)

var listtrades = cli.Command{
	Name:   "listtrades",
	Usage:  "list all processed trades",
	Action: listTradesAction,
}

func listTradesAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	resp, err := client.ListTrades(
		context.Background(), &pboperator.ListTradesRequest{},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}
