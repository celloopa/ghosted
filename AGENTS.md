# GHOSTED - Agent Instructions

A terminal-based job application tracker built with Go and Charm libraries.

## Quick Start for Agents

This app is designed to be easily edited by AI agents. You can add/update job applications from job posting text or screenshots.

### CLI Commands (Recommended)

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

### Direct JSON Editing

Data file: `~/.local/share/ghosted/applications.json`

You can read and write this file directly. The TUI will pick up changes on next launch.

## JSON Schema

```json
{
  "id": "uuid-string (auto-generated on add)",
  "company": "string (required)",
  "position": "string (required)",
  "status": "saved|applied|screening|interview|offer|accepted|rejected|withdrawn",
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

## Local Files Directory

The `local/` folder (gitignored) contains reference materials:

```
local/
├── postings/           # Job posting files (txt, md, png screenshots)
├── resumes/            # Resume versions (PDF, DOCX, MD)
└── cover-letters/      # Cover letter versions
```

**How to use:**

1. **When adding a job** - Check `local/postings/` for the source file, extract details
2. **When setting resume_version** - List `local/resumes/` to see available versions, use filename
3. **When setting cover_letter** - List `local/cover-letters/` to see available versions, use filename

## Extracting Data from Job Postings

When given a job posting (text or screenshot), extract:

1. **company** - Company name
2. **position** - Job title
3. **location** - City, State or "Remote"
4. **remote** - true if remote/hybrid mentioned
5. **salary_min/salary_max** - Parse salary range if mentioned (as integers, e.g., 150000)
6. **job_url** - URL if provided
7. **notes** - Key requirements, tech stack, or other relevant details

Example:
```bash
./ghosted add --json '{"company":"TechCorp","position":"Senior Backend Engineer","location":"New York, NY","remote":true,"salary_min":180000,"salary_max":220000,"notes":"Go, Kubernetes, PostgreSQL. 5+ years exp required."}'
```

## Status Workflow

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

**Statuses:** saved, applied, screening, interview, offer, accepted, rejected, withdrawn

## Environment

- **Data file:** `~/.local/share/ghosted/applications.json`
- **Override:** Set `GHOSTED_DATA` environment variable

## Project Structure

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
│       ├── styles.go       # Styling
│       └── keys.go         # Key bindings
└── schema.json             # JSON Schema specification
```

## Building

```bash
go build -o ghosted
./ghosted          # Launch TUI
./ghosted help     # Show CLI commands
```
