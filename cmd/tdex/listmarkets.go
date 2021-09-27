package main

import (
	"context"

	pboperator "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/operator"

	"github.com/urfave/cli/v2"
)

var listmarkets = cli.Command{
	Name:   "listmarkets",
	Usage:  "list all created markets",
	Action: listMarketsAction,
}

func listMarketsAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	resp, err := client.ListMarkets(
		context.Background(), &pboperator.ListMarketsRequest{},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}
