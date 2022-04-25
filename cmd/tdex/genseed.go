package main

import (
	"context"
	"fmt"
	"strings"

	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v1"

	"github.com/urfave/cli/v2"
)

var genseed = cli.Command{
	Name:   "genseed",
	Usage:  "generate a mnemonic seed",
	Action: genSeedAction,
}

func genSeedAction(ctx *cli.Context) error {
	client, cleanup, err := getUnlockerClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	resp, err := client.GenSeed(
		context.Background(),
		&tdexv1.GenSeedRequest{},
	)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(strings.Join(resp.GetSeedMnemonic(), " "))

	return nil
}
