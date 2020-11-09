package main

import (
	"context"
	"fmt"

	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"

	"github.com/urfave/cli/v2"
)

var openmarket = cli.Command{
	Name:   "open",
	Usage:  "open a market",
	Action: openMarketAction,
}

func openMarketAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	_, err = client.OpenMarket(
		context.Background(), &pboperator.OpenMarketRequest{
			Market: &pbtypes.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
		},
	)
	if err != nil {
		return err
	}

	fmt.Println("market is open")
	return nil
}
