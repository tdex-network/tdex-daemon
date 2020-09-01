package wallet

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"

	"golang.org/x/crypto/scrypt"
)

// EncryptOpts is the struct given to Encrypt method
type EncryptOpts struct {
	PlainText  string
	Passphrase string
}

func (o EncryptOpts) validate() error {
	if len(o.PlainText) <= 0 {
		return ErrNullPlainText
	}
	if len(o.Passphrase) <= 0 {
		return ErrNullPassphrase
	}
	return nil
}

// Encrypt encrypts (with AES-128) a plaintext with the provided passphrase
func Encrypt(opts EncryptOpts) (string, error) {
	if err := opts.validate(); err != nil {
		return "", err
	}

	key, salt, err := DeriveKey([]byte(opts.Passphrase), nil)
	if err != nil {
		return "", err
	}

	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(opts.PlainText), nil)
	ciphertext = append(ciphertext, salt...)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptOpts is the struct given to Decrypt method
type DecryptOpts struct {
	CypherText string
	Passphrase string
}

func (o DecryptOpts) validate() error {
	if len(o.CypherText) <= 0 {
		return ErrNullCypherText
	}
	if _, err := base64.StdEncoding.DecodeString(o.CypherText); err != nil {
		return ErrInvalidCypherText
	}
	if len(o.Passphrase) <= 0 {
		return ErrNullPassphrase
	}
	return nil
}

// Decrypt decrypts (with AES-128) a cyphertext with the provided passphrase
func Decrypt(opts DecryptOpts) (string, error) {
	if err := opts.validate(); err != nil {
		return "", err
	}

	data, _ := base64.StdEncoding.DecodeString(opts.CypherText)
	salt, data := data[len(data)-32:], data[:len(data)-32]

	key, _, err := DeriveKey([]byte(opts.Passphrase), salt)
	if err != nil {
		return "", err
	}

	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return "", err
	}
	nonce, text := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, text, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// DeriveKey derives a 32 byte array key from a custom passhprase
func DeriveKey(passphrase, salt []byte) ([]byte, []byte, error) {
	if salt == nil {
		salt = make([]byte, 32)
		if _, err := rand.Read(salt); err != nil {
			return nil, nil, err
		}
	}
	// 2^20 = 1048576 recommended length for key-stretching
	// check the doc for other recommended values:
	// https://godoc.org/golang.org/x/crypto/scrypt
	key, err := scrypt.Key(passphrase, salt, 1048576, 8, 1, 32)
	if err != nil {
		return nil, nil, err
	}
	return key, salt, nil
}
