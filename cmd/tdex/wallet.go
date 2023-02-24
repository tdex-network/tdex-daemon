package main

import (
	"github.com/urfave/cli/v2"
)

var walletAccount = cli.Command{
	Name: "wallet",
	Subcommands: []*cli.Command{
		{
			Name:   "balance",
			Action: walletBalanceAction,
		},
		{
			Name:   "receive",
			Action: walletReceiveAction,
		},
		{
			Name:   "send",
			Action: walletSendAction,
		},
	},
}

func walletBalanceAction(ctx *cli.Context) error {
	printDeprecatedWarn("")
	return nil
}

func walletReceiveAction(ctx *cli.Context) error {
	printDeprecatedWarn("")
	return nil
}

func walletSendAction(ctx *cli.Context) error {
	printDeprecatedWarn("")
	return nil
}
