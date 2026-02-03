package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"github.com/techquestsdev/howto/internal/prompt"
	"github.com/techquestsdev/howto/internal/provider"
	"github.com/techquestsdev/howto/internal/terminal"
	"github.com/techquestsdev/howto/internal/ui"
)

var (
	modelFlag    string
	providerFlag string
	dryRunFlag   bool
	timeoutFlag  time.Duration
)

var rootCmd = &cobra.Command{
	Use:   "howto <query>",
	Short: "Get command-line suggestions from AI",
	Long: `Howto is a CLI tool that uses AI to suggest shell commands based on natural language queries.

Supports multiple AI providers:
  - OpenAI (GPT-4, GPT-3.5)
  - Anthropic (Claude)
  - Google Gemini
  - DeepSeek
  - GitHub Copilot

Environment Variables:
  OPENAI_API_KEY      OpenAI API key
  ANTHROPIC_API_KEY   Anthropic API key
  GEMINI_API_KEY      Google Gemini API key
  DEEPSEEK_API_KEY    DeepSeek API key
  GITHUB_TOKEN        GitHub token (for Copilot)
  HOWTO_MODEL         Override default model for the provider
  HOWTO_PROVIDER      Force a specific provider
  HOWTO_TIMEOUT       Request timeout (e.g., "30s", "1m") - default: 30s`,
	Version: "1.0.0",
	Args:    cobra.MinimumNArgs(1),
	RunE:    runHowto,
}

var listProvidersCmd = &cobra.Command{
	Use:   "providers",
	Short: "List available AI providers and their status",
	RunE:  runListProviders,
}

func runHowto(cmd *cobra.Command, args []string) error {
	query := strings.Join(args, " ")

	// Detect or use specified provider
	p, apiKey, err := getProvider()
	if err != nil {
		return err
	}

	// Override model if specified
	model := modelFlag
	if model == "" {
		model = p.DefaultModel
	}

	// Generate the prompt
	promptText := prompt.Generate(query)

	// Create context with timeout
	timeout := provider.GetTimeout(timeoutFlag)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Query the AI
	response, err := p.Query(ctx, apiKey, model, promptText)
	if err != nil {
		ui.PrintError(fmt.Sprintf("Failed to query %s: %v", p.Name, err))

		return errors.Wrap(err, "failed to query provider")
	}

	// Sanitize the command
	command := prompt.SanitizeCommand(response)

	if dryRunFlag {
		ui.PrintInfo(fmt.Sprintf("Provider: %s (model: %s)", p.Name, model))
		fmt.Println(command)

		return nil
	}

	// Insert the command into the terminal
	terminal.InsertInput(command)

	return nil
}

func getProvider() (*provider.Provider, string, error) {
	if providerFlag != "" {
		p, apiKey, err := provider.GetByName(providerFlag)
		if err != nil {
			ui.PrintError(fmt.Sprintf("Provider '%s' not found or not configured", providerFlag))

			return nil, "", errors.Wrap(err, "failed to get provider")
		}

		return p, apiKey, nil
	}

	p, apiKey := provider.Detect()
	if p == nil {
		ui.PrintError("No API key found")
		ui.PrintInfo("Set one of: OPENAI_API_KEY, ANTHROPIC_API_KEY, GEMINI_API_KEY, DEEPSEEK_API_KEY, GITHUB_TOKEN")

		return nil, "", errors.New("no provider configured")
	}

	return p, apiKey, nil
}

func runListProviders(cmd *cobra.Command, args []string) error {
	providers := provider.ListAll()

	ui.PrintHeader("Available Providers")

	headers := []string{"Provider", "Status", "Default Model", "Env Variable"}
	rows := make([][]string, 0, len(providers))

	for _, p := range providers {
		status := "Not configured"
		if p.Configured {
			status = "Ready"
		}

		rows = append(rows, []string{p.Name, status, p.DefaultModel, p.EnvVar})
	}

	ui.PrintTable(headers, rows)

	return nil
}

// Execute is the main entry point for the CLI.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return errors.Wrap(err, "failed to execute root command")
	}

	return nil
}

func init() {
	rootCmd.Flags().StringVarP(&modelFlag, "model", "m", "", "Override the default model")
	rootCmd.Flags().StringVarP(&providerFlag, "provider", "p", "", "Force a specific provider")
	rootCmd.Flags().BoolVarP(&dryRunFlag, "dry-run", "d", false, "Print command without inserting into terminal")
	rootCmd.Flags().DurationVarP(&timeoutFlag, "timeout", "t", 0, "Request timeout (e.g., 30s, 1m) - default: 30s")

	rootCmd.AddCommand(listProvidersCmd)
}
