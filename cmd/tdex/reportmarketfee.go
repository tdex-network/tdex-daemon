package main

import (
	"context"

	pboperator "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"

	"github.com/urfave/cli/v2"
)

var reportmarketfee = cli.Command{
	Name:   "reportmarketfee",
	Usage:  "return a report of the collected fees for a market.",
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

	reply, err := client.ReportMarketFee(
		context.Background(), &pboperator.ReportMarketFeeRequest{
			Market: &pbtypes.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(reply)

	return nil
}
