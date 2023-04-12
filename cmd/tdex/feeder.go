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
			removePriceFeed, infoPriceFeed, listPriceFeeds, listSources,
		},
	}
	addPriceFeed = &cli.Command{
		Name:   "add",
		Usage:  "add a new price feed",
		Action: addPriceFeedAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "source",
				Usage:    "price source to use, check 'sources' command more info",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "ticker",
				Usage:    "ticker of the market for the selected price source",
				Required: true,
			},
		},
	}
	startPriceFeed = &cli.Command{
		Name:   "start",
		Usage:  "starts price feed",
		Action: startPriceFeedAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "id",
				Usage:    "id of the price feed to start",
				Required: true,
			},
		},
	}
	stopPriceFeed = &cli.Command{
		Name:   "stop",
		Usage:  "stops price feed",
		Action: stopPriceFeedAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "id",
				Usage:    "id of the price feed to stop",
				Required: true,
			},
		},
	}
	updatePriceFeed = &cli.Command{
		Name:   "update",
		Usage:  "updates a price feed source and/or ticker",
		Action: updatePriceFeedAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "id",
				Usage:    "id of the price feed to update",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "source",
				Usage: "price source to be updated, check 'sources' command for more info",
			},
			&cli.StringFlag{
				Name:  "ticker",
				Usage: "ticker of the market to be updated",
			},
		},
	}
	removePriceFeed = &cli.Command{
		Name:   "remove",
		Usage:  "removes a price feed",
		Action: removePriceFeedAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "id",
				Usage:    "id of the price feed to remove",
				Required: true,
			},
		},
	}
	infoPriceFeed = &cli.Command{
		Name:   "info",
		Usage:  "get info about a price feed",
		Action: getPriceFeedInfoAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "id",
				Usage: "id of the price feed to retrieve info",
			},
		},
	}
	listPriceFeeds = &cli.Command{
		Name:   "list",
		Usage:  "lists all price feeds",
		Action: listPriceFeedsAction,
	}
	listSources = &cli.Command{
		Name:   "sources",
		Usage:  "lists supported price feed sources",
		Action: listSourcesAction,
	}
)

func addPriceFeedAction(ctx *cli.Context) error {
	client, cleanup, err := getFeederClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	source := ctx.String("source")
	ticker := ctx.String("ticker")

	baseAsset, quoteAsset, err := getMarketFromState()
	if err != nil {
		return err
	}

	reply, err := client.AddPriceFeed(ctx.Context, &daemonv2.AddPriceFeedRequest{
		Market: &tdexv2.Market{
			BaseAsset:  baseAsset,
			QuoteAsset: quoteAsset,
		},
		Source: source,
		Ticker: ticker,
	})
	if err != nil {
		return err
	}

	fmt.Println("")
	fmt.Println("price feed id:", reply.GetId())
	return nil
}

func startPriceFeedAction(ctx *cli.Context) error {
	client, cleanup, err := getFeederClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	id := ctx.String("id")

	if _, err := client.StartPriceFeed(
		ctx.Context, &daemonv2.StartPriceFeedRequest{Id: id},
	); err != nil {
		return err
	}

	fmt.Println("")
	fmt.Println("price feed started")
	return nil
}

func stopPriceFeedAction(ctx *cli.Context) error {
	client, cleanup, err := getFeederClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	id := ctx.String("id")

	if _, err := client.StopPriceFeed(
		ctx.Context, &daemonv2.StopPriceFeedRequest{Id: id},
	); err != nil {
		return err
	}

	fmt.Println("")
	fmt.Println("price feed stopped")
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

	if _, err := client.UpdatePriceFeed(
		ctx.Context, &daemonv2.UpdatePriceFeedRequest{
			Id:     id,
			Source: source,
			Ticker: ticker,
		},
	); err != nil {
		return err
	}

	fmt.Println("")
	fmt.Println("price feed updated")
	return nil
}

func removePriceFeedAction(ctx *cli.Context) error {
	client, cleanup, err := getFeederClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	id := ctx.String("id")

	if _, err := client.RemovePriceFeed(
		ctx.Context, &daemonv2.RemovePriceFeedRequest{Id: id},
	); err != nil {
		return err
	}

	fmt.Println("")
	fmt.Println("price feed removed")
	return nil
}

func getPriceFeedInfoAction(ctx *cli.Context) error {
	client, cleanup, err := getFeederClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	id := ctx.String("id")

	reply, err := client.GetPriceFeed(ctx.Context, &daemonv2.GetPriceFeedRequest{
		Id: id,
	})
	if err != nil {
		return err
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

	reply, err := client.ListPriceFeeds(
		ctx.Context, &daemonv2.ListPriceFeedsRequest{},
	)
	if err != nil {
		return err
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
		return err
	}

	printRespJSON(reply)
	return nil
}
