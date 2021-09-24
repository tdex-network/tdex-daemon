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
	printDeprecatedWarn("tdex balance fee")
	return nil
}

func printDeprecatedWarn(newCmd string) {
	colorYellow := "\033[33m"
	fmt.Println(fmt.Sprintf(
		"%sWarning: this command is deprecated and will be removed in the next "+
			"version.\nInstead, use the new command '%s'", string(colorYellow), newCmd,
	))
}
