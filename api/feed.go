package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

type FeedRequest struct {
	FeedType   string
	FID        uint64
	FilterType string
	ParentURL  string
	FIDs       []uint64
	Cursor     string
	Limit      uint64
	ViewerFID  uint64
}

type FeedResponse struct {
	Casts []*Cast
}

func (c *Client) GetFeed(r FeedRequest) (*FeedResponse, error) {
	viewer := GetSigner().FID
	if r.FID == 0 {
		r.FID = viewer
	}
	u := c.buildEndpoint("/feed")

	q := url.Values{}

	if r.FeedType != "" {
		q.Set("feed_type", r.FeedType)
	}
	if r.FilterType != "" {
		q.Set("filter_type", r.FilterType)
	}
	if r.ParentURL != "" {
		q.Set("parent_url", r.ParentURL)
	}
	if r.FIDs != nil {
		for _, fid := range r.FIDs {
			q.Add("fids", fmt.Sprintf("%d", fid))
		}
	}
	if r.FeedType == "following" {
		if r.FID == 0 {
			q.Set("fid", fmt.Sprintf("%d", viewer))
		} else {
			q.Set("fid", fmt.Sprintf("%d", r.FID))
		}
	}

	if r.Cursor != "" {
		q.Set("cursor", r.Cursor)
	}
	if r.Limit != 0 {
		q.Set("limit", fmt.Sprintf("%d", r.Limit))
	}
	q.Set("viewer_fid", fmt.Sprintf("%d", viewer))

	u += "?" + q.Encode()

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, u, nil)
	if err != nil {
		log.Println("failed to create request: ", err)
		return nil, err
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("api_key", c.apiKey)
	res, err := c.c.Do(req)
	if err != nil {
		log.Println("failed to get feed: ", err)
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		log.Println("failed to get feed: ", res.Status)
		return nil, fmt.Errorf("failed to get feed: %s", res.Status)
	}

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
