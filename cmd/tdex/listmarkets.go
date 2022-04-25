package main

import (
	"context"

	daemonv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v1"

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
		context.Background(), &daemonv1.ListMarketsRequest{},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}
