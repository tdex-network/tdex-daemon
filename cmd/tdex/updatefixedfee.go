package main

import (
	"context"
	"fmt"

	pboperator "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"

	"github.com/urfave/cli/v2"
)

var updateFixedfee = cli.Command{
	Name:  "fixedfee",
	Usage: "updates the current market fixed fee",
	Flags: []cli.Flag{
		&cli.Int64Flag{
			Name:  "base_fee",
			Usage: "set the fixed fee for base asset",
		},
		&cli.Int64Flag{
			Name:  "quote_fee",
			Usage: "set the fixed fee for quote asset",
		},
	},
	Action: updateFixedFeeAction,
}

func updateFixedFeeAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	baseFee := ctx.Int64("base_fee")
	quoteFee := ctx.Int64("quote_fee")
	req := &pboperator.UpdateMarketFixedFeeRequest{
		Market: &pbtypes.Market{
			BaseAsset:  baseAsset,
			QuoteAsset: quoteAsset,
		},
		Fixed: &pbtypes.Fixed{
			BaseFee:  baseFee,
			QuoteFee: quoteFee,
		},
	}

	if _, err := client.UpdateMarketFixedFee(
		context.Background(), req,
	); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("market fees have been updated")
	return nil
}
