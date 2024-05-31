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

func (c *Client) React(cast string, t ReactionType) error {
	s := GetSigner()
	if s == nil {
		log.Println("no signer found")
		return errors.New("no signer found")
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
