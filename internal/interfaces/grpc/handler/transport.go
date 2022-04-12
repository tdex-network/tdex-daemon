package grpchandler

import (
	"context"

	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/go/tdex/v1"
)

type transportHandler struct {
	tdexv1.UnimplementedTransportServer
}

func NewTransportHandler() tdexv1.TransportServer {
	return newTransportHandler()
}

func newTransportHandler() *transportHandler {
	return &transportHandler{}
}

func (t transportHandler) SupportedContentTypes(
	context.Context,
	*tdexv1.SupportedContentTypesRequest,
) (*tdexv1.SupportedContentTypesReply, error) {
	return &tdexv1.SupportedContentTypesReply{
		AcceptedTypes: []tdexv1.ContentType{
			tdexv1.ContentType_JSON,
			tdexv1.ContentType_GRPC,
			tdexv1.ContentType_GRPCWEB,
			tdexv1.ContentType_GRPCWEBTEXT,
		},
	}, nil
}
