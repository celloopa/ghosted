# jobtrack

A terminal-based job application tracker built with Go and [Charm](https://charm.sh) libraries.

![License](https://img.shields.io/badge/license-MIT-blue.svg)

## Features

- **Interactive TUI** - Full keyboard-driven interface with vim-style navigation
- **CLI Commands** - Script-friendly commands for automation and AI agents
- **JSON Storage** - Human-readable data format, easy to backup and version control
- **Status Pipeline** - Track applications from saved → applied → interview → offer
- **Search & Filter** - Quickly find applications by company, position, or status
- **Quick Actions** - Change status with single keystrokes (1-8)

## Installation

### From Source

```bash
git clone https://github.com/yourusername/jobtrack.git
cd jobtrack
go build -o jobtrack
./jobtrack
```

### Go Install

```bash
go install github.com/yourusername/jobtrack@latest
```

## Usage

### Interactive TUI

```bash
jobtrack
```

**Keyboard Shortcuts:**

| Key | Action |
|-----|--------|
| `j`/`k` or arrows | Navigate up/down |
| `a` | Add new application |
| `e` | Edit selected |
| `d` | Delete selected |
| `Enter` | View details |
| `1-8` | Quick status change |
| `/` | Search |
| `f` | Filter by status |
| `c` | Clear filters |
| `?` | Toggle help |
| `q` | Quit |

**Status Keys:**
1. Saved
2. Applied
3. Screening
4. Interview
5. Offer
6. Accepted
7. Rejected
8. Withdrawn

### CLI Commands

```bash
# Add new application
jobtrack add --json '{"company":"Acme","position":"Engineer","salary_min":150000}'

# List all applications
jobtrack list
jobtrack list --json

# Get single application (supports partial ID)
jobtrack get abc123
jobtrack get abc123 --json

# Update application
jobtrack update abc123 --json '{"status":"interview","notes":"Phone screen scheduled"}'

# Delete application
jobtrack delete abc123

# Help
jobtrack help
```

## Data Storage

By default, data is stored at:
```
~/.local/share/jobtrack/applications.json
```

Override with environment variable:
```bash
export JOBTRACK_DATA=/path/to/your/applications.json
```

### Sample Data

New installations are seeded with 3 sample applications to help you get started. Delete them with `d` in the TUI or start fresh:

```bash
rm ~/.local/share/jobtrack/applications.json
```

## JSON Schema

```json
{
  "company": "string (required)",
  "position": "string (required)",
  "status": "saved|applied|screening|interview|offer|accepted|rejected|withdrawn",
  "date_applied": "2025-01-15T00:00:00Z",
  "salary_min": 150000,
  "salary_max": 200000,
  "job_url": "https://...",
  "location": "City, State",
  "remote": true,
  "contact_name": "string",
  "contact_email": "string",
  "resume_version": "string",
  "cover_letter": "string",
  "notes": "string",
  "interviews": [
    {
      "date": "2025-01-20T14:00:00Z",
      "type": "phone|video|onsite|technical",
      "notes": "string",
      "with_whom": "string"
    }
  ]
}
```

See [schema.json](schema.json) for the complete JSON Schema specification.

## Local Files (Optional)

For organizing job-related documents, create a `local/` directory:

```
local/
├── postings/       # Job posting files (txt, md, png)
├── resumes/        # Resume versions
└── cover-letters/  # Cover letter templates
```

This directory is gitignored by default.

## Development

### Project Structure

```
├── main.go                 # Entry point, CLI commands
├── internal/
│   ├── model/
│   │   └── application.go  # Data structures, status constants
│   ├── store/
│   │   └── json.go         # JSON persistence, CRUD operations
│   └── tui/
│       ├── app.go          # Main TUI controller
│       ├── list.go         # List view
│       ├── detail.go       # Detail view
│       ├── form.go         # Add/edit form
│       ├── styles.go       # Lip Gloss styling
│       └── keys.go         # Key bindings
├── samples/
│   └── applications.json   # Sample data for new users
└── schema.json             # JSON Schema specification
```

### Building

```bash
go build -o jobtrack
```

### Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions welcome! Please feel free to submit a Pull Request.
