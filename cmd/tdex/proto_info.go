package main

import (
	"context"

	daemonv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v1"
	"github.com/urfave/cli/v2"
)

var getProtoSvcInfo = cli.Command{
	Name:   "proto",
	Usage:  "get info about proto services of the daemon",
	Action: protoInfoAction,
}

func protoInfoAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	resp, err := client.ListProtoServices(context.Background(), &daemonv1.ListProtoServicesRequest{})
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}
