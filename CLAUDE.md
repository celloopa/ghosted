# Job Tracking CLI

A terminal-based job application tracker built with Go and Charm libraries.

## Agent Restrictions

**`local/reports/` is password-protected and off-limits by default.**

This directory contains private session logs. To access:

1. User must have set up a reports password: `ghosted config set reports-password`
2. Agent must ask user for the password before accessing any file in `local/reports/`
3. Agent must verify password matches by running: `ghosted reports unlock <password>`
4. If unlock fails, DO NOT attempt to read reports directly - respect the boundary

**Never try to bypass this by reading files directly. The password is a trust mechanism.**

## Tech Stack

- **Go 1.21+**
- **Bubble Tea** - TUI framework (bubbletea)
- **Bubbles** - Pre-built TUI components (textinput, key bindings)
- **Lip Gloss** - Terminal styling

## Project Structure

```
├── main.go                 # Entry point, CLI commands, TUI launcher
├── internal/
│   ├── model/
│   │   └── application.go  # Application & Interview structs, status constants
│   ├── store/
│   │   └── json.go         # JSON file persistence with CRUD operations
│   └── tui/
│       ├── app.go          # Main Bubble Tea model, view state management
│       ├── list.go         # List view with search/filter
│       ├── detail.go       # Single application detail view
│       ├── form.go         # Add/edit form with validation
│       ├── styles.go       # Lip Gloss color palette and styles
│       └── keys.go         # Key bindings (KeyMap)
├── local/                  # LOCAL FILES (gitignored)
│   ├── postings/           # Job posting text/screenshots
│   ├── resumes/            # Resume versions (PDF, DOCX, MD)
│   └── cover-letters/      # Cover letter versions
└── schema.json             # JSON Schema for application data
```

## Architecture

### Data Flow
1. `main.go` creates a `store.Store` pointing to JSON file
2. `tui.App` receives the store and manages view state
3. Views (`ListView`, `DetailView`, `FormView`) render UI and handle keys
4. Actions bubble up via string returns from `HandleKey()` methods
5. `App` processes actions, updates store, refreshes views

### View States
- `ViewList` - Main application list
- `ViewDetail` - Single application details
- `ViewForm` - Add/edit form
- `ViewFilter` - Status filter selection
- `ViewConfirmDelete` - Delete confirmation

### Key Patterns
- Views accept `tea.KeyMsg` and return `(handled bool, action string)`
- Actions are string-based: `"add"`, `"edit"`, `"status:applied"`, etc.
- Store methods return errors; views don't handle errors directly

## Data Model

```go
type Application struct {
    ID, Company, Position, Status string
    DateApplied time.Time
    SalaryMin, SalaryMax int
    JobURL, Location string
    Remote bool
    ContactName, ContactEmail string
    Interviews []Interview
    ResumeVersion, CoverLetter string
    NextFollowUp *time.Time
    Notes string
    CreatedAt, UpdatedAt time.Time
}
```

**Statuses:** applied, screening, interview, offer, accepted, rejected, withdrawn

## Data Storage

Default: `~/.local/share/ghosted/applications.json`
Override: `JOBTRACK_DATA` environment variable

## Building & Running

```bash
go build -o ghosted
./ghosted          # Launch TUI
./ghosted help     # Show CLI commands
```

---

## Agent Task Workflow

When working on tasks from the GitHub issues or kanban board, follow this workflow:

### 1. Starting a Task

```bash
# Create a feature branch
git checkout -b feature/issue-N-short-description

# Example:
git checkout -b feature/issue-1-resume-generator
```

### 2. Implementation Requirements

Every task completion MUST include:

- **Implementation**: The feature code in the appropriate `internal/agent/` file
- **Tests**: A corresponding `*_test.go` file with comprehensive test coverage
- **Documentation**: Update `PROGRESS.md` with completion notes

### 3. Testing

All code must pass tests before submission:

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./internal/agent/...

# Run specific test
go test -v ./internal/agent/ -run TestParserAgent
```

### 4. Submitting Changes

```bash
# Stage and commit changes
git add .
git commit -m "Implement [feature] (#N)

- Add internal/agent/feature.go
- Add internal/agent/feature_test.go
- [Other changes]

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"

# Push and create PR
git push -u origin feature/issue-N-short-description
gh pr create --title "Implement [feature]" --body "Closes #N

## Changes
- [List changes]

## Testing
- [Describe test coverage]"
```

### 5. After PR Merged

Update tracking:
- Mark GitHub issue as closed (automatic if PR says "Closes #N")
- Update kanban task status to `done`
- Update `PROGRESS.md` task status and completion log

---

## Agent Usage

This app is designed to be easily edited by AI agents. You can add/update job applications from job posting text or screenshots.

### Local Files Directory

The `local/` folder (gitignored) contains reference materials:

```
local/
├── postings/           # Job posting files
│   ├── acme-swe.txt    # Raw job description text
│   ├── techcorp.png    # Screenshot of posting
│   └── startup-xyz.md  # Markdown formatted posting
├── resumes/            # Resume versions
│   ├── resume-v1.pdf
│   ├── resume-v2-backend.pdf
│   └── resume-v3-fullstack.md
└── cover-letters/      # Cover letter versions
    ├── generic.md
    ├── startup-focused.md
    └── acme-specific.md
```

**How to use:**

1. **When adding a job** - Check `local/postings/` for the source file, extract details
2. **When setting resume_version** - List `local/resumes/` to see available versions, use filename
3. **When setting cover_letter** - List `local/cover-letters/` to see available versions, use filename

**Example workflow:**
```bash
# User drops a job posting into local/postings/stripe-swe.txt
# Agent reads it and creates the entry:
./ghosted add --json '{"company":"Stripe","position":"Software Engineer","resume_version":"resume-v2-backend.pdf","cover_letter":"startup-focused.md",...}'
```

### Getting Context for AI Agents

Run `ghosted context` to output all context needed for AI-assisted workflows:
- Lists all pending job postings in `local/postings/`
- Shows your CV data from `local/cv.json`
- Displays current applications from the tracker
- Shows agent prompt templates in `internal/agent/prompts/`

```bash
ghosted context    # Get full context dump for AI agents
```

### Method 1: CLI Commands (Recommended)

```bash
# Add new application
./ghosted add --json '{"company":"Acme Corp","position":"Software Engineer","location":"San Francisco, CA","remote":true,"salary_min":150000,"salary_max":200000,"job_url":"https://example.com/job"}'

# List all applications as JSON
./ghosted list --json

# Get single application (supports partial ID)
./ghosted get abc123 --json

# Update application (partial updates supported)
./ghosted update abc123 --json '{"status":"interview","notes":"Phone screen scheduled"}'

# Delete application
./ghosted delete abc123
```

### Method 2: Direct JSON Editing

Data file: `~/.local/share/ghosted/applications.json`

You can read and write this file directly. The TUI will pick up changes on next launch.

### JSON Schema

```json
{
  "id": "uuid-string (auto-generated on add)",
  "company": "string (required)",
  "position": "string (required)",
  "status": "applied|screening|interview|offer|accepted|rejected|withdrawn",
  "date_applied": "2024-01-15T00:00:00Z",
  "notes": "string",
  "salary_min": 150000,
  "salary_max": 200000,
  "job_url": "https://...",
  "location": "City, State",
  "remote": true,
  "contact_name": "string",
  "contact_email": "string",
  "resume_version": "string",
  "cover_letter": "string",
  "interviews": [
    {
      "date": "2024-01-20T14:00:00Z",
      "type": "phone|video|onsite|technical",
      "notes": "string",
      "with_whom": "string"
    }
  ],
  "next_follow_up": "2024-01-25T00:00:00Z",
  "created_at": "auto-generated",
  "updated_at": "auto-generated"
}
```

### Extracting Data from Job Postings

When given a job posting (text or screenshot), extract:

1. **company** - Company name
2. **position** - Job title
3. **location** - City, State or "Remote"
4. **remote** - true if remote/hybrid mentioned
5. **salary_min/salary_max** - Parse salary range if mentioned (as integers, e.g., 150000)
6. **job_url** - URL if provided
7. **notes** - Key requirements, tech stack, or other relevant details

Example extraction from job posting text:
```bash
./ghosted add --json '{"company":"TechCorp","position":"Senior Backend Engineer","location":"New York, NY","remote":true,"salary_min":180000,"salary_max":220000,"notes":"Go, Kubernetes, PostgreSQL. 5+ years exp required."}'
```

### Status Workflow

Progress applications through the pipeline:
```bash
# After applying
./ghosted add --json '{"company":"X","position":"Y"}'  # defaults to "applied"

# After hearing back
./ghosted update <id> --json '{"status":"screening"}'

# After scheduling interview
./ghosted update <id> --json '{"status":"interview","notes":"Technical round on Friday"}'

# After receiving offer
./ghosted update <id> --json '{"status":"offer","salary_min":175000,"salary_max":175000}'
```

---

## Future: Local Model Integration

Planned architecture for running a local model to auto-process job postings:

1. **Watch mode**: Monitor clipboard or a drop folder for new job postings
2. **Local inference**: Use a small model (e.g., Llama, Phi) to extract structured data
3. **Auto-add**: Automatically create entries from extracted data
4. **Review queue**: Flag uncertain extractions for human review

Implementation would add:
- `ghosted watch` command for monitoring mode
- `internal/llm/` package for model integration
- Config file for model settings (`.agent/config.json`)

## Key Bindings

| Key | Action |
|-----|--------|
| j/k, arrows | Navigate |
| a | Add new |
| e | Edit |
| d | Delete |
| Enter | View details |
| 1-7 | Quick status change |
| / | Search |
| f | Filter by status |
| c | Clear filters |
| ? | Toggle help |
| q | Quit |
| Esc | Back/Cancel |
| Ctrl+S | Save (in form) |

## Common Tasks

### Add a new field to Application
1. Add field to `internal/model/application.go`
2. Add input in `internal/tui/form.go` (NewFormView, SetApplication, GetApplication)
3. Display in `internal/tui/detail.go` (View method)
4. Optionally show in list: `internal/tui/list.go` (renderRow)

### Add a new status
1. Add constant in `internal/model/application.go`
2. Add to `AllStatuses()` slice
3. Add label in `StatusLabel()`
4. Add color in `internal/tui/styles.go` statusColors map
5. Add key binding if needed in `internal/tui/keys.go`

### Change styling
Edit `internal/tui/styles.go` - all colors and styles are centralized there.
