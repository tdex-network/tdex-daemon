package main

import (
	"context"

	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	"github.com/urfave/cli/v2"
)

var reloadtxos = cli.Command{
	Name:   "reloadutxos",
	Usage:  "reload all utxos",
	Action: reloadUtxos,
}

func reloadUtxos(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	resp, err := client.ReloadUtxos(
		context.Background(), &pboperator.ReloadUtxosRequest{},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}
