package main

import (
	"context"

	daemonv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v1"
	"github.com/urfave/cli/v2"
)

var listdeposits = cli.Command{
	Name:  "listdeposits",
	Usage: "list all deposits for a wallet account",
	Flags: []cli.Flag{
		&cli.Uint64Flag{
			Name:     "account_index",
			Usage:    "the index of the account for which listing deposits",
			Required: true,
		},
		&cli.Uint64Flag{
			Name:  "page",
			Usage: "the number of the page to be listed. If omitted, the entire list is returned",
		},
		&cli.Uint64Flag{
			Name:  "page_size",
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

	accountIndex := ctx.Int64("account_index")
	pageNumber := ctx.Int64("page")
	pageSize := ctx.Int64("page_size")
	var page *daemonv1.Page
	if pageNumber > 0 {
		page = &daemonv1.Page{
			PageNumber: pageNumber,
			PageSize:   pageSize,
		}
	}

	resp, err := client.ListDeposits(
		context.Background(), &daemonv1.ListDepositsRequest{
			AccountIndex: accountIndex,
			Page:         page,
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}
