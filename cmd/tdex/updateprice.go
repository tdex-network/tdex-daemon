package main

import (
	"context"
	"fmt"

	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"

	"github.com/urfave/cli/v2"
)

var updateprice = cli.Command{
	Name:  "price",
	Usage: "updates the current market price to be used for future trades",
	Flags: []cli.Flag{
		&cli.Float64Flag{
			Name:     "base_price",
			Usage:    "set the base price",
			Required: true,
		},
		&cli.Float64Flag{
			Name:     "quote_price",
			Usage:    "set the quote price",
			Required: true,
		},
	},
	Action: updatePriceAction,
}

func updatePriceAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	_, err = client.UpdateMarketPrice(
		context.Background(), &pboperator.UpdateMarketPriceRequest{
			Market: &pbtypes.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
			Price: &pbtypes.Price{
				BasePrice:  float32(ctx.Float64("base_price")),
				QuotePrice: float32(ctx.Float64("quote_price")),
			},
		},
	)
	if err != nil {
		return err
	}

	fmt.Println("price has been updates")
	return nil
}
