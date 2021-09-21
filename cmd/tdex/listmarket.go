package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var listmarket = cli.Command{
	Name:   "listmarket",
	Usage:  "list all created markets",
	Action: listMarketAction,
}

func listMarketAction(ctx *cli.Context) error {
	return fmt.Errorf(
		"this command is deprecated and will be removed in the next version.\n" +
			"Instead, use the new command 'tdex listmarkets'",
	)
}
