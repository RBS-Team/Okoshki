package ml

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func NewClient() *Client {
	return &Client{
		BaseURL: "http://ml-service:8010",
		HTTP: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

func (c *Client) Classify(ctx context.Context, text string) (*Response, error) {
	body, _ := json.Marshal(Request{Text: text})

	req, _ := http.NewRequestWithContext(
		ctx,
		"POST",
		c.BaseURL+"/classify",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result Response
	json.NewDecoder(resp.Body).Decode(&result)

	return &result, nil
}
