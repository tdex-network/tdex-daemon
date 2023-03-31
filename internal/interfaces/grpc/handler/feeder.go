package grpchandler

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/core/application/feeder"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"
	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v1"
)

type feederHandler struct {
	feederSvc feeder.Service
}

func NewFeederHandler(
	feederSvc feeder.Service,
) daemonv2.FeederServiceServer {
	return &feederHandler{
		feederSvc,
	}
}

func (f *feederHandler) AddPriceFeed(
	ctx context.Context,
	req *daemonv2.AddPriceFeedRequest,
) (*daemonv2.AddPriceFeedResponse, error) {
	mkt, err := parseMarket(req.GetMarket())
	id, err := f.feederSvc.AddPriceFeed(
		ctx, mkt, req.GetSource(), req.GetTicker(),
	)
	if err != nil {
		return nil, err
	}

	return &daemonv2.AddPriceFeedResponse{
		Id: id,
	}, nil
}

func (f *feederHandler) StartPriceFeed(
	ctx context.Context,
	req *daemonv2.StartPriceFeedRequest,
) (*daemonv2.StartPriceFeedResponse, error) {
	if err := f.feederSvc.StartFeed(ctx, req.GetId()); err != nil {
		return nil, err
	}

	return &daemonv2.StartPriceFeedResponse{}, nil
}

func (f *feederHandler) StopPriceFeed(
	ctx context.Context,
	req *daemonv2.StopPriceFeedRequest,
) (*daemonv2.StopPriceFeedResponse, error) {
	if err := f.feederSvc.StopFeed(ctx, req.GetId()); err != nil {
		return nil, err
	}

	return &daemonv2.StopPriceFeedResponse{}, nil
}

func (f *feederHandler) UpdatePriceFeed(
	ctx context.Context,
	req *daemonv2.UpdatePriceFeedRequest,
) (*daemonv2.UpdatePriceFeedResponse, error) {
	if err := f.feederSvc.UpdatePriceFeed(
		ctx, req.GetId(), req.GetSource(), req.GetTicker(),
	); err != nil {
		return nil, err
	}

	return &daemonv2.UpdatePriceFeedResponse{}, nil
}

func (f *feederHandler) RemovePriceFeed(
	ctx context.Context,
	req *daemonv2.RemovePriceFeedRequest,
) (*daemonv2.RemovePriceFeedResponse, error) {
	if err := f.feederSvc.RemovePriceFeed(ctx, req.GetId()); err != nil {
		return nil, err
	}

	return &daemonv2.RemovePriceFeedResponse{}, nil
}

func (f *feederHandler) GetPriceFeed(
	ctx context.Context,
	request *daemonv2.GetPriceFeedRequest,
) (*daemonv2.GetPriceFeedResponse, error) {
	priceFeed, err := f.feederSvc.GetPriceFeed(
		ctx,
		request.GetMarket().GetBaseAsset(),
		request.GetMarket().GetQuoteAsset(),
	)
	if err != nil {
		return nil, err
	}

	return &daemonv2.GetPriceFeedResponse{
		Feed: &daemonv2.PriceFeed{
			Id: priceFeed.GetId(),
			Market: &tdexv1.Market{
				BaseAsset:  priceFeed.GetMarket().GetBaseAsset(),
				QuoteAsset: priceFeed.GetMarket().GetBaseAsset(),
			},
			Source:  priceFeed.GetSource(),
			Ticker:  priceFeed.GetTicker(),
			Started: priceFeed.IsStarted(),
		},
	}, nil
}

func (f *feederHandler) ListSupportedPriceSources(
	ctx context.Context,
	req *daemonv2.ListSupportedPriceSourcesRequest,
) (*daemonv2.ListSupportedPriceSourcesResponse, error) {
	sources := f.feederSvc.ListSources(ctx)
	return &daemonv2.ListSupportedPriceSourcesResponse{
		Sources: sources,
	}, nil
}

func (f *feederHandler) ListPriceFeeds(
	ctx context.Context,
	req *daemonv2.ListPriceFeedsRequest,
) (*daemonv2.ListPriceFeedsResponse, error) {
	priceFeeds, err := f.feederSvc.ListPriceFeeds(ctx)
	if err != nil {
		return nil, err
	}

	feeds := make([]*daemonv2.PriceFeed, len(priceFeeds))
	for i, feed := range priceFeeds {
		feeds[i] = &daemonv2.PriceFeed{
			Id: feed.GetId(),
			Market: &tdexv1.Market{
				BaseAsset:  feed.GetMarket().GetBaseAsset(),
				QuoteAsset: feed.GetMarket().GetQuoteAsset(),
			},
			Source:  feed.GetSource(),
			Ticker:  feed.GetTicker(),
			Started: feed.IsStarted(),
		}
	}

	return &daemonv2.ListPriceFeedsResponse{
		Feeds: feeds,
	}, nil
}
