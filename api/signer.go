package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Signer struct {
	SignerUUID string `json:"signer_uuid"`
	PublicKey  string `json:"public_key"`
	Status     string `json:"status"`
}

func (c *Client) CreateSigner() (*Signer, error) {
	url := fmt.Sprintf("%s/signer", c.baseURL)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("api_key", c.apiKey)

	res, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	resp := &Signer{}
	err = json.NewDecoder(res.Body).Decode(resp)
	if err != nil {
		return nil, err
	}
	log.Printf("signer: %+v\n", resp)
	return resp, nil
}
