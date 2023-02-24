package main

import (
	"context"
	"fmt"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"

	"github.com/urfave/cli/v2"
)

var removewebhook = cli.Command{
	Name:  "removewebhook",
	Usage: "remove a webhook",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "id",
			Usage: "the id of the webhook to remove",
			Value: "",
		},
	},
	Action: removeWebhookAction,
}

func removeWebhookAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	hookID := ctx.String("id")

	if _, err := client.RemoveWebhook(
		context.Background(), &daemonv2.RemoveWebhookRequest{
			Id: hookID,
		},
	); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("removed hook with id:", hookID)
	return nil
}
