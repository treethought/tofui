package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
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
	FIDs       []uint64
	Cursor     string
	Limit      uint64
	ViewerFID  uint64
}

type FeedResponse struct {
	Casts     []*Cast
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

func (c *Client) GetUserByFID(fid uint64) (*User, error) {
	log.Println("get user by fid: ", fid)
	key := fmt.Sprintf("user:%d", fid)
	cached, err := db.GetDB().Get([]byte(key))
	if err == nil {
		u := &User{}
		if err := json.Unmarshal(cached, u); err != nil {
			log.Fatal("failed to unmarshal cached user: ", err)
		}
		log.Println("got cached user: ", u.Username)
		return u, nil
	}
	url := c.buildEndpoint(fmt.Sprintf("/user/bulk?fids=%d&viewer_fid=%d", fid, GetSigner().FID))
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, url, nil)
	if err != nil {
		log.Println("failed to create request: ", err)
		return nil, err
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("api_key", c.apiKey)
	res, err := c.c.Do(req)
	if err != nil {
		log.Println("failed to get user: ", err)
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		log.Println("failed to get user: ", res.Status)
		return nil, fmt.Errorf("failed to get user: %s", res.Status)
	}
	resp := &BulkUsersResponse{}
	if err = json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, err
	}
	if len(resp.Users) == 0 {
		return nil, fmt.Errorf("user not found")
	}
	user := resp.Users[0]
	d, _ := json.Marshal(user)
	if err := db.GetDB().Set([]byte(key), []byte(d)); err != nil {
		log.Println("failed to cache user: ", err)
	}
	return user, nil
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
