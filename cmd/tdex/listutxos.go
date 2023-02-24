package main

import (
	"context"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"
	"github.com/urfave/cli/v2"
)

var listutxos = cli.Command{
	Name:  "utxos",
	Usage: "get a list of all utxos for a wallet account",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "account_name",
			Usage:    "the name of the wallet account for which listing utxos",
			Required: true,
		},
		&cli.Int64Flag{
			Name:  "page",
			Usage: "the number of the page to be listed. If omitted, the entire list is returned",
		},
		&cli.Int64Flag{
			Name:  "page_size",
			Usage: "the size of the page",
			Value: 10,
		},
	},
	Action: listUtxosAction,
}

func listUtxosAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	accountName := ctx.String("account_name")
	pageNumber := ctx.Int64("page")
	pageSize := ctx.Int64("page_size")
	var page *daemonv2.Page
	if pageNumber > 0 {
		page = &daemonv2.Page{
			Number: pageNumber,
			Size:   pageSize,
		}
	}

	resp, err := client.ListUtxos(
		context.Background(), &daemonv2.ListUtxosRequest{
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
