package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
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
	Casts []Cast
}

type Cast struct {
	Hash         string
	ThreadHash   string
	ParentHash   string
	ParentURL    string
	ParentAuthor struct {
		FIID int32
	}
	Author struct {
		FIID        int32
		Username    string
		DisplayName string `json:"display_name"`
		PfpURL      string `json:"pfp_url"`
		Profile     struct {
			Bio struct {
				Text string
			}
		}
		FollowerCount  int32
		FollowingCount int32
		Verifications  []string
		ActiveStatus   string
	}
	Text      string
	Timestamp time.Time
	Embeds    []struct {
		URL string
	}
	Reactions struct {
		Likes   []Reaction
		Recasts []Reaction
	}
	// Replies struct {
	//    Count int32 `json:"count",string`
	//  }
}

func (c Cast) HumanTime() string {
	return c.Timestamp.Format("Jan 2 15:04")
}

type Reaction struct {
	FIID int32
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
