package ports

const AnyTopic = "*"
const UnspecifiedTopic = ""

type Subscription interface {
	Topic() string
	Id() string
	IsSecured() bool
	NotifyAt() string
}

// PubSubStore defines the methods to manage the internal store of a
// SecurePubSub service.
type PubSubStore interface {
	// Init initialize the store with an optional encryption password.
	Init(password string) error
	// IsLocked returns whether the store is locked.
	IsLocked() bool
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
	// Subscribe adds a new subscription for the requested topic.
	Subscribe(topic, endpoint, secret string) (string, error)
	// SubscribeWithID adds a subscription for the requested topic by using the
	// given id instead of assinging a new one.
	SubscribeWithID(id, topic, endpoint, secret string) (string, error)
	// Unsubscribe removes some client defined by its id for a topic.
	Unsubscribe(topic, id string) error
	// ListSubscriptionsForTopic returns the info of all clients subscribed for
	// a certain topic.
	ListSubscriptionsForTopic(topic string) []Subscription
	// Publish publishes a message for a certain topic. All clients subscribed
	// for such topic will receive the message.
	Publish(topic string, message string) error
}
