package application

type Topic interface {
	Code() int
	Label() string
}

// PubSubStore defines the methods to manage the internal store of a
// SecurePubSub service.
type PubSubStore interface {
	// Init initialize the store with an optional encryption password.
	Init(password string) error
	// Lock locks the store.
	Lock()
	// UnlockStore unlocks the internal store.
	Unlock(password string) error
	// ChangePassword allows to change the encryption password.
	ChangePassword(oldPwd, newPwd string) error
	// Close should be used to gracefully close the connection with the store.
	Close() error
}

// SecurePubSub defines the methods of a pubsub service and its internal store.
// This service might need to store some data in memory or on the disk, for
// this reason along with the typical methods of a pubsub service, allows to
// manage an internal optionally encrypted storage.
type SecurePubSub interface {
	// Store returns the internal store.
	Store() PubSubStore
	// Subscribes some client for a topic.
	Subscribe(topic string, args ...interface{}) (string, error)
	// Unsubscribe removes some client defined by its id for a topic.
	Unsubscribe(topic, id string) error
	// ListSubscriptionsForTopic returns the info of all clients subscribed for
	// a certain topic.
	ListSubscriptionsForTopic(topic string) []interface{}
	// Publish publishes a message for a certain topic. All clients subscribed
	// for such topic will receive the message.
	Publish(topic string, message string) error
	// TopicsByCode returns the all the topics supported by the service mapped by their
	// code.
	TopicsByCode() map[int]Topic
	// TopicsByLabel returns the all the topics supported by the service mapped by their
	// label.
	TopicsByLabel() map[string]Topic
}
