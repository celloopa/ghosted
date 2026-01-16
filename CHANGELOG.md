# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

## [0.7.0-beta] - 2026-01-16

### Added

- **Unified Fetch Command** - Auto-detects job postings vs CV
  - `ghosted fetch <url|domain>` now handles both job postings and CVs
  - **Job posting fetch**: Any URL with a path → saves to `local/postings/`
    - Support for Lever, Greenhouse, Workday, LinkedIn, Ashby, and generic HTML
    - Auto-detect company name and position from page content
    - Convert HTML to clean markdown with metadata header
  - **CV fetch**: Bare domain or `/cv.json` path → saves to `local/cv.json`
    - Fetches JSON Resume format from `{domain}/cv.json`
    - Extracts name and label from JSON Resume basics
  - **TUI fetch view**: Press `f` from the list view to fetch via the TUI
    - URL input with auto-detection
    - Async fetch with loading state and result display
  ```bash
  ghosted fetch https://jobs.lever.co/company/job-id   # Job posting
  ghosted fetch cello.design                           # CV from domain/cv.json
  ghosted fetch https://example.com/cv.json            # CV from explicit URL
  ghosted fetch --output acme-swe.md https://example.com/job
  ```

- **Agent Pipeline** ([#13](https://github.com/celloopa/ghosted/pull/13))
  - Resume Generator Agent with CV loading, Typst output, and skill matching
  - Cover Letter Generator Agent with experience extraction and 3-section structure
  - Hiring Manager Review Agent with weighted scoring (70+ approval threshold)
  - Tracker Integration Agent with store integration and job type detection
  - 67 comprehensive tests covering all agent functionality

- **Apply Command** - Full pipeline CLI command
  - New `ghosted apply <posting>` runs the complete document generation pipeline
  - Parser → Resume → Cover Letter → Reviewer → Tracker integration
  - Flags: `--dry-run` (generate without tracking), `--auto-approve` (skip review)
  ```bash
  ghosted apply local/postings/acme-swe.md
  ghosted apply --dry-run local/postings/test.md
  ghosted apply --auto-approve local/postings/acme-swe.md
  ```

- **Context Command** - AI agent context dump
  - New `ghosted context` outputs all context needed for AI-assisted workflows
  - Shows pending postings, CV data, current applications, and agent prompts
  - Organized by job type folders with file listings

- **Compile Command** - Typst compilation with tracker update
  - New `ghosted compile <id|dir>` compiles .typ files to PDF
  - Automatically updates tracker with resume/cover letter references
  - Opens output folder after compilation
  ```bash
  ghosted compile abc123
  ghosted compile local/applications/swe/acme/
  ```

- **CV Fetch Command** ([#9](https://github.com/celloopa/ghosted/issues/9))
  - New `ghosted cv fetch <website>` command to fetch CV from remote websites
  - Automatically constructs URL as `https://<website>/cv.json`
  - Validates JSON structure before saving
  - Creates timestamped backup of existing CV before overwriting
  - Displays name and title from fetched CV
  ```bash
  ghosted cv fetch cello.design
  # Fetches https://cello.design/cv.json → local/cv.json
  ```

- **TUI Improvements**
  - Help view is now a centered overlay dialog
  - Expanded help with CLI commands and tips

- **Agent Prompt Templates** ([#8](https://github.com/celloopa/ghosted/pull/8))
  - 5 agent prompt templates for the multi-agent document generation pipeline
  - Prompts located in `internal/agent/prompts/`:
    - `parser.md` - Job posting parser (extracts structured JSON)
    - `resume.md` - Resume generator (creates tailored Typst resumes)
    - `cover.md` - Cover letter generator (creates personalized Typst cover letters)
    - `reviewer.md` - Hiring manager reviewer (scores documents with feedback)
    - `tracker.md` - Tracker integration (generates ghosted CLI commands)

### Changed

- **TUI key bindings**: `s` for filter by status, `f` for fetch (previously `f` for filter)
