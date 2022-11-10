// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: tdex-daemon/v2/wallet.proto

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

// WalletServiceClient is the client API for WalletService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type WalletServiceClient interface {
	// GenSeed is the first method that should be used to instantiate a new tdexd
	// instance. This method allows a caller to generate a new HD Wallet.
	// Once the seed is obtained and verified by the user, the InitWallet
	// method should be used to commit the newly generated seed, and create the
	// wallet.
	GenSeed(ctx context.Context, in *GenSeedRequest, opts ...grpc.CallOption) (*GenSeedResponse, error)
	// InitWallet is used when tdexd is starting up for the first time to fully
	// initialize the daemon and its internal wallet.
	// The wallet in the tdexd context is a database file on the disk that can be
	// found in the configured data directory.
	// At the very least a mnemonic and a wallet password must be provided to this
	// RPC. The latter will be used to encrypt sensitive material on disk.
	// Once initialized the wallet is locked and since the password is never stored
	// on the disk, it's required to pass it into the Unlock RPC request to be able
	// to manage the daemon for operations like depositing funds or opening a market.
	InitWallet(ctx context.Context, in *InitWalletRequest, opts ...grpc.CallOption) (WalletService_InitWalletClient, error)
	// UnlockWallet is used at startup of tdexd to provide a password to unlock
	// the wallet.
	UnlockWallet(ctx context.Context, in *UnlockWalletRequest, opts ...grpc.CallOption) (*UnlockWalletResponse, error)
	// LockWallet can be used to lock tdexd and disable any operation but those
	// provided by this service.
	LockWallet(ctx context.Context, in *LockWalletRequest, opts ...grpc.CallOption) (*LockWalletResponse, error)
	// ChangePassword changes the password of the encrypted wallet. This RPC
	// requires the internal wallet to be locked. It doesn't change the wallet state
	// in any case, therefore, like after calling InitWallet, it is required to
	// unlock the walket with UnlockWallet RPC after this operation succeeds.
	ChangePassword(ctx context.Context, in *ChangePasswordRequest, opts ...grpc.CallOption) (*ChangePasswordResponse, error)
	// GetStatus is useful for external applications interacting with tdexd to know
	// whether its ready, meaning that also the wallet, operator trade services
	// are able to serve requests.
	// Restarting tdexd or initiliazing it by restoring an existing wallet can be
	// time-expensive operations causing tdexd to not be ready until they haven't
	// finished.
	GetStatus(ctx context.Context, in *GetStatusRequest, opts ...grpc.CallOption) (*GetStatusResponse, error)
	// GetInfo returns info about the configuration and the internal wallet of tdexd.
	GetInfo(ctx context.Context, in *GetInfoRequest, opts ...grpc.CallOption) (*GetInfoResponse, error)
}

type walletServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewWalletServiceClient(cc grpc.ClientConnInterface) WalletServiceClient {
	return &walletServiceClient{cc}
}

func (c *walletServiceClient) GenSeed(ctx context.Context, in *GenSeedRequest, opts ...grpc.CallOption) (*GenSeedResponse, error) {
	out := new(GenSeedResponse)
	err := c.cc.Invoke(ctx, "/tdex_daemon.v2.WalletService/GenSeed", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *walletServiceClient) InitWallet(ctx context.Context, in *InitWalletRequest, opts ...grpc.CallOption) (WalletService_InitWalletClient, error) {
	stream, err := c.cc.NewStream(ctx, &WalletService_ServiceDesc.Streams[0], "/tdex_daemon.v2.WalletService/InitWallet", opts...)
	if err != nil {
		return nil, err
	}
	x := &walletServiceInitWalletClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type WalletService_InitWalletClient interface {
	Recv() (*InitWalletResponse, error)
	grpc.ClientStream
}

type walletServiceInitWalletClient struct {
	grpc.ClientStream
}

func (x *walletServiceInitWalletClient) Recv() (*InitWalletResponse, error) {
	m := new(InitWalletResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *walletServiceClient) UnlockWallet(ctx context.Context, in *UnlockWalletRequest, opts ...grpc.CallOption) (*UnlockWalletResponse, error) {
	out := new(UnlockWalletResponse)
	err := c.cc.Invoke(ctx, "/tdex_daemon.v2.WalletService/UnlockWallet", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *walletServiceClient) LockWallet(ctx context.Context, in *LockWalletRequest, opts ...grpc.CallOption) (*LockWalletResponse, error) {
	out := new(LockWalletResponse)
	err := c.cc.Invoke(ctx, "/tdex_daemon.v2.WalletService/LockWallet", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *walletServiceClient) ChangePassword(ctx context.Context, in *ChangePasswordRequest, opts ...grpc.CallOption) (*ChangePasswordResponse, error) {
	out := new(ChangePasswordResponse)
	err := c.cc.Invoke(ctx, "/tdex_daemon.v2.WalletService/ChangePassword", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *walletServiceClient) GetStatus(ctx context.Context, in *GetStatusRequest, opts ...grpc.CallOption) (*GetStatusResponse, error) {
	out := new(GetStatusResponse)
	err := c.cc.Invoke(ctx, "/tdex_daemon.v2.WalletService/GetStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *walletServiceClient) GetInfo(ctx context.Context, in *GetInfoRequest, opts ...grpc.CallOption) (*GetInfoResponse, error) {
	out := new(GetInfoResponse)
	err := c.cc.Invoke(ctx, "/tdex_daemon.v2.WalletService/GetInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WalletServiceServer is the server API for WalletService service.
// All implementations should embed UnimplementedWalletServiceServer
// for forward compatibility
type WalletServiceServer interface {
	// GenSeed is the first method that should be used to instantiate a new tdexd
	// instance. This method allows a caller to generate a new HD Wallet.
	// Once the seed is obtained and verified by the user, the InitWallet
	// method should be used to commit the newly generated seed, and create the
	// wallet.
	GenSeed(context.Context, *GenSeedRequest) (*GenSeedResponse, error)
	// InitWallet is used when tdexd is starting up for the first time to fully
	// initialize the daemon and its internal wallet.
	// The wallet in the tdexd context is a database file on the disk that can be
	// found in the configured data directory.
	// At the very least a mnemonic and a wallet password must be provided to this
	// RPC. The latter will be used to encrypt sensitive material on disk.
	// Once initialized the wallet is locked and since the password is never stored
	// on the disk, it's required to pass it into the Unlock RPC request to be able
	// to manage the daemon for operations like depositing funds or opening a market.
	InitWallet(*InitWalletRequest, WalletService_InitWalletServer) error
	// UnlockWallet is used at startup of tdexd to provide a password to unlock
	// the wallet.
	UnlockWallet(context.Context, *UnlockWalletRequest) (*UnlockWalletResponse, error)
	// LockWallet can be used to lock tdexd and disable any operation but those
	// provided by this service.
	LockWallet(context.Context, *LockWalletRequest) (*LockWalletResponse, error)
	// ChangePassword changes the password of the encrypted wallet. This RPC
	// requires the internal wallet to be locked. It doesn't change the wallet state
	// in any case, therefore, like after calling InitWallet, it is required to
	// unlock the walket with UnlockWallet RPC after this operation succeeds.
	ChangePassword(context.Context, *ChangePasswordRequest) (*ChangePasswordResponse, error)
	// GetStatus is useful for external applications interacting with tdexd to know
	// whether its ready, meaning that also the wallet, operator trade services
	// are able to serve requests.
	// Restarting tdexd or initiliazing it by restoring an existing wallet can be
	// time-expensive operations causing tdexd to not be ready until they haven't
	// finished.
	GetStatus(context.Context, *GetStatusRequest) (*GetStatusResponse, error)
	// GetInfo returns info about the configuration and the internal wallet of tdexd.
	GetInfo(context.Context, *GetInfoRequest) (*GetInfoResponse, error)
}

// UnimplementedWalletServiceServer should be embedded to have forward compatible implementations.
type UnimplementedWalletServiceServer struct {
}

func (UnimplementedWalletServiceServer) GenSeed(context.Context, *GenSeedRequest) (*GenSeedResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GenSeed not implemented")
}
func (UnimplementedWalletServiceServer) InitWallet(*InitWalletRequest, WalletService_InitWalletServer) error {
	return status.Errorf(codes.Unimplemented, "method InitWallet not implemented")
}
func (UnimplementedWalletServiceServer) UnlockWallet(context.Context, *UnlockWalletRequest) (*UnlockWalletResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnlockWallet not implemented")
}
func (UnimplementedWalletServiceServer) LockWallet(context.Context, *LockWalletRequest) (*LockWalletResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LockWallet not implemented")
}
func (UnimplementedWalletServiceServer) ChangePassword(context.Context, *ChangePasswordRequest) (*ChangePasswordResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangePassword not implemented")
}
func (UnimplementedWalletServiceServer) GetStatus(context.Context, *GetStatusRequest) (*GetStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStatus not implemented")
}
func (UnimplementedWalletServiceServer) GetInfo(context.Context, *GetInfoRequest) (*GetInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInfo not implemented")
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

func _WalletService_GenSeed_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GenSeedRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServiceServer).GenSeed(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tdex_daemon.v2.WalletService/GenSeed",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServiceServer).GenSeed(ctx, req.(*GenSeedRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WalletService_InitWallet_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(InitWalletRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(WalletServiceServer).InitWallet(m, &walletServiceInitWalletServer{stream})
}

type WalletService_InitWalletServer interface {
	Send(*InitWalletResponse) error
	grpc.ServerStream
}

type walletServiceInitWalletServer struct {
	grpc.ServerStream
}

func (x *walletServiceInitWalletServer) Send(m *InitWalletResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _WalletService_UnlockWallet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UnlockWalletRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServiceServer).UnlockWallet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tdex_daemon.v2.WalletService/UnlockWallet",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServiceServer).UnlockWallet(ctx, req.(*UnlockWalletRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WalletService_LockWallet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LockWalletRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServiceServer).LockWallet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tdex_daemon.v2.WalletService/LockWallet",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServiceServer).LockWallet(ctx, req.(*LockWalletRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WalletService_ChangePassword_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChangePasswordRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServiceServer).ChangePassword(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tdex_daemon.v2.WalletService/ChangePassword",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServiceServer).ChangePassword(ctx, req.(*ChangePasswordRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WalletService_GetStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServiceServer).GetStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tdex_daemon.v2.WalletService/GetStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServiceServer).GetStatus(ctx, req.(*GetStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WalletService_GetInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WalletServiceServer).GetInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tdex_daemon.v2.WalletService/GetInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WalletServiceServer).GetInfo(ctx, req.(*GetInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// WalletService_ServiceDesc is the grpc.ServiceDesc for WalletService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var WalletService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tdex_daemon.v2.WalletService",
	HandlerType: (*WalletServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GenSeed",
			Handler:    _WalletService_GenSeed_Handler,
		},
		{
			MethodName: "UnlockWallet",
			Handler:    _WalletService_UnlockWallet_Handler,
		},
		{
			MethodName: "LockWallet",
			Handler:    _WalletService_LockWallet_Handler,
		},
		{
			MethodName: "ChangePassword",
			Handler:    _WalletService_ChangePassword_Handler,
		},
		{
			MethodName: "GetStatus",
			Handler:    _WalletService_GetStatus_Handler,
		},
		{
			MethodName: "GetInfo",
			Handler:    _WalletService_GetInfo_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "InitWallet",
			Handler:       _WalletService_InitWallet_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "tdex-daemon/v2/wallet.proto",
}