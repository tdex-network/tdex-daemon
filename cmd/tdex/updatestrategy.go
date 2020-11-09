package main

import (
	"context"
	"fmt"

	pboperator "github.com/tdex-network/tdex-protobuf/generated/go/operator"
	pbtypes "github.com/tdex-network/tdex-protobuf/generated/go/types"

	"github.com/urfave/cli/v2"
)

var updatestrategy = cli.Command{
	Name:  "strategy",
	Usage: "updates the current market making strategy, either automated or pluggable market making",
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
		&cli.BoolFlag{
			Name:  "pluggable",
			Usage: "set the strategy as pluggable",
			Value: false,
		},
	},
	Action: updateStrategyAction,
}

func updateStrategyAction(ctx *cli.Context) error {
	client, cleanup, err := getOperatorClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	strategy := pboperator.StrategyType_BALANCED
	if ctx.Bool("pluggable") {
		strategy = pboperator.StrategyType_PLUGGABLE
	}

	_, err = client.UpdateMarketStrategy(
		context.Background(), &pboperator.UpdateMarketStrategyRequest{
			Market: &pbtypes.Market{
				BaseAsset:  ctx.String("base_asset"),
				QuoteAsset: ctx.String("quote_asset"),
			},
			StrategyType: strategy,
		},
	)
	if err != nil {
		return err
	}

	fmt.Println("strategy has been update")
	return nil
}
