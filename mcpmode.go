package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Run MCP mode - start MCP server

func runMCPMode() {
	fmt.Fprintf(os.Stderr, "🔌 Starting MCP server...\n")

	cs, err := initCalendarServiceFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize calendar service: %v\n", err)
		os.Exit(1)
	}

	// Create MCP server
	s := server.NewMCPServer(
		"google-calendar-agenda",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// Create the get-daily-agenda tool
	tool := mcp.NewTool("get_daily_agenda",
		mcp.WithDescription("Get today's calendar agenda from Google Calendar"),
	)

	// Add tool handler
	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		events, err := cs.getTodaysEvents()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error getting calendar events: %v", err)), nil
		}

		agenda := formatEventsForDisplay(events)
		return mcp.NewToolResultText(agenda), nil
	})

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		os.Exit(1)
	}
}
