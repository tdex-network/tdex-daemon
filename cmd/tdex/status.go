package main

import (
	"context"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"
	"github.com/urfave/cli/v2"
)

var status = cli.Command{
	Name:   "status",
	Usage:  "get info about the status of the daemon",
	Action: getStatusAction,
}

func getStatusAction(ctx *cli.Context) error {
	client, cleanup, err := getWalletClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.GetStatus(context.Background(), &daemonv2.GetStatusRequest{})
	if err != nil {
		return err
	}

	printRespJSON(reply)
	return nil
}
