package main

import (
	"context"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"
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
		context.Background(), &daemonv2.ReloadUtxosRequest{},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}
