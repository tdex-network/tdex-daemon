package webhookpubsub

import "github.com/tdex-network/tdex-daemon/pkg/securestore"

var (
	hooksBucket         = []byte("hooks")
	hooksByActionBucket = []byte("hooksbyaction")

	// separator equivalent character is ÿ.
	// Should be fine to use such value  since it's not used for Secret (jwt
	// base64-encoded token), nor for Endpoint (http url).
	separator = []byte{255}
)

// webhookStore is wrapper around SecureStorage that implements the
// application.PubSubStore interface.
type webhookStore struct {
	store securestore.SecureStorage
}

func (ws webhookStore) IsLocked() bool {
	return ws.store.IsLocked()
}

func (ws webhookStore) Init(password string) error {
	if err := ws.Unlock(password); err != nil {
		return err
	}
	defer ws.Lock()

	if err := ws.store.CreateBucket(hooksBucket); err != nil {
		return err
	}
	if err := ws.store.CreateBucket(hooksByActionBucket); err != nil {
		return err
	}
	return nil
}

func (ws webhookStore) Lock() {
	ws.store.Lock()
}

func (ws webhookStore) Unlock(password string) error {
	pwd := []byte(password)
	return ws.store.CreateUnlock(&pwd)
}

func (ws webhookStore) ChangePassword(oldPwd, newPwd string) error {
	return ws.store.ChangePassword([]byte(oldPwd), []byte(newPwd))
}

func (ws webhookStore) Close() error {
	return ws.store.Close()
}

func (ws webhookStore) db() securestore.SecureStorage {
	return ws.store
}
