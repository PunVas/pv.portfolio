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
	// Always log the submission for review as requested
	log.Printf("[contact-log] FROM: %s | MSG: %s", source, message)

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

		bodyBytes, _ := json.Marshal(payload)

		maxRetries := 5
		httpClient := &http.Client{Timeout: 10 * time.Second}

		for i := 0; i < maxRetries; i++ {
			req, err := http.NewRequest("POST", c.WebhookURL, bytes.NewBuffer(bodyBytes))
			if err != nil {
				log.Printf("[discord] request error: %v", err)
				return
			}

			// Add headers to look less like a generic bot
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

			resp, err := httpClient.Do(req)
			if err != nil {
				wait := time.Duration(1<<i) * time.Second
				log.Printf("[discord] network error (attempt %d): %v. retrying in %v...", i+1, err, wait)
				time.Sleep(wait)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusTooManyRequests {
				var retryAfter time.Duration = 5 * time.Second // fallback

				// 1. Check Retry-After header (seconds)
				if val := resp.Header.Get("Retry-After"); val != "" {
					if s, err := time.ParseDuration(val + "s"); err == nil {
						retryAfter = s
					}
				}

				// 2. Try parsing body
				var discordErr struct {
					RetryAfter float64 `json:"retry_after"`
				}
				if err := json.NewDecoder(resp.Body).Decode(&discordErr); err == nil && discordErr.RetryAfter > 0 {
					retryAfter = time.Duration(discordErr.RetryAfter * float64(time.Second))
				}

				log.Printf("[discord] RATE LIMITED (429). Discord says wait: %v. Attempt %d/%d", retryAfter, i+1, maxRetries)
				time.Sleep(retryAfter + (500 * time.Millisecond)) // Add jitter
				continue
			}

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				log.Printf("[discord] unexpected status: %d", resp.StatusCode)
			} else {
				log.Printf("[discord] message sent successfully to %s", source)
			}
			return
		}
		log.Printf("[discord] CRITICAL: failed to send message after %d attempts", maxRetries)
	}()
}
