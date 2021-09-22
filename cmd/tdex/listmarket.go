package main

import (
	"github.com/urfave/cli/v2"
)

var listmarket = cli.Command{
	Name:   "listmarket",
	Usage:  "list all created markets",
	Action: listMarketAction,
}

func listMarketAction(ctx *cli.Context) error {
	printDeprecatedWarn("tdex listmarkets")
	return nil
}
