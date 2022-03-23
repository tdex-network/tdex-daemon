package main

import (
	"context"

	daemonv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/go/tdex-daemon/v1"
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
		context.Background(), &daemonv1.ReloadUtxosRequest{},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}
