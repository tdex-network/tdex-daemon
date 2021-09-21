package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var closemarket = cli.Command{
	Name:   "close",
	Usage:  "close a market",
	Action: closeMarketAction,
}

func closeMarketAction(ctx *cli.Context) error {
	return fmt.Errorf(
		"this command is deprecated and will be removed in the next version.\n" +
			"Instead, use the new command 'tdex market close'",
	)
}
