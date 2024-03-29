// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package tdex_daemonv2

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// FeederServiceClient is the client API for FeederService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FeederServiceClient interface {
	// AddPriceFeed creates a new price feed for the given market.
	AddPriceFeed(ctx context.Context, in *AddPriceFeedRequest, opts ...grpc.CallOption) (*AddPriceFeedResponse, error)
	// StartPriceFeed starts the price feed with the given id.
	StartPriceFeed(ctx context.Context, in *StartPriceFeedRequest, opts ...grpc.CallOption) (*StartPriceFeedResponse, error)
	// StopPriceFeed stops the price feed with the given id.
	StopPriceFeed(ctx context.Context, in *StopPriceFeedRequest, opts ...grpc.CallOption) (*StopPriceFeedResponse, error)
	// UpdatePriceFeed allows to change source and/or ticker of the given price feed.
	UpdatePriceFeed(ctx context.Context, in *UpdatePriceFeedRequest, opts ...grpc.CallOption) (*UpdatePriceFeedResponse, error)
	// RemovePriceFeed removes the price feed with the given id.
	RemovePriceFeed(ctx context.Context, in *RemovePriceFeedRequest, opts ...grpc.CallOption) (*RemovePriceFeedResponse, error)
	// GetPriceFeed returns the price feed for the given market.
	GetPriceFeed(ctx context.Context, in *GetPriceFeedRequest, opts ...grpc.CallOption) (*GetPriceFeedResponse, error)
	// ListPriceFeeds returns the list of price feeds of all markets.
	ListPriceFeeds(ctx context.Context, in *ListPriceFeedsRequest, opts ...grpc.CallOption) (*ListPriceFeedsResponse, error)
	// ListSupportedPriceSources returns the list of supported price sources.
	ListSupportedPriceSources(ctx context.Context, in *ListSupportedPriceSourcesRequest, opts ...grpc.CallOption) (*ListSupportedPriceSourcesResponse, error)
}

type feederServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewFeederServiceClient(cc grpc.ClientConnInterface) FeederServiceClient {
	return &feederServiceClient{cc}
}

func (c *feederServiceClient) AddPriceFeed(ctx context.Context, in *AddPriceFeedRequest, opts ...grpc.CallOption) (*AddPriceFeedResponse, error) {
	out := new(AddPriceFeedResponse)
	err := c.cc.Invoke(ctx, "/tdex_daemon.v2.FeederService/AddPriceFeed", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *feederServiceClient) StartPriceFeed(ctx context.Context, in *StartPriceFeedRequest, opts ...grpc.CallOption) (*StartPriceFeedResponse, error) {
	out := new(StartPriceFeedResponse)
	err := c.cc.Invoke(ctx, "/tdex_daemon.v2.FeederService/StartPriceFeed", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *feederServiceClient) StopPriceFeed(ctx context.Context, in *StopPriceFeedRequest, opts ...grpc.CallOption) (*StopPriceFeedResponse, error) {
	out := new(StopPriceFeedResponse)
	err := c.cc.Invoke(ctx, "/tdex_daemon.v2.FeederService/StopPriceFeed", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *feederServiceClient) UpdatePriceFeed(ctx context.Context, in *UpdatePriceFeedRequest, opts ...grpc.CallOption) (*UpdatePriceFeedResponse, error) {
	out := new(UpdatePriceFeedResponse)
	err := c.cc.Invoke(ctx, "/tdex_daemon.v2.FeederService/UpdatePriceFeed", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *feederServiceClient) RemovePriceFeed(ctx context.Context, in *RemovePriceFeedRequest, opts ...grpc.CallOption) (*RemovePriceFeedResponse, error) {
	out := new(RemovePriceFeedResponse)
	err := c.cc.Invoke(ctx, "/tdex_daemon.v2.FeederService/RemovePriceFeed", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *feederServiceClient) GetPriceFeed(ctx context.Context, in *GetPriceFeedRequest, opts ...grpc.CallOption) (*GetPriceFeedResponse, error) {
	out := new(GetPriceFeedResponse)
	err := c.cc.Invoke(ctx, "/tdex_daemon.v2.FeederService/GetPriceFeed", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *feederServiceClient) ListPriceFeeds(ctx context.Context, in *ListPriceFeedsRequest, opts ...grpc.CallOption) (*ListPriceFeedsResponse, error) {
	out := new(ListPriceFeedsResponse)
	err := c.cc.Invoke(ctx, "/tdex_daemon.v2.FeederService/ListPriceFeeds", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *feederServiceClient) ListSupportedPriceSources(ctx context.Context, in *ListSupportedPriceSourcesRequest, opts ...grpc.CallOption) (*ListSupportedPriceSourcesResponse, error) {
	out := new(ListSupportedPriceSourcesResponse)
	err := c.cc.Invoke(ctx, "/tdex_daemon.v2.FeederService/ListSupportedPriceSources", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FeederServiceServer is the server API for FeederService service.
// All implementations should embed UnimplementedFeederServiceServer
// for forward compatibility
type FeederServiceServer interface {
	// AddPriceFeed creates a new price feed for the given market.
	AddPriceFeed(context.Context, *AddPriceFeedRequest) (*AddPriceFeedResponse, error)
	// StartPriceFeed starts the price feed with the given id.
	StartPriceFeed(context.Context, *StartPriceFeedRequest) (*StartPriceFeedResponse, error)
	// StopPriceFeed stops the price feed with the given id.
	StopPriceFeed(context.Context, *StopPriceFeedRequest) (*StopPriceFeedResponse, error)
	// UpdatePriceFeed allows to change source and/or ticker of the given price feed.
	UpdatePriceFeed(context.Context, *UpdatePriceFeedRequest) (*UpdatePriceFeedResponse, error)
	// RemovePriceFeed removes the price feed with the given id.
	RemovePriceFeed(context.Context, *RemovePriceFeedRequest) (*RemovePriceFeedResponse, error)
	// GetPriceFeed returns the price feed for the given market.
	GetPriceFeed(context.Context, *GetPriceFeedRequest) (*GetPriceFeedResponse, error)
	// ListPriceFeeds returns the list of price feeds of all markets.
	ListPriceFeeds(context.Context, *ListPriceFeedsRequest) (*ListPriceFeedsResponse, error)
	// ListSupportedPriceSources returns the list of supported price sources.
	ListSupportedPriceSources(context.Context, *ListSupportedPriceSourcesRequest) (*ListSupportedPriceSourcesResponse, error)
}

// UnimplementedFeederServiceServer should be embedded to have forward compatible implementations.
type UnimplementedFeederServiceServer struct {
}

func (UnimplementedFeederServiceServer) AddPriceFeed(context.Context, *AddPriceFeedRequest) (*AddPriceFeedResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddPriceFeed not implemented")
}
func (UnimplementedFeederServiceServer) StartPriceFeed(context.Context, *StartPriceFeedRequest) (*StartPriceFeedResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartPriceFeed not implemented")
}
func (UnimplementedFeederServiceServer) StopPriceFeed(context.Context, *StopPriceFeedRequest) (*StopPriceFeedResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StopPriceFeed not implemented")
}
func (UnimplementedFeederServiceServer) UpdatePriceFeed(context.Context, *UpdatePriceFeedRequest) (*UpdatePriceFeedResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdatePriceFeed not implemented")
}
func (UnimplementedFeederServiceServer) RemovePriceFeed(context.Context, *RemovePriceFeedRequest) (*RemovePriceFeedResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemovePriceFeed not implemented")
}
func (UnimplementedFeederServiceServer) GetPriceFeed(context.Context, *GetPriceFeedRequest) (*GetPriceFeedResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPriceFeed not implemented")
}
func (UnimplementedFeederServiceServer) ListPriceFeeds(context.Context, *ListPriceFeedsRequest) (*ListPriceFeedsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListPriceFeeds not implemented")
}
func (UnimplementedFeederServiceServer) ListSupportedPriceSources(context.Context, *ListSupportedPriceSourcesRequest) (*ListSupportedPriceSourcesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListSupportedPriceSources not implemented")
}

// UnsafeFeederServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FeederServiceServer will
// result in compilation errors.
type UnsafeFeederServiceServer interface {
	mustEmbedUnimplementedFeederServiceServer()
}

func RegisterFeederServiceServer(s grpc.ServiceRegistrar, srv FeederServiceServer) {
	s.RegisterService(&FeederService_ServiceDesc, srv)
}

func _FeederService_AddPriceFeed_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddPriceFeedRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FeederServiceServer).AddPriceFeed(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tdex_daemon.v2.FeederService/AddPriceFeed",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FeederServiceServer).AddPriceFeed(ctx, req.(*AddPriceFeedRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FeederService_StartPriceFeed_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StartPriceFeedRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FeederServiceServer).StartPriceFeed(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tdex_daemon.v2.FeederService/StartPriceFeed",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FeederServiceServer).StartPriceFeed(ctx, req.(*StartPriceFeedRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FeederService_StopPriceFeed_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopPriceFeedRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FeederServiceServer).StopPriceFeed(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tdex_daemon.v2.FeederService/StopPriceFeed",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FeederServiceServer).StopPriceFeed(ctx, req.(*StopPriceFeedRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FeederService_UpdatePriceFeed_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdatePriceFeedRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FeederServiceServer).UpdatePriceFeed(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tdex_daemon.v2.FeederService/UpdatePriceFeed",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FeederServiceServer).UpdatePriceFeed(ctx, req.(*UpdatePriceFeedRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FeederService_RemovePriceFeed_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemovePriceFeedRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FeederServiceServer).RemovePriceFeed(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tdex_daemon.v2.FeederService/RemovePriceFeed",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FeederServiceServer).RemovePriceFeed(ctx, req.(*RemovePriceFeedRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FeederService_GetPriceFeed_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPriceFeedRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FeederServiceServer).GetPriceFeed(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tdex_daemon.v2.FeederService/GetPriceFeed",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FeederServiceServer).GetPriceFeed(ctx, req.(*GetPriceFeedRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FeederService_ListPriceFeeds_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListPriceFeedsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FeederServiceServer).ListPriceFeeds(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tdex_daemon.v2.FeederService/ListPriceFeeds",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FeederServiceServer).ListPriceFeeds(ctx, req.(*ListPriceFeedsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FeederService_ListSupportedPriceSources_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListSupportedPriceSourcesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FeederServiceServer).ListSupportedPriceSources(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tdex_daemon.v2.FeederService/ListSupportedPriceSources",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FeederServiceServer).ListSupportedPriceSources(ctx, req.(*ListSupportedPriceSourcesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// FeederService_ServiceDesc is the grpc.ServiceDesc for FeederService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var FeederService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tdex_daemon.v2.FeederService",
	HandlerType: (*FeederServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddPriceFeed",
			Handler:    _FeederService_AddPriceFeed_Handler,
		},
		{
			MethodName: "StartPriceFeed",
			Handler:    _FeederService_StartPriceFeed_Handler,
		},
		{
			MethodName: "StopPriceFeed",
			Handler:    _FeederService_StopPriceFeed_Handler,
		},
		{
			MethodName: "UpdatePriceFeed",
			Handler:    _FeederService_UpdatePriceFeed_Handler,
		},
		{
			MethodName: "RemovePriceFeed",
			Handler:    _FeederService_RemovePriceFeed_Handler,
		},
		{
			MethodName: "GetPriceFeed",
			Handler:    _FeederService_GetPriceFeed_Handler,
		},
		{
			MethodName: "ListPriceFeeds",
			Handler:    _FeederService_ListPriceFeeds_Handler,
		},
		{
			MethodName: "ListSupportedPriceSources",
			Handler:    _FeederService_ListSupportedPriceSources_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "tdex-daemon/v2/feeder.proto",
}
