package main

import (
	"context"

	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"

	"github.com/urfave/cli/v2"
)

var listmarket = cli.Command{
	Name:   "listmarket",
	Usage:  "list all created markets",
	Action: listMarketAction,
}

func listMarketAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	resp, err := client.ListMarket(
		context.Background(), &pboperator.ListMarketRequest{},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}
