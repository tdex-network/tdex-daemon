package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var balancefee = cli.Command{
	Name:   "balancefee",
	Usage:  "check the balance of the fee account.",
	Action: balanceFeeAccountAction,
}

func balanceFeeAccountAction(ctx *cli.Context) error {
	return fmt.Errorf(
		"this command is deprecated and will be removed in the next version.\n" +
			"Instead, use the new command 'tdex fee balance'",
	)
}
