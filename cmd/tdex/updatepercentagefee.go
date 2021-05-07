package main

import (
	"context"
	"fmt"

	pboperator "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"

	"github.com/urfave/cli/v2"
)

var updatePercentagefee = cli.Command{
	Name:  "percentagefee",
	Usage: "updates the current market percentage fee",
	Flags: []cli.Flag{
		&cli.Int64Flag{
			Name:  "basis_point",
			Usage: "set the fee basis point",
		},
	},
	Action: updatePercentageFeeAction,
}

func updatePercentageFeeAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	basisPoint := ctx.Int64("basis_point")
	req := &pboperator.UpdateMarketPercentageFeeRequest{
		Market: &pbtypes.Market{
			BaseAsset:  baseAsset,
			QuoteAsset: quoteAsset,
		},
		BasisPoint: basisPoint,
	}

	if _, err := client.UpdateMarketPercentageFee(
		context.Background(), req,
	); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("market fees have been updated")
	return nil
}
