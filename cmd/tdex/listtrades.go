package main

import (
	"context"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"
	tdexv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v2"

	"github.com/urfave/cli/v2"
)

var listtrades = cli.Command{
	Name:  "trades",
	Usage: "get a list of all trades for a market",
	Flags: []cli.Flag{
		&cli.Uint64Flag{
			Name:  "page",
			Usage: "the number of the page to be listed. If omitted, the entire list is returned",
		},
		&cli.Uint64Flag{
			Name:  "page-size",
			Usage: "the size of the page",
			Value: 10,
		},
	},
	Action: listTradesAction,
}

func listTradesAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	pageNumber := ctx.Int64("page")
	pageSize := ctx.Int64("page-size")
	var page *daemonv2.Page
	if pageNumber > 0 {
		page = &daemonv2.Page{
			Number: pageNumber,
			Size:   pageSize,
		}
	}

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	resp, err := client.ListTrades(
		context.Background(), &daemonv2.ListTradesRequest{
			Market: &tdexv2.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
			Page: page,
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}
