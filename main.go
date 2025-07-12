package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: agenda-mcp <mode> [options]")
		fmt.Println("Modes:")
		fmt.Println("  text [YYYY-MM-DD] - Display agenda (today's agenda if no date specified)")
		fmt.Println("  mcp               - Start MCP server to provide agenda tool")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  agenda-mcp text           # Show today's agenda")
		fmt.Println("  agenda-mcp text 2024-12-25   # Show agenda for Christmas")
		fmt.Println("")
		fmt.Println("Environment variables for MCP mode:")
		fmt.Println("  client_id     - Google OAuth client ID")
		fmt.Println("  project_id    - Google project ID")
		fmt.Println("  client_secret - Google OAuth client secret")
		os.Exit(1)
	}

	mode := os.Args[1]

	switch mode {
	case "text":
		// Check if a date parameter was provided
		var dateStr string
		if len(os.Args) >= 3 {
			dateStr = os.Args[2]
		}
		runTextMode(dateStr)
	case "mcp":
		runMCPMode()
	default:
		// Default to text mode with no date (today)
		runTextMode("")
	}
}
