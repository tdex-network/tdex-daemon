package main

import (
	"context"
	"fmt"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"

	"github.com/urfave/cli/v2"
)

var (
	webhook = cli.Command{
		Name:  "webhook",
		Usage: "add or remove webhooks",
		Subcommands: []*cli.Command{
			webhookAddCmd, webhookRemoveCmd,
		},
	}
	listwebhooks = cli.Command{
		Name:  "webhooks",
		Usage: "list all webhooks, optionally filtered by target event",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "trade_settled_event",
				Usage: "triggers the webhook endpoint whenever a trade is settled",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "account_low_balance_event",
				Usage: "triggers the webhook endpoint whenever a wallet account's balance goes under a threshold configured at startup",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "account_withdraw_event",
				Usage: "triggers the webhook endpoint whenever a withdrawal from a wallet account is made",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "account_deposit_event",
				Usage: "triggers the webhook endpoint whenever a deposit to a wallet account is made",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "any_event",
				Usage: "triggers the webhook endpoint whenever any event occurs",
				Value: false,
			},
		},
		Action: listWebhooksAction,
	}

	webhookAddCmd = &cli.Command{
		Name:  "add",
		Usage: "add a (secured) webhook endpoint called whenever a target event occurs",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "endpoint",
				Usage: "the webhook endpoint to be called whenever the target event occurs",
				Value: "",
			},
			&cli.StringFlag{
				Name: "secret",
				Usage: "the eventual secret to use to generate an OAuth token for " +
					"authenticating requests to the webhook endpoint",
				Value: "",
			},
			&cli.BoolFlag{
				Name:  "trade_settled_event",
				Usage: "triggers the webhook endpoint whenever a trade is settled",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "account_low_balance_event",
				Usage: "triggers the webhook endpoint whenever a wallet account's balance goes under a threshold configured at startup",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "account_withdraw_event",
				Usage: "triggers the webhook endpoint whenever a withdrawal from a wallet account is made",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "account_deposit_event",
				Usage: "triggers the webhook endpoint whenever a deposit to a wallet account is made",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "any_event",
				Usage: "triggers the webhook endpoint whenever any event occurs",
				Value: false,
			},
		},
		Action: addWebhookAction,
	}

	webhookRemoveCmd = &cli.Command{
		Name:  "remove",
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
)

func addWebhookAction(ctx *cli.Context) error {
	client, cleanup, err := getWebhookClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	endpoint := ctx.String("endpoint")
	secret := ctx.String("secret")
	event, err := parseEvent(ctx)
	if err != nil {
		return err
	}
	if event == daemonv2.WebhookEvent_WEBHOOK_EVENT_UNSPECIFIED {
		return fmt.Errorf("missing event")
	}

	reply, err := client.AddWebhook(
		context.Background(), &daemonv2.AddWebhookRequest{
			Endpoint: endpoint,
			Event:    event,
			Secret:   secret,
		},
	)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("webhook id:", reply.GetId())
	return nil
}

func removeWebhookAction(ctx *cli.Context) error {
	client, cleanup, err := getWebhookClient(ctx)
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
	fmt.Println("removed webhook with id:", hookID)
	return nil
}

func listWebhooksAction(ctx *cli.Context) error {
	client, cleanup, err := getWebhookClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	event, err := parseEvent(ctx)
	if err != nil {
		return err
	}

	reply, err := client.ListWebhooks(
		context.Background(), &daemonv2.ListWebhooksRequest{
			Event: event,
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(reply)

	return nil
}

func parseEvent(ctx *cli.Context) (daemonv2.WebhookEvent, error) {
	event := daemonv2.WebhookEvent_WEBHOOK_EVENT_UNSPECIFIED
	events := []bool{
		ctx.Bool("trade_settled_event"),
		ctx.Bool("account_low_balance_event"),
		ctx.Bool("account_withdraw_event"),
		ctx.Bool("account_deposit_event"),
		ctx.Bool("any_event"),
	}
	trues := 0
	for _, e := range events {
		if e {
			trues++
		}
	}
	if trues > 1 {
		return -1, fmt.Errorf("only one event can be set for a webhook")
	}
	switch {
	case ctx.Bool("trade_settled_event"):
		event = daemonv2.WebhookEvent_WEBHOOK_EVENT_TRADE_SETTLED
	case ctx.Bool("account_low_balance_event"):
		event = daemonv2.WebhookEvent_WEBHOOK_EVENT_ACCOUNT_LOW_BALANCE
	case ctx.Bool("account_withdraw_event"):
		event = daemonv2.WebhookEvent_WEBHOOK_EVENT_ACCOUNT_WITHDRAW
	case ctx.Bool("account_deposit_event"):
		event = daemonv2.WebhookEvent_WEBHOOK_EVENT_ACCOUNT_DEPOSIT
	case ctx.Bool("any_event"):
		event = daemonv2.WebhookEvent_WEBHOOK_EVENT_ANY
	}

	return event, nil
}
