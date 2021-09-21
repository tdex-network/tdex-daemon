package main

import (
	"fmt"

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
	return fmt.Errorf(
		"this command is deprecated and will be removed in the next version.\n" +
			"Instead, use the new command 'tdex market strategy'",
	)
}
