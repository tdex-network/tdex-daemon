package main

import (
	"context"
	"fmt"

	daemonv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/go/tdex-daemon/v1"

	"github.com/urfave/cli/v2"
)

var listwebhooks = cli.Command{
	Name:  "listwebhooks",
	Usage: "list all webhook registered for some action",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "action",
			Usage: "the action to filter hooks by",
		},
	},
	Action: listWebhooksAction,
}

func listWebhooksAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	action, ok := daemonv1.ActionType_value[ctx.String("action")]
	if !ok {
		return fmt.Errorf("unknown action type")
	}

	reply, err := client.ListWebhooks(
		context.Background(), &daemonv1.ListWebhooksRequest{
			Action: daemonv1.ActionType(action),
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(reply)

	return nil
}
