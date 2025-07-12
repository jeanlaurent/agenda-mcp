# Daily Agenda from Google Calendar

A simple Go program that fetches and displays your daily agenda from Google Calendar.

## Setup Instructions

### 1. Enable Google Calendar API

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Google Calendar API:
   - Go to "APIs & Services" > "Library"
   - Search for "Google Calendar API"
   - Click on it and press "Enable"

### 2. Create Credentials

1. Go to "APIs & Services" > "Credentials"
2. Click "Create Credentials" > "OAuth client ID"
3. If prompted, configure the OAuth consent screen first:
   - Choose "External" user type
   - Fill in the required information (app name, user support email, developer contact)
   - Add your email to test users
4. For Application type, choose "Desktop application"
5. Give it a name (e.g., "Daily Agenda App")
6. **Important**: Add `http://localhost:8080` to the "Authorized redirect URIs"
7. Download the JSON file and save it as `credentials.json` in this directory

### 3. Install Dependencies

```bash
go mod tidy
```

### 4. Run the Program

The program supports three modes:

#### Authentication Mode

Set up OAuth credentials (run this first):

```bash
go run main.go auth
```

#### Test Mode (Display Today's Agenda)

```bash
go run main.go test
```

#### MCP Server Mode

```bash
go run main.go mcp
```

On first run (auth mode):

1. The program will automatically open your browser (or show you a URL if it can't)
2. Sign in with your Google account
3. Grant permission to read your calendar
4. The authorization will complete automatically - you'll see a success page
5. Your credentials are saved for future runs

After authentication, you can use `test` mode to display your agenda or `mcp` mode to run as an MCP server.

## Features

- 📅 Shows today's events in chronological order
- 🕐 Displays event times (or "All day" for full-day events)
- 🎨 Shows event colors with your actual category names (Focus Time, Internal Group Meetings, External Meetings, Personal, etc.)
- 📍 Shows event locations if available
- 📝 Displays event descriptions (truncated to 100 characters)
- 👥 Lists attendees with their response status (✅ accepted, ❌ declined, ❓ tentative, ⏳ pending)
- 👁️ Shows event visibility settings
- ⏱️ Displays transparency settings (busy/free)
- 🎉 Friendly message when no events are scheduled
- 🔌 **MCP Server**: Exposes calendar data via Model Context Protocol for integration with LLM applications

## MCP Server Mode

When run in MCP mode (`go run main.go mcp`), the program acts as a Model Context Protocol server that can be integrated with LLM applications like Claude, providing a `get_daily_agenda` tool.

### MCP Integration

Add this to your MCP client configuration:

```json
{
  "mcpServers": {
    "google-calendar": {
      "command": "agenda-mcp",
      "args": ["mcp"]
    }
  }
}
```

This allows LLM applications to fetch your daily Google Calendar agenda using the `get_daily_agenda` tool.

## File Structure

- `main.go` - Main program
- `credentials.json` - Google API credentials (you need to create this)
- `token.json` - OAuth token (automatically created after first authorization)
- `go.mod` - Go module dependencies

## Security Notes

- Keep `credentials.json` and `token.json` private
- Don't commit these files to version control
- The program only requests read-only access to your calendar
