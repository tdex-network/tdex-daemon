package main

import (
	"github.com/urfave/cli/v2"
)

var closemarket = cli.Command{
	Name:   "close",
	Usage:  "close a market",
	Action: closeMarketAction,
}

func closeMarketAction(ctx *cli.Context) error {
	printDeprecatedWarn("tdex market close")
	return nil
}
