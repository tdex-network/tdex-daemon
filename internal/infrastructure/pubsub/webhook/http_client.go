package webhookpubsub

import (
	"io"
	"net/http"
	"strings"
	"time"
)

type client struct {
	*http.Client
}

func newHTTPClient(requestTimeout time.Duration) *client {
	return &client{&http.Client{Timeout: requestTimeout}}
}

func (c *client) post(url, bodyString string, header map[string]string) (int, string, error) {
	body := strings.NewReader(bodyString)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return 0, "", err
	}

	for key, value := range header {
		req.Header.Set(key, value)
	}

	return c.doRequest(req)
}

func (c *client) doRequest(req *http.Request) (int, string, error) {
	rs, err := c.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer rs.Body.Close()

	bodyBytes, err := io.ReadAll(rs.Body)
	if err != nil {
		return -1, "", err
	}
	return rs.StatusCode, string(bodyBytes), nil
}
