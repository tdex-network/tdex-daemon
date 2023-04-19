// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: tdex-daemon/v1/wallet.proto

package tdex_daemonv1

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

const (
	WalletService_WalletAddress_FullMethodName = "/tdex_daemon.v1.WalletService/WalletAddress"
	WalletService_WalletBalance_FullMethodName = "/tdex_daemon.v1.WalletService/WalletBalance"
	WalletService_SendToMany_FullMethodName    = "/tdex_daemon.v1.WalletService/SendToMany"
)

// WalletServiceClient is the client API for WalletService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type WalletServiceClient interface {
	// WalletAddress returns a Liquid confidential p2wpkh address (BLECH32)
	WalletAddress(ctx context.Context, in *WalletAddressRequest, opts ...grpc.CallOption) (*WalletAddressResponse, error)
	// WalletBalance returns total unspent outputs (confirmed and unconfirmed),
	// all confirmed unspent outputs and all unconfirmed unspent outputs under
	// controll of the wallet.
	WalletBalance(ctx context.Context, in *WalletBalanceRequest, opts ...grpc.CallOption) (*WalletBalanceResponse, error)
	// SendToMany sends funds to many outputs
	SendToMany(ctx context.Context, in *SendToManyRequest, opts ...grpc.CallOption) (*SendToManyResponse, error)
}

type walletServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewWalletServiceClient(cc grpc.ClientConnInterface) WalletServiceClient {
	return &walletServiceClient{cc}
}

func (c *walletServiceClient) WalletAddress(ctx context.Context, in *WalletAddressRequest, opts ...grpc.CallOption) (*WalletAddressResponse, error) {
	out := new(WalletAddressResponse)
	err := c.cc.Invoke(ctx, WalletService_WalletAddress_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *walletServiceClient) WalletBalance(ctx context.Context, in *WalletBalanceRequest, opts ...grpc.CallOption) (*WalletBalanceResponse, error) {
	out := new(WalletBalanceResponse)
	err := c.cc.Invoke(ctx, WalletService_WalletBalance_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *walletServiceClient) SendToMany(ctx context.Context, in *SendToManyRequest, opts ...grpc.CallOption) (*SendToManyResponse, error) {
	out := new(SendToManyResponse)
	err := c.cc.Invoke(ctx, WalletService_SendToMany_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WalletServiceServer is the server API for WalletService service.
// All implementations should embed UnimplementedWalletServiceServer
// for forward compatibility
type WalletServiceServer interface {
	// WalletAddress returns a Liquid confidential p2wpkh address (BLECH32)
	WalletAddress(context.Context, *WalletAddressRequest) (*WalletAddressResponse, error)
	// WalletBalance returns total unspent outputs (confirmed and unconfirmed),
	// all confirmed unspent outputs and all unconfirmed unspent outputs under
	// controll of the wallet.
	WalletBalance(context.Context, *WalletBalanceRequest) (*WalletBalanceResponse, error)
	// SendToMany sends funds to many outputs
	SendToMany(context.Context, *SendToManyRequest) (*SendToManyResponse, error)
}

// UnimplementedWalletServiceServer should be embedded to have forward compatible implementations.
type UnimplementedWalletServiceServer struct {
}

func (UnimplementedWalletServiceServer) WalletAddress(context.Context, *WalletAddressRequest) (*WalletAddressResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WalletAddress not implemented")
}
func (UnimplementedWalletServiceServer) WalletBalance(context.Context, *WalletBalanceRequest) (*WalletBalanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WalletBalance not implemented")
}
func (UnimplementedWalletServiceServer) SendToMany(context.Context, *SendToManyRequest) (*SendToManyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendToMany not implemented")
}

// UnsafeWalletServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to WalletServiceServer will
// result in compilation errors.
type UnsafeWalletServiceServer interface {
	mustEmbedUnimplementedWalletServiceServer()
}

func RegisterWalletServiceServer(s grpc.ServiceRegistrar, srv WalletServiceServer) {
	s.RegisterService(&WalletService_ServiceDesc, srv)
}

func _WalletService_WalletAddress_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WalletAddressRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServiceServer).WalletAddress(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WalletService_WalletAddress_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServiceServer).WalletAddress(ctx, req.(*WalletAddressRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WalletService_WalletBalance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WalletBalanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServiceServer).WalletBalance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WalletService_WalletBalance_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServiceServer).WalletBalance(ctx, req.(*WalletBalanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WalletService_SendToMany_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendToManyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServiceServer).SendToMany(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WalletService_SendToMany_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServiceServer).SendToMany(ctx, req.(*SendToManyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// WalletService_ServiceDesc is the grpc.ServiceDesc for WalletService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var WalletService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tdex_daemon.v1.WalletService",
	HandlerType: (*WalletServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "WalletAddress",
			Handler:    _WalletService_WalletAddress_Handler,
		},
		{
			MethodName: "WalletBalance",
			Handler:    _WalletService_WalletBalance_Handler,
		},
		{
			MethodName: "SendToMany",
			Handler:    _WalletService_SendToMany_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "tdex-daemon/v1/wallet.proto",
}
