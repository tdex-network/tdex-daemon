package main

import (
	"github.com/urfave/cli/v2"
)

var openmarket = cli.Command{
	Name:   "open",
	Usage:  "open a market",
	Action: openMarketAction,
}

func openMarketAction(ctx *cli.Context) error {
	printDeprecatedWarn("tdex market open")
	return nil
}
