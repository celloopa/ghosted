# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

- **Compile Command** - New `ghosted compile` command to compile `.typ` files to PDFs with descriptive names and auto-update tracker
  ```bash
  ghosted compile abc123                           # By app ID
  ghosted compile local/applications/swe/acme/    # By directory
  ```

- **Open Folder Feature** - Press `o` in TUI detail view to open documents folder in system file manager (cross-platform: macOS, Windows, Linux)

- **Documents Directory Tracking** - New `documents_dir` field in Application model to track location of generated documents

- **TUI URL Input Dialog** ([#10](https://github.com/celloopa/ghosted/issues/10))
  - Press `u` in list view to open URL input dialog
  - Fetches job posting and launches Claude Code with agent context
  - Supports Lever, Greenhouse, Workday, LinkedIn, Ashby, and generic URLs

- **Fetch Command** ([#14](https://github.com/celloopa/ghosted/pull/14))
  - New `ghosted fetch <url>` command to fetch job postings from URLs
  - Support for multiple job boards: Lever, Greenhouse, Workday, LinkedIn, Ashby
  - Auto-detect company name and position from page content
  - Convert HTML to clean markdown with metadata header
  - Save to `local/postings/` directory
  ```bash
  ghosted fetch https://jobs.lever.co/company/job-id
  ghosted fetch --output acme-swe.md https://example.com/job
  ```

- **Agent Pipeline** ([#13](https://github.com/celloopa/ghosted/pull/13))
  - Resume Generator Agent with CV loading, Typst output, and skill matching
  - Cover Letter Generator Agent with experience extraction and 3-section structure
  - Hiring Manager Review Agent with weighted scoring (70+ approval threshold)
  - Tracker Integration Agent with store integration and job type detection
  - 67 comprehensive tests covering all agent functionality

- **Agent Prompt Templates** ([#8](https://github.com/celloopa/ghosted/pull/8))
  - 5 agent prompt templates for the multi-agent document generation pipeline
  - Prompts located in `internal/agent/prompts/`:
    - `parser.md` - Job posting parser (extracts structured JSON)
    - `resume.md` - Resume generator (creates tailored Typst resumes)
    - `cover.md` - Cover letter generator (creates personalized Typst cover letters)
    - `reviewer.md` - Hiring manager reviewer (scores documents with feedback)
    - `tracker.md` - Tracker integration (generates ghosted CLI commands)
