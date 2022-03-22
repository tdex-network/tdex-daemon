package main

import (
	"context"

	daemonv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/go/tdex-daemon/v1"
	"github.com/urfave/cli/v2"
)

var getwalletinfo = cli.Command{
	Name:   "info",
	Usage:  "get info about the internal wallet of the daemon",
	Action: infoAction,
}

func infoAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	resp, err := client.GetInfo(context.Background(), &daemonv1.GetInfoRequest{})
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}
