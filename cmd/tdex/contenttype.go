package main

import (
	"context"

	"github.com/tdex-network/tdex-protobuf/generated/go/transport"

	"github.com/urfave/cli/v2"
)

var contentType = cli.Command{
	Name:   "transport",
	Usage:  "list available content types",
	Action: listContentTypes,
}

func listContentTypes(ctx *cli.Context) error {
	client, cleanup, err := getTransportClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	resp, err := client.SupportedContentTypes(
		context.Background(), &transport.SupportedContentTypesRequest{},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}
