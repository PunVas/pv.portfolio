package discord

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
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

		maxRetries := 3
		backoff := 2 * time.Second

		for i := 0; i < maxRetries; i++ {
			resp, err := http.Post(c.WebhookURL, "application/json", bytes.NewBuffer(body))
			if err != nil {
				log.Printf("[discord] error sending message (attempt %d): %v", i+1, err)
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusTooManyRequests {
				log.Printf("[discord] rate limited (429) on attempt %d, retrying in %v...", i+1, backoff)
				time.Sleep(backoff)
				backoff *= 2
				continue
			}

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				log.Printf("[discord] unexpected status: %d", resp.StatusCode)
			} else {
				log.Printf("[discord] message sent successfully to %s", source)
			}
			return
		}
		log.Printf("[discord] failed to send message after %d attempts", maxRetries)
	}()
}
