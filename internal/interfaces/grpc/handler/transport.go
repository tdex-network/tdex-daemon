package grpchandler

import (
	"context"

	tdexv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex/v2"
)

type transportHandler struct{}

func NewTransportHandler() tdexv2.TransportServiceServer {
	return newTransportHandler()
}

func newTransportHandler() *transportHandler {
	return &transportHandler{}
}

func (t transportHandler) SupportedContentTypes(
	context.Context, *tdexv2.SupportedContentTypesRequest,
) (*tdexv2.SupportedContentTypesResponse, error) {
	return &tdexv2.SupportedContentTypesResponse{
		AcceptedTypes: []tdexv2.ContentType{
			tdexv2.ContentType_CONTENT_TYPE_JSON,
			tdexv2.ContentType_CONTENT_TYPE_GRPC,
			tdexv2.ContentType_CONTENT_TYPE_GRPCWEB,
			tdexv2.ContentType_CONTENT_TYPE_GRPCWEBTEXT,
		},
	}, nil
}
