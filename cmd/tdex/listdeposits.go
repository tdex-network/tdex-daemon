package main

import (
	"context"

	pboperator "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/operator"
	"github.com/urfave/cli/v2"
)

var listdeposits = cli.Command{
	Name:   "listdeposits",
	Usage:  "list deposits",
	Action: listDeposits,
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "account_index",
			Usage: "account index for which deposits should be listed",
		},
		&cli.IntFlag{
			Name:  "page_number",
			Usage: "page to be listed",
			Value: 1,
		},
		&cli.IntFlag{
			Name:  "page_size",
			Usage: "size of the page",
			Value: 10,
		},
	},
}

func listDeposits(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	accountIndex := ctx.Int("account_index")
	pageNumber := ctx.Int("page_number")
	pageSize := ctx.Int("page_size")

	resp, err := client.ListDeposits(
		context.Background(), &pboperator.ListDepositsRequest{
			AccountIndex: int64(accountIndex),
			Page: &pboperator.Page{
				PageNumber: int64(pageNumber),
				PageSize:   int64(pageSize),
			},
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}
