package main

import (
	"context"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"

	"github.com/urfave/cli/v2"
)

var listmarkets = cli.Command{
	Name:   "markets",
	Usage:  "get a list of all markets",
	Action: listMarketsAction,
}

func listMarketsAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	resp, err := client.ListMarkets(
		context.Background(), &daemonv2.ListMarketsRequest{},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}
