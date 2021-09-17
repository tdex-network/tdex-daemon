package main

import (
	"context"

	pboperator "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"

	"github.com/urfave/cli/v2"
)

var reportmarketfee = cli.Command{
	Name:  "reportmarketfee",
	Usage: "get a report of the fees collected for the trades of a market.",
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
	},
	Action: reportMarketFeeAction,
}

func reportMarketFeeAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	pageNumber := ctx.Int64("page")
	pageSize := ctx.Int64("page_size")
	var page *pboperator.Page
	if pageNumber > 0 {
		page = &pboperator.Page{
			PageNumber: pageNumber,
			PageSize:   pageSize,
		}
	}

	reply, err := client.ReportMarketFee(
		context.Background(), &pboperator.ReportMarketFeeRequest{
			Market: &pbtypes.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
			Page: page,
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(reply)

	return nil
}
