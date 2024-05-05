package api

import (
	"net/http"
)

var (
	client *Client
)

func GetClient() *Client {
	return client
}

type Client struct {
	c       *http.Client
	apiKey  string
	baseURL string
}

func NewClient(url, apiKey string) *Client {
	client = &Client{
		c:       http.DefaultClient,
		apiKey:  apiKey,
		baseURL: url,
	}
	return client
}

func (c *Client) buildEndpoint(path string) string {
	return c.baseURL + path
}
