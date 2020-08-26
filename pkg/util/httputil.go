package util

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var client = &http.Client{Timeout: 30 * time.Second}

// NewHTTPRequest function builds http call
// @param method <string>: http method
// @param url <string>: URL http to call
// @return <string>, error
func NewHTTPRequest(method string, url string, bodyString string, header map[string]string) (int, string, error) {
	switch method {
	case "GET":
		return get(url, header)
	case "LIST":
		return list(url, header)
	case "DELETE":
		return delete(url, header)
	case "POST":
		return post(url, bodyString, header)
	default:
		return 0, "", fmt.Errorf("verb not supported %s", method)
	}
}

func get(url string, header map[string]string) (int, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, "", err
	}

	for key, value := range header {
		req.Header.Set(key, value)
	}

	rs, err := client.Do(req)

	// process response
	if err != nil {
		return 0, "", err
	}
	defer rs.Body.Close()

	bodyBytes, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		return 0, "", err
	}

	return rs.StatusCode, string(bodyBytes), nil
}

func list(url string, header map[string]string) (int, string, error) {
	req, err := http.NewRequest("LIST", url, nil)
	if err != nil {
		return 0, "", err
	}

	for key, value := range header {
		req.Header.Set(key, value)
	}

	rs, err := client.Do(req)

	// process response
	if err != nil {
		return 0, "", err
	}
	defer rs.Body.Close()

	bodyBytes, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		return 0, "", err
	}

	return rs.StatusCode, string(bodyBytes), nil
}

func delete(url string, header map[string]string) (int, string, error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return 0, "", err
	}

	for key, value := range header {
		req.Header.Set(key, value)
	}

	rs, err := client.Do(req)

	// process response
	if err != nil {
		return 0, "", err
	}
	defer rs.Body.Close()

	bodyBytes, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		return 0, "", err
	}

	return rs.StatusCode, string(bodyBytes), nil
}

func post(url string, bodyString string, header map[string]string) (int, string, error) {
	body := strings.NewReader(bodyString)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return 0, "", err
	}

	for key, value := range header {
		req.Header.Set(key, value)
	}

	rs, err := client.Do(req)
	if err != nil {
		return 0, "", errors.New("Failed to create named key request: " + err.Error())
	}
	defer rs.Body.Close()

	bodyBytes, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		return 0, "", errors.New("Failed to parse response body: " + err.Error())
	}

	return rs.StatusCode, string(bodyBytes), nil
}
