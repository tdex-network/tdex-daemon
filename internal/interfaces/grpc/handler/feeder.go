package grpchandler

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"
)

type feederHandler struct {
	feederSvc application.FeederService
}

func NewFeederHandler(
	feederSvc application.FeederService,
) daemonv2.FeederServiceServer {
	return &feederHandler{
		feederSvc,
	}
}

func (f *feederHandler) AddPriceFeed(
	ctx context.Context, req *daemonv2.AddPriceFeedRequest,
) (*daemonv2.AddPriceFeedResponse, error) {
	mkt, err := parseMarket(req.GetMarket())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

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
	ctx context.Context, req *daemonv2.StartPriceFeedRequest,
) (*daemonv2.StartPriceFeedResponse, error) {
	id, err := parseId(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := f.feederSvc.StartPriceFeed(ctx, id); err != nil {
		return nil, err
	}

	return &daemonv2.StartPriceFeedResponse{}, nil
}

func (f *feederHandler) StopPriceFeed(
	ctx context.Context, req *daemonv2.StopPriceFeedRequest,
) (*daemonv2.StopPriceFeedResponse, error) {
	id, err := parseId(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := f.feederSvc.StopPriceFeed(ctx, id); err != nil {
		return nil, err
	}

	return &daemonv2.StopPriceFeedResponse{}, nil
}

func (f *feederHandler) UpdatePriceFeed(
	ctx context.Context, req *daemonv2.UpdatePriceFeedRequest,
) (*daemonv2.UpdatePriceFeedResponse, error) {
	id, err := parseId(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if req.GetSource() == "" && req.GetTicker() == "" {
		return nil, status.Error(
			codes.InvalidArgument, "missing source and/or ticker",
		)
	}

	if err := f.feederSvc.UpdatePriceFeed(
		ctx, id, req.GetSource(), req.GetTicker(),
	); err != nil {
		return nil, err
	}

	return &daemonv2.UpdatePriceFeedResponse{}, nil
}

func (f *feederHandler) RemovePriceFeed(
	ctx context.Context, req *daemonv2.RemovePriceFeedRequest,
) (*daemonv2.RemovePriceFeedResponse, error) {
	id, err := parseId(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := f.feederSvc.RemovePriceFeed(ctx, id); err != nil {
		return nil, err
	}

	return &daemonv2.RemovePriceFeedResponse{}, nil
}

func (f *feederHandler) GetPriceFeed(
	ctx context.Context, req *daemonv2.GetPriceFeedRequest,
) (*daemonv2.GetPriceFeedResponse, error) {
	id, err := parseId(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	priceFeed, err := f.feederSvc.GetPriceFeed(ctx, id)
	if err != nil {
		return nil, err
	}

	return &daemonv2.GetPriceFeedResponse{
		Feed: priceFeedInfo{priceFeed}.toProto(),
	}, nil
}

func (f *feederHandler) ListSupportedPriceSources(
	ctx context.Context, _ *daemonv2.ListSupportedPriceSourcesRequest,
) (*daemonv2.ListSupportedPriceSourcesResponse, error) {
	sources := f.feederSvc.ListSources(ctx)
	return &daemonv2.ListSupportedPriceSourcesResponse{
		Sources: sources,
	}, nil
}

func (f *feederHandler) ListPriceFeeds(
	ctx context.Context, _ *daemonv2.ListPriceFeedsRequest,
) (*daemonv2.ListPriceFeedsResponse, error) {
	priceFeeds, err := f.feederSvc.ListPriceFeeds(ctx)
	if err != nil {
		return nil, err
	}

	return &daemonv2.ListPriceFeedsResponse{
		Feeds: priceFeedsInfo(priceFeeds).toProto(),
	}, nil
}
