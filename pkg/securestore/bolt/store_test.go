package boltsecurestore_test

import (
	"os"
	"securestore"
	boltsecurestore "securestore/bolt"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	password = []byte("password")
)

func TestNewSecureStore(t *testing.T) {
	store, clean, err := newTestStore()
	require.NoError(t, err)

	t.Cleanup(clean)
	require.NotNil(t, store)
}

func TestCreateUnlock(t *testing.T) {
	store, clean, err := newTestStore()
	require.NoError(t, err)

	t.Cleanup(clean)

	_, err = store.GetAllFromBucket(nil)
	require.EqualError(t, err, boltsecurestore.ErrStoreLocked.Error())

	err = store.CreateUnlock(&password)
	require.NoError(t, err)

	// ensures that the securestore does nothing if already unlocked.
	err = store.CreateUnlock(&password)
	require.NoError(t, err)

	_, err = store.GetAllFromBucket(nil)
	require.NoError(t, err)
}

func TestFailingCreate(t *testing.T) {
	store, clean, err := newTestStore()
	require.NoError(t, err)

	t.Cleanup(clean)

	tests := []struct {
		name        string
		password    *[]byte
		expectedErr error
	}{
		{
			name:        "missing password",
			password:    nil,
			expectedErr: boltsecurestore.ErrPasswordRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.CreateUnlock(tt.password)
			require.EqualError(t, err, tt.expectedErr.Error())
		})
	}
}

func TestFailingUnlock(t *testing.T) {
	store, clean, err := newTestStoreUnlocked()
	require.NoError(t, err)

	t.Cleanup(clean)

	store.Lock()

	tests := []struct {
		name        string
		password    []byte
		expectedErr error
	}{
		{
			name:        "missing password",
			password:    nil,
			expectedErr: boltsecurestore.ErrPasswordRequired,
		},
		{
			name:        "wrong password",
			password:    []byte("wrongpassword"),
			expectedErr: boltsecurestore.ErrInvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pwd *[]byte
			if tt.password != nil {
				pwd = &tt.password
			}
			err := store.CreateUnlock(pwd)
			require.EqualError(t, err, tt.expectedErr.Error())
		})
	}
}

func TestAddToGetFromBucket(t *testing.T) {
	store, clean, err := newTestStoreUnlocked()
	require.NoError(t, err)

	t.Cleanup(clean)

	key := []byte("key")
	value := []byte("value")
	err = store.AddToBucket(nil, key, value)
	require.NoError(t, err)

	t.Run("data found", func(t *testing.T) {
		val, err := store.GetFromBucket(nil, key)
		require.NoError(t, err)
		require.Equal(t, value, val)
	})
	t.Run("data not found", func(t *testing.T) {
		val, err := store.GetFromBucket(nil, []byte("notfound"))
		require.NoError(t, err)
		require.Nil(t, val)
	})
}

func TestFailingAddToBucket(t *testing.T) {
	store, clean, err := newTestStoreUnlocked()
	require.NoError(t, err)

	t.Cleanup(clean)

	tests := []struct {
		name        string
		bucketKey   []byte
		key         []byte
		value       []byte
		expectedErr error
	}{
		{
			name:        "missing bucket",
			bucketKey:   []byte("test"),
			key:         []byte("test"),
			value:       []byte("test"),
			expectedErr: boltsecurestore.ErrBucketNotFound,
		},
		{
			name:        "missing data key",
			bucketKey:   nil,
			key:         nil,
			value:       []byte("test"),
			expectedErr: boltsecurestore.ErrMissingDataKey,
		},
		{
			name:        "forbidden data key",
			bucketKey:   nil,
			key:         []byte("enckey"),
			value:       nil,
			expectedErr: boltsecurestore.ErrForbiddenDataKey,
		},
		{
			name:        "missing data",
			bucketKey:   nil,
			key:         []byte("test"),
			value:       nil,
			expectedErr: boltsecurestore.ErrMissingData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.AddToBucket(tt.bucketKey, tt.key, tt.value)
			require.EqualError(t, err, tt.expectedErr.Error())
		})
	}

	t.Run("store locked", func(t *testing.T) {
		store.Lock()

		key := []byte("test")
		value := []byte("test")
		err := store.AddToBucket(nil, key, value)
		require.EqualError(t, err, boltsecurestore.ErrStoreLocked.Error())
	})
}

func TestFailingGetFromBucket(t *testing.T) {
	store, clean, err := newTestStoreUnlocked()
	require.NoError(t, err)

	t.Cleanup(clean)

	tests := []struct {
		name        string
		bucketKey   []byte
		key         []byte
		expectedErr error
	}{
		{
			name:        "missing bucket",
			bucketKey:   []byte("test"),
			key:         []byte("test"),
			expectedErr: boltsecurestore.ErrBucketNotFound,
		},
		{
			name:        "missing data key",
			bucketKey:   nil,
			key:         nil,
			expectedErr: boltsecurestore.ErrMissingDataKey,
		},
		{
			name:        "forbidden data key",
			bucketKey:   nil,
			key:         []byte("enckey"),
			expectedErr: boltsecurestore.ErrForbiddenDataKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := store.GetFromBucket(tt.bucketKey, tt.key)
			require.EqualError(t, err, tt.expectedErr.Error())
			require.Nil(t, value)
		})
	}

	t.Run("store locked", func(t *testing.T) {
		store.Lock()

		value, err := store.GetFromBucket(nil, []byte("test"))
		require.EqualError(t, err, boltsecurestore.ErrStoreLocked.Error())
		require.Nil(t, value)
	})
}
func TestCreateListBucket(t *testing.T) {
	store, clean, err := newTestStoreUnlocked()
	require.NoError(t, err)

	t.Cleanup(clean)

	bucketKey := []byte("bucketkey")
	err = store.CreateBucket(bucketKey)
	require.NoError(t, err)

	bucketKeys, err := store.ListBuckets()
	require.NoError(t, err)
	require.Len(t, bucketKeys, 1)
}

func TestFailingCreateBucket(t *testing.T) {
	store, clean, err := newTestStoreUnlocked()
	require.NoError(t, err)

	t.Cleanup(clean)

	tests := []struct {
		name        string
		bucketKey   []byte
		expectedErr error
	}{
		{
			name:        "missing bucket key",
			bucketKey:   nil,
			expectedErr: boltsecurestore.ErrMissingBucketKey,
		},
		{
			name:        "forbidden bucket key",
			bucketKey:   []byte("enckey"),
			expectedErr: boltsecurestore.ErrForbiddenBucketKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.CreateBucket(tt.bucketKey)
			require.EqualError(t, err, tt.expectedErr.Error())
		})
	}

	t.Run("store locked", func(t *testing.T) {
		store.Lock()

		err := store.CreateBucket([]byte("test"))
		require.EqualError(t, err, boltsecurestore.ErrStoreLocked.Error())
	})
}

func TestFailingListBuckets(t *testing.T) {
	store, clean, err := newTestStoreUnlocked()
	require.NoError(t, err)

	t.Cleanup(clean)

	t.Run("store locked", func(t *testing.T) {
		store.Lock()

		buckets, err := store.ListBuckets()
		require.EqualError(t, err, boltsecurestore.ErrStoreLocked.Error())
		require.Nil(t, buckets)
	})
}

func TestRemoveFromBucket(t *testing.T) {
	store, clean, err := newTestStoreUnlocked()
	require.NoError(t, err)

	t.Cleanup(clean)

	// populate the db with some key/value and nested bucket
	dataKey := []byte("test")
	dataValue := []byte("value")
	bucketKey := []byte("nested")

	err = store.AddToBucket(nil, dataKey, dataValue)
	require.NoError(t, err)
	err = store.CreateBucket(bucketKey)
	require.NoError(t, err)
	err = store.AddToBucket(bucketKey, dataKey, dataValue)
	require.NoError(t, err)

	err = store.RemoveFromBucket(nil, dataKey)
	require.NoError(t, err)
	err = store.RemoveFromBucket(bucketKey, dataKey)
	require.NoError(t, err)

	_, err = store.GetFromBucket(nil, dataKey)
	require.EqualError(t, err, boltsecurestore.ErrDataNotFound.Error())
	_, err = store.GetFromBucket(bucketKey, dataKey)
	require.EqualError(t, err, boltsecurestore.ErrDataNotFound.Error())
}

func TestFailingRemoveFromBucket(t *testing.T) {
	store, clean, err := newTestStoreUnlocked()
	require.NoError(t, err)

	t.Cleanup(clean)

	tests := []struct {
		name        string
		bucketKey   []byte
		key         []byte
		expectedErr error
	}{
		{
			name:        "missing data key",
			bucketKey:   nil,
			key:         nil,
			expectedErr: boltsecurestore.ErrMissingDataKey,
		},
		{
			name:        "forbidden data key",
			bucketKey:   nil,
			key:         []byte("enckey"),
			expectedErr: boltsecurestore.ErrForbiddenDataKey,
		},
		{
			name:        "missing bucket",
			bucketKey:   []byte("test"),
			key:         []byte("test"),
			expectedErr: boltsecurestore.ErrBucketNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.RemoveFromBucket(tt.bucketKey, tt.key)
			require.EqualError(t, err, tt.expectedErr.Error())
		})
	}

	t.Run("store locked", func(t *testing.T) {
		store.Lock()
		err := store.RemoveFromBucket(nil, []byte("test"))
		require.EqualError(t, err, boltsecurestore.ErrStoreLocked.Error())
	})
}

func TestRemoveBucket(t *testing.T) {
	store, clean, err := newTestStoreUnlocked()
	require.NoError(t, err)

	t.Cleanup(clean)

	// populate the db with a non-empty nested bucket
	dataKey := []byte("test")
	dataValue := []byte("value")
	bucketKey := []byte("nested")

	err = store.CreateBucket(bucketKey)
	require.NoError(t, err)
	err = store.AddToBucket(bucketKey, dataKey, dataValue)
	require.NoError(t, err)

	err = store.RemoveBucket(bucketKey)
	require.NoError(t, err)

	_, err = store.GetFromBucket(bucketKey, dataKey)
	require.EqualError(t, err, boltsecurestore.ErrBucketNotFound.Error())
}

func TestFailingRemoveBucket(t *testing.T) {
	store, clean, err := newTestStoreUnlocked()
	require.NoError(t, err)

	t.Cleanup(clean)

	tests := []struct {
		name        string
		bucketKey   []byte
		expectedErr error
	}{
		{
			name:        "missing bucket key",
			bucketKey:   nil,
			expectedErr: boltsecurestore.ErrMissingBucketKey,
		},
		{
			name:        "forbidden bucket key",
			bucketKey:   []byte("enckey"),
			expectedErr: boltsecurestore.ErrForbiddenBucketKey,
		},
		{
			name:        "missing bucket",
			bucketKey:   []byte("test"),
			expectedErr: boltsecurestore.ErrBucketNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.RemoveBucket(tt.bucketKey)
			require.EqualError(t, err, tt.expectedErr.Error())
		})
	}
}

func TestChangePassword(t *testing.T) {
	store, clean, err := newTestStoreUnlocked()
	require.NoError(t, err)

	t.Cleanup(clean)

	dataKey := []byte("toplevel")
	dataValue := []byte("value")

	bucketKey := []byte("nested")
	nestedDataKey := []byte("key")
	nestedDataValue := []byte("value")

	// populate the db with some entries and nested buckets
	err = store.AddToBucket(nil, dataKey, dataValue)
	require.NoError(t, err)

	err = store.CreateBucket(bucketKey)
	require.NoError(t, err)

	err = store.AddToBucket(bucketKey, nestedDataKey, nestedDataValue)
	require.NoError(t, err)

	val, err := store.GetFromBucket(nil, dataKey)
	require.NoError(t, err)
	require.Equal(t, dataValue, val)

	nestedVal, err := store.GetFromBucket(bucketKey, nestedDataKey)
	require.NoError(t, err)
	require.Equal(t, nestedDataValue, nestedVal)

	password := []byte("password")
	newPassword := []byte("newpassword")

	err = store.ChangePassword(password, newPassword)
	require.NoError(t, err)

	val, err = store.GetFromBucket(nil, dataKey)
	require.NoError(t, err)
	require.Equal(t, dataValue, val)

	nestedVal, err = store.GetFromBucket(bucketKey, nestedDataKey)
	require.NoError(t, err)
	require.Equal(t, nestedDataValue, nestedVal)
}

func TestFailingChangePassword(t *testing.T) {
	store, clean, err := newTestStoreUnlocked()
	require.NoError(t, err)

	t.Cleanup(clean)

	tests := []struct {
		oldPwd      []byte
		newPwd      []byte
		expectedErr error
	}{
		{
			oldPwd:      nil,
			newPwd:      []byte("test"),
			expectedErr: boltsecurestore.ErrPasswordRequired,
		},
		{
			oldPwd:      []byte("test"),
			newPwd:      nil,
			expectedErr: boltsecurestore.ErrPasswordRequired,
		},
		{
			oldPwd:      nil,
			newPwd:      nil,
			expectedErr: boltsecurestore.ErrPasswordRequired,
		},
	}

	for _, tt := range tests {
		err := store.ChangePassword(tt.oldPwd, tt.newPwd)
		require.EqualError(t, err, tt.expectedErr.Error())
	}
}

func newTestStoreUnlocked() (securestore.SecureStorage, func(), error) {
	store, clean, err := newTestStore()
	if err != nil {
		return nil, nil, err
	}
	if err := store.CreateUnlock(&password); err != nil {
		return nil, nil, err
	}
	return store, clean, nil
}

func newTestStore() (securestore.SecureStorage, func(), error) {
	dir, filename := "test", "test.db"
	store, err := boltsecurestore.NewSecureStorage(dir, filename)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		store.Close()
		os.RemoveAll(dir)
	}
	return store, cleanup, nil
}
