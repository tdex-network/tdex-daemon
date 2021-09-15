package main

import (
	"context"

	pboperator "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"

	"github.com/urfave/cli/v2"
)

var listtrades = cli.Command{
	Name:  "listtrades",
	Usage: "list all trades for a market, or for all markets",
	Flags: []cli.Flag{
		&cli.Uint64Flag{
			Name:  "page",
			Usage: "the number of the page to be listed. If omitted, the entire list is returned",
		},
		&cli.Uint64Flag{
			Name:  "page_size",
			Usage: "the size of the page",
			Value: 10,
		},
		&cli.BoolFlag{
			Name:  "all",
			Usage: "to list all trades, not filtered by any market",
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

	allTrades := ctx.Bool("all")
	pageNumber := ctx.Int64("page")
	pageSize := ctx.Int64("page_size")
	var page *pboperator.Page
	if pageNumber > 0 {
		page = &pboperator.Page{
			PageNumber: pageNumber,
			PageSize:   pageSize,
		}
	}

	baseAsset, quoteAsset, _ := getMarketFromState()
	var market *pbtypes.Market
	if baseAsset != "" && !allTrades {
		market = &pbtypes.Market{
			BaseAsset:  baseAsset,
			QuoteAsset: quoteAsset,
		}
	}

	resp, err := client.ListTrades(
		context.Background(), &pboperator.ListTradesRequest{
			Market: market,
			Page:   page,
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}
