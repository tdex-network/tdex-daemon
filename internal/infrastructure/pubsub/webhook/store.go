package pubsub

import "github.com/tdex-network/tdex-daemon/pkg/securestore"

var (
	subsBucket        = []byte("subscriptions")
	subsByEventBucket = []byte("subscriptionsbyevent")

	// separator equivalent character is ÿ.
	// Should be fine to use such value  since it's not used for Secret (jwt
	// base64-encoded token), nor for Endpoint (http url).
	separator = []byte{255}
)

type store struct {
	store securestore.SecureStorage
}

func (s store) IsLocked() bool {
	return s.store.IsLocked()
}

func (s store) Init(password string) error {
	if err := s.Unlock(password); err != nil {
		return err
	}
	defer s.Lock()

	if err := s.store.CreateBucket(subsBucket); err != nil {
		return err
	}
	if err := s.store.CreateBucket(subsByEventBucket); err != nil {
		return err
	}
	return nil
}

func (s store) Lock() {
	s.store.Lock()
}

func (s store) Unlock(password string) error {
	pwd := []byte(password)
	return s.store.CreateUnlock(&pwd)
}

func (s store) ChangePassword(oldPwd, newPwd string) error {
	return s.store.ChangePassword([]byte(oldPwd), []byte(newPwd))
}

func (s store) Close() error {
	return s.store.Close()
}

func (s store) db() securestore.SecureStorage {
	return s.store
}
