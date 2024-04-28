package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Client struct {
	c       *http.Client
	apiKey  string
	baseURL string
}

func NewClient(url, apiKey string) *Client {
	return &Client{
		c:       http.DefaultClient,
		apiKey:  apiKey,
		baseURL: url,
	}
}

type FeedRequest struct {
	FeedType   string
	FID        int32
	FilterType string
	ParentURL  string
	FIDs       []int32
	Cursor     string
	Limit      int32
}

type FeedResponse struct {
	Casts []*Cast
}

func (c *Client) buildEndpoint(path string) string {
	return c.baseURL + path
}

func (c *Client) GetFeed(r FeedRequest) (*FeedResponse, error) {
	url := c.buildEndpoint(
		fmt.Sprintf("/feed?feed_type=%s&fid=%d&filter_type=%s&parent_url=%s&fids=%v&cursor=%s&limit=%d",
			r.FeedType, r.FID, r.FilterType, r.ParentURL, r.FIDs, r.Cursor, r.Limit))
	log.Println(url)
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, url, nil)
	if err != nil {
		log.Println("failed to create request: ", err)
		return nil, err
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("api_key", c.apiKey)
	log.Println("making request")
	res, err := c.c.Do(req)
	if err != nil {
		log.Println("failed to get feed: ", err)
		return nil, err
	}
	log.Println("got response: ", res.Status)

	defer res.Body.Close()
	resp := &FeedResponse{}
	err = json.NewDecoder(res.Body).Decode(resp)
	if err != nil {
		return nil, err
	}
	// d, _ := json.MarshalIndent(resp, "", "  ")
	// log.Println(string(d))
	return resp, nil
}
