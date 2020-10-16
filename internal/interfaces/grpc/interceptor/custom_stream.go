package interceptor

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ServerStreamWrapper struct {
	stream grpc.ServerStream
	ctx    context.Context
}

func WrapServerStream(stream grpc.ServerStream, ctx context.Context) *ServerStreamWrapper {
	return &ServerStreamWrapper{stream: stream, ctx: ctx}
}

func (ssw *ServerStreamWrapper) SetHeader(md metadata.MD) error {
	return ssw.stream.SetHeader(md)
}

func (ssw *ServerStreamWrapper) SendHeader(md metadata.MD) error {
	return ssw.stream.SendHeader(md)
}

func (ssw *ServerStreamWrapper) SetTrailer(md metadata.MD) {
	ssw.stream.SetTrailer(md)
}

func (ssw *ServerStreamWrapper) Context() context.Context {
	return ssw.ctx
}

func (ssw *ServerStreamWrapper) SendMsg(m interface{}) error {
	return ssw.stream.SendMsg(m)
}

func (ssw *ServerStreamWrapper) RecvMsg(m interface{}) error {
	return ssw.stream.RecvMsg(m)
}
