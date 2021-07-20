package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/btcsuite/btcutil"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	tlsCertKey  = "tls_cert_path"
	passwordKey = "password_path"
)

var (
	defaultDatadir     = btcutil.AppDataDir("tdex-daemon", false)
	defaultTLSCertPath = filepath.Join(defaultDatadir, "tls", "cert.pem")

	tlsCertFlag = pflag.String(
		tlsCertKey,
		defaultTLSCertPath,
		"the path of the TLS certificate file",
	)
	passwordFlag = pflag.String(
		passwordKey,
		"",
		"the path of the file containing the password to unlock the wallet",
	)
)

type fileProvider struct {
	pwdPath     string
	tlsCertPath string
}

func NewFileProvider() (provider, error) {
	pwdPath := viper.GetString(passwordKey)
	if pwdPath == "" {
		return nil, fmt.Errorf("invalid flag: %s must not be null", passwordKey)
	}
	if !pathExists(pwdPath) {
		return nil, fmt.Errorf("invalid flag: %s must be an existing path", passwordKey)
	}

	tlsCertPath := viper.GetString(tlsCertKey)
	if tlsCertPath == "" {
		return nil, fmt.Errorf("invalid flag: %s must not be null", tlsCertKey)
	}

	return &fileProvider{pwdPath, tlsCertPath}, nil
}

func (fp *fileProvider) Password() ([]byte, error) {
	pwd, err := ioutil.ReadFile(fp.pwdPath)
	if err != nil {
		return nil, err
	}
	return bytes.TrimFunc(pwd, func(r rune) bool {
		return r == 10 || r == 32
	}), nil
}

func (fp *fileProvider) TLSCertificate() ([]byte, error) {
	if !pathExists(fp.tlsCertPath) {
		return nil, nil
	}
	return ioutil.ReadFile(fp.tlsCertPath)
}
