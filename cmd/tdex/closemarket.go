package main

import (
	"context"
	"fmt"

	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"

	"github.com/urfave/cli/v2"
)

var closemarket = cli.Command{
	Name:  "close",
	Usage: "close a market",
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
	},
	Action: closeMarketAction,
}

func closeMarketAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	_, err = client.CloseMarket(
		context.Background(), &pboperator.CloseMarketRequest{
			Market: &pbtypes.Market{
				BaseAsset:  ctx.String("base_asset"),
				QuoteAsset: ctx.String("quote_asset"),
			},
		},
	)
	if err != nil {
		return err
	}

	fmt.Println("market is closed")
	return nil
}
