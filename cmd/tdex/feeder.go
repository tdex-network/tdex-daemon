package main

import (
	"fmt"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"
	tdexv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v2"

	"github.com/urfave/cli/v2"
)

var (
	feeder = cli.Command{
		Name:  "feeder",
		Usage: "manage price feeds",
		Subcommands: []*cli.Command{
			addPriceFeed, startPriceFeed, stopPriceFeed, updatePriceFeed,
			removePriceFeed, getPriceFeed, listPriceFeeds, listSources,
		},
	}
	addPriceFeed = &cli.Command{
		Name:   "add",
		Usage:  "add a new price feed",
		Action: addPriceFeedAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "base_asset",
				Usage: "base asset of the market for which the price feed is created",
			},
			&cli.StringFlag{
				Name:  "quote_asset",
				Usage: "quote asset of the market for which the price feed is created",
			},
			&cli.StringFlag{
				Name:  "source",
				Usage: "price source to use, e.g. kraken, bitfinex, coinbase etc",
			},
			&cli.StringFlag{
				Name:  "ticker",
				Usage: "ticker of the market, e.g. XBT/USDT, XBT/EUR etc",
			},
		},
	}
	startPriceFeed = &cli.Command{
		Name:   "start",
		Usage:  "starts price feed",
		Action: startPriceFeedAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "id",
				Usage: "id of the price feed to start",
			},
		},
	}
	stopPriceFeed = &cli.Command{
		Name:   "stop",
		Usage:  "stops price feed",
		Action: stopPriceFeedAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "id",
				Usage: "id of the price feed to start",
			},
		},
	}
	updatePriceFeed = &cli.Command{
		Name:   "update",
		Usage:  "update a price feed source and ticker",
		Action: updatePriceFeedAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "id",
				Usage: "id of the price feed to start",
			},
			&cli.StringFlag{
				Name:  "source",
				Usage: "price source to use, e.g. kraken, bitfinex, coinbase etc",
			},
			&cli.StringFlag{
				Name:  "ticker",
				Usage: "ticker of the market, e.g. XBT/USDT, XBT/EUR etc",
			},
		},
	}
	removePriceFeed = &cli.Command{
		Name:   "remove",
		Usage:  "remove a price feed",
		Action: removePriceFeedAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "id",
				Usage: "id of the price feed to start",
			},
		},
	}
	getPriceFeed = &cli.Command{
		Name:   "get",
		Usage:  "get a price feed",
		Action: getPriceFeedAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "base_asset",
				Usage: "base asset of the market for which the price feed is created",
			},
			&cli.StringFlag{
				Name:  "quote_asset",
				Usage: "quote asset of the market for which the price feed is created",
			},
		},
	}
	listPriceFeeds = &cli.Command{
		Name:   "list",
		Usage:  "list price feeds",
		Action: listPriceFeedsAction,
	}
	listSources = &cli.Command{
		Name:   "sources",
		Usage:  "list price feed sources",
		Action: listSourcesAction,
	}
)

func addPriceFeedAction(ctx *cli.Context) error {
	client, cleanup, err := getFeederClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset := ctx.String("base_asset")
	quoteAsset := ctx.String("quote_asset")
	source := ctx.String("source")
	ticker := ctx.String("ticker")

	if baseAsset == "" || quoteAsset == "" || source == "" || ticker == "" {
		return cli.Exit("base_asset, quote_asset, source and ticker are required", 1)
	}

	if _, err := client.AddPriceFeed(ctx.Context, &daemonv2.AddPriceFeedRequest{
		Market: &tdexv2.Market{
			BaseAsset:  baseAsset,
			QuoteAsset: quoteAsset,
		},
		Source: source,
		Ticker: ticker,
	}); err != nil {
		return cli.Exit(err, 1)
	}

	fmt.Println("Price feed added")
	return nil
}

func startPriceFeedAction(ctx *cli.Context) error {
	client, cleanup, err := getFeederClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	id := ctx.String("id")

	if id == "" {
		return cli.Exit("id is required", 1)
	}

	if _, err := client.StartPriceFeed(ctx.Context, &daemonv2.StartPriceFeedRequest{
		Id: id,
	}); err != nil {
		return cli.Exit(err, 1)
	}

	fmt.Println("Price feed started")
	return nil
}

func stopPriceFeedAction(ctx *cli.Context) error {
	client, cleanup, err := getFeederClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	id := ctx.String("id")
	if id == "" {
		return cli.Exit("id is required", 1)
	}

	if _, err := client.StopPriceFeed(ctx.Context, &daemonv2.StopPriceFeedRequest{
		Id: id,
	}); err != nil {
		return cli.Exit(err, 1)
	}

	fmt.Println("Price feed stopped")
	return nil
}

func updatePriceFeedAction(ctx *cli.Context) error {
	client, cleanup, err := getFeederClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	id := ctx.String("id")
	source := ctx.String("source")
	ticker := ctx.String("ticker")
	if id == "" || source == "" || ticker == "" {
		return cli.Exit("id, source and ticker are required", 1)
	}

	if _, err := client.UpdatePriceFeed(ctx.Context, &daemonv2.UpdatePriceFeedRequest{
		Id:     id,
		Source: source,
		Ticker: ticker,
	}); err != nil {
		return cli.Exit(err, 1)
	}

	fmt.Println("Price feed updated")
	return nil
}

func removePriceFeedAction(ctx *cli.Context) error {
	client, cleanup, err := getFeederClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	id := ctx.String("id")
	if id == "" {
		return cli.Exit("id is required", 1)
	}

	if _, err := client.RemovePriceFeed(ctx.Context, &daemonv2.RemovePriceFeedRequest{
		Id: id,
	}); err != nil {
		return cli.Exit(err, 1)
	}

	fmt.Println("Price feed removed")
	return nil
}

func getPriceFeedAction(ctx *cli.Context) error {
	client, cleanup, err := getFeederClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	baseAsset := ctx.String("base_asset")
	quoteAsset := ctx.String("quote_asset")
	if baseAsset == "" || quoteAsset == "" {
		return cli.Exit("base_asset and quote_asset are required", 1)
	}

	reply, err := client.GetPriceFeed(ctx.Context, &daemonv2.GetPriceFeedRequest{
		Market: &tdexv2.Market{
			BaseAsset:  baseAsset,
			QuoteAsset: quoteAsset,
		},
	})
	if err != nil {
		return cli.Exit(err, 1)
	}

	printRespJSON(reply)
	return nil
}

func listPriceFeedsAction(ctx *cli.Context) error {
	client, cleanup, err := getFeederClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.ListPriceFeeds(ctx.Context, &daemonv2.ListPriceFeedsRequest{})
	if err != nil {
		return cli.Exit(err, 1)
	}

	printRespJSON(reply)
	return nil
}

func listSourcesAction(ctx *cli.Context) error {
	client, cleanup, err := getFeederClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	reply, err := client.ListSupportedPriceSources(ctx.Context,
		&daemonv2.ListSupportedPriceSourcesRequest{})
	if err != nil {
		return cli.Exit(err, 1)
	}

	printRespJSON(reply)
	return nil
}
