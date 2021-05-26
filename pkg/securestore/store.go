/**
*
* Copyright (C) 2015-2018 Lightning Labs and The Lightning Network Developers
*
* Permission is hereby granted, free of charge, to any person obtaining a copy
* of this software and associated documentation files (the "Software"), to deal
* in the Software without restriction, including without limitation the rights
* to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
* copies of the Software, and to permit persons to whom the Software is
* furnished to do so, subject to the following conditions:
*
* The above copyright notice and this permission notice shall be included in
* all copies or substantial portions of the Software.
*
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
* IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
* FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
* LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
* OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
* THE SOFTWARE.
*
**/

// This package takes inspiration by:
//   * github.com/lightningnetwork/lnd/macaroons
//   * github.com/lightningnetwork/lnd/kvdb
// This is an adaptation of the great work already done, with the purpose of
// extending the concept of RootKeyStorage to a more general secure key/value
// storage, ie. a DB where the value associated to a key are stored encrypted.

package securestore

// SecureStorage interface defines the methods for a key/value DB that secures
// its content by encrypting the values of the entries.
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
