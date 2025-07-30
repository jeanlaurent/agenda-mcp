package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// Run test mode - show agenda for specified date or today
func runTextMode(dateStr string) {
	fmt.Println("🔐 Running authentication flow...")
	cs, err := initCalendarService()
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}
	fmt.Println("✅ Authentication successful! Token saved.")

	var events []CalendarEvent

	if dateStr == "" {
		// No date specified, use today
		fmt.Println("📅 Fetching today's agenda...")
		events, err = cs.getTodaysEvents()
		if err != nil {
			log.Fatalf("Failed to get today's events: %v", err)
		}
		fmt.Print(formatEventsForDisplay(events))
	} else {
		// Date specified, use the provided date
		fmt.Printf("📅 Fetching agenda for %s...\n", dateStr)
		events, err = cs.getEventForDay(dateStr)
		if err != nil {
			log.Fatalf("Failed to get events for %s: %v", dateStr, err)
		}
		fmt.Print(formatEventsForDisplayForDate(events, dateStr))
	}
}

// Format events for display
func formatEventsForDisplay(events []CalendarEvent) string {
	now := time.Now()
	var output strings.Builder

	output.WriteString(fmt.Sprintf("📅 Daily Agenda for %s\n", now.Format("Monday, January 2, 2006")))
	output.WriteString(strings.Repeat("=", 50) + "\n\n")

	if len(events) == 0 {
		output.WriteString("🎉 No events scheduled for today!")
		return output.String()
	}

	for i, event := range events {
		output.WriteString(fmt.Sprintf("%d. ", i+1))

		if event.IsAllDay {
			output.WriteString(fmt.Sprintf("🗓️  %s (All day) %s %s\n", event.Summary, event.ColorEmoji, event.ColorName))
		} else {
			output.WriteString(fmt.Sprintf("🕐 %s", event.StartTime))
			if event.EndTime != "" && event.EndTime != event.StartTime {
				output.WriteString(fmt.Sprintf(" - %s", event.EndTime))
			}
			output.WriteString(fmt.Sprintf(" | %s %s %s\n", event.Summary, event.ColorEmoji, event.ColorName))
		}

		if event.Location != "" {
			output.WriteString(fmt.Sprintf("   📍 %s\n", event.Location))
		}

		if event.Description != "" {
			desc := event.Description
			if len(desc) > 100 {
				desc = desc[:100] + "..."
			}
			output.WriteString(fmt.Sprintf("   📝 %s\n", desc))
		}

		output.WriteString("\n")
	}

	return output.String()
}

// Format events for display for a specific date
func formatEventsForDisplayForDate(events []CalendarEvent, dateStr string) string {
	targetDate, err := time.Parse("2006-01-02", dateStr)
	var output strings.Builder

	if err != nil {
		output.WriteString(fmt.Sprintf("📅 Daily Agenda for %s\n", dateStr))
	} else {
		output.WriteString(fmt.Sprintf("📅 Daily Agenda for %s\n", targetDate.Format("Monday, January 2, 2006")))
	}
	output.WriteString(strings.Repeat("=", 50) + "\n\n")

	if len(events) == 0 {
		output.WriteString("🎉 No events scheduled for this day!")
		return output.String()
	}

	for i, event := range events {
		output.WriteString(fmt.Sprintf("%d. ", i+1))

		if event.IsAllDay {
			output.WriteString(fmt.Sprintf("🗓️  %s (All day) %s %s\n", event.Summary, event.ColorEmoji, event.ColorName))
		} else {
			output.WriteString(fmt.Sprintf("🕐 %s", event.StartTime))
			if event.EndTime != "" && event.EndTime != event.StartTime {
				output.WriteString(fmt.Sprintf(" - %s", event.EndTime))
			}
			output.WriteString(fmt.Sprintf(" | %s %s %s\n", event.Summary, event.ColorEmoji, event.ColorName))
		}

		if event.Location != "" {
			output.WriteString(fmt.Sprintf("   📍 %s\n", event.Location))
		}

		if event.Description != "" {
			desc := event.Description
			if len(desc) > 100 {
				desc = desc[:100] + "..."
			}
			output.WriteString(fmt.Sprintf("   📝 %s\n", desc))
		}

		output.WriteString("\n")
	}

	return output.String()
}
