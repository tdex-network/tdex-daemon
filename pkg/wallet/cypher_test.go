package wallet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptDecrypt(t *testing.T) {
	plaintext := "super secret message"
	passphrase := "supersecurekey"

	encOpts := EncryptOpts{
		PlainText:  plaintext,
		Passphrase: passphrase,
	}
	cyphertext, err := Encrypt(encOpts)
	if err != nil {
		t.Fatal(err)
		return
	}

	decOpts := DecryptOpts{
		CypherText: cyphertext,
		Passphrase: passphrase,
	}
	revealedtext, err := Decrypt(decOpts)
	assert.Equal(t, plaintext, revealedtext)
}

func TestFailingEncrypt(t *testing.T) {
	tests := []struct {
		opts EncryptOpts
		err  error
	}{
		{
			opts: EncryptOpts{
				PlainText:  "",
				Passphrase: "supersecurekey",
			},
			err: ErrNullPlainText,
		},
		{
			opts: EncryptOpts{
				PlainText:  "super secret message",
				Passphrase: "",
			},
			err: ErrNullPassphrase,
		},
	}
	for _, tt := range tests {
		_, err := Encrypt(tt.opts)
		assert.Equal(t, tt.err, err)
	}
}

func TestFailingDecrypt(t *testing.T) {
	tests := []struct {
		opts DecryptOpts
		err  error
	}{
		{
			opts: DecryptOpts{
				CypherText: "",
				Passphrase: "supersecurekey",
			},
			err: ErrNullCypherText,
		},
		{
			opts: DecryptOpts{
				CypherText: "supersecretmessage",
				Passphrase: "supersecurekey",
			},
			err: ErrInvalidCypherText,
		},
		{
			opts: DecryptOpts{
				CypherText: "fUzjTyxipK6fGrGXTLYFCb6oFHEOtqfdJTvXM5XMBx+YbK1EgFv+1PqkmZ2A3skaIyqQ0jJjA4gzKGw/dxtK0rRKL0ud8bq8BPImQvXAaYk=",
				Passphrase: "",
			},
			err: ErrNullPassphrase,
		},
	}
	for _, tt := range tests {
		_, err := Decrypt(tt.opts)
		assert.Equal(t, tt.err, err)
	}
}
