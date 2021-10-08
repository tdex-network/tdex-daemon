package tdexdconnect

import (
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/url"
)

// Encode encodes the given args as query parameters of the returned *url.URL.
func Encode(
	rpcServerAddr string, certBytes, macBytes []byte,
) (*url.URL, error) {
	u := url.URL{Scheme: "tdexdconnect", Host: rpcServerAddr}
	q := u.Query()

	if len(certBytes) > 0 {
		block, _ := pem.Decode(certBytes)
		if block == nil || block.Type != "CERTIFICATE" {
			return nil, fmt.Errorf("failed to decode PEM block containing certificate")
		}
		certificate := base64.RawURLEncoding.EncodeToString(block.Bytes)
		q.Add("cert", certificate)
	}

	if len(macBytes) > 0 {
		macaroon := base64.RawURLEncoding.EncodeToString(macBytes)
		q.Add("macaroon", macaroon)
	}

	u.RawQuery = q.Encode()
	return &u, nil
}

// EncodeToString encodes the given args into a base64 string URL.
func EncodeToString(
	rpcServerAddr string, certBytes, macBytes []byte,
) (string, error) {
	u, err := Encode(rpcServerAddr, certBytes, macBytes)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// Decode decodes a base64 string URL and returns its query parameters.
func Decode(
	connectUrl string,
) (rpcAddress string, certBytes, macBytes []byte, err error) {
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

	rpcAddress = u.Host
	certBytes = cBytes
	macBytes = mBytes
	return
}
