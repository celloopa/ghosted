# Multi-Agent Document Generation Pipeline - Progress Tracker

> **Last Updated:** 2026-01-16 (Added compile command and open folder feature)
> **Project:** ghosted
> **Kanban Project ID:** `b666852b-0ef9-4ee0-8d91-a7f341697897`
> **GitHub Repo:** `celloopa/ghosted`

## Overview

Building a multi-agent pipeline that automates job application document generation:

```
[Job Posting] → [Parser] → [Resume Gen] → [Cover Letter Gen] → [Reviewer] → [Tracker]
```

When a user drops a job posting into `local/postings/`, agents will:
1. Parse the posting and extract structured data
2. Generate a tailored resume using Typst
3. Generate a tailored cover letter using Typst
4. Have a "hiring manager" agent review and score the documents
5. On approval, add to the ghosted tracker with document references

---

## Task Progress

### Phase 1: Foundation

| Status | Task | ID | Notes |
|--------|------|----|-------|
| `[x]` | Design multi-agent pipeline architecture | `72b33fb7-8924-4b3f-8c38-11001f93436d` | ✅ Created `internal/agent/config.go`, `pipeline.go`, and runtime config |
| `[x]` | Create agent prompt templates | `88ab2620-aeae-4842-a45c-e6453f3ee1b0` | Prompts for each agent role |

### Phase 2: Core Agents

| Status | Task | ID | Notes |
|--------|------|----|-------|
| `[x]` | Implement Posting Parser Agent | `bcaee5cf-e414-4d4f-b3c4-61a6da2ed451` | ✅ Created `parser.go` with tests |
| `[x]` | Implement Resume Generator Agent | `c4f2f83d-2f81-49d3-95d5-4af1dcc9f98c` | ✅ `resume_generator.go` + 14 tests |
| `[x]` | Implement Cover Letter Generator Agent | `e28db1da-93ee-418f-a0a7-a88605f3c398` | ✅ `coverletter_generator.go` + 13 tests |
| `[x]` | Implement Hiring Manager Review Agent | `05b75d49-34ad-47a8-b019-532d73b8469d` | ✅ `reviewer.go` + 17 tests |

### Phase 3: Integration

| Status | Task | ID | Notes |
|--------|------|----|-------|
| `[x]` | Implement Tracker Integration | `1dd2fb3f-6555-4baf-b7bb-d77056c1968d` | ✅ `tracker.go` + 23 tests |
| `[x]` | Add `ghosted apply` CLI command | `b9615c5d-fd52-418c-89da-4bec8c724f83` | ✅ Implemented with --dry-run, --auto-approve |
| `[x]` | Add TUI URL input dialog | `b1fb39cb-634a-46e3-8f28-9815036dffaf` | ✅ Press 'u' to fetch URL and launch Claude |
| `[ ]` | Add watch mode for automatic processing | `2cdc1317-ddc0-408b-ab3d-b6fb92e2887b` | Nice-to-have: monitor folder |

### Phase 4: Agent Automation & Training Data (After Phase 3)

*These tasks depend on Phase 3 completion. Priority order within phase:*

| Status | Priority | Task | ID | Notes |
|--------|----------|------|----|-------|
| `[ ]` | P0 | Add --non-interactive flag to ghosted apply | `bdfff0fc` | Blocker for AI agent usage |
| `[ ]` | P1 | Persist intermediate outputs (parsed.json, review.json) | `225c8b1b` | Debugging + training data |
| `[ ]` | P2 | Add session logging (optional, off by default) | `ae201fb8` | Privacy-first, user opt-in only |
| `[ ]` | P3 | Add ghosted export-training command | `e1cdba88` | Export training pairs |
| `[ ]` | P4 | Research and test 8B parameter models | `94f2594a` | Local model evaluation |

---

## GitHub Issues

All remaining tasks are tracked as GitHub issues. Each issue includes full implementation details and testing requirements.

| Issue | Title | Phase |
|-------|-------|-------|
| [#1](https://github.com/celloopa/ghosted/issues/1) | Implement Resume Generator Agent | Core |
| [#2](https://github.com/celloopa/ghosted/issues/2) | Implement Cover Letter Generator Agent | Core |
| [#3](https://github.com/celloopa/ghosted/issues/3) | Implement Hiring Manager Review Agent | Core |
| [#4](https://github.com/celloopa/ghosted/issues/4) | Implement Tracker Integration | Integration |
| [#5](https://github.com/celloopa/ghosted/issues/5) | Add `ghosted apply` CLI command | Integration |
| [#6](https://github.com/celloopa/ghosted/issues/6) | Add watch mode for automatic processing | Nice-to-have |
| [#7](https://github.com/celloopa/ghosted/issues/7) | Create agent prompt templates | Foundation |

### Workflow

1. Pick an issue from the list above
2. Create a feature branch: `git checkout -b feature/issue-N-short-description`
3. Implement the feature with tests
4. Ensure all tests pass: `go test ./...`
5. Submit a pull request referencing the issue
6. Update task status in kanban and PROGRESS.md

---

## File Structure (Planned)

```
ghosted/
├── main.go                          # Add 'apply' and 'watch' commands
├── internal/
│   ├── agent/
│   │   ├── pipeline.go              # ✅ Orchestrator - runs all agents in sequence
│   │   ├── config.go                # ✅ Agent configuration loader
│   │   ├── parser.go                # ✅ Posting Parser Agent
│   │   ├── parser_test.go           # ✅ Parser agent tests
│   │   ├── resume_generator.go      # ✅ Resume Generator Agent
│   │   ├── resume_generator_test.go # ✅ 14 tests
│   │   ├── coverletter_generator.go # ✅ Cover Letter Generator Agent
│   │   ├── coverletter_generator_test.go # ✅ 13 tests
│   │   ├── reviewer.go              # ✅ Hiring Manager Review Agent
│   │   ├── reviewer_test.go         # ✅ 17 tests
│   │   ├── tracker.go               # ✅ Tracker Integration
│   │   ├── tracker_test.go          # ✅ 23 tests
│   │   └── prompts/                 # ✅ Agent prompt templates
│   │       ├── parser.md            # ✅ Parser agent prompt
│   │       ├── resume.md            # ✅ Resume generator prompt
│   │       ├── cover.md             # ✅ Cover letter generator prompt
│   │       ├── reviewer.md          # ✅ Reviewer agent prompt
│   │       └── tracker.md           # ✅ Tracker integration prompt
│   └── ...
├── local/
│   ├── postings/                    # Drop job postings here
│   │   ├── processed/               # Successfully processed postings
│   │   └── needs-review/            # Rejected - needs manual review
│   ├── resumes/                     # Generated resume PDFs
│   ├── cover-letters/               # Generated cover letter PDFs
│   └── document-generation/
│       ├── .agent/
│       │   └── config.json          # Runtime configuration (references local files)
│       ├── cv.json                  # Source CV data
│       ├── resume.typ               # Base resume template
│       └── coverletter.typ          # Base cover letter template
└── PROGRESS.md                      # This file
```

---

## Data Schemas

### Parsed Job Posting (Parser → Generators)

```json
{
  "company": "Twitch",
  "position": "Software Engineer",
  "team": "BITS Team",
  "location": "Seattle, WA",
  "remote": false,
  "salary_min": null,
  "salary_max": null,
  "job_url": null,
  "requirements": ["1+ years experience", "..."],
  "bonus_skills": ["Golang", "AWS", "..."],
  "keywords": ["live streaming", "interactive", "..."],
  "tech_stack": ["Go", "TypeScript", "React"],
  "company_values": ["community", "collaboration"]
}
```

### Review Result (Reviewer → Tracker)

```json
{
  "approved": true,
  "overall_score": 85,
  "resume_review": {
    "score": 88,
    "strengths": ["..."],
    "weaknesses": ["..."],
    "suggestions": ["..."]
  },
  "cover_letter_review": {
    "score": 82,
    "strengths": ["..."],
    "weaknesses": ["..."],
    "suggestions": ["..."]
  },
  "recommendation": "Approve with minor edits"
}
```

---

## CLI Commands (Planned)

```bash
# Run full pipeline on a posting
ghosted apply local/postings/twitch-swe.md

# Options
ghosted apply --auto-approve posting.md    # Skip review, add directly
ghosted apply --dry-run posting.md         # Generate docs, don't add to tracker
ghosted apply --reviewer-only resume.pdf cover.pdf posting.md

# Watch mode (Phase 3)
ghosted watch                              # Monitor postings folder
ghosted watch --auto-approve               # Auto-approve all
```

---

## Current Focus

**Completed:**
- `ghosted fetch <url>` command - fetches job postings from URLs
- `ghosted apply <posting>` command - runs full pipeline with --dry-run support

**Next task to work on:** Phase 4 improvements:
1. Add `--non-interactive` flag to ghosted apply (`bdfff0fc`) - for AI agent usage

**Blockers:** None - Phase 3 complete!

**Questions/Decisions:**
- [x] Should agents be Claude Code subagents or Go code with LLM API calls? → **Claude Code subagents**
- [x] Approval threshold: 70/100 score? → **Yes, implemented in reviewer.go**
- [x] Keep rejected postings or delete them? → **Save feedback to `{company}-feedback.json`, keep posting for revision**
- [x] URL support for apply? → **Fetch-first workflow: `ghosted fetch` then `ghosted apply`**
- [ ] Small model fine-tuning strategy? → **Collect training data first, then test 8B models per step**

---

## Agent Interaction Logging (Optional, Privacy-First)

### Design Principles

- **OFF by default** - no logging unless user explicitly enables
- **No annoying prompts** - user must seek out the option
- **Local-first** - sharing requires deliberate action
- **Documented in help** - not in dialogs or banners

### Three Modes

| Mode | Command | Behavior |
|------|---------|----------|
| **Off** (default) | n/a | No logging whatsoever |
| **Local** | `ghosted config set logging local` | Saves to `local/reports/`, never leaves machine |
| **Share** | `ghosted config set logging share` | Opt-in feedback to developer (requires explicit acknowledgment) |

### Report Location (when local logging enabled)

```
local/reports/
├── YYYY-MM-DD-{company}-{role}-session.json   # Machine-readable
└── YYYY-MM-DD-{company}-{role}-session.md     # Human-readable
```

### Use Cases

- **Off**: Default for all users, complete privacy
- **Local**: Power users iterating on prompts, self-hosted fine-tuning
- **Share**: Users who want to help improve ghosted (must actively enable)

### Manual Session Log (Developer Reference)

- **2026-01-16**: Microsoft UX Engineer II
  - Score: 78/100 (Approved)
  - Automation rate: 43% (3/7 steps)
  - Key gap: `ghosted apply` requires TTY
  - Reports: `local/reports/2026-01-16-microsoft-ux_engineer-session.*`
  - *Note: This was logged manually for framework development*

---

## Training Data Pipeline (Planned)

### Phase 1: Collection

Collect high-quality input/output pairs from manual orchestrator runs:

| Agent | Input | Output | Format |
|-------|-------|--------|--------|
| Parser | posting.md | parsed.json | JSON extraction |
| Resume | parsed.json + cv.json | resume.typ | Typst generation |
| Cover | parsed.json + cv.json + resume.typ | cover-letter.typ | Typst generation |
| Reviewer | all docs | review.json | Scoring + reasoning |

### Phase 2: Export

```bash
# Planned command
ghosted export-training local/applications/ux-design/microsoft-ux_engineer/
```

Output format:
```json
{
  "step": "parser",
  "input": "...",
  "output": "...",
  "metadata": { "company", "position", "review_score" }
}
```

### Phase 3: Fine-tuning

Target models (8B parameter class):
- **Qwen2.5-7B-Instruct** - Strong at structured output
- **Llama-3.1-8B-Instruct** - Good general reasoning
- **Mistral-7B-Instruct-v0.3** - Fast inference

Training strategy:
1. Start with Parser (most constrained task)
2. Add Resume/Cover (template-guided generation)
3. Reviewer last (requires most reasoning)

### Phase 4: Integration

Replace Claude API calls with local model inference:
- Use Ollama or llama.cpp for serving
- Orchestrator (Claude) only handles error recovery and complex decisions
- Target: 80%+ automation with local models

---

## Session Notes (for context reset)

### What's Done
1. **All 4 core agents implemented** with 80+ tests:
   - `resume_generator.go` - CV loading, Typst output, skill matching
   - `coverletter_generator.go` - Experience extraction, consistency with resume
   - `reviewer.go` - Scoring (70+ approval), detailed feedback
   - `tracker.go` - Store integration, file organization

2. **`ghosted fetch` command implemented**:
   - `internal/fetch/fetcher.go` - URL fetching, HTML parsing, markdown conversion
   - Supports: Lever, Greenhouse, Workday, LinkedIn, Ashby, generic
   - Saves to `local/postings/` with metadata header
   - Usage: `ghosted fetch https://jobs.lever.co/company/123`

### What's Next
1. **Phase 4 improvements** - Now that the pipeline is complete:
   - Add `--non-interactive` flag to `ghosted apply` for agent automation
   - Persist intermediate outputs (parsed.json, review.json)
   - Add session logging (optional, privacy-first)

2. **Test the full TUI workflow**:
   ```bash
   ./ghosted           # Launch TUI
   # Press 'u' to open URL input dialog
   # Paste a job posting URL and press Enter
   # Claude Code launches with full context
   ```

---

## Completed Work Log

### 2026-01-16: Compile Command and Open Folder Feature

**New CLI command: `ghosted compile`**
- Compiles `.typ` files to PDFs with descriptive names
- Auto-updates tracker with PDF filenames
- Usage: `ghosted compile <app-id>` or `ghosted compile <documents-dir>`
- Naming: `{company}-{position}-resume.pdf`, `{company}-{position}-cover.pdf`

**New field: `documents_dir`**
- Added to Application model for tracking document folder location
- Enables the "open folder" TUI feature
- Set via: `ghosted add --json '{"documents_dir":"local/applications/swe/company/"}'`
- Updated via: `ghosted update <id> --json '{"documents_dir":"..."}'`

**New TUI keybinding: `o` (open folder)**
- Press `o` in detail view to open documents folder in system file manager
- Cross-platform: macOS (open), Windows (explorer), Linux (xdg-open)
- Shows "No documents folder set" if `documents_dir` is empty

**Files modified:**
- `main.go` - Added `cmdCompile()` function with Typst compilation
- `internal/model/application.go` - Added `DocumentsDir` field
- `internal/tui/app.go` - Added `openFolder()` function and handler
- `internal/tui/detail.go` - Added folder display and `o` keybinding
- `internal/tui/keys.go` - Added `OpenFolder` key binding
- `internal/agent/prompts/tracker.md` - Updated workflow to use compile command
- `CLAUDE.md` - Documented new field and keybindings

**Agent workflow updated:**
```bash
# 1. Add to tracker with documents_dir
ghosted add --json '{"company":"Acme","position":"SWE","documents_dir":"local/applications/swe/acme/"}'

# 2. Compile .typ to PDF (required step)
ghosted compile local/applications/swe/acme/
```

---

### 2026-01-16: TUI URL Input Dialog for Agent Pipeline

**Files created/modified:**
- `internal/tui/urlinput.go` - New URLInputView component with:
  - Text input for job posting URLs
  - URL validation (http/https prefix)
  - Centered dialog overlay
  - Usage hints for supported job boards
  - "What happens next" workflow description
- `internal/tui/app.go` - Added:
  - `ViewURLInput` view state
  - `urlInputView` field on App struct
  - `fetchResultMsg` and `claudeLaunchedMsg` message types
  - `handleURLInputKey()` - Key handling for URL input view
  - `fetchPosting()` - Async fetch of job posting URL
  - `launchClaude()` - Launch Claude Code with agent context via `tea.ExecProcess`
- `internal/tui/keys.go` - Added:
  - `URLInput` key binding (`u` key)
  - Updated `ShortHelp()` and `FullHelp()` to include new binding
- `internal/tui/list.go` - Added:
  - Handler for `urlinput` action
  - "AI Agent" section in help dialog with `u` keybinding

**User workflow:**
1. Press `u` in list view
2. URL input dialog appears (centered overlay)
3. Paste job posting URL and press Enter
4. ghosted fetches the posting to `local/postings/`
5. Claude Code is launched with full context:
   - CV data from `local/cv.json`
   - Posting list from `local/postings/`
   - Prompt templates from `internal/agent/prompts/`
   - Instructions to run the agent pipeline

**Technical notes:**
- Uses `tea.ExecProcess` for proper terminal handoff to Claude
- Fetches context via `ghosted context` command
- Builds full prompt with workflow instructions for Claude
- Returns to TUI after Claude exits

---

### 2026-01-16: All Core Agents Implemented (Tasks 1-4)

**Files created:**
- `internal/agent/resume_generator.go` - Resume Generator Agent
- `internal/agent/resume_generator_test.go` - 14 tests
- `internal/agent/coverletter_generator.go` - Cover Letter Generator Agent
- `internal/agent/coverletter_generator_test.go` - 13 tests
- `internal/agent/reviewer.go` - Hiring Manager Review Agent
- `internal/agent/reviewer_test.go` - 17 tests
- `internal/agent/tracker.go` - Tracker Integration Agent
- `internal/agent/tracker_test.go` - 23 tests

**Total: 67 new tests (80+ total in agent package)**

**Resume Generator Agent features:**
- CV loading from JSON Resume format
- Typst template loading
- Skill matching between CV and job requirements
- Output path generation with job-type folders
- PDF compilation via Typst CLI
- Prompt generation for AI-based tailoring

**Cover Letter Generator Agent features:**
- CV and resume loading for consistency
- Relevant experience extraction (scored by job match)
- Three-section structure (Hook, Bridge, Value)
- Typst output with modern-cv template

**Hiring Manager Review Agent features:**
- Document loading and validation
- Weighted scoring (60% resume, 40% cover letter)
- Requirement match analysis (met/missing/bonus)
- Approval threshold: 70/100
- Recommendation categories: Approve, Approve with edits, Revise, Not recommended
- Detailed feedback with strengths, weaknesses, suggestions

**Tracker Integration Agent features:**
- Application creation in ghosted store
- Notes generation from tech stack, score, requirements
- Job type auto-detection (fe-dev, ux-design, product-design, swe)
- Application folder path generation
- Posting archival to processed folder
- Rejection feedback file creation

**Run tests:**
```bash
go test -v ./internal/agent/...
```

---

### 2026-01-16: Agent Prompt Templates (Task 7)

**Files created:**
- `internal/agent/prompts/parser.md` - Job posting parser prompt
- `internal/agent/prompts/resume.md` - Resume generator prompt with Typst format
- `internal/agent/prompts/cover.md` - Cover letter generator prompt with Typst format
- `internal/agent/prompts/reviewer.md` - Hiring manager review prompt with scoring criteria
- `internal/agent/prompts/tracker.md` - Tracker integration prompt for ghosted CLI

**Key features:**
- Parser: Structured JSON output with requirements, tech_stack, keywords, company_values
- Resume: Typst template with tailoring principles, keyword optimization, skills highlighting
- Cover Letter: Three-section structure (Hook, Bridge, Value) with tone guidelines
- Reviewer: Scoring criteria (40% requirements, 30% experience, 20% communication, 10% culture)
- Tracker: CLI command generation with file organization and validation

---

### 2026-01-16: Posting Parser Agent (Task 2)

**Files created/modified:**
- `internal/agent/parser.go` - ParserAgent implementation with:
  - File type detection (`.md`, `.txt`, `.png`, `.jpg`, `.jpeg`)
  - Content reading for text and image markers
  - System/user prompt generation for AI parsing
  - JSON parsing and validation
  - Output schema definition
- `internal/agent/parser_test.go` - Test coverage for all parser methods
- `internal/agent/config.go` - Updated `ParsedPosting` struct with new fields:
  - `Team`, `BonusSkills`, `Keywords`, `CompanyValues`
- `internal/agent/pipeline.go` - Updated `runParserStep()` to use ParserAgent
  - Added `extractBasicInfo()` fallback for basic text extraction

**Test coverage:**
- `TestParserAgent_SupportedExtensions` - File type detection
- `TestParserAgent_IsImageFile` - Image identification
- `TestParserAgent_ReadPosting` - File reading
- `TestParserAgent_ParseJSON` - JSON parsing with validation
- `TestParserAgent_GetSystemPrompt` - Prompt content verification
- `TestExtractBasicInfo` - Basic info extraction fallback

---

### 2026-01-16: Pipeline Architecture (Task 1)

**Files created:**
- `internal/agent/config.go` - Agent types, config structs, data models
- `internal/agent/pipeline.go` - Pipeline orchestrator with state management
- `local/document-generation/.agent/config.json` - Runtime configuration

**Key decisions implemented:**
- Sequential pipeline: Parser → Resume → Cover → Reviewer → Tracker
- JSON intermediate format between agents
- State persistence for resume/retry capability
- Typst for PDF generation

**Data types defined:**
- `ParsedPosting` - Structured job posting data
- `GeneratedDocuments` - Paths to generated .typ and .pdf files
- `ReviewResult` - Approval status, score, and feedback

---

## Status Legend

- `[ ]` - Not started
- `[~]` - In progress
- `[x]` - Completed
- `[!]` - Blocked

---

## How to Update This File

When working on a task:
1. Change status from `[ ]` to `[~]`
2. Add notes about approach/blockers
3. Update "Current Focus" section

When completing a task:
1. Change status from `[~]` to `[x]`
2. Update "Current Focus" to next task
3. Add completion notes if relevant

Use vibe kanban MCP to sync:
```
mcp__vibe_kanban__update_task(task_id, status="inprogress"|"done")
```
