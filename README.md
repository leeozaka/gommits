# Git Commits App

A standalone application for analyzing Git commits and exporting changed files to CSV with a user-friendly terminal UI.

## Project Structure

```
go-commits-app/
├── cmd/
│   └── app/
│       └── main.go         # Application entry point
├── internal/
│   ├── git/
│   │   └── git.go          # Git operations
│   ├── models/
│   │   └── commit.go       # Data models
│   └── ui/
│       └── ui.go           # Terminal UI implementation
└── pkg/
    └── utils/
        └── csv.go          # CSV export utilities
```

## Installation

### Prerequisites

- Go 1.18 or higher
- Git installed and available in your PATH

### Building from Source

1. Clone this repository:
   ```bash
   git clone https://github.com/username/go-commits-app.git
   cd go-commits-app
   ```

2. Build the application:
   ```bash
   go build -o go-commits-app ./cmd/app
   ```

## Usage

1. Run the application:
   ```bash
   ./go-commits-app
   ```

2. Follow the interactive UI to:
   - Select a Git repository
   - Enter an author name or email
   - Configure options
   - View commit results
   - Export to CSV

## Navigation

- **Enter**: Proceed to next step
- **Tab**: Auto-complete current directory or toggle options
- **Alt+Backspace**: Go back to previous screen
- **Esc**: Quit the application