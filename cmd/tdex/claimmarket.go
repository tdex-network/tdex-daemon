package main

import (
	"context"
	"fmt"

	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"

	"github.com/urfave/cli/v2"
)

var claimmarket = cli.Command{
	Name:  "claimmarket",
	Usage: "claim deposits for a market",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "outpoints",
			Usage: "list of outpoints referring to utxos [{\"hash\": <string>, \"index\": <number>}]",
		},
	},
	Action: claimMarketAction,
}

func claimMarketAction(ctx *cli.Context) error {
	outpoints, err := parseOutpoints(ctx.String("outpoints"))
	if err != nil {
		return err
	}

	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	if _, err := client.ClaimMarketDeposit(
		context.Background(), &pboperator.ClaimMarketDepositRequest{
			Market: &pbtypes.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
			},
			Outpoints: outpoints,
		},
	); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("market is funded")

	return nil
}
