package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: agenda-mcp <mode>")
		fmt.Println("Modes:")
		fmt.Println("  text - Display today's agenda")
		fmt.Println("  mcp  - Start MCP server to provide agenda tool")
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
		runTextMode()
	case "mcp":
		runMCPMode()
	default:
		runTextMode()
	}
}
