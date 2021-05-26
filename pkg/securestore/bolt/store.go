package boltsecurestore

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/btcsuite/btcwallet/snacl"
	"github.com/btcsuite/btcwallet/walletdb"
	"github.com/tdex-network/tdex-daemon/pkg/securestore"
	"github.com/tdex-network/tdex-daemon/pkg/securestore/kvdb"
)

const (
	// RootKeyLen is the length of a root key.
	RootKeyLen = 32
)

var (
	// RootKeyBucketName is the name of the root key store bucket.
	RootKeyBucketName = []byte("root")

	// encryptionKeyID is the name of the database key that stores the
	// encryption key, encrypted with a salted + hashed password. The
	// format is 32 bytes of salt, and the rest is encrypted key.
	encryptionKeyID = []byte("enckey")
)

type boltSecureStorage struct {
	db kvdb.Backend

	encKeyMtx sync.RWMutex
	encKey    *snacl.SecretKey
}

// NewSecureStorage creates a bolt instance of the SecureStorage interface.
func NewSecureStorage(datadir, filename string) (securestore.SecureStorage, error) {
	if _, err := os.Stat(datadir); os.IsNotExist(err) {
		os.Mkdir(datadir, os.ModeDir|0755)
	}

	db, err := kvdb.Create(
		kvdb.BoltBackendName,
		path.Join(datadir, filename),
		true,
		kvdb.DefaultDBTimeout,
	)
	if err != nil {
		return nil, err
	}

	// If the store's bucket doesn't exist, create it.
	if err := kvdb.Update(db, func(tx kvdb.RwTx) error {
		_, err := tx.CreateTopLevelBucket(RootKeyBucketName)
		return err
	}, func() {}); err != nil {
		return nil, err
	}

	// Return the DB wrapped in a SecureStorage object.
	return &boltSecureStorage{db: db, encKey: nil}, nil
}

// IsLocked returns whether the store is locked by checking if the encryption
// key is stored in-memory.
func (s *boltSecureStorage) IsLocked() bool {
	return s.encKey == nil
}

// Lock eventually locks the store by flushing the in-memory encryption key.
func (s *boltSecureStorage) Lock() {
	if !s.IsLocked() {
		s.encKey.Zero()
		s.encKey = nil
	}
}

// CreateUnlock sets an encryption key if one is not already set, otherwise it
// checks if the password is correct for the stored encryption key.
func (s *boltSecureStorage) CreateUnlock(password *[]byte) error {
	// Is the store is already unlocked there's nothing to do here.
	if !s.IsLocked() {
		return nil
	}

	if password == nil {
		return ErrPasswordRequired
	}

	s.encKeyMtx.Lock()
	defer s.encKeyMtx.Unlock()

	return kvdb.Update(s.db, func(tx kvdb.RwTx) error {
		bucket := tx.ReadWriteBucket(RootKeyBucketName)
		if bucket == nil {
			return ErrRootKeyBucketNotFound
		}

		dbKey := bucket.Get(encryptionKeyID)
		if len(dbKey) > 0 {
			// A key is already stored, so try to unlock with the password.
			encKey := &snacl.SecretKey{}
			if err := encKey.Unmarshal(dbKey); err != nil {
				return err
			}

			if err := encKey.DeriveKey(password); err != nil {
				return ErrInvalidPassword
			}

			s.encKey = encKey
			return nil
		}

		// The encryption key is not yet stored, so create a new one.
		encKey, err := snacl.NewSecretKey(
			password, snacl.DefaultN, snacl.DefaultR, snacl.DefaultP,
		)
		if err != nil {
			return err
		}

		if err := bucket.Put(encryptionKeyID, encKey.Marshal()); err != nil {
			return err
		}

		s.encKey = encKey
		return nil
	}, func() {})
}

// ChangePassword decrypts the store (included the root key) with the old
// password and then encrypts it again with the new password.
func (s *boltSecureStorage) ChangePassword(oldPw, newPw []byte) error {
	// The store must be already unlocked. This ensures that there already is a
	// key in the DB.
	if s.IsLocked() {
		return ErrStoreLocked
	}

	if oldPw == nil || newPw == nil {
		return ErrPasswordRequired
	}

	encKeyNew, err := snacl.NewSecretKey(
		&newPw, snacl.DefaultN, snacl.DefaultR, snacl.DefaultP,
	)
	if err != nil {
		return err
	}

	// Check that old password is correct.
	if err := kvdb.View(s.db, func(tx kvdb.RTx) error {
		bucket := tx.ReadBucket(RootKeyBucketName)
		if bucket == nil {
			return ErrRootKeyBucketNotFound
		}
		dbKey := bucket.Get(encryptionKeyID)
		// The encryption key must be present otherwise we are in the wrong
		// state to change the password.
		if len(dbKey) <= 0 {
			return ErrEncKeyNotFound
		}

		encKeyOld := &snacl.SecretKey{}
		if err := encKeyOld.Unmarshal(dbKey); err != nil {
			return err
		}

		return encKeyOld.DeriveKey(&oldPw)
	}, func() {}); err != nil {
		return err
	}

	// Efficiently decrypt DB with old password and encrypt is again with the
	// new one.
	if err := s.updateEncryptedDb(encKeyNew); err != nil {
		return err
	}

	// Finally, store the new encryption key parameters in the DB
	// as well.
	return kvdb.Update(s.db, func(tx kvdb.RwTx) error {
		bucket := tx.ReadWriteBucket(RootKeyBucketName)
		if bucket == nil {
			return ErrRootKeyBucketNotFound
		}

		if err := bucket.Put(encryptionKeyID, encKeyNew.Marshal()); err != nil {
			return err
		}

		s.encKey = encKeyNew
		return nil
	}, func() {})
}

// CreateBucket creates a nested bucket into the root one.
func (s *boltSecureStorage) CreateBucket(key []byte) error {
	if s.IsLocked() {
		return ErrStoreLocked
	}

	if len(key) <= 0 {
		return ErrMissingBucketKey
	}
	if bytes.Equal(key, encryptionKeyID) {
		return ErrForbiddenBucketKey
	}

	return kvdb.Update(s.db, func(tx kvdb.RwTx) error {
		bucket := tx.ReadWriteBucket(RootKeyBucketName)
		if bucket == nil {
			return ErrRootKeyBucketNotFound
		}
		_, err := bucket.CreateBucketIfNotExists(key)
		return err
	}, func() {})
}

// AddToBucket stores the provided data encrypted into the given bucket.
// If the bucket key is nil, the key/value entry is added to the root one.
func (s *boltSecureStorage) AddToBucket(bucketKey, key, value []byte) error {
	if s.IsLocked() {
		return ErrStoreLocked
	}

	if len(key) <= 0 {
		return ErrMissingDataKey
	}
	if bytes.Equal(key, encryptionKeyID) {
		return ErrForbiddenDataKey
	}
	if len(value) <= 0 {
		return ErrMissingData
	}

	return kvdb.Update(s.db, func(tx kvdb.RwTx) error {
		bucket := tx.ReadWriteBucket(RootKeyBucketName)
		if bucket == nil {
			return ErrRootKeyBucketNotFound
		}

		if len(bucketKey) > 0 {
			// If the bucket key is not nil, data must be added to the nested bucket.
			bucket = bucket.NestedReadWriteBucket(bucketKey)
			if bucket == nil {
				return ErrBucketNotFound
			}
		}

		// Encrypt value with encryption key.
		encryptedValue, err := s.encKey.Encrypt(value)
		if err != nil {
			return err
		}

		return bucket.Put(key, encryptedValue)
	}, func() {})
}

// GetFromBucket retrieves data for the given key and bucket. If the bucket key
// is nil, data is retrieved from the root bucket.
func (s *boltSecureStorage) GetFromBucket(bucketKey, key []byte) ([]byte, error) {
	if s.IsLocked() {
		return nil, ErrStoreLocked
	}

	if len(key) <= 0 {
		return nil, ErrMissingDataKey
	}
	if bytes.Equal(key, encryptionKeyID) {
		return nil, ErrForbiddenDataKey
	}

	var value []byte
	if err := kvdb.View(s.db, func(tx kvdb.RTx) error {
		bucket := tx.ReadBucket(RootKeyBucketName)
		if bucket == nil {
			return ErrRootKeyBucketNotFound
		}

		if len(bucketKey) > 0 {
			bucket = bucket.NestedReadBucket(bucketKey)
			if bucket == nil {
				return ErrBucketNotFound
			}
		}

		encryptedValue := bucket.Get(key)
		if len(encryptedValue) <= 0 {
			return nil
		}

		v, err := s.encKey.Decrypt(encryptedValue)
		if err != nil {
			return err
		}

		value = make([]byte, len(v))
		copy(value[:], v)
		return nil
	}, func() {}); err != nil {
		return nil, err
	}

	return value, nil
}

// GetAllFromBucket returns all data stored in the given bucket.
// If the bucket key is nil
func (s *boltSecureStorage) GetAllFromBucket(bucketKey []byte) (map[string][]byte, error) {
	if s.IsLocked() {
		return nil, ErrStoreLocked
	}

	res := make(map[string][]byte)
	if err := kvdb.View(s.db, func(tx kvdb.RTx) error {
		bucket := tx.ReadBucket(RootKeyBucketName)
		if bucket == nil {
			return ErrRootKeyBucketNotFound
		}

		if len(bucketKey) > 0 {
			bucket = bucket.NestedReadBucket(bucketKey)
			if bucket == nil {
				return walletdb.ErrBucketNotFound
			}
		}

		return bucket.ForEach(func(k, v []byte) error {
			if !bytes.Equal(k, encryptionKeyID) && v != nil {
				key := string(k)
				value, err := s.encKey.Decrypt(v)
				if err != nil {
					return err
				}
				res[key] = value
			}
			return nil
		})
	}, func() {}); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *boltSecureStorage) ListBuckets() ([][]byte, error) {
	if s.IsLocked() {
		return nil, ErrStoreLocked
	}

	var bucketKeys [][]byte
	if err := kvdb.View(s.db, func(tx walletdb.ReadTx) error {
		bucket := tx.ReadBucket(RootKeyBucketName)
		if bucket == nil {
			return ErrRootKeyBucketNotFound
		}

		return bucket.ForEach(func(key, value []byte) error {
			if value == nil {
				bucketKey := make([]byte, len(key))
				copy(bucketKey[:], key)
				bucketKeys = append(bucketKeys, bucketKey)
			}
			return nil
		})
	}, func() {}); err != nil {
		return nil, err
	}

	return bucketKeys, nil
}

// Close closes the underlying database and zeroes the encryption key stored
// in memory.
func (s *boltSecureStorage) Close() error {
	s.encKeyMtx.Lock()
	defer s.encKeyMtx.Unlock()

	s.Lock()

	return s.db.Close()
}

// RemoveFromBucket removes the entry identified by the given key for the given
// bucket. If bucket key is nil, the entry is removed from the root bucket.
func (s *boltSecureStorage) RemoveFromBucket(bucketKey, key []byte) error {
	if s.IsLocked() {
		return ErrStoreLocked
	}

	if len(key) <= 0 {
		return ErrMissingDataKey
	}
	if bytes.Equal(key, encryptionKeyID) {
		return ErrForbiddenDataKey
	}

	return kvdb.Update(s.db, func(tx kvdb.RwTx) error {
		bucket := tx.ReadWriteBucket(RootKeyBucketName)
		if bucket == nil {
			return ErrRootKeyBucketNotFound
		}

		if len(bucketKey) > 0 {
			bucket = bucket.NestedReadWriteBucket(bucketKey)
			if bucket == nil {
				return ErrBucketNotFound
			}
		}

		return bucket.Delete(key)
	}, func() {})
}

func (s *boltSecureStorage) RemoveBucket(key []byte) error {
	if s.IsLocked() {
		return ErrStoreLocked
	}

	if len(key) <= 0 {
		return ErrMissingBucketKey
	}
	if bytes.Equal(key, encryptionKeyID) {
		return ErrForbiddenBucketKey
	}

	return kvdb.Update(s.db, func(tx kvdb.RwTx) error {
		bucket := tx.ReadWriteBucket(RootKeyBucketName)
		if bucket == nil {
			return ErrRootKeyBucketNotFound
		}

		return bucket.DeleteNestedBucket(key)
	}, func() {})
}

type bucketData struct {
	bucketKey []byte
	data      map[string][]byte
	err       error
}

func (s *boltSecureStorage) updateEncryptedDb(newKey *snacl.SecretKey) error {
	buckets, err := s.ListBuckets()
	if err != nil {
		return err
	}
	// nil key is used for the top level bucket
	buckets = append(buckets, nil)

	wg := &sync.WaitGroup{}
	wg.Add(len(buckets))
	chData := make(chan bucketData)

	go func() {
		wg.Wait()
		close(chData)
	}()

	for _, bucketKey := range buckets {
		go s.getAllFromBucket(bucketKey, wg, chData)
	}

	allBucketData := make(map[string][]byte)
	for d := range chData {
		if d.err != nil {
			return err
		}

		for k, v := range d.data {
			key := fmt.Sprintf("%s_%s", string(d.bucketKey), k)
			allBucketData[key] = v
		}
	}

	wg = &sync.WaitGroup{}
	wg.Add(len(allBucketData))

	for k := range allBucketData {
		k := k
		v := allBucketData[k]
		go func() {
			defer wg.Done()
			kk := strings.Split(k, "_")
			bucketKey := []byte(kk[0])
			key := []byte(kk[1])

			kvdb.Batch(s.db, func(tx kvdb.RwTx) error {
				bucket := tx.ReadWriteBucket(RootKeyBucketName)
				if len(bucketKey) > 0 {
					bucket = bucket.NestedReadWriteBucket(bucketKey)
				}

				encryptedValue, _ := newKey.Encrypt(v)
				bucket.Put(key, encryptedValue)
				return nil
			})
		}()
	}

	wg.Wait()

	return nil
}

func (s *boltSecureStorage) getAllFromBucket(
	key []byte, wg *sync.WaitGroup, chData chan bucketData,
) {
	defer wg.Done()

	data, err := s.GetAllFromBucket(key)
	if err != nil {
		chData <- bucketData{err: err}
		return
	}
	chData <- bucketData{bucketKey: key, data: data}
}
