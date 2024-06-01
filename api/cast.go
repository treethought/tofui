package api

import (
	"context"
	"errors"
	"fmt"
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
	FID   int32  `json:"fid"`
	FName string `json:"fname"`
}

type Reactions struct {
	LikesCount   uint       `json:"likes_count"`
	RecastsCount uint       `json:"recasts_count"`
	Likes        []Reaction `json:"likes"`
	Recasts      []Reaction `json:"recasts"`
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
	Reactions Reactions `json:"reactions"`
	Replies   struct {
		Count int32 `json:"count"`
	}
	DirectReplies []*Cast `json:"direct_replies"`
	ViewerContext struct {
		Liked    bool `json:"liked"`
		Recasted bool `json:"recasted"`
	} `json:"viewer_context"`
}

func (c Cast) HumanTime() string {
	return c.Timestamp.Format("Jan 2 15:04")
}

type CastClient struct {
	c *Client
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

type ConversationResponse struct {
	Conversation *struct {
		Cast Cast `json:"cast"`
	} `json:"conversation"`
}

type Conversation struct {
	Cast
}

func (c *Client) GetCastWithReplies(hash string) (*Cast, error) {
	path := "/cast/conversation"
	opts := []RequestOption{
		WithQuery("identifier", hash),
		WithQuery("type", "hash"),
		WithQuery("reply_depth", "10"),
	}
	signer := GetSigner()
	if signer != nil {
		opts = append(opts, WithQuery("viewer_fid", fmt.Sprintf("%d", signer.FID)))
	}

	var resp ConversationResponse
	if err := c.doRequestInto(context.TODO(), path, &resp, opts...); err != nil {
		return nil, err
	}
	if resp.Conversation == nil {
		return nil, errors.New("no replies found")
	}
	return &resp.Conversation.Cast, nil
}
