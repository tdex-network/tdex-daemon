package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var openmarket = cli.Command{
	Name:   "open",
	Usage:  "open a market",
	Action: openMarketAction,
}

func openMarketAction(ctx *cli.Context) error {
	return fmt.Errorf(
		"this command is deprecated and will be removed in the next version.\n" +
			"Instead, use the new command 'tdex market open'",
	)
}
