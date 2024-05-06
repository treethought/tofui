package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/treethought/castr/db"
)

type ChannelsResponse struct {
	Channels []*Channel
	Next     struct {
		Cursor *string `json:"cursor"`
	} `json:"next"`
}
type ChannelResponse struct {
	Channel       *Channel      `json:"channel"`
	ViewerContext ViewerContext `json:"viewer_context"`
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
	Lead          User   `json:"lead"`
	Hosts         []User `json:"hosts"`
}

func (c *Client) GetUserChannels(fid uint64, active bool, opts ...RequestOption) ([]*Channel, error) {
	var path string
	if active {
		path = "/channel/user"
	} else {
		path = "/user/channels"
	}
	opts = append(opts, WithFID(fid))

	var resp ChannelsResponse
	if err := c.doRequestInto(context.TODO(), path, &resp, opts...); err != nil {
		return nil, err
	}
	if resp.Channels == nil {
		return nil, errors.New("no channels found")
	}
	return resp.Channels, nil
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
	path := "/channel"

	opts := []RequestOption{WithQuery("id", pu), WithQuery("type", "parent_url")}

	var resp ChannelResponse
	err = c.doRequestInto(context.TODO(), path, &resp, opts...)
	if err != nil {
		return nil, err
	}
	if resp.Channel.Name == "" {
		return nil, errors.New("channel name empty")
	}

	d, _ := json.Marshal(resp.Channel)
	if err := db.GetDB().Set([]byte(key), []byte(d)); err != nil {
		log.Println("failed to cache channel: ", err)
	}
	return resp.Channel, nil

}

func (c *Client) FetchAllChannels() error {
	var resp ChannelsResponse
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
