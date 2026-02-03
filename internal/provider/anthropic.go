package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	pkgerrors "github.com/cockroachdb/errors"
)

// AnthropicRequest represents an Anthropic API request.
type AnthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"maxTokens"`
	Messages  []AnthropicMessage `json:"messages"`
}

// AnthropicMessage represents a message in Anthropic format.
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicResponse represents an Anthropic API response.
type AnthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *Provider) queryAnthropic(ctx context.Context, apiKey, model, promptText string) (string, error) {
	requestBody := AnthropicRequest{
		Model:     model,
		MaxTokens: 1000,
		Messages:  []AnthropicMessage{{Role: "user", Content: promptText}},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", pkgerrors.Wrap(err, "failed to marshal request")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.Endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", pkgerrors.Wrap(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", apiKey)
	req.Header.Set("Anthropic-Version", "2023-06-01")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", pkgerrors.New("request timed out")
		}

		return "", pkgerrors.Wrap(err, "failed to send request")
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", pkgerrors.Wrap(err, "failed to read response")
	}

	var anthropicResp AnthropicResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return "", pkgerrors.Wrapf(err, "failed to parse response: %s", string(body))
	}

	if anthropicResp.Error != nil {
		return "", pkgerrors.Newf("API error: %s", anthropicResp.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return "", pkgerrors.Newf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	if len(anthropicResp.Content) > 0 && anthropicResp.Content[0].Type == "text" {
		return anthropicResp.Content[0].Text, nil
	}

	return "", pkgerrors.New("no response from Anthropic")
}
