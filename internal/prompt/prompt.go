package prompt

import (
	"fmt"
	"runtime"
	"strings"
)

// Generate creates the prompt for the AI provider.
func Generate(query string) string {
	return fmt.Sprintf(`You are a command line assistant that helps users with shell commands.
User wants assistance with the following task:

%s

Instructions:
- Respond with a single command that achieves the desired result
- The command should be suitable for %s operating system
- Output ONLY the command, without any explanation
- Do not include any quotes, backticks, or markdown formatting
- If the task requires multiple commands, chain them with && or ;
- If you're unsure, provide the most common/standard approach
`, query, userOS())
}

// SanitizeCommand cleans up the AI response to extract just the command.
func SanitizeCommand(cmd string) string {
	cmd = strings.TrimSpace(cmd)

	// Remove markdown code blocks
	if strings.HasPrefix(cmd, "```") {
		lines := strings.Split(cmd, "\n")
		if len(lines) > 2 {
			// Remove first and last lines (``` markers)
			lines = lines[1 : len(lines)-1]
			cmd = strings.Join(lines, "\n")
		}
	}

	// Remove inline backticks
	cmd = strings.Trim(cmd, "`")

	// Remove any leading language identifiers (bash, sh, zsh, etc.)
	for _, lang := range []string{"bash", "sh", "zsh", "shell", "cmd", "powershell"} {
		if strings.HasPrefix(strings.ToLower(cmd), lang+"\n") {
			cmd = strings.TrimPrefix(cmd, lang+"\n")
			cmd = strings.TrimPrefix(cmd, strings.ToUpper(lang)+"\n")
		}
	}

	// Replace newlines with spaces for multi-line commands
	cmd = strings.ReplaceAll(cmd, "\n", " ")

	// Clean up multiple spaces
	for strings.Contains(cmd, "  ") {
		cmd = strings.ReplaceAll(cmd, "  ", " ")
	}

	return strings.TrimSpace(cmd)
}

func userOS() string {
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
