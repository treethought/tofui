package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/treethought/tofui/config"
)

var (
	client     *Client
	clientOnce sync.Once
)

type NeynarError struct {
	message string
	status  int
	path    string
	error   error
}

func (e NeynarError) Error() string {
	if e.error != nil {
		return fmt.Sprintf("%s: %s", e.path, e.error)
	}
	if e.status != 0 {
		return fmt.Sprintf("%s %d: %s", e.path, e.status, e.message)
	}
	return fmt.Sprintf("%s: %s", e.path, e.message)
}

type Client struct {
	c              *http.Client
	apiKey         string
	baseURL        string
	clientID       string
	persistantOpts []RequestOption
}

func NewClient(cfg *config.Config) *Client {
	clientOnce.Do(func() {
		client = &Client{
			c:        http.DefaultClient,
			apiKey:   cfg.Neynar.APIKey,
			baseURL:  cfg.Neynar.BaseUrl,
			clientID: cfg.Neynar.ClientID,
		}
	})
	return client
}

func (c *Client) SetOptions(opts ...RequestOption) {
	c.persistantOpts = opts
}

func (c *Client) buildEndpoint(path string) string {
	return c.baseURL + path
}

func (c *Client) SetAPIKey(key string) {
	c.apiKey = key
}

func (c *Client) doPostRequest(ctx context.Context, path string, body io.Reader, opts ...RequestOption) (*http.Response, error) {
	url := c.buildEndpoint(path)

	log.Println("sending request to: ", url)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		log.Println("failed to create request: ", err)
		return nil, err
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("api_key", c.apiKey)
	req.Header.Add("content-type", "application/json")

	for _, opt := range c.persistantOpts {
		log.Println("applying persistant option")
		opt(req)
	}

	for _, opt := range opts {
		opt(req)
	}
	return c.c.Do(req)
}

func (c *Client) doRequest(ctx context.Context, path string, opts ...RequestOption) (*http.Response, error) {
	url := c.buildEndpoint(path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("api_key", c.apiKey)

	for _, opt := range c.persistantOpts {
		opt(req)
	}

	for _, opt := range opts {
		opt(req)
	}
	return c.c.Do(req)
}

func (c *Client) doRequestInto(ctx context.Context, path string, v interface{}, opts ...RequestOption) error {
	res, err := c.doRequest(ctx, path, opts...)
	if err != nil {
		return NeynarError{"failed to create request", 0, path, err}
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		d, _ := io.ReadAll(res.Body)
		return NeynarError{string(d), res.StatusCode, path, nil}
	}

	if err := json.NewDecoder(res.Body).Decode(v); err != nil {
		return NeynarError{"failed to decode response", res.StatusCode, path, err}
	}
	return nil
}

func (c *Client) doPostInto(ctx context.Context, path string, body interface{}, v interface{}, opts ...RequestOption) error {
	data, err := json.Marshal(body)
	if err != nil {
		return NeynarError{"failed to marshal body", 0, path, err}
	}
	log.Println("sending payload: ", string(data))

	r := bytes.NewReader(data)
	resp, err := c.doPostRequest(ctx, path, r, opts...)
	if err != nil {
		return NeynarError{"failed to create request", 0, path, err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		d, _ := io.ReadAll(resp.Body)
		return NeynarError{string(d), resp.StatusCode, path, nil}
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return NeynarError{"failed to decode response", resp.StatusCode, path, err}
	}
	return nil
}

type RequestOption func(*http.Request)

func setQueryParam(r *http.Request, key, value string) {
	q := r.URL.Query()
	q.Add(key, value)
	r.URL.RawQuery = q.Encode()
}

func WithQuery(key, value string) RequestOption {
	return func(r *http.Request) {
		setQueryParam(r, key, value)
	}
}

func WithLimit(limit int) RequestOption {
	return WithQuery("limit", fmt.Sprintf("%d", limit))
}

func WithFID(fid uint64) RequestOption {
	return WithQuery("fid", fmt.Sprintf("%d", fid))
}
