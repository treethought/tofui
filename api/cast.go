package api

import "time"

type Profile struct {
	Bio struct {
		Text string
	}
}

type Author struct {
	FID            int32    `json:"fid"`
	Username       string   `json:"username"`
	DisplayName    string   `json:"display_name"`
	PfpURL         string   `json:"pfp_url"`
	Profile        Profile  `json:"profile"`
	FollowerCount  int32    `json:"follower_count"`
	FollowingCount int32    `json:"following_count"`
	Verifications  []string `json:"verifications"`
	ActiveStatus   string   `json:"active_status"`
}

type Embed struct {
	URL string `json:"url"`
}

type Reaction struct {
	FIID int32
}

type Reactions struct {
	Likes   []Reaction `json:"likes"`
	Recasts []Reaction `json:"recasts"`
}

type Cast struct {
	Hash         string `json:"hash"`
	ThreadHash   string `json:"thread_hash"`
	ParentHash   string `json:"parent_hash"`
	ParentURL    string `json:"parent_url"`
	ParentAuthor struct {
		FID int32
	} `json:"parent_author"`
	Author    Author    `json:"author"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
	Embeds    []Embed   `json:"embeds"`
	Reactions Reaction  `json:"reactions"`
	// Replies struct {
	// 	Count int32 `json:"count,string"`
	// }
}

func (c Cast) HumanTime() string {
	return c.Timestamp.Format("Jan 2 15:04")
}
