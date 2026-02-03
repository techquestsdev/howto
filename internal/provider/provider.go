package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	pkgerrors "github.com/cockroachdb/errors"
)

// DefaultTimeout is the default timeout for API requests.
const DefaultTimeout = 30 * time.Second

// TimeoutEnvVar is the environment variable to override the default timeout.
const TimeoutEnvVar = "HOWTO_TIMEOUT"

// Provider represents an AI provider configuration.
type Provider struct {
	Name         string
	Endpoint     string
	DefaultModel string
	EnvVar       string
	AuthType     AuthType
	Configured   bool
}

// AuthType defines how the provider authenticates requests.
type AuthType int

const (
	// AuthBearer uses "Authorization: Bearer <token>" header.
	AuthBearer AuthType = iota
	// AuthAPIKey uses a custom API key header or query param.
	AuthAPIKey
	// AuthCLI uses an external CLI tool (like gh copilot).
	AuthCLI
)

// ProviderInfo contains provider information for display.
type ProviderInfo struct {
	Name         string
	DefaultModel string
	EnvVar       string
	Configured   bool
}

// ChatRequest represents a chat completion request.
type ChatRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"maxTokens"`
}

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse represents a chat completion response.
type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *APIError `json:"error,omitempty"`
}

// APIError represents an API error response.
type APIError struct {
	Message string `json:"message"`
}

// Providers configuration.
var (
	OpenAI = &Provider{
		Name:         "OpenAI",
		Endpoint:     "https://api.openai.com/v1/chat/completions",
		DefaultModel: "gpt-4o",
		EnvVar:       "OPENAI_API_KEY",
		AuthType:     AuthBearer,
	}

	Anthropic = &Provider{
		Name:         "Anthropic",
		Endpoint:     "https://api.anthropic.com/v1/messages",
		DefaultModel: "claude-sonnet-4-20250514",
		EnvVar:       "ANTHROPIC_API_KEY",
		AuthType:     AuthAPIKey,
	}

	Gemini = &Provider{
		Name:         "Gemini",
		Endpoint:     "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions",
		DefaultModel: "gemini-2.0-flash",
		EnvVar:       "GEMINI_API_KEY",
		AuthType:     AuthBearer,
	}

	DeepSeek = &Provider{
		Name:         "DeepSeek",
		Endpoint:     "https://api.deepseek.com/chat/completions",
		DefaultModel: "deepseek-chat",
		EnvVar:       "DEEPSEEK_API_KEY",
		AuthType:     AuthBearer,
	}

	GitHubCopilot = &Provider{
		Name:         "GitHub Copilot",
		Endpoint:     "", // Uses gh CLI
		DefaultModel: "gpt-4",
		EnvVar:       "", // No env var needed, uses gh auth
		AuthType:     AuthCLI,
	}
)

// apiProviders are providers that use API keys.
var apiProviders = []*Provider{
	OpenAI,
	Anthropic,
	Gemini,
	DeepSeek,
}

// Detect automatically detects the first available provider.
func Detect() (*Provider, string) {
	// First check API-based providers
	for _, p := range apiProviders {
		if key := os.Getenv(p.EnvVar); key != "" {
			p.Configured = true

			return p, key
		}
	}

	// Then check if GitHub Copilot CLI is available
	if IsCopilotAvailable() {
		GitHubCopilot.Configured = true

		return GitHubCopilot, ""
	}

	return nil, ""
}

// GetByName returns a provider by name.
func GetByName(name string) (*Provider, string, error) {
	// Check API providers
	for _, p := range apiProviders {
		if p.Name == name {
			key := os.Getenv(p.EnvVar)
			if key == "" {
				return nil, "", pkgerrors.Newf("provider %s requires %s to be set", name, p.EnvVar)
			}

			p.Configured = true

			return p, key, nil
		}
	}

	// Check GitHub Copilot
	if name == GitHubCopilot.Name || name == "Copilot" || name == "copilot" {
		if !IsCopilotAvailable() {
			return nil, "", pkgerrors.New(
				"GitHub Copilot CLI not available. Install with: gh extension install github/gh-copilot",
			)
		}

		GitHubCopilot.Configured = true

		return GitHubCopilot, "", nil
	}

	return nil, "", pkgerrors.Newf("unknown provider: %s", name)
}

// ListAll returns information about all providers.
func ListAll() []ProviderInfo {
	result := make([]ProviderInfo, 0, len(apiProviders)+1)

	// API-based providers
	for _, p := range apiProviders {
		info := ProviderInfo{
			Name:         p.Name,
			DefaultModel: p.DefaultModel,
			EnvVar:       p.EnvVar,
			Configured:   os.Getenv(p.EnvVar) != "",
		}
		result = append(result, info)
	}

	// GitHub Copilot (CLI-based)
	result = append(result, ProviderInfo{
		Name:         GitHubCopilot.Name,
		DefaultModel: GitHubCopilot.DefaultModel,
		EnvVar:       "gh copilot (CLI)",
		Configured:   IsCopilotAvailable(),
	})

	return result
}

// GetTimeout returns the configured timeout duration.
// It checks the HOWTO_TIMEOUT environment variable first,
// then falls back to the provided default or DefaultTimeout.
func GetTimeout(flagTimeout time.Duration) time.Duration {
	// Flag takes precedence
	if flagTimeout > 0 {
		return flagTimeout
	}

	// Check environment variable
	if envTimeout := os.Getenv(TimeoutEnvVar); envTimeout != "" {
		if d, err := time.ParseDuration(envTimeout); err == nil {
			return d
		}
	}

	return DefaultTimeout
}

// Query sends a chat completion request to the provider.
func (p *Provider) Query(ctx context.Context, apiKey, model, promptText string) (string, error) {
	if p.AuthType == AuthCLI {
		return p.queryCopilot(ctx, apiKey, model, promptText)
	}

	if p.Name == "Anthropic" {
		return p.queryAnthropic(ctx, apiKey, model, promptText)
	}

	return p.queryOpenAICompatible(ctx, apiKey, model, promptText)
}

func (p *Provider) queryOpenAICompatible(ctx context.Context, apiKey, model, promptText string) (string, error) {
	requestBody := ChatRequest{
		Model:     model,
		Messages:  []Message{{Role: "user", Content: promptText}},
		MaxTokens: 1000,
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
	req.Header.Set("Authorization", "Bearer "+apiKey)

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

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", pkgerrors.Wrap(err, "failed to parse response")
	}

	if chatResp.Error != nil {
		return "", pkgerrors.Newf("API error: %s", chatResp.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return "", pkgerrors.Newf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	if len(chatResp.Choices) > 0 {
		return chatResp.Choices[0].Message.Content, nil
	}

	return "", pkgerrors.Newf("no response from %s", p.Name)
}
