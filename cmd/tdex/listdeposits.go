package main

import (
	"context"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"
	"github.com/urfave/cli/v2"
)

var listdeposits = cli.Command{
	Name:  "deposits",
	Usage: "get a list of all deposits for a wallet account",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "account-name",
			Usage:    "the name of the account for which listing deposits",
			Required: true,
		},
		&cli.Int64Flag{
			Name:  "page",
			Usage: "the number of the page to be listed. If omitted, the entire list is returned",
		},
		&cli.Int64Flag{
			Name:  "page-size",
			Usage: "the size of the page",
			Value: 10,
		},
	},
	Action: listDepositsAction,
}

func listDepositsAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	accountName := ctx.String("account-name")
	pageNumber := ctx.Int64("page")
	pageSize := ctx.Int64("page-size")
	var page *daemonv2.Page
	if pageNumber > 0 {
		page = &daemonv2.Page{
			Number: pageNumber,
			Size:   pageSize,
		}
	}

	resp, err := client.ListDeposits(
		context.Background(), &daemonv2.ListDepositsRequest{
			AccountName: accountName,
			Page:        page,
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)
	return nil
}
