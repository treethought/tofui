package api

import (
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

func (c *Client) GetFeed(r FeedRequest) (*FeedResponse, error) {
	url := fmt.Sprintf("%s/feed?api_key=%s&fid=%d", c.baseURL, c.apiKey, r.FID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("accept", "application/json")
	res, err := c.c.Do(req)
	if err != nil {
		log.Println("failed to get feed: ", err)
		return nil, err
	}

	defer res.Body.Close()
	resp := &FeedResponse{}
	err = json.NewDecoder(res.Body).Decode(resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
