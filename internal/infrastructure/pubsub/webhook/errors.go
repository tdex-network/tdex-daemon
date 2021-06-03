package webhookpubsub

import "errors"

var (
	// ErrNullSecureStore specifies that a SecureStorage is required.
	ErrNullSecureStore = errors.New("secure store must not be null")
	// ErrNullHTTPClient specifies that a HTTP client is required.
	ErrNullHTTPClient = errors.New("http client must not be null")
	// ErrUnknownWebhookAction specifies that the given string does not represent
	// any known action.
	ErrUnknownWebhookAction = errors.New("action is unknown")
	// ErrInvalidArgs specifies that the provided args do not properly represent a
	// Webhook.
	ErrInvalidArgs = errors.New(
		"webhook info must be serialized as a stringified JSON",
	)
	// ErrInvalidArgType specifies that the provided arg is not of the expected
	// type.
	ErrInvalidArgType = errors.New("arg type must be string")
	// ErrInvalidTopic is returned whenever attempting to subscribe to an unknown
	// topic.
	ErrInvalidTopic = errors.New("topic is invalid")
)
