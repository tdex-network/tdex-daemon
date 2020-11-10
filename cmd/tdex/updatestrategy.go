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

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	strategy := pboperator.StrategyType_BALANCED
	if ctx.Bool("pluggable") {
		strategy = pboperator.StrategyType_PLUGGABLE
	}

	_, err = client.UpdateMarketStrategy(
		context.Background(), &pboperator.UpdateMarketStrategyRequest{
			Market: &pbtypes.Market{
				BaseAsset:  baseAsset,
				QuoteAsset: quoteAsset,
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
