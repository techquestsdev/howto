package ui

import (
	"fmt"

	"github.com/fatih/color"
)

// OutputType defines different output types.
type OutputType int

const (
	// OutputSuccess indicates a successful operation.
	OutputSuccess OutputType = iota
	// OutputError indicates an error.
	OutputError
	// OutputWarning indicates a warning.
	OutputWarning
	// OutputInfo indicates informational output.
	OutputInfo
)

// Print prints colored output.
func Print(outputType OutputType, message string) {
	switch outputType {
	case OutputSuccess:
		color.Green(message)
	case OutputError:
		color.Red(message)
	case OutputWarning:
		color.Yellow(message)
	case OutputInfo:
		color.Cyan(message)
	}
}

// PrintHeader prints a formatted header.
func PrintHeader(title string) {
	color.Cyan("\n=== %s ===\n", title)
}

// PrintSuccess prints a success message.
func PrintSuccess(message string) {
	Print(OutputSuccess, "✓ "+message)
}

// PrintError prints an error message.
func PrintError(message string) {
	Print(OutputError, "✗ "+message)
}

// PrintWarning prints a warning message.
func PrintWarning(message string) {
	Print(OutputWarning, "⚠ "+message)
}

// PrintInfo prints an info message.
func PrintInfo(message string) {
	Print(OutputInfo, "ℹ "+message)
}

// PrintTable prints a simple table.
func PrintTable(headers []string, rows [][]string) {
	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Print headers
	for i, header := range headers {
		fmt.Printf("%-*s", colWidths[i], header)

		if i < len(headers)-1 {
			fmt.Print("  ")
		}
	}

	fmt.Println()

	// Print separator
	for i := range headers {
		for range colWidths[i] {
			fmt.Print("-")
		}

		if i < len(headers)-1 {
			fmt.Print("  ")
		}
	}

	fmt.Println()

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			fmt.Printf("%-*s", colWidths[i], cell)

			if i < len(row)-1 {
				fmt.Print("  ")
			}
		}

		fmt.Println()
	}
}
