package prompt

import (
	"runtime"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		query    string
		wantOS   string
		contains []string
	}{
		{
			name:   "simple query",
			query:  "list files",
			wantOS: expectedOS(),
			contains: []string{
				"list files",
				"single command",
				expectedOS(),
			},
		},
		{
			name:   "complex query",
			query:  "find all go files larger than 1MB",
			wantOS: expectedOS(),
			contains: []string{
				"find all go files larger than 1MB",
				"shell commands",
			},
		},
		{
			name:   "empty query",
			query:  "",
			wantOS: expectedOS(),
			contains: []string{
				expectedOS(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := Generate(tt.query)

			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("Generate() = %q, want to contain %q", got, want)
				}
			}
		})
	}
}

func TestSanitizeCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple command",
			input: "ls -la",
			want:  "ls -la",
		},
		{
			name:  "command with surrounding whitespace",
			input: "  ls -la  ",
			want:  "ls -la",
		},
		{
			name:  "command with markdown code block",
			input: "```bash\nls -la\n```",
			want:  "ls -la",
		},
		{
			name:  "command with markdown code block no language",
			input: "```\nls -la\n```",
			want:  "ls -la",
		},
		{
			name:  "command with inline backticks",
			input: "`ls -la`",
			want:  "ls -la",
		},
		{
			name:  "command with language prefix",
			input: "bash\nls -la",
			want:  "ls -la",
		},
		{
			name:  "command with sh prefix",
			input: "sh\nls -la",
			want:  "ls -la",
		},
		{
			name:  "command with zsh prefix",
			input: "zsh\nls -la",
			want:  "ls -la",
		},
		{
			name:  "multi-line command becomes single line",
			input: "ls -la\npwd",
			want:  "ls -la pwd",
		},
		{
			name:  "multiple spaces collapsed",
			input: "ls   -la    /tmp",
			want:  "ls -la /tmp",
		},
		{
			name:  "complex markdown response",
			input: "```bash\nfind . -name \"*.go\" -type f\n```",
			want:  "find . -name \"*.go\" -type f",
		},
		{
			name:  "command with triple backticks inline",
			input: "```ls -la```",
			want:  "ls -la",
		},
		{
			name:  "powershell prefix",
			input: "powershell\nGet-ChildItem",
			want:  "Get-ChildItem",
		},
		{
			name:  "cmd prefix",
			input: "cmd\ndir /s",
			want:  "dir /s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := SanitizeCommand(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUserOS(t *testing.T) {
	t.Parallel()

	got := userOS()

	switch runtime.GOOS {
	case "darwin":
		if got != "macOS" {
			t.Errorf("userOS() = %q, want %q", got, "macOS")
		}
	case "linux":
		if got != "Linux" {
			t.Errorf("userOS() = %q, want %q", got, "Linux")
		}
	case "windows":
		if got != "Windows" {
			t.Errorf("userOS() = %q, want %q", got, "Windows")
		}
	default:
		if got != runtime.GOOS {
			t.Errorf("userOS() = %q, want %q", got, runtime.GOOS)
		}
	}
}

// expectedOS returns the expected OS string for the current platform.
func expectedOS() string {
	switch runtime.GOOS {
	case "darwin":
		return "macOS"
	case "linux":
		return "Linux"
	case "windows":
		return "Windows"
	default:
		return runtime.GOOS
	}
}
