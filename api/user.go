package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/treethought/castr/db"
)

type Profile struct {
	Bio struct {
		Text string
	}
}

type VerifiedAddresses struct {
	EthAddresses []string `json:"eth_addresses"`
	SolAddresses []string `json:"sol_addresses"`
}

type ViewerContext struct {
	Following  bool `json:"following"`
	FollowedBy bool `json:"followed_by"`
}

type BulkUsersResponse struct {
	Users []*User `json:"users"`
}

type User struct {
	FID               uint64            `json:"fid"`
	Username          string            `json:"username"`
	DisplayName       string            `json:"display_name"`
	PfpURL            string            `json:"pfp_url"`
	Profile           Profile           `json:"profile"`
	FollowerCount     int32             `json:"follower_count"`
	FollowingCount    int32             `json:"following_count"`
	Verifications     []string          `json:"verifications"`
	VerifiedAddresses VerifiedAddresses `json:"verified_addresses"`
	ActiveStatus      string            `json:"active_status"`
	PowerBadge        bool              `json:"power_badge"`
	ViewerContext     ViewerContext     `json:"viewer_context"`
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
	signer := GetSigner()
	var viewer uint64
	if signer != nil {
		viewer = signer.FID
	}
	url := c.buildEndpoint(fmt.Sprintf("/user/bulk?fids=%d", fid))
	if viewer != 0 {
		url += fmt.Sprintf("&viewer_fid=%d", viewer)
	}

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
