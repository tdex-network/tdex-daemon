package boltsecurestore

import (
	"fmt"

	"github.com/btcsuite/btcwallet/walletdb"
)

var (
	// ErrStoreLocked specifies that the store must be unlocked to perform the
	// requested operation.
	ErrStoreLocked = fmt.Errorf("store is locked")

	// ErrPasswordRequired specifies that a password is required to create/unlock
	// the store.
	ErrPasswordRequired = fmt.Errorf("password must not be null")
	// ErrInvalidPassword is returned when trying to unlock the store with an
	// incorrect password.
	ErrInvalidPassword = fmt.Errorf("password is not valid")

	// ErrRootKeyBucketNotFound specifies that there is no root bucket which
	// can/should happen only if the store has been corrupted or was initialized
	// incorrectly.
	ErrRootKeyBucketNotFound = fmt.Errorf("root key bucket not found")
	// ErrEncKeyNotFound specifies that there was no encryption key found
	// even if one was expected to be generated.
	ErrEncKeyNotFound = fmt.Errorf("store encryption key not found")

	// ErrBucketNotFound specifies that there is no such bucket to
	// read/add data from/to.
	ErrBucketNotFound = walletdb.ErrBucketNotFound
	// ErrMissingBucketKey specifies that a bucket key is required to perform
	// the requested operation.
	ErrMissingBucketKey = fmt.Errorf("missing bucket key")
	// ErrForbiddenBucketKey is used when the bucket key uses encryptionKeyID as
	// its value.
	ErrForbiddenBucketKey = fmt.Errorf("bucket key is not allowed")

	// ErrDataNotFound specifies that no data has been found for a given key.
	ErrDataNotFound = fmt.Errorf("data not found")
	// ErrMissingDataKey specifies that a data key is required to perform the
	// requested operation.
	ErrMissingDataKey = fmt.Errorf("missing data key")
	// ErrForbiddenDataKey is used when the data key used encryptionKeyID as its
	// value.
	ErrForbiddenDataKey = fmt.Errorf("data key is not allowed")
	// ErrMissingData specifies that the data value is required to perform a
	// write operation.
	ErrMissingData = fmt.Errorf("missing data to add")
)
