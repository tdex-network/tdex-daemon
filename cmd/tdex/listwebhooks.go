package main

import (
	"context"
	"fmt"

	pboperator "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/operator"

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

	action, ok := pboperator.ActionType_value[ctx.String("action")]
	if !ok {
		return fmt.Errorf("unknown action type")
	}

	reply, err := client.ListWebhooks(
		context.Background(), &pboperator.ListWebhooksRequest{
			Action: pboperator.ActionType(action),
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(reply)

	return nil
}
