package api

import (
	"context"
	"errors"
	"log"
	"time"
)

type Embed struct {
	URL    string `json:"url"`
	CastId struct {
		Hash string `json:"hash"`
		FID  int32  `json:"fid"`
	}
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
	Author    User      `json:"author"`
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
type CastPayload struct {
	SignerUUID      string  `json:"signer_uuid"`
	Text            string  `json:"text"`
	Parent          string  `json:"parent"`
	ChannelID       string  `json:"channel_id"`
	Idem            string  `json:"idem"`
	ParentAuthorFID int32   `json:"parent_author_fid"`
	Embeds          []Embed `json:"embeds"`
}

type PostCastResponse struct {
	Success bool
	Cast    Cast
}

func (c *Client) PostCast(text string) (*PostCastResponse, error) {
	s := GetSigner()
	if s == nil {
		return nil, errors.New("no signer found")
	}
	payload := CastPayload{
		Text:       text,
		SignerUUID: s.UUID,
	}
	log.Println("posting cast: ", text)

	var resp PostCastResponse
	if err := c.doPostInto(context.TODO(), "/cast", payload, &resp); err != nil {
		log.Println("failed to post cast: ", err)
		return nil, err
	}
	if !resp.Success {
		return nil, errors.New("failed to post cast")
	}

	return &resp, nil
}
