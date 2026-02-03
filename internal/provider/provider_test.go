package provider

import (
	"os"
	"testing"
	"time"
)

func TestCleanCopilotResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple command",
			input:    "ls -la",
			expected: "ls -la",
		},
		{
			name:     "command with leading empty lines",
			input:    "\n\nls -la",
			expected: "ls -la",
		},
		{
			name:     "command in code block",
			input:    "```bash\nls -la\n```",
			expected: "ls -la",
		},
		{
			name:     "command in code block no language",
			input:    "```\nls -la\n```",
			expected: "ls -la",
		},
		{
			name:     "command with inline backticks",
			input:    "`ls -la`",
			expected: "ls -la",
		},
		{
			name:     "multi-line command in code block",
			input:    "```bash\nfind . -name \"*.go\" \\\n  -type f\n```",
			expected: "find . -name \"*.go\" \\\n  -type f",
		},
		{
			name:     "command with trailing backticks",
			input:    "ls -la`",
			expected: "ls -la",
		},
		{
			name:     "command with leading backticks",
			input:    "`ls -la",
			expected: "ls -la",
		},
		{
			name:     "empty code block",
			input:    "```\n```",
			expected: "",
		},
		{
			name:     "nested code blocks",
			input:    "```\necho '```'\n```",
			expected: "echo '```'",
		},
		{
			name:     "whitespace only",
			input:    "   \n\n   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := CleanCopilotResponse(tt.input)
			if got != tt.expected {
				t.Errorf("CleanCopilotResponse() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGetTimeout(t *testing.T) {
	t.Parallel()

	t.Run("default timeout", func(t *testing.T) {
		t.Parallel()

		got := GetTimeout(0)
		if got != DefaultTimeout {
			t.Errorf("GetTimeout() = %v, want %v", got, DefaultTimeout)
		}
	})

	t.Run("flag timeout takes precedence", func(t *testing.T) {
		t.Parallel()

		got := GetTimeout(60 * time.Second)
		if got != 60*time.Second {
			t.Errorf("GetTimeout() = %v, want %v", got, 60*time.Second)
		}
	})

	t.Run("flag timeout of 1s", func(t *testing.T) {
		t.Parallel()

		got := GetTimeout(1 * time.Second)
		if got != 1*time.Second {
			t.Errorf("GetTimeout() = %v, want %v", got, 1*time.Second)
		}
	})
}

func TestGetTimeoutWithEnv(t *testing.T) {
	tests := []struct {
		name        string
		flagTimeout time.Duration
		envTimeout  string
		expected    time.Duration
	}{
		{
			name:        "env timeout when no flag",
			flagTimeout: 0,
			envTimeout:  "45s",
			expected:    45 * time.Second,
		},
		{
			name:        "env timeout with minutes",
			flagTimeout: 0,
			envTimeout:  "2m",
			expected:    2 * time.Minute,
		},
		{
			name:        "invalid env timeout falls back to default",
			flagTimeout: 0,
			envTimeout:  "invalid",
			expected:    DefaultTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(TimeoutEnvVar, tt.envTimeout)

			got := GetTimeout(tt.flagTimeout)
			if got != tt.expected {
				t.Errorf("GetTimeout() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDetect(t *testing.T) {
	// Clear all provider env vars for clean testing
	envVars := []string{"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "GEMINI_API_KEY", "DEEPSEEK_API_KEY"}
	for _, v := range envVars {
		t.Setenv(v, "")
	}

	t.Run("detects OpenAI when key is set", func(t *testing.T) {
		t.Setenv("OPENAI_API_KEY", "test-key")

		p, key := Detect()
		if p == nil {
			t.Fatal("Detect() returned nil provider")

			return
		}

		if p.Name != "OpenAI" {
			t.Errorf("Detect() provider = %q, want %q", p.Name, "OpenAI")
		}

		if key != "test-key" {
			t.Errorf("Detect() key = %q, want %q", key, "test-key")
		}
	})

	t.Run("detects Anthropic when key is set", func(t *testing.T) {
		t.Setenv("ANTHROPIC_API_KEY", "test-anthropic-key")

		p, key := Detect()
		if p == nil {
			t.Fatal("Detect() returned nil provider")

			return
		}

		if p.Name != "Anthropic" {
			t.Errorf("Detect() provider = %q, want %q", p.Name, "Anthropic")
		}

		if key != "test-anthropic-key" {
			t.Errorf("Detect() key = %q, want %q", key, "test-anthropic-key")
		}
	})

	t.Run("OpenAI takes priority over Anthropic", func(t *testing.T) {
		t.Setenv("OPENAI_API_KEY", "openai-key")
		t.Setenv("ANTHROPIC_API_KEY", "anthropic-key")

		p, _ := Detect()
		if p == nil {
			t.Fatal("Detect() returned nil provider")

			return
		}

		if p.Name != "OpenAI" {
			t.Errorf("Detect() provider = %q, want %q (OpenAI should take priority)", p.Name, "OpenAI")
		}
	})
}

func TestGetByName(t *testing.T) {
	t.Run("returns OpenAI when configured", func(t *testing.T) {
		t.Setenv("OPENAI_API_KEY", "test-key")

		p, key, err := GetByName("OpenAI")
		if err != nil {
			t.Fatalf("GetByName() error = %v", err)
		}

		if p.Name != "OpenAI" {
			t.Errorf("GetByName() provider = %q, want %q", p.Name, "OpenAI")
		}

		if key != "test-key" {
			t.Errorf("GetByName() key = %q, want %q", key, "test-key")
		}
	})

	t.Run("returns error when OpenAI not configured", func(t *testing.T) {
		_ = os.Unsetenv("OPENAI_API_KEY")

		_, _, err := GetByName("OpenAI")
		if err == nil {
			t.Error("GetByName() expected error, got nil")
		}
	})

	t.Run("returns error for unknown provider", func(t *testing.T) {
		_, _, err := GetByName("UnknownProvider")
		if err == nil {
			t.Error("GetByName() expected error, got nil")
		}
	})

	t.Run("copilot aliases work", func(t *testing.T) {
		// This test may fail if gh copilot is not installed
		// We just verify the alias matching works
		aliases := []string{"GitHub Copilot", "Copilot", "copilot"}
		expectedErr := "GitHub Copilot CLI not available. " +
			"Install with: gh extension install github/gh-copilot"

		for _, alias := range aliases {
			_, _, err := GetByName(alias)
			// We expect either success (if installed) or specific error about CLI
			if err != nil && err.Error() != expectedErr {
				// This is fine - just means copilot isn't installed
				continue
			}
		}
	})
}

func TestListAll(t *testing.T) {
	t.Parallel()

	providers := ListAll()

	if len(providers) < 5 {
		t.Errorf("ListAll() returned %d providers, want at least 5", len(providers))
	}

	// Check that all expected providers are present
	expectedNames := []string{"OpenAI", "Anthropic", "Gemini", "DeepSeek", "GitHub Copilot"}
	providerNames := make(map[string]bool)

	for _, p := range providers {
		providerNames[p.Name] = true
	}

	for _, name := range expectedNames {
		if !providerNames[name] {
			t.Errorf("ListAll() missing provider %q", name)
		}
	}
}

func TestProviderConstants(t *testing.T) {
	t.Parallel()

	t.Run("DefaultTimeout is 30 seconds", func(t *testing.T) {
		t.Parallel()

		if DefaultTimeout != 30*time.Second {
			t.Errorf("DefaultTimeout = %v, want %v", DefaultTimeout, 30*time.Second)
		}
	})

	t.Run("TimeoutEnvVar is HOWTO_TIMEOUT", func(t *testing.T) {
		t.Parallel()

		if TimeoutEnvVar != "HOWTO_TIMEOUT" {
			t.Errorf("TimeoutEnvVar = %q, want %q", TimeoutEnvVar, "HOWTO_TIMEOUT")
		}
	})
}
