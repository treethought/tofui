package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/treethought/tofui/db"
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

func (c *Client) GetUserByFID(fid uint64, viewer uint64) (*User, error) {
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

	path := "/user/bulk"

	opts := []RequestOption{
		WithQuery("fids", fmt.Sprintf("%d", fid)),
	}
	if viewer != 0 {
		opts = append(opts, WithQuery("viewer_fid", fmt.Sprintf("%d", viewer)))
	}

	var resp BulkUsersResponse
	if err := c.doRequestInto(context.TODO(), path, &resp, opts...); err != nil {
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
