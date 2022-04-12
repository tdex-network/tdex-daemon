package grpchandler

import (
	"context"

	tdexv1 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/go/tdex/v1"
)

type transportHandler struct{}

func NewTransportHandler() tdexv1.TransportServiceServer {
	return newTransportHandler()
}

func newTransportHandler() *transportHandler {
	return &transportHandler{}
}

func (t transportHandler) SupportedContentTypes(
	context.Context, *tdexv1.SupportedContentTypesRequest,
) (*tdexv1.SupportedContentTypesResponse, error) {
	return &tdexv1.SupportedContentTypesResponse{
		AcceptedTypes: []tdexv1.ContentType{
			tdexv1.ContentType_CONTENT_TYPE_JSON,
			tdexv1.ContentType_CONTENT_TYPE_GRPC,
			tdexv1.ContentType_CONTENT_TYPE_GRPCWEB,
			tdexv1.ContentType_CONTENT_TYPE_GRPCWEBTEXT,
		},
	}, nil
}
