# Howto

AI-powered command-line assistant that suggests shell commands from natural language queries.

[![CI](https://github.com/techquestsdev/howto/actions/workflows/ci.yml/badge.svg)](https://github.com/techquestsdev/howto/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/techquestsdev/howto?style=flat-square&logo=go&logoColor=white)](https://github.com/techquestsdev/howto/blob/main/go.mod)
[![Go Report Card](https://goreportcard.com/badge/github.com/techquestsdev/howto)](https://goreportcard.com/report/github.com/techquestsdev/howto)
[![codecov](https://img.shields.io/codecov/c/github/techquestsdev/howto?style=flat-square)](https://codecov.io/gh/techquestsdev/howto)
[![Release](https://img.shields.io/github/v/release/techquestsdev/howto?style=flat-square&include_prereleases)](https://github.com/techquestsdev/howto/releases)
[![License](https://img.shields.io/github/license/techquestsdev/howto?style=flat-square)](https://github.com/techquestsdev/howto/blob/main/LICENSE)

## Features

- **Multiple AI Providers**: OpenAI, Anthropic (Claude), Google Gemini, DeepSeek, and GitHub Copilot
- **Terminal Integration**: Commands are inserted directly into your terminal for review before execution
- **Cross-Platform**: Works on macOS, Linux, and Windows
- **Auto-Detection**: Automatically detects available providers from environment variables
- **Model Override**: Use any model your provider supports

## Installation

### Homebrew (macOS/Linux)

```bash
brew install techquestsdev/tap/howto
```

### Go Install

```bash
go install github.com/techquestsdev/howto@latest
```

### From Source

```bash
git clone https://github.com/techquestsdev/howto.git
cd howto
make build
```

## Configuration

Set an API key for your preferred provider:

```bash
# OpenAI
export OPENAI_API_KEY=sk-...

# Anthropic (Claude)
export ANTHROPIC_API_KEY=sk-ant-...

# Google Gemini
export GEMINI_API_KEY=...

# DeepSeek
export DEEPSEEK_API_KEY=...

# GitHub Copilot (uses gh CLI, no env var needed)
# See "GitHub Copilot Setup" section below
```

## Usage

### Basic Usage

```bash
howto "find all .go files modified in the last 7 days"
# Output: find . -name "*.go" -mtime -7

howto "compress directory foo to tar.gz"
# Output: tar -czvf foo.tar.gz foo

howto "show disk usage sorted by size"
# Output: du -sh * | sort -hr
```

### Options

```bash
# Dry run (print command without inserting into terminal)
howto -d "list docker containers"

# Use a specific model
howto -m gpt-4-turbo "count lines of code"

# Force a specific provider
howto -p Anthropic "show memory usage"
```

### List Available Providers

```bash
howto providers
```

Output:
```
=== Available Providers ===

Provider        Status          Default Model             Env Variable
--------        ------          -------------             ------------
OpenAI          Ready           gpt-4o                    OPENAI_API_KEY
Anthropic       Not configured  claude-sonnet-4-20250514  ANTHROPIC_API_KEY
Gemini          Ready           gemini-2.0-flash          GEMINI_API_KEY
DeepSeek        Not configured  deepseek-chat             DEEPSEEK_API_KEY
GitHub Copilot  Ready           gpt-4                     gh copilot (CLI)
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `OPENAI_API_KEY` | OpenAI API key |
| `ANTHROPIC_API_KEY` | Anthropic API key |
| `GEMINI_API_KEY` | Google Gemini API key |
| `DEEPSEEK_API_KEY` | DeepSeek API key |
| `HOWTO_MODEL` | Override default model for auto-detected provider |
| `HOWTO_PROVIDER` | Force a specific provider |

## GitHub Copilot Setup

GitHub Copilot integration uses the official GitHub CLI extension (no API key needed):

```bash
# 1. Install GitHub CLI (if not already installed)
brew install gh  # macOS
# or see https://cli.github.com/ for other platforms

# 2. Install the Copilot extension
gh extension install github/gh-copilot

# 3. Authenticate with GitHub
gh auth login

# 4. Verify it works
gh copilot suggest "list files"
```

> **Note**: Requires an active GitHub Copilot subscription.

## How It Works

1. You provide a natural language description of what you want to do
2. Howto sends your query to the configured AI provider
3. The AI returns a shell command appropriate for your OS
4. The command is inserted into your terminal's input buffer
5. You can review and edit before pressing Enter to execute

> **Note**: On Windows, commands are printed to stdout instead of being injected into the terminal.

## Provider Priority

When multiple providers are configured, howto uses them in this order:
1. OpenAI
2. Anthropic
3. Gemini
4. DeepSeek
5. GitHub Copilot

Use the `--provider` flag to override the automatic selection.

## Development

```bash
# Run tests
make test

# Run linter
make lint

# Build binary
make build

# Run all checks
make check
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Credits

Inspired by the original [howto](https://github.com/antonmedv/howto) by Anton Medvedev.

---

### Made with ❤️ and Go
