package api

import (
	"context"
	"errors"
	"log"
)

type ReactionType string

const (
	Like   ReactionType = "like"
	Recast ReactionType = "recast"
)

type ReactionRequest struct {
	SignerUUID   string       `json:"signer_uuid"`
	ReactionType ReactionType `json:"reaction_type"`
	Target       string       `json:"target"`
}

type ReactionResponse struct {
	Success bool
	Message string
}

func (c *Client) React(s *Signer, cast string, t ReactionType) error {
	if s == nil {
		return errors.New("signer required")
	}

	var payload = ReactionRequest{
		SignerUUID:   s.UUID,
		ReactionType: t,
		Target:       cast,
	}

	log.Println("reacting to cast: ", cast, " with type: ", t)
	var resp ReactionResponse
	if err := c.doPostInto(context.TODO(), "/reaction", payload, &resp); err != nil {
		log.Println("failed to react: ", err)
		return err
	}
	log.Println("got reaction response: ", resp)

	if !resp.Success {
		return errors.New(resp.Message)
	}
	return nil
}
