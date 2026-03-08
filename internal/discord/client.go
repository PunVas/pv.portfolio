package discord

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type Client struct {
	WebhookURL string
}

func New() *Client {
	return &Client{WebhookURL: os.Getenv("DISCORD_WEBHOOK_URL")}
}

func (c *Client) Send(message, source string) {
	if c.WebhookURL == "" {
		log.Println("[discord] skip: DISCORD_WEBHOOK_URL not set")
		return
	}

	go func() {
		payload := map[string]interface{}{
			"username":   source,
			"avatar_url": "https://cdn-icons-png.flaticon.com/512/9344/9344186.png",
			"content":    message,
		}

		body, _ := json.Marshal(payload)
		resp, err := http.Post(c.WebhookURL, "application/json", bytes.NewBuffer(body))
		if err != nil {
			log.Printf("[discord] error sending message: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			log.Printf("[discord] unexpected status: %d", resp.StatusCode)
		}
	}()
}
