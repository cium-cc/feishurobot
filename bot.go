package feishubot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient defines the interface for an HTTP client.
// This allows for easy testing with mock clients.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is a Feishu custom bot client for sending messages via webhook.
//
// Example usage:
//
//	client := feishubot.NewClient(webhookURL, secret)
//	message := feishubot.NewTextMessage("Hello, World!")
//	resp, err := client.Send(context.Background(), message)
//	if err != nil {
//	    // handle error
//	}
type Client struct {
	WebhookURL string
	Secret     string
	HTTPClient HTTPClient
}

// Response represents the response from the Feishu webhook API.
type Response struct {
	Code          int         `json:"code"`
	Msg           string      `json:"msg"`
	Data          interface{} `json:"data"`
	StatusCode    int         `json:"StatusCode,omitempty"`    // Deprecated: Use Code instead
	StatusMessage string      `json:"StatusMessage,omitempty"` // Deprecated: Use Msg instead
}

// NewClient creates a new Feishu bot client.
//
// Parameters:
//   - webhookURL: The full webhook URL for your custom bot
//   - secret: Optional secret for signature verification. If empty, no signature will be sent.
//
// The default HTTP client has a 30 second timeout. For custom timeout settings,
// use SetHTTPClient after creating the client.
func NewClient(webhookURL string, secret string) *Client {
	return &Client{
		WebhookURL: webhookURL,
		Secret:     secret,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetHTTPClient sets a custom HTTP client for the bot client.
// This is useful for testing or for custom timeout configurations.
func (c *Client) SetHTTPClient(client HTTPClient) {
	c.HTTPClient = client
}

// Send sends a message to the Feishu webhook.
//
// If a secret is configured, the timestamp and signature will be automatically
// added to the message for security verification.
//
// Parameters:
//   - ctx: Context for the request, can be used for cancellation
//   - msg: The message to send
//
// Returns:
//   - The API response
//   - An error if the request fails or returns a non-zero code
func (c *Client) Send(ctx context.Context, msg *Message) (*Response, error) {
	// Clone the message to avoid modifying the original
	msgCopy := *msg

	// Add signature if secret is configured
	if c.Secret != "" {
		timestamp := time.Now().Unix()
		sign, err := GenSign(c.Secret, timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to generate signature: %w", err)
		}
		msgCopy.Timestamp = timestamp
		msgCopy.Sign = sign
	}

	// Marshal message to JSON
	body, err := json.Marshal(msgCopy)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response
	var apiResp Response
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for API errors
	if apiResp.Code != 0 {
		return &apiResp, fmt.Errorf("API error (code %d): %s", apiResp.Code, apiResp.Msg)
	}

	return &apiResp, nil
}
