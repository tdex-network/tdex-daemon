package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/btcsuite/btcutil"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	datadirKey  = "datadir"
	passwordKey = "password-filepath"
)

var (
	defaultDatadir = btcutil.AppDataDir("tdex-daemon", false)

	datadirFlag = pflag.String(
		datadirKey,
		defaultDatadir,
		"specify a daemon's datadir different from the default one",
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

	datadir := viper.GetString(datadirKey)
	if datadir == "" {
		return nil, fmt.Errorf("invalid flag: %s must not be null", datadirKey)
	}
	if !pathExists(datadir) {
		return nil, fmt.Errorf("invalid flag: %s must be an existing path", datadirKey)
	}
	tlsCertPath := filepath.Join(datadir, "tls", "cert.pem")

	return &fileProvider{pwdPath, tlsCertPath}, nil
}

func (fp *fileProvider) Password() ([]byte, error) {
	return ioutil.ReadFile(fp.pwdPath)
}

func (fp *fileProvider) TLSCertificate() ([]byte, error) {
	if !pathExists(fp.tlsCertPath) {
		return nil, nil
	}
	return ioutil.ReadFile(fp.tlsCertPath)
}
