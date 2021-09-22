package main

import (
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
	printDeprecatedWarn("tdex market strategy")
	return nil
}
