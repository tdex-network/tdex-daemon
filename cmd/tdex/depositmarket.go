package main

import (
	"context"

	pboperator "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"

	"github.com/urfave/cli/v2"
)

var depositmarket = cli.Command{
	Name:  "depositmarket",
	Usage: "get a deposit address for a given market or create a new one",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "base_asset",
			Usage: "the base asset hash of an existent market",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "quote_asset",
			Usage: "the base asset hash of an existent market",
			Value: "",
		},
		&cli.IntFlag{
			Name:  "num_of_addresses",
			Usage: "the number of addresses to generate for the market",
		},
	},
	Action: depositMarketAction,
}

func depositMarketAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	numOfAddresses := ctx.Int64("num_of_addresses")
	resp, err := client.DepositMarket(
		context.Background(),
		&pboperator.DepositMarketRequest{
			Market: &pbtypes.Market{
				BaseAsset:  ctx.String("base_asset"),
				QuoteAsset: ctx.String("quote_asset"),
			},
			NumOfAddresses: numOfAddresses,
		},
	)
	if err != nil {
		return err
	}

	printRespJSON(resp)

	return nil
}
