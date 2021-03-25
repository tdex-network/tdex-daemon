package main

import (
	"context"

	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	"github.com/urfave/cli/v2"
)

var listutxos = cli.Command{
	Name:   "listutxos",
	Usage:  "list all utxos",
	Action: listUtxos,
}

func listUtxos(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	resp, err := client.ListUtxos(
		context.Background(), &pboperator.ListUtxosRequest{},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}
