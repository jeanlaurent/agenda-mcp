package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"golang.org/x/oauth2"
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

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	// Start a local HTTP server to handle the callback
	codeCh := make(chan string)
	errCh := make(chan error)

	server := &http.Server{Addr: ":8080"}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no authorization code received")
			return
		}

		// Send success page
		fmt.Fprintf(w, `
			<html>
			<head><title>Authorization Success</title></head>
			<body style="font-family: Arial, sans-serif; text-align: center; padding: 50px;">
				<h1 style="color: green;">✅ Authorization Successful!</h1>
				<p>You can now close this window and return to the terminal.</p>
			</body>
			</html>
		`)

		codeCh <- code
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Update the redirect URI to match our local server
	config.RedirectURL = "http://localhost:8080"

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Opening browser for authorization...\n")
	fmt.Printf("If the browser doesn't open automatically, go to: %v\n", authURL)

	// Try to open the browser automatically
	openBrowser(authURL)

	var authCode string
	select {
	case authCode = <-codeCh:
		fmt.Println("✅ Authorization code received successfully!")
	case err := <-errCh:
		log.Fatalf("Error during authorization: %v", err)
	case <-time.After(5 * time.Minute):
		log.Fatalf("Authorization timed out after 5 minutes")
	}

	// Shutdown the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// openBrowser tries to open the URL in a browser
func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Printf("Could not open browser automatically: %v\n", err)
	}
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
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
	if strings.HasPrefix(hexColor, "#") {
		hexColor = hexColor[1:]
	}

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

// Retrieve a token from existing file, don't start OAuth flow
func getClientFromExistingToken(config *oauth2.Config) (*http.Client, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("unable to determine executable path: %v", err)
	}
	tokenPath := filepath.Join(filepath.Dir(execPath), "token.json")
	tok, err := tokenFromFile(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("unable to load existing token from %s: %v. Please run 'agenda-mcp auth' first", tokenPath, err)
	}
	// Use the config passed as parameter to create the client with the loaded token
	return config.Client(context.Background(), tok), nil
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

// Run auth mode - just authenticate and exit
func runAuthMode() {
	fmt.Println("🔐 Running authentication flow...")
	_, err := initCalendarService("credentials.json")
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}
	fmt.Println("✅ Authentication successful! Token saved.")
}

// Run test mode - show today's agenda
func runTestMode() {
	fmt.Println("📅 Fetching today's agenda...")
	cs, err := initCalendarService("credentials.json")
	if err != nil {
		log.Fatalf("Failed to initialize calendar service: %v", err)
	}

	events, err := cs.getTodaysEvents()
	if err != nil {
		log.Fatalf("Failed to get today's events: %v", err)
	}

	fmt.Print(formatEventsForDisplay(events))
}

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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: agenda-mcp <mode>")
		fmt.Println("Modes:")
		fmt.Println("  auth - Run OAuth authentication flow and exit")
		fmt.Println("  test - Display today's agenda (current behavior)")
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
	case "auth":
		runAuthMode()
	case "test":
		runTestMode()
	case "mcp":
		runMCPMode()
	default:
		fmt.Printf("Unknown mode: %s\n", mode)
		fmt.Println("Valid modes: auth, test, mcp")
		os.Exit(1)
	}
}
