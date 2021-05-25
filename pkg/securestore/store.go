package securestore

// SecureStorage interface defines the methods for a key/value DB that secures
// its content by encrypting the values of the pairs.
type SecureStorage interface {
	// Lock locks the DB once unlocked.
	Lock()
	// Close closes the connection to the DB.
	Close() (err error)
	// IsLocked returns whether the DB is (un)locked.
	IsLocked() (locked bool)
	// CreateUnlock creates or unlocks the DB with a password.
	CreateUnlock(password *[]byte) (err error)
	// ChangePassword allows to change the password for unlocking the DB.
	ChangePassword(oldPw, newPw []byte) (err error)
	// CreateBucket creates a nested bucket (a collection of key/value pairs).
	CreateBucket(key []byte) (err error)
	// AddToBucket adds the key/value entry to some bucket.
	AddToBucket(bucketKey, key, value []byte) (err error)
	// GetFromBucket retrieves a key/value entry from some bucket.
	GetFromBucket(bucketKey, key []byte) (value []byte, err error)
	// GetAllFromBucket retrieves all key/value pairs contained by a bucket.
	GetAllFromBucket(bucketKey []byte) (valuesByKey map[string][]byte, err error)
	// ListBuckets returns the list of all buckets in the DB.
	ListBuckets() (bucketKeys [][]byte, err error)
	// RemoveFromBucket removes a key/value pair from a bucket.
	RemoveFromBucket(bucketKey, key []byte) (err error)
	// RemoveBucket removes a bucket from the root one.
	RemoveBucket(bucketKey []byte) (err error)
}
