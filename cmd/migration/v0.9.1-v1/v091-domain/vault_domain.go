package v091domain

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math"
	"math/big"
	"runtime/debug"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/scrypt"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcutil/base58"
	"github.com/vulpemventures/go-elements/network"
)

const (
	RootPath = "m/84'/0'"
)

type Vault struct {
	EncryptedMnemonic      string
	PassphraseHash         []byte
	Accounts               map[int]*Account
	AccountAndKeyByAddress map[string]AccountAndKey
	Network                *network.Network
}

func (v Vault) IsValidPassword(password string) bool {
	return bytes.Equal(v.PassphraseHash, btcutil.Hash160([]byte(password)))
}

func (v Vault) MasterKey(password string) (*hdkeychain.ExtendedKey, error) {
	mnemonic, err := Decrypt(v.EncryptedMnemonic, password)
	if err != nil {
		return nil, err
	}

	seed := generateSeedFromMnemonic(strings.Split(string(mnemonic), " "))
	rp, err := parseRootDerivationPath(RootPath)
	if err != nil {
		return nil, err
	}

	signingMasterKey, err := generateSigningMasterKey(seed, rp)
	if err != nil {
		return nil, err
	}

	masterKey, err := hdkeychain.NewKeyFromString(
		base58.Encode(signingMasterKey),
	)
	if err != nil {
		return nil, err
	}

	return masterKey, nil
}

type AccountAndKey struct {
	AccountIndex int
	BlindingKey  []byte
}

type Account struct {
	AccountIndex           int
	LastExternalIndex      int
	LastInternalIndex      int
	DerivationPathByScript map[string]string
}

type AddressInfo struct {
	AccountIndex   int
	Address        string
	BlindingKey    []byte
	DerivationPath string
	Script         string
}

func Xpub(
	account uint32,
	masterKey *hdkeychain.ExtendedKey,
) (string, error) {
	step := account + hdkeychain.HardenedKeyStart
	extPrivKey, err := masterKey.Derive(step)
	if err != nil {
		return "", err
	}

	xpub, err := extPrivKey.Neuter()
	if err != nil {
		return "", err
	}

	return xpub.String(), nil
}

func generateSigningMasterKey(
	seed []byte, derivationPath derivationPath,
) ([]byte, error) {
	hdNode, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}
	for _, step := range derivationPath {
		hdNode, err = hdNode.Derive(step)
		if err != nil {
			return nil, err
		}
	}
	return base58.Decode(hdNode.String()), nil
}

type derivationPath []uint32

func parseRootDerivationPath(strPath string) (derivationPath, error) {
	path, err := parseDerivationPath(strPath, true)
	if err != nil {
		return nil, err
	}
	if len(path) != 2 {
		return nil, ErrInvalidRootPathLen
	}
	if path[0] < hdkeychain.HardenedKeyStart || path[1] < hdkeychain.HardenedKeyStart {
		return nil, ErrInvalidRootPath
	}
	return path, nil
}

func (path derivationPath) String() string {
	if len(path) <= 0 {
		return ""
	}

	result := "m"
	for _, component := range path {
		var hardened bool
		if component >= hdkeychain.HardenedKeyStart {
			component -= hdkeychain.HardenedKeyStart
			hardened = true
		}
		result = fmt.Sprintf("%s/%d", result, component)
		if hardened {
			result += "'"
		}
	}
	return result
}

func parseDerivationPath(
	strPath string, checkAbsolutePath bool,
) (derivationPath, error) {
	if strPath == "" {
		return nil, ErrMissingDerivationPath
	}

	elems := strings.Split(strPath, "/")
	if containsEmptyString(elems) {
		return nil, ErrMalformedDerivationPath
	}
	if checkAbsolutePath {
		if elems[0] != "m" {
			return nil, ErrRequiredAbsoluteDerivationPath
		}
	}
	if len(elems) < 2 {
		return nil, ErrMalformedDerivationPath
	}
	if strings.TrimSpace(elems[0]) == "m" {
		elems = elems[1:]
	}

	path := make(derivationPath, 0)
	for _, elem := range elems {
		elem = strings.TrimSpace(elem)
		var value uint32

		if strings.HasSuffix(elem, "'") {
			value = hdkeychain.HardenedKeyStart
			elem = strings.TrimSpace(strings.TrimSuffix(elem, "'"))
		}

		// use big int for convertion
		bigval, ok := new(big.Int).SetString(elem, 0)
		if !ok {
			return nil, fmt.Errorf("invalid elem '%s' in path", elem)
		}

		max := math.MaxUint32 - value
		if bigval.Sign() < 0 || bigval.Cmp(big.NewInt(int64(max))) > 0 {
			if value == 0 {
				return nil, fmt.Errorf("elem %v must be in range [0, %d]", bigval, max)
			}
			return nil, fmt.Errorf("elem %v must be in hardened range [0, %d]", bigval, max)
		}
		value += uint32(bigval.Uint64())

		path = append(path, value)
	}

	return path, nil
}

func containsEmptyString(composedPath []string) bool {
	for _, s := range composedPath {
		if s == "" {
			return true
		}
	}
	return false
}

func generateSeedFromMnemonic(mnemonic []string) []byte {
	m := strings.Join(mnemonic, " ")
	return bip39.NewSeed(m, "")
}

var (
	ErrMissingDerivationPath          = fmt.Errorf("missing derivation path")
	ErrInvalidRootPathLen             = fmt.Errorf(`invalid root path length, must be in the form m/purpose'/coin_type'`)
	ErrInvalidRootPath                = fmt.Errorf("root path must contain only hardended values")
	ErrRequiredAbsoluteDerivationPath = fmt.Errorf("path must be an absolute derivation starting with 'm/'")
	ErrMalformedDerivationPath        = fmt.Errorf("path must not start or end with a '/'")
)

func Decrypt(encryptedMnemonic, password string) (string, error) {
	defer debug.FreeOSMemory()

	data, _ := base64.StdEncoding.DecodeString(encryptedMnemonic)
	salt, data := data[len(data)-32:], data[:len(data)-32]

	key, _, err := deriveKey([]byte(password), salt)
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
		return "", fmt.Errorf("invalid password")
	}
	return string(plaintext), nil
}

func Encrypt(text, password string) (string, error) {
	defer debug.FreeOSMemory()

	key, salt, err := deriveKey([]byte(password), nil)
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

	ciphertext := gcm.Seal(nonce, nonce, []byte(text), nil)
	ciphertext = append(ciphertext, salt...)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func deriveKey(password, salt []byte) ([]byte, []byte, error) {
	if salt == nil {
		salt = make([]byte, 32)
		if _, err := rand.Read(salt); err != nil {
			return nil, nil, err
		}
	}
	// 2^20 = 1048576 recommended length for key-stretching
	// check the doc for other recommended values:
	// https://godoc.org/golang.org/x/crypto/scrypt
	key, err := scrypt.Key(password, salt, 1048576, 8, 1, 32)
	if err != nil {
		return nil, nil, err
	}
	return key, salt, nil
}
