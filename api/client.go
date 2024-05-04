package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/treethought/castr/db"
)

var (
	client *Client
	signer *Signer
	once   sync.Once
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

func SetSigner(s *Signer) {
	once.Do(func() {
		signer = s
		d, _ := json.Marshal(s)
		if err := db.GetDB().Set([]byte("signer"), d); err != nil {
			log.Fatal("failed to save signer: ", err)
		}
	})
}
func GetSigner() *Signer {
	if signer == nil {
		d, err := db.GetDB().Get([]byte("signer"))
		if err != nil {
			log.Println("no signer found in db")
			return nil
		}
		signer = &Signer{}
		if err = json.Unmarshal(d, signer); err != nil {
			log.Println("failed to unmarshal signer: ", err)
			return nil
		}
	}
	return signer
}

type Signer struct {
	FID  uint64
	UUID string
}

type FeedRequest struct {
	FeedType   string
	FID        uint64
	FilterType string
	ParentURL  string
	FIDs       []int32
	Cursor     string
	Limit      int32
}

type FeedResponse struct {
	Casts []*Cast
}

type ChannelResponse struct {
	Channels []Channel
	Next     struct {
		Cursor *string `json:"cursor"`
	} `json:"next"`
}

type Channel struct {
	ID            string `json:"id"`
	URL           string `json:"url"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	FollowerCount int32  `json:"follower_count"`
	Object        string `json:"object"`
	ImageURL      string `json:"image_url"`
	CreatedAt     uint   `json:"created_at"`
	ParentURL     string `json:"parent_url"`
	// TODO lead/hosts

}

func (c *Client) buildEndpoint(path string) string {
	return c.baseURL + path
}

func (c *Client) FetchAllChannels() error {
	var resp ChannelResponse
	var res *http.Response

	defer db.GetDB().Set([]byte("channelsloaded"), []byte(fmt.Sprintf("%d", time.Now().Unix())))

	for {
		if res != nil {
			res.Body.Close()
		}
		url := c.buildEndpoint(fmt.Sprintf("/channel/list?limit=50"))
		if resp.Next.Cursor != nil {
			url += fmt.Sprintf("&cursor=%s", *resp.Next.Cursor)
		}
		req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		req.Header.Add("accept", "application/json")
		req.Header.Add("api_key", c.apiKey)

		res, err = c.c.Do(req)
		if err != nil {
			return err
		}

		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to get channels: %s", res.Status)
		}

		if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
			return err
		}

		for _, ch := range resp.Channels {
			key := fmt.Sprintf("channel:%s", ch.ParentURL)
			d, err := json.Marshal(ch)
			if err != nil {
				log.Println("failed to marshal channel: ", err)
				continue
			}
			if err := db.GetDB().Set([]byte(key), []byte(d)); err != nil {
				log.Println("failed to cache channel: ", err)
			}
		}

		if resp.Next.Cursor == nil {
			break
		}
	}

	return nil
}

func (c *Client) GetChannelByParentURL(pu string) (*Channel, error) {
	key := fmt.Sprintf("channel:%s", pu)
	cached, err := db.GetDB().Get([]byte(key))
	if err == nil {
		ch := &Channel{}
		if err := json.Unmarshal(cached, ch); err != nil {
			log.Fatal("failed to unmarshal cached channel: ", err)
		}
		return ch, nil
	}

	// TODO viewer FID
	url := c.buildEndpoint(fmt.Sprintf("/channel?id=%s&type=parent_url", pu))
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("api_key", c.apiKey)
	res, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get channel: %s", res.Status)
	}

	resp := &Channel{}
	if err = json.NewDecoder(res.Body).Decode(resp); err != nil {
		return nil, err
	}
	if resp.Name == "" {
		return nil, fmt.Errorf("channel name empty")
	}

	d, _ := json.Marshal(resp)
	if err := db.GetDB().Set([]byte(key), []byte(d)); err != nil {
		log.Println("failed to cache channel: ", err)
	}
	return resp, nil
}

func (c *Client) GetFeed(r FeedRequest) (*FeedResponse, error) {
	if r.FID == 0 {
		r.FID = GetSigner().FID
	}
	url := c.buildEndpoint(
		fmt.Sprintf("/feed?feed_type=%s&fid=%d&filter_type=%s&parent_url=%s&fids=%v&cursor=%s&limit=%d",
			r.FeedType, r.FID, r.FilterType, r.ParentURL, r.FIDs, r.Cursor, r.Limit))
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, url, nil)
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
