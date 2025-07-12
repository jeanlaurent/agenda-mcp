package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Custom color mapping based on your actual Google Calendar categories
var customColorMap = map[string]map[string]string{
	"1":  {"name": "Lavender", "emoji": "🟣"},                // Default fallback
	"2":  {"name": "Focus Time", "emoji": "🟢"},              // Green
	"3":  {"name": "Internal Group Meetings", "emoji": "🟣"}, // Purple
	"4":  {"name": "External Meetings", "emoji": "🔴"},       // Red
	"5":  {"name": "1:1", "emoji": "🟡"},                     // Yellow/Orange
	"6":  {"name": "Estaff", "emoji": "🔴"},                  // Light red/orange
	"7":  {"name": "Personal", "emoji": "🔵"},                // Blue
	"8":  {"name": "Reminders", "emoji": "⚫"},               // Gray
	"9":  {"name": "Travel", "emoji": "🟠"},                  // Orange/Red
	"10": {"name": "NotWorking", "emoji": "🟢"},              // Green
	"11": {"name": "External Meetings", "emoji": "🔴"},       // Red
}

type CalendarService struct {
	service          *calendar.Service
	colorDefinitions map[string]calendar.ColorDefinition
}

// CalendarEvent represents a simplified calendar event
type CalendarEvent struct {
	Summary     string
	StartTime   string
	EndTime     string
	Location    string
	Description string
	ColorName   string
	ColorEmoji  string
	IsAllDay    bool
}

func formatTime(timeStr string) string {
	if timeStr == "" {
		return "All day"
	}

	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return timeStr
	}

	return t.Format("15:04")
}

func getColorInfo(colorId string, colorDefinitions map[string]calendar.ColorDefinition) (string, string) {
	if colorId == "" {
		return "Default", "⚪"
	}

	// Use our custom mapping based on your actual calendar categories
	if color, exists := customColorMap[colorId]; exists {
		return color["name"], color["emoji"]
	}

	// If not in our custom mapping, try to get info from API color definitions
	if colorDef, exists := colorDefinitions[colorId]; exists {
		emoji := getEmojiFromHex(colorDef.Background)
		return fmt.Sprintf("Category %s", colorId), emoji
	}

	return fmt.Sprintf("Color %s", colorId), "🎨"
}

func getEmojiFromHex(hexColor string) string {
	// Simple color mapping based on hex values
	// This is a basic implementation - you might want to enhance this
	if hexColor == "" {
		return "⚪"
	}

	// Remove # if present
	hexColor = strings.TrimPrefix(hexColor, "#")

	// Basic color detection based on hex values
	switch {
	case strings.HasPrefix(hexColor, "f") || strings.HasPrefix(hexColor, "e") || strings.HasPrefix(hexColor, "d"):
		if strings.Contains(hexColor, "0") || strings.Contains(hexColor, "1") || strings.Contains(hexColor, "2") {
			return "🔴" // Red-ish
		}
		return "🟡" // Yellow-ish
	case strings.HasPrefix(hexColor, "a") || strings.HasPrefix(hexColor, "b") || strings.HasPrefix(hexColor, "c"):
		return "🟢" // Green-ish
	case strings.HasPrefix(hexColor, "7") || strings.HasPrefix(hexColor, "8") || strings.HasPrefix(hexColor, "9"):
		return "🔵" // Blue-ish
	case strings.HasPrefix(hexColor, "4") || strings.HasPrefix(hexColor, "5") || strings.HasPrefix(hexColor, "6"):
		return "🟣" // Purple-ish
	case strings.HasPrefix(hexColor, "1") || strings.HasPrefix(hexColor, "2") || strings.HasPrefix(hexColor, "3"):
		return "⚫" // Dark
	default:
		return "🎨"
	}
}

func fetchCalendarColors(srv *calendar.Service) map[string]calendar.ColorDefinition {
	colors, err := srv.Colors.Get().Do()
	if err != nil {
		log.Printf("Unable to retrieve calendar colors: %v", err)
		return make(map[string]calendar.ColorDefinition)
	}

	return colors.Event
}

// Initialize calendar service with credentials file path (for auth and test modes)
func initCalendarService(credentialsPath string) (*CalendarService, error) {
	ctx := context.Background()
	b, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Calendar client: %v", err)
	}

	colorDefinitions := fetchCalendarColors(srv)

	return &CalendarService{
		service:          srv,
		colorDefinitions: colorDefinitions,
	}, nil
}

// Initialize calendar service from environment variables (for MCP mode)
func initCalendarServiceFromEnv() (*CalendarService, error) {
	ctx := context.Background()

	// Get required environment variables
	clientID := os.Getenv("client_id")
	projectID := os.Getenv("project_id")

	// Validate required environment variables
	if clientID == "" {
		return nil, fmt.Errorf("client_id environment variable is required")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project_id environment variable is required")
	}

	// client_secret is also required for OAuth
	clientSecret := os.Getenv("client_secret")
	if clientSecret == "" {
		return nil, fmt.Errorf("client_secret environment variable is required")
	}

	// Create credentials JSON structure in memory
	credentialsJSON := map[string]interface{}{
		"installed": map[string]interface{}{
			"client_id":                   clientID,
			"project_id":                  projectID,
			"auth_uri":                    "https://accounts.google.com/o/oauth2/auth",
			"token_uri":                   "https://oauth2.googleapis.com/token",
			"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
			"client_secret":               clientSecret,
			"redirect_uris":               []string{"http://localhost:8080"},
		},
	}

	// Convert to JSON bytes
	credentialsBytes, err := json.Marshal(credentialsJSON)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal credentials JSON: %v", err)
	}

	// Use google.ConfigFromJSON just like the file-based version
	config, err := google.ConfigFromJSON(credentialsBytes, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials to config: %v", err)
	}

	// Get client from existing token (don't start OAuth flow)
	client, err := getClientFromExistingToken(config)
	if err != nil {
		return nil, err
	}

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Calendar client: %v", err)
	}

	colorDefinitions := fetchCalendarColors(srv)

	return &CalendarService{
		service:          srv,
		colorDefinitions: colorDefinitions,
	}, nil
}

// Get today's events
func (cs *CalendarService) getTodaysEvents() ([]CalendarEvent, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	events, err := cs.service.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(startOfDay.Format(time.RFC3339)).
		TimeMax(endOfDay.Format(time.RFC3339)).OrderBy("startTime").Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve events: %v", err)
	}

	var calendarEvents []CalendarEvent
	for _, item := range events.Items {
		var startTime, endTime string
		isAllDay := false

		if item.Start.DateTime != "" {
			startTime = formatTime(item.Start.DateTime)
		} else {
			startTime = "All day"
			isAllDay = true
		}

		if item.End.DateTime != "" {
			endTime = formatTime(item.End.DateTime)
		}

		colorName, colorEmoji := getColorInfo(item.ColorId, cs.colorDefinitions)

		calendarEvents = append(calendarEvents, CalendarEvent{
			Summary:     item.Summary,
			StartTime:   startTime,
			EndTime:     endTime,
			Location:    item.Location,
			Description: item.Description,
			ColorName:   colorName,
			ColorEmoji:  colorEmoji,
			IsAllDay:    isAllDay,
		})
	}

	return calendarEvents, nil
}
