package main

import (
	"context"
	"fmt"

	pboperator "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/operator"

	"github.com/urfave/cli/v2"
)

var addwebhook = cli.Command{
	Name:  "addwebhook",
	Usage: "add a webhook registered for some event",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "endpoint",
			Usage: "the endpoint where to notify the webhook",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "secret",
			Usage: "the eventual secret to authenticate requests",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "action",
			Usage: "the action for which the webhook gets notified",
		},
	},
	Action: addWebhookAction,
}

func addWebhookAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	endpoint := ctx.String("endpoint")
	secret := ctx.String("secret")
	action, ok := pboperator.ActionType_value[ctx.String("action")]
	if !ok {
		return fmt.Errorf("unknown action type")
	}

	reply, err := client.AddWebhook(
		context.Background(), &pboperator.AddWebhookRequest{
			Endpoint: endpoint,
			Action:   pboperator.ActionType(action),
			Secret:   secret,
		},
	)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("hook id:", reply.GetId())
	return nil
}
