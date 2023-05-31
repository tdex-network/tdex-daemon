package grpchandler

import (
	"context"

	"github.com/tdex-network/tdex-daemon/internal/core/application"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"
)

type webhookHandler struct {
	webhookSvc application.OperatorService
}

// NewWebhookHandler is a constructor function returning an protobuf WebhookServer.
func NewWebhookHandler(
	webhookSvc application.OperatorService,
) daemonv2.WebhookServiceServer {
	return newWebhookHandler(webhookSvc)
}

func newWebhookHandler(
	webhookSvc application.OperatorService,
) *webhookHandler {
	return &webhookHandler{webhookSvc}
}

func (h *webhookHandler) AddWebhook(
	ctx context.Context, req *daemonv2.AddWebhookRequest,
) (*daemonv2.AddWebhookResponse, error) {
	return h.addWebhook(ctx, req)
}

func (h *webhookHandler) RemoveWebhook(
	ctx context.Context, req *daemonv2.RemoveWebhookRequest,
) (*daemonv2.RemoveWebhookResponse, error) {
	return h.removeWebhook(ctx, req)
}
func (h *webhookHandler) ListWebhooks(
	ctx context.Context, req *daemonv2.ListWebhooksRequest,
) (*daemonv2.ListWebhooksResponse, error) {
	return h.listWebhooks(ctx, req)
}

func (h *webhookHandler) addWebhook(
	ctx context.Context, req *daemonv2.AddWebhookRequest,
) (*daemonv2.AddWebhookResponse, error) {
	webhook, err := parseWebhook(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	hookID, err := h.webhookSvc.AddWebhook(ctx, webhook)
	if err != nil {
		return nil, err
	}
	return &daemonv2.AddWebhookResponse{Id: hookID}, nil
}

func (h *webhookHandler) removeWebhook(
	ctx context.Context, req *daemonv2.RemoveWebhookRequest,
) (*daemonv2.RemoveWebhookResponse, error) {
	if err := h.webhookSvc.RemoveWebhook(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &daemonv2.RemoveWebhookResponse{}, nil
}

func (h *webhookHandler) listWebhooks(
	ctx context.Context, req *daemonv2.ListWebhooksRequest,
) (*daemonv2.ListWebhooksResponse, error) {
	event, err := parseWebhookEvent(req.GetEvent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	hooks, err := h.webhookSvc.ListWebhooks(ctx, event)
	if err != nil {
		return nil, err
	}
	return &daemonv2.ListWebhooksResponse{
		WebhookInfo: webhooksInfo(hooks).toProto(),
	}, nil
}
