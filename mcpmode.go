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
	todayTool := mcp.NewTool("get_todays_agenda",
		mcp.WithDescription("Get today's agenda from Google Calendar for the user"),
	)

	// Add tool handler for today's agenda
	s.AddTool(todayTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		events, err := cs.getTodaysEvents()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error getting calendar events: %v", err)), nil
		}

		agenda := formatEventsForDisplay(events)
		return mcp.NewToolResultText(agenda), nil
	})

	// Create the get-agenda-for-date tool with date parameter
	dateTool := mcp.NewTool("get_agenda_for_date",
		mcp.WithDescription("Get agenda from Google Calendar for a specific date"),
		mcp.WithString("date",
			mcp.Required(),
			mcp.Description("Date in YYYY-MM-DD format (e.g., 2024-12-25)"),
			mcp.Pattern("^\\d{4}-\\d{2}-\\d{2}$"),
		),
	)

	// Add tool handler for specific date agenda
	s.AddTool(dateTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract date parameter from request using the correct method
		dateStr, err := request.RequireString("date")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Missing required parameter 'date': %v", err)), nil
		}

		// Get events for the specified date
		events, err := cs.getEventForDay(dateStr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error getting calendar events for %s: %v", dateStr, err)), nil
		}

		agenda := formatEventsForDisplayForDate(events, dateStr)
		return mcp.NewToolResultText(agenda), nil
	})

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		os.Exit(1)
	}
}
