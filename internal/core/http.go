package core

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
)

type HTTPClient struct {
	Client  *http.Client
	BaseURL string
}

func NewHTTPClient(baseURL string, isCompress bool) *HTTPClient {
	if isCompress {
		return &HTTPClient{
			Client:  http.DefaultClient,
			BaseURL: baseURL,
		}
	}

	return &HTTPClient{
		Client: &http.Client{
			Transport: &http.Transport{
				DisableCompression: true,
			},
		},
		BaseURL: baseURL,
	}
}

func (c *HTTPClient) URLRequest(t *testing.T, method, path string) (*http.Response, string) {
	req, err := http.NewRequest(method, c.BaseURL+path, nil)
	require.NoError(t, err)

	resp, err := c.Client.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func (c *HTTPClient) JSONRequest(t *testing.T, method, path, reqBody string) (*http.Response, string) {
	req, err := http.NewRequest(method, c.BaseURL+path, bytes.NewBuffer([]byte(reqBody)))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Client.Do(req)
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(body)
}
