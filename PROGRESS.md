# Multi-Agent Document Generation Pipeline - Progress Tracker

> **Last Updated:** 2026-01-16
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
| `[ ]` | Implement Resume Generator Agent | `c4f2f83d-2f81-49d3-95d5-4af1dcc9f98c` | Input: JSON + cv.json → Output: .typ + .pdf |
| `[ ]` | Implement Cover Letter Generator Agent | `e28db1da-93ee-418f-a0a7-a88605f3c398` | Input: JSON + profile → Output: .typ + .pdf |
| `[ ]` | Implement Hiring Manager Review Agent | `05b75d49-34ad-47a8-b019-532d73b8469d` | Score docs, approve/reject with feedback |

### Phase 3: Integration

| Status | Task | ID | Notes |
|--------|------|----|-------|
| `[ ]` | Implement Tracker Integration | `1dd2fb3f-6555-4baf-b7bb-d77056c1968d` | Add to ghosted, organize files |
| `[ ]` | Add `ghosted apply` CLI command | `b9615c5d-fd52-418c-89da-4bec8c724f83` | User-facing command |
| `[ ]` | Add watch mode for automatic processing | `2cdc1317-ddc0-408b-ab3d-b6fb92e2887b` | Nice-to-have: monitor folder |

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
│   │   ├── resume_generator.go      # Resume Generator Agent
│   │   ├── coverletter_generator.go # Cover Letter Generator Agent
│   │   ├── reviewer.go              # Hiring Manager Review Agent
│   │   └── tracker.go               # Tracker Integration
│   └── ...
├── local/
│   ├── postings/                    # Drop job postings here
│   │   ├── processed/               # Successfully processed postings
│   │   └── needs-review/            # Rejected - needs manual review
│   ├── resumes/                     # Generated resume PDFs
│   ├── cover-letters/               # Generated cover letter PDFs
│   └── document-generation/
│       ├── .agent/
│       │   ├── config.json          # ✅ Runtime configuration
│       │   └── prompts/
│       │       ├── parser.md        # Parser agent prompt
│       │       ├── resume.md        # Resume generator prompt
│       │       ├── coverletter.md   # Cover letter generator prompt
│       │       └── reviewer.md      # Reviewer agent prompt
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

**Next task to work on:** Implement Resume Generator Agent (`c4f2f83d`)

**Blockers:** None

**Questions/Decisions:**
- [x] Should agents be Claude Code subagents or Go code with LLM API calls? → **Claude Code subagents**
- [ ] Approval threshold: 70/100 score?
- [ ] Keep rejected postings or delete them?

---

## Completed Work Log

### 2026-01-16: Agent Prompt Templates (Task 7)

**Files created:**
- `local/document-generation/.agent/prompts/parser.md` - Job posting parser prompt
- `local/document-generation/.agent/prompts/resume.md` - Resume generator prompt with Typst format
- `local/document-generation/.agent/prompts/cover.md` - Cover letter generator prompt with Typst format
- `local/document-generation/.agent/prompts/reviewer.md` - Hiring manager review prompt with scoring criteria
- `local/document-generation/.agent/prompts/tracker.md` - Tracker integration prompt for ghosted CLI

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
