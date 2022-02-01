package grpchandler

import (
	"context"

	"github.com/tdex-network/tdex-protobuf/generated/go/transport"
)

type transportHandler struct {
	transport.UnimplementedTransportServer
}

func NewTransportHandler() transport.TransportServer {
	return newTransportHandler()
}

func newTransportHandler() *transportHandler {
	return &transportHandler{}
}

func (t transportHandler) SupportedContentTypes(
	context.Context,
	*transport.SupportedContentTypesRequest,
) (*transport.SupportedContentTypesReply, error) {
	return &transport.SupportedContentTypesReply{
		AcceptedTypes: []transport.ContentType{
			//transport.ContentType_JSON,
			transport.ContentType_GRPC,
			transport.ContentType_GRPCWEB,
			transport.ContentType_GRPCWEBTEXT,
		},
	}, nil
}
