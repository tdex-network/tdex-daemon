package main

import (
	"fmt"
	"github.com/magiconair/properties/assert"
	"github.com/urfave/cli/v2"
	"testing"
)

func TestFragmentation(t *testing.T) {
	assetValuePair := AssetValuePair{
		BaseValue:  100000000,
		QuoteValue: 400000000000000,
	}

	baseFragments, quoteFragments := fragmentUnspents(assetValuePair)

	fmt.Println(baseFragments)
	fmt.Println(quoteFragments)

	baseSum := uint64(0)
	for _, v := range baseFragments {
		baseSum += v
	}

	quoteSum := uint64(0)
	for _, v := range quoteFragments {
		quoteSum += v
	}

	assert.Equal(t, baseSum, assetValuePair.BaseValue)
	assert.Equal(t, quoteSum, assetValuePair.QuoteValue)
}

func TestDepositMarketCli(t *testing.T) {

	rpc := cli.StringFlag{
		Name:  "rpcserver",
		Usage: "tdexd daemon address host:port",
		Value: "localhost:9000",
	}

	app := cli.NewApp()

	app.Version = "0.0.1"
	app.Name = "tdex operator CLI"
	app.Usage = "Command line interface for tdexd daemon operators"
	app.Flags = []cli.Flag{
		&rpc,
	}

	app.Commands = append(
		app.Commands,
		&depositmarket,
	)

	err := app.Run([]string{"", "depositmarket"})
	if err != nil {
		t.Fatal(err)
	}

}
