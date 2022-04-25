package main

import (
	"context"

	daemonv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v1"
	"github.com/urfave/cli/v2"
)

var status = cli.Command{
	Name:   "status",
	Usage:  "returns info about the status of the daemon",
	Action: getStatusAction,
}

func getStatusAction(ctx *cli.Context) error {
	client, cleanup, err := getUnlockerClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.IsReady(context.Background(), &daemonv1.IsReadyRequest{})
	if err != nil {
		return err
	}

	printRespJSON(reply)

	return nil
}
