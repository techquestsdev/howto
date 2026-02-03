package provider

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"

	pkgerrors "github.com/cockroachdb/errors"
)

// queryCopilot uses the official GitHub Copilot CLI (gh copilot).
// This requires:
// 1. GitHub CLI (gh) to be installed
// 2. Active GitHub Copilot subscription
// 3. Being authenticated with: gh auth login
//
// The modern gh copilot CLI uses:
//   - `-p` or `--prompt` for non-interactive mode
//   - `-s` or `--silent` for output only (no stats)
//   - `--model` for model selection
func (p *Provider) queryCopilot(ctx context.Context, _ string, model string, promptText string) (string, error) {
	// Check if gh is available
	ghPath, err := exec.LookPath("gh")
	if err != nil {
		return "", pkgerrors.New("GitHub CLI (gh) not found. Install it from https://cli.github.com/")
	}

	// Build the command args
	// Use -p for prompt mode (non-interactive) and -s for silent (only output, no stats)
	// The prompt asks for a shell command specifically
	shellPrompt := "Output only a shell command (no explanation, no markdown, no backticks) that: " + promptText

	args := []string{"copilot", "--", "-p", shellPrompt, "-s"}

	// Add model if specified and not the default
	if model != "" && model != "gpt-4" {
		args = append(args, "--model", model)
	}

	cmd := exec.CommandContext(ctx, ghPath, args...)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", handleCopilotError(ctx, err, stderr.String())
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", pkgerrors.New("no suggestion from GitHub Copilot")
	}

	// Clean up the response - remove any markdown formatting that might slip through
	result = CleanCopilotResponse(result)

	return result, nil
}

func handleCopilotError(ctx context.Context, err error, stderrStr string) error {
	// Check for context cancellation/timeout
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return pkgerrors.New("request timed out")
	}

	if errors.Is(ctx.Err(), context.Canceled) {
		return pkgerrors.New("request canceled")
	}

	// Check for common errors
	if strings.Contains(stderrStr, "not installed") || strings.Contains(stderrStr, "extension") {
		return pkgerrors.New("GitHub Copilot CLI not available. Ensure gh copilot works")
	}

	if strings.Contains(stderrStr, "auth") || strings.Contains(stderrStr, "login") {
		return pkgerrors.New("Not authenticated with GitHub. Run: gh auth login")
	}

	if strings.Contains(stderrStr, "subscription") {
		return pkgerrors.New("GitHub Copilot subscription required")
	}

	return pkgerrors.Wrapf(err, "gh copilot failed: %s", stderrStr)
}

// CleanCopilotResponse removes markdown formatting from Copilot's response.
func CleanCopilotResponse(response string) string {
	lines := strings.Split(response, "\n")

	var cleaned []string

	inCodeBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines at start
		if len(cleaned) == 0 && trimmed == "" {
			continue
		}

		// Handle code blocks
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock

			continue
		}

		// If we're in a code block or it's a regular line, include it
		if inCodeBlock || !strings.HasPrefix(trimmed, "```") {
			cleaned = append(cleaned, line)
		}
	}

	result := strings.Join(cleaned, "\n")

	// Remove inline backticks
	result = strings.Trim(result, "`")

	return strings.TrimSpace(result)
}

// IsCopilotAvailable checks if GitHub Copilot CLI is available.
func IsCopilotAvailable() bool {
	// Check if gh is installed
	ghPath, err := exec.LookPath("gh")
	if err != nil {
		return false
	}

	// Check if copilot works (the modern gh copilot is built-in, not an extension)
	cmd := exec.Command(ghPath, "copilot", "--", "--version")

	return cmd.Run() == nil
}
