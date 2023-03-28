package grpchandler

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"

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
	id, err := f.feederSvc.AddPriceFeed(
		ctx,
		ports.AddPriceFeedReq{
			MarketBaseAsset:  req.GetMarket().GetBaseAsset(),
			MarketQuoteAsset: req.GetMarket().GetQuoteAsset(),
			Source:           req.GetSource(),
			Ticker:           req.GetTicker(),
		},
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
	request *daemonv2.UpdatePriceFeedRequest,
) (*daemonv2.UpdatePriceFeedResponse, error) {
	if err := f.feederSvc.UpdatePriceFeed(
		ctx,
		ports.UpdatePriceFeedReq{
			ID:     request.GetId(),
			Source: request.GetSource(),
			Ticker: request.GetTicker(),
		},
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
			Id: priceFeed.ID,
			Market: &tdexv1.Market{
				BaseAsset:  priceFeed.MarketBaseAsset,
				QuoteAsset: priceFeed.MarketQuoteAsset,
			},
			Source: priceFeed.Source,
			Ticker: priceFeed.Ticker,
			On:     false,
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
