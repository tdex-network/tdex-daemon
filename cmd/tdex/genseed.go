package main

import (
	"context"
	"fmt"
	"strings"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"

	"github.com/urfave/cli/v2"
)

var genseed = cli.Command{
	Name:   "genseed",
	Usage:  "generate a mnemonic seed",
	Action: genSeedAction,
}

func genSeedAction(ctx *cli.Context) error {
	client, cleanup, err := getWalletClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	resp, err := client.GenSeed(context.Background(), &daemonv2.GenSeedRequest{})
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(strings.Join(resp.GetSeedMnemonic(), " "))

	return nil
}
