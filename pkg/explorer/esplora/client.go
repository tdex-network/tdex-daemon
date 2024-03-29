package esplora

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	*http.Client
}

func NewHTTPClient(requestTimeout time.Duration) *Client {
	return &Client{&http.Client{Timeout: requestTimeout}}
}

// NewHTTPRequest function builds http call
// @param method <string>: http method
// @param url <string>: URL http to call
// @return <string>, error
func (s *Client) NewHTTPRequest(
	method, url, bodyString string,
	header map[string]string,
) (int, string, error) {
	switch method {
	case "GET":
		return s.get(url, header)
	case "LIST":
		return s.list(url, header)
	case "DELETE":
		return s.delete(url, header)
	case "POST":
		return s.post(url, bodyString, header)
	default:
		return 0, "", fmt.Errorf("verb not supported %s", method)
	}
}

func (s *Client) get(url string, header map[string]string) (int, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, "", err
	}

	for key, value := range header {
		req.Header.Set(key, value)
	}

	rs, err := s.Do(req)

	// process response
	if err != nil {
		return 0, "", err
	}
	defer rs.Body.Close()

	bodyBytes, err := io.ReadAll(rs.Body)
	if err != nil {
		return 0, "", err
	}

	return rs.StatusCode, string(bodyBytes), nil
}

func (s *Client) list(url string, header map[string]string) (int, string, error) {
	req, err := http.NewRequest("LIST", url, nil)
	if err != nil {
		return 0, "", err
	}

	for key, value := range header {
		req.Header.Set(key, value)
	}

	rs, err := s.Do(req)

	// process response
	if err != nil {
		return 0, "", err
	}
	defer rs.Body.Close()

	bodyBytes, err := io.ReadAll(rs.Body)
	if err != nil {
		return 0, "", err
	}

	return rs.StatusCode, string(bodyBytes), nil
}

func (s *Client) delete(url string, header map[string]string) (int, string, error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return 0, "", err
	}

	for key, value := range header {
		req.Header.Set(key, value)
	}

	rs, err := s.Do(req)

	// process response
	if err != nil {
		return 0, "", err
	}
	defer rs.Body.Close()

	bodyBytes, err := io.ReadAll(rs.Body)
	if err != nil {
		return 0, "", err
	}

	return rs.StatusCode, string(bodyBytes), nil
}

func (s *Client) post(url, bodyString string, header map[string]string) (int, string, error) {
	body := strings.NewReader(bodyString)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return 0, "", err
	}

	for key, value := range header {
		req.Header.Set(key, value)
	}

	rs, err := s.Do(req)
	if err != nil {
		return 0, "", errors.New("Failed to create named key request: " + err.Error())
	}
	defer rs.Body.Close()

	bodyBytes, err := io.ReadAll(rs.Body)
	if err != nil {
		return 0, "", errors.New("Failed to parse response body: " + err.Error())
	}

	return rs.StatusCode, string(bodyBytes), nil
}
