# Daily Agenda from Google Calendar

A simple Go program that fetches and displays your daily agenda from Google Calendar.

## Prerequisites

This project uses [Task](https://taskfile.dev/) for task automation. Install it first:

```bash
# macOS
brew install go-task/tap/go-task

# Linux
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin

# Windows
choco install go-task

# Or install via Go
go install github.com/go-task/task/v3/cmd/task@latest
```

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
task mod-tidy
```

### 4. Build the Application

```bash
task build
```

## Usage

The program supports 2 modes:

### Text Mode (Display Calendar Agenda)

Display today's agenda:

```bash
task run-text
# or directly: ./agenda-mcp text
```

Display agenda for a specific date:

```bash
./agenda-mcp text 2024-12-25
```

The date must be in YYYY-MM-DD format. If no date is provided, today's agenda is displayed.

### MCP Server Mode

```bash
task run-mcp
```

On first run (auth mode):

1. The program will automatically open your browser (or show you a URL if it can't)
2. Sign in with your Google account
3. Grant permission to read your calendar
4. The authorization will complete automatically - you'll see a success page
5. Your credentials are saved for future runs

After authentication, you can use `task run-test` to display your agenda or `task run-mcp` to run as an MCP server.

## Available Tasks

Run `task --list` to see all available tasks:

- `task build` - Build the local binary
- `task clean` - Clean build artifacts
- `task mod-tidy` - Tidy and verify go modules
- `task run-text` - Show today's agenda in the command line
- `task run-mcp` - Start MCP server
- `task inspector` - Run the npx MCP inspector

### MCP Inspector

For debugging and testing the MCP server, you can use the MCP inspector:

```bash
task inspector
```

This will start the MCP inspector at http://localhost:8080 for testing the MCP server functionality.

## Alternative: Direct Go Commands

If you prefer not to use Task, you can run the commands directly:

```bash
# Install dependencies
go mod tidy

# Run modes directly
go run main.go text              # Show today's agenda
go run main.go text 2024-12-25   # Show agenda for specific date
go run main.go mcp               # Start MCP server
```

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

When run in MCP mode (`task run-mcp`), the program acts as a Model Context Protocol server that can be integrated with LLM applications like Claude, providing two calendar tools:

### Available MCP Tools

1. **`get_todays_agenda`** - Get today's calendar agenda from Google Calendar
2. **`get_agenda_for_date`** - Get calendar agenda for a specific date (YYYY-MM-DD format)

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

This allows LLM applications to:

- Fetch your daily Google Calendar agenda using the `get_todays_agenda` tool
- Get calendar events for any specific date using the `get_agenda_for_date` tool with a date parameter (e.g., "2024-12-25")

## File Structure

- `main.go` - Main program
- `Taskfile.yml` - Task automation configuration
- `credentials.json` - Google API credentials (you need to create this)
- `token.json` - OAuth token (automatically created after first authorization)
- `go.mod` - Go module dependencies

## Security Notes

- Keep `credentials.json` and `token.json` private
- Don't commit these files to version control
- The program only requests read-only access to your calendar
