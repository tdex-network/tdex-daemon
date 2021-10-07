package tdexdconnect

import (
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/url"
)

// Encode encodes the given args as query parameters of the returned *url.URL.
func Encode(
	rpcServerAddr, network string, certBytes, macBytes []byte,
) (*url.URL, error) {
	u := url.URL{Scheme: "tdexdconnect", Host: rpcServerAddr}
	q := u.Query()

	block, _ := pem.Decode(certBytes)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("failed to decode PEM block containing certificate")
	}
	certificate := base64.RawURLEncoding.EncodeToString(block.Bytes)
	q.Add("cert", certificate)

	if macBytes != nil {
		macaroon := base64.RawURLEncoding.EncodeToString(macBytes)
		q.Add("macaroon", macaroon)
	}

	q.Add("net", network)
	u.RawQuery = q.Encode()
	return &u, nil
}

// EncodeToString encodes the given args into a base64 string URL.
func EncodeToString(
	rpcServerAddr, network string, certBytes, macBytes []byte,
) (string, error) {
	u, err := Encode(rpcServerAddr, network, certBytes, macBytes)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// Decode decodes a base64 string URL and returns its query parameters.
func Decode(
	connectUrl string,
) (rpcAddress, network string, certBytes, macBytes []byte, err error) {
	u, err := url.Parse(connectUrl)
	if err != nil {
		return
	}

	certificate := u.Query().Get("cert")
	cBytes, err := base64.RawURLEncoding.DecodeString(certificate)
	if err != nil {
		err = fmt.Errorf("failed to decode certificate: %s", err)
		return
	}

	macaroon := u.Query().Get("macaroon")
	mBytes, err := base64.RawURLEncoding.DecodeString(macaroon)
	if err != nil {
		err = fmt.Errorf("failed to decode macaroon: %s", err)
		return
	}

	network = u.Query().Get("net")
	rpcAddress = u.Host
	certBytes = []byte(cBytes)
	macBytes = []byte(mBytes)
	return
}
