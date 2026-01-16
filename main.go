package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/celloopa/ghosted/internal/agent"
	"github.com/celloopa/ghosted/internal/fetch"
	"github.com/celloopa/ghosted/internal/model"
	"github.com/celloopa/ghosted/internal/store"
	"github.com/celloopa/ghosted/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Determine data file location
	dataPath := getDataPath()

	// Initialize store
	s, err := store.New(dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing store: %v\n", err)
		os.Exit(1)
	}

	// If no args or just the binary name, run TUI
	if len(os.Args) < 2 {
		runTUI(s)
		return
	}

	// Handle subcommands
	switch os.Args[1] {
	case "add":
		cmdAdd(s, os.Args[2:])
	case "list":
		cmdList(s, os.Args[2:])
	case "get":
		cmdGet(s, os.Args[2:])
	case "update":
		cmdUpdate(s, os.Args[2:])
	case "delete":
		cmdDelete(s, os.Args[2:])
	case "fetch":
		cmdFetch(os.Args[2:])
	case "context":
		cmdContext(s)
	case "apply":
		cmdApply(s, os.Args[2:])
	case "compile":
		cmdCompile(s, os.Args[2:])
	case "upgrade":
		cmdUpgrade()
	case "help", "--help", "-h":
		printHelp()
	default:
		// Unknown command, run TUI
		runTUI(s)
	}
}

func runTUI(s *store.Store) {
	app := tui.New(s)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`
     .-.
    (o o)  GHOSTED
    | O |  job application tracker
    |   |  for the perpetually ghosted
    '~~~'

Usage:
  ghosted              Launch interactive TUI
  ghosted <command>    Run a command

Commands:
  add --json '<json>'   Add a new application from JSON
  list [--json]         List all applications (--json for JSON output)
  get <id> [--json]     Get application by ID
  update <id> --json '<json>'  Update application fields
  delete <id>           Delete an application
  fetch <url> [--output name]  Fetch job posting from URL
  apply <posting> [flags]      Run full pipeline on a job posting
  compile <id|dir>             Compile .typ to PDF and update tracker
  context               Show context for AI agents (postings, CV, applications)
  upgrade               Update ghosted to the latest version
  help                  Show this help

Environment:
  GHOSTED_DATA         Path to data file (default: ~/.local/share/ghosted/applications.json)

Examples:
  ghosted add --json '{"company":"Acme Corp","position":"Software Engineer"}'
  ghosted list --json
  ghosted update abc123 --json '{"status":"interview"}'
  ghosted delete abc123
  ghosted fetch https://jobs.lever.co/company/job-id
  ghosted fetch --output acme-swe.md https://example.com/job
  ghosted apply local/postings/acme-swe.md
  ghosted apply --dry-run local/postings/test.md
  ghosted apply --auto-approve local/postings/acme-swe.md
  ghosted compile abc123
  ghosted compile local/applications/swe/acme/

Apply Command Flags:
  --dry-run       Generate documents without adding to tracker
  --auto-approve  Skip review confirmation step

─────────────────────────────────────────────────────────────────────────────────
AI AGENT WORKFLOW
─────────────────────────────────────────────────────────────────────────────────

Ghosted is designed to work with AI agents (Claude, GPT, etc.) for automated
job application workflows. Run 'ghosted context' to get all necessary context.

Folder Structure:
  local/
  ├── cv.json                          # Your master CV (source of truth)
  ├── postings/                        # Job postings to process
  │   └── *.md                         # Fetched or manually added postings
  └── applications/{job-type}/{company}/
      ├── posting.md                   # Copy of the job posting
      ├── resume.typ / resume.pdf      # Generated resume
      └── cover-letter.typ / .pdf      # Generated cover letter

Job Types: fe-dev, swe, ux-design, product-design

Agent Pipeline:
  1. Parser      - Extract structured data from job posting
  2. Resume      - Generate tailored resume (Typst → PDF)
  3. Cover       - Generate personalized cover letter
  4. Reviewer    - Score documents (70+ to approve)
  5. Tracker     - Add application to ghosted

Quick Start for Agents:
  1. ghosted context                   # Get full context (CV, postings, apps)
  2. ghosted fetch <url>               # Fetch a job posting
  3. Read internal/agent/prompts/*.md  # Agent prompt templates
  4. Generate documents to local/applications/{type}/{company}/
  5. ghosted add --json '{...}'        # Add to tracker

Prompt Templates: internal/agent/prompts/
  parser.md   - Job posting parser
  resume.md   - Resume generator
  cover.md    - Cover letter generator
  reviewer.md - Hiring manager reviewer
  tracker.md  - Tracker integration`)
}

// cmdAdd adds a new application from JSON input
func cmdAdd(s *store.Store, args []string) {
	if len(args) < 2 || args[0] != "--json" {
		fmt.Fprintln(os.Stderr, "Usage: ghosted add --json '<json>'")
		os.Exit(1)
	}

	jsonData := args[1]
	var app model.Application
	if err := json.Unmarshal([]byte(jsonData), &app); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Set defaults if not provided
	if app.Status == "" {
		app.Status = model.StatusApplied
	}
	// DateApplied defaults are handled by the store based on status

	// Validate required fields
	if app.Company == "" {
		fmt.Fprintln(os.Stderr, "Error: company is required")
		os.Exit(1)
	}
	if app.Position == "" {
		fmt.Fprintln(os.Stderr, "Error: position is required")
		os.Exit(1)
	}

	created, err := s.Add(app)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error adding application: %v\n", err)
		os.Exit(1)
	}

	// Output the created application as JSON
	output, _ := json.MarshalIndent(created, "", "  ")
	fmt.Println(string(output))
}

// cmdList lists all applications
func cmdList(s *store.Store, args []string) {
	apps := s.List()

	// Check for --json flag
	jsonOutput := false
	for _, arg := range args {
		if arg == "--json" {
			jsonOutput = true
			break
		}
	}

	if jsonOutput {
		output, _ := json.MarshalIndent(apps, "", "  ")
		fmt.Println(string(output))
	} else {
		// Simple text output
		if len(apps) == 0 {
			fmt.Println("No applications found.")
			return
		}
		for _, app := range apps {
			date := "—"
			if app.DateApplied != nil {
				date = app.DateApplied.Format("2006-01-02")
			}
			fmt.Printf("[%s] %s @ %s - %s (%s)\n",
				app.ID[:8],
				app.Position,
				app.Company,
				model.StatusLabel(app.Status),
				date,
			)
		}
	}
}

// cmdGet gets a single application by ID
func cmdGet(s *store.Store, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: ghosted get <id> [--json]")
		os.Exit(1)
	}

	id := args[0]
	app, err := s.GetByID(id)
	if err != nil {
		// Try partial ID match
		apps := s.List()
		for _, a := range apps {
			if len(a.ID) >= len(id) && a.ID[:len(id)] == id {
				app = a
				err = nil
				break
			}
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Application not found: %s\n", id)
		os.Exit(1)
	}

	// Check for --json flag
	jsonOutput := false
	for _, arg := range args[1:] {
		if arg == "--json" {
			jsonOutput = true
			break
		}
	}

	if jsonOutput {
		output, _ := json.MarshalIndent(app, "", "  ")
		fmt.Println(string(output))
	} else {
		fmt.Printf("ID:       %s\n", app.ID)
		fmt.Printf("Company:  %s\n", app.Company)
		fmt.Printf("Position: %s\n", app.Position)
		fmt.Printf("Status:   %s\n", model.StatusLabel(app.Status))
		if app.DateApplied != nil {
			fmt.Printf("Applied:  %s\n", app.DateApplied.Format("2006-01-02"))
		} else {
			fmt.Printf("Applied:  Not sent\n")
		}
		if app.Location != "" {
			fmt.Printf("Location: %s\n", app.Location)
		}
		if app.Remote {
			fmt.Println("Remote:   Yes")
		}
		if app.SalaryRange() != "" {
			fmt.Printf("Salary:   %s\n", app.SalaryRange())
		}
		if app.JobURL != "" {
			fmt.Printf("URL:      %s\n", app.JobURL)
		}
		if app.Notes != "" {
			fmt.Printf("Notes:    %s\n", app.Notes)
		}
	}
}

// cmdUpdate updates an existing application
func cmdUpdate(s *store.Store, args []string) {
	if len(args) < 3 || args[1] != "--json" {
		fmt.Fprintln(os.Stderr, "Usage: ghosted update <id> --json '<json>'")
		os.Exit(1)
	}

	id := args[0]
	jsonData := args[2]

	// Get existing application
	app, err := s.GetByID(id)
	if err != nil {
		// Try partial ID match
		apps := s.List()
		for _, a := range apps {
			if len(a.ID) >= len(id) && a.ID[:len(id)] == id {
				app = a
				err = nil
				break
			}
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Application not found: %s\n", id)
		os.Exit(1)
	}

	// Parse update JSON into a map to handle partial updates
	var updates map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &updates); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Apply updates
	if v, ok := updates["company"].(string); ok {
		app.Company = v
	}
	if v, ok := updates["position"].(string); ok {
		app.Position = v
	}
	if v, ok := updates["status"].(string); ok {
		app.Status = v
		// Auto-set date when transitioning to non-saved status
		if app.DateApplied == nil && v != model.StatusSaved {
			now := time.Now()
			app.DateApplied = &now
		}
	}
	if v, ok := updates["notes"].(string); ok {
		app.Notes = v
	}
	if v, ok := updates["location"].(string); ok {
		app.Location = v
	}
	if v, ok := updates["remote"].(bool); ok {
		app.Remote = v
	}
	if v, ok := updates["job_url"].(string); ok {
		app.JobURL = v
	}
	if v, ok := updates["salary_min"].(float64); ok {
		app.SalaryMin = int(v)
	}
	if v, ok := updates["salary_max"].(float64); ok {
		app.SalaryMax = int(v)
	}
	if v, ok := updates["contact_name"].(string); ok {
		app.ContactName = v
	}
	if v, ok := updates["contact_email"].(string); ok {
		app.ContactEmail = v
	}
	if v, ok := updates["resume_version"].(string); ok {
		app.ResumeVersion = v
	}
	if v, ok := updates["cover_letter"].(string); ok {
		app.CoverLetter = v
	}
	if v, ok := updates["documents_dir"].(string); ok {
		app.DocumentsDir = v
	}
	if v, ok := updates["date_applied"].(string); ok {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			parsedDate := t
			app.DateApplied = &parsedDate
		}
	}

	if err := s.Update(app); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating application: %v\n", err)
		os.Exit(1)
	}

	// Output the updated application
	output, _ := json.MarshalIndent(app, "", "  ")
	fmt.Println(string(output))
}

// cmdDelete deletes an application
func cmdDelete(s *store.Store, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: ghosted delete <id>")
		os.Exit(1)
	}

	id := args[0]

	// Try to get the app first (supports partial ID)
	app, err := s.GetByID(id)
	if err != nil {
		// Try partial ID match
		apps := s.List()
		for _, a := range apps {
			if len(a.ID) >= len(id) && a.ID[:len(id)] == id {
				app = a
				id = a.ID
				err = nil
				break
			}
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Application not found: %s\n", id)
		os.Exit(1)
	}

	if err := s.Delete(id); err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting application: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Deleted: %s @ %s\n", app.Position, app.Company)
}

// cmdFetch fetches a job posting from a URL and saves it locally
func cmdFetch(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: ghosted fetch <url> [--output name.md]")
		os.Exit(1)
	}

	var urlArg string
	var outputName string

	// Parse arguments
	for i := 0; i < len(args); i++ {
		if args[i] == "--output" || args[i] == "-o" {
			if i+1 < len(args) {
				outputName = args[i+1]
				i++
			}
		} else if fetch.IsURL(args[i]) {
			urlArg = args[i]
		}
	}

	if urlArg == "" {
		fmt.Fprintln(os.Stderr, "Error: URL is required")
		fmt.Fprintln(os.Stderr, "Usage: ghosted fetch <url> [--output name.md]")
		os.Exit(1)
	}

	// Default output directory
	outputDir := "local/postings"

	// Create fetcher and fetch
	f := fetch.NewFetcher(outputDir)

	fmt.Printf("Fetching: %s\n", urlArg)

	result, err := f.Fetch(urlArg, outputName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching URL: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Saved to: %s\n", result.OutputPath)
	if result.Company != "" {
		fmt.Printf("Company:  %s\n", result.Company)
	}
	if result.Position != "" {
		fmt.Printf("Position: %s\n", result.Position)
	}
	fmt.Printf("Size:     %d bytes\n", result.ContentSize)
	fmt.Println("\nNext step: ghosted apply", result.OutputPath)
}

func getDataPath() string {
	// Check for environment variable override
	if path := os.Getenv("GHOSTED_DATA"); path != "" {
		return path
	}

	// Default to ~/.local/share/ghosted/applications.json
	// or current directory if home not available
	home, err := os.UserHomeDir()
	if err != nil {
		return "data/applications.json"
	}

	return filepath.Join(home, ".local", "share", "ghosted", "applications.json")
}

// cmdContext outputs context information for AI agents
func cmdContext(s *store.Store) {
	fmt.Println(`
═══════════════════════════════════════════════════════════════════════════════
                           GHOSTED AGENT CONTEXT
═══════════════════════════════════════════════════════════════════════════════`)

	// Check for CV
	fmt.Println("📄 CANDIDATE CV")
	fmt.Println("───────────────────────────────────────────────────────────────────────────────")
	cvPath := "local/cv.json"
	if cvData, err := os.ReadFile(cvPath); err == nil {
		fmt.Println("Path: local/cv.json")
		fmt.Println("Content:")
		fmt.Println(string(cvData))
	} else {
		fmt.Println("⚠️  No CV found at local/cv.json")
		fmt.Println("   Create this file with your master resume data in JSON format.")
		fmt.Println(`   Example: {"name": "Your Name", "email": "...", "experience": [...]}`)
	}
	fmt.Println()

	// List postings
	fmt.Println("📬 PENDING POSTINGS")
	fmt.Println("───────────────────────────────────────────────────────────────────────────────")
	postingsDir := "local/postings"
	if entries, err := os.ReadDir(postingsDir); err == nil {
		count := 0
		for _, entry := range entries {
			if !entry.IsDir() && (filepath.Ext(entry.Name()) == ".md" || filepath.Ext(entry.Name()) == ".txt") {
				count++
				fmt.Printf("  • %s/%s\n", postingsDir, entry.Name())
			}
		}
		if count == 0 {
			fmt.Println("  (no postings found)")
		}
		fmt.Printf("\nTo add a posting: ghosted fetch <url>\n")
	} else {
		fmt.Println("  (local/postings directory not found)")
		fmt.Println("  Create it with: mkdir -p local/postings")
	}
	fmt.Println()

	// List applications by job type
	fmt.Println("📁 APPLICATIONS BY JOB TYPE")
	fmt.Println("───────────────────────────────────────────────────────────────────────────────")
	appsDir := "local/applications"
	jobTypes := []string{"fe-dev", "swe", "ux-design", "product-design"}
	foundAny := false
	for _, jobType := range jobTypes {
		typeDir := filepath.Join(appsDir, jobType)
		if entries, err := os.ReadDir(typeDir); err == nil {
			companies := []string{}
			for _, entry := range entries {
				if entry.IsDir() {
					companies = append(companies, entry.Name())
				}
			}
			if len(companies) > 0 {
				foundAny = true
				fmt.Printf("\n  %s/\n", jobType)
				for _, company := range companies {
					fmt.Printf("    └── %s/\n", company)
					// List files in company folder
					companyDir := filepath.Join(typeDir, company)
					if files, err := os.ReadDir(companyDir); err == nil {
						for _, f := range files {
							if !f.IsDir() {
								fmt.Printf("        • %s\n", f.Name())
							}
						}
					}
				}
			}
		}
	}
	if !foundAny {
		fmt.Println("  (no applications yet)")
		fmt.Println("  Structure: local/applications/{job-type}/{company}/")
	}
	fmt.Println()

	// Show tracker entries
	fmt.Println("📊 TRACKER ENTRIES")
	fmt.Println("───────────────────────────────────────────────────────────────────────────────")
	apps := s.List()
	if len(apps) == 0 {
		fmt.Println("  (no applications in tracker)")
	} else {
		// Group by status
		byStatus := make(map[string][]model.Application)
		for _, app := range apps {
			byStatus[app.Status] = append(byStatus[app.Status], app)
		}
		for _, status := range model.AllStatuses() {
			if statusApps, ok := byStatus[status]; ok && len(statusApps) > 0 {
				fmt.Printf("\n  %s (%d)\n", model.StatusLabel(status), len(statusApps))
				for _, app := range statusApps {
					fmt.Printf("    [%s] %s @ %s\n", app.ID[:8], app.Position, app.Company)
				}
			}
		}
	}
	fmt.Println()

	// Show prompt templates
	fmt.Println("📝 AGENT PROMPTS")
	fmt.Println("───────────────────────────────────────────────────────────────────────────────")
	promptsDir := "internal/agent/prompts"
	prompts := []struct {
		file string
		desc string
	}{
		{"parser.md", "Extract structured JSON from job postings"},
		{"resume.md", "Generate tailored Typst resumes"},
		{"cover.md", "Generate personalized cover letters"},
		{"reviewer.md", "Score documents (70+ to approve)"},
		{"tracker.md", "Generate ghosted CLI commands"},
	}
	for _, p := range prompts {
		path := filepath.Join(promptsDir, p.file)
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("  ✓ %s - %s\n", path, p.desc)
		} else {
			fmt.Printf("  ✗ %s (not found)\n", path)
		}
	}
	fmt.Println()

	fmt.Println(`═══════════════════════════════════════════════════════════════════════════════
WORKFLOW: fetch posting → parse → generate resume/cover → review → add to tracker
═══════════════════════════════════════════════════════════════════════════════`)
}

// cmdApply runs the full pipeline on a job posting
func cmdApply(s *store.Store, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: ghosted apply <posting-file> [--dry-run] [--auto-approve]")
		os.Exit(1)
	}

	// Parse arguments
	var postingPath string
	dryRun := false
	autoApprove := false

	for _, arg := range args {
		switch arg {
		case "--dry-run":
			dryRun = true
		case "--auto-approve":
			autoApprove = true
		default:
			if postingPath == "" && !isFlag(arg) {
				postingPath = arg
			}
		}
	}

	if postingPath == "" {
		fmt.Fprintln(os.Stderr, "Error: posting file is required")
		fmt.Fprintln(os.Stderr, "Usage: ghosted apply <posting-file> [--dry-run] [--auto-approve]")
		os.Exit(1)
	}

	// Check if file exists
	if _, err := os.Stat(postingPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", postingPath)
		os.Exit(1)
	}

	// Create pipeline config path
	configPath := filepath.Join("local", "document-generation", ".agent", "config.json")

	// For dry run, don't pass the store (prevents tracker entry)
	var pipelineStore *store.Store
	if !dryRun {
		pipelineStore = s
	}

	// Create pipeline
	pipeline, err := agent.NewPipeline(configPath, pipelineStore)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating pipeline: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Running pipeline on: %s\n", postingPath)
	if dryRun {
		fmt.Println("Mode: dry-run (no tracker entry will be created)")
	}
	if autoApprove {
		fmt.Println("Mode: auto-approve (skipping review confirmation)")
	}
	fmt.Println()

	// Run pipeline
	if err := pipeline.Run(postingPath); err != nil {
		fmt.Fprintf(os.Stderr, "\nPipeline failed: %v\n", err)
		fmt.Println("\n" + pipeline.GetStatus())
		os.Exit(1)
	}

	// Output status
	fmt.Println("\n" + pipeline.GetStatus())

	if dryRun {
		fmt.Println("\nDry run complete. No application was added to tracker.")
	} else {
		fmt.Println("\nApplication added to tracker. Run 'ghosted list' to view.")
	}
}

// isFlag checks if an argument is a flag
func isFlag(arg string) bool {
	return len(arg) > 0 && arg[0] == '-'
}

// cmdUpgrade updates ghosted to the latest version
func cmdUpgrade() {
	fmt.Println("Upgrading ghosted to latest version...")

	// Check if go is available
	if _, err := exec.LookPath("go"); err != nil {
		fmt.Fprintln(os.Stderr, "Error: Go is not installed or not in PATH")
		fmt.Fprintln(os.Stderr, "Install Go from https://go.dev/dl/ or use your package manager")
		os.Exit(1)
	}

	// Run go install to get the latest version
	cmd := exec.Command("go", "install", "github.com/celloopa/ghosted@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error upgrading: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully upgraded ghosted!")

	// Show where it was installed
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		home, _ := os.UserHomeDir()
		gopath = filepath.Join(home, "go")
	}
	binPath := filepath.Join(gopath, "bin", "ghosted")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}
	fmt.Printf("Installed to: %s\n", binPath)
}

// cmdCompile compiles .typ files to PDFs and updates the tracker
// Usage: ghosted compile <app-id-or-dir> [--generate-missing]
func cmdCompile(s *store.Store, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: ghosted compile <app-id> [--generate-missing]")
		fmt.Fprintln(os.Stderr, "       ghosted compile <documents-dir> [--generate-missing]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Compiles .typ files to PDFs with descriptive names and updates tracker.")
		fmt.Fprintln(os.Stderr, "Use --generate-missing to create missing cover letters via AI.")
		os.Exit(1)
	}

	target := args[0]
	generateMissing := false
	for _, arg := range args[1:] {
		if arg == "--generate-missing" {
			generateMissing = true
		}
	}

	var app *model.Application
	var documentsDir string

	// Check if target is an app ID or a directory
	if info, err := os.Stat(target); err == nil && info.IsDir() {
		// It's a directory path
		documentsDir = target
		// Try to find matching application
		apps := s.List()
		for i := range apps {
			if apps[i].DocumentsDir == target || apps[i].DocumentsDir == target+"/" {
				app = &apps[i]
				break
			}
		}
	} else {
		// Try as app ID
		found, err := s.GetByID(target)
		if err != nil {
			// Try partial ID match
			apps := s.List()
			for i := range apps {
				if len(apps[i].ID) >= len(target) && apps[i].ID[:len(target)] == target {
					found = apps[i]
					err = nil
					break
				}
			}
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Application not found: %s\n", target)
			os.Exit(1)
		}
		app = &found
		documentsDir = app.DocumentsDir
	}

	if documentsDir == "" {
		fmt.Fprintln(os.Stderr, "Error: No documents_dir set for this application")
		fmt.Fprintln(os.Stderr, "Set it with: ghosted update <id> --json '{\"documents_dir\":\"local/applications/...\"}'")
		os.Exit(1)
	}

	// Check if typst is available
	typstPath, err := exec.LookPath("typst")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: typst not found in PATH")
		fmt.Fprintln(os.Stderr, "Install from: https://typst.app or 'brew install typst'")
		os.Exit(1)
	}

	// Generate base name from company/position or directory name
	baseName := generateBaseName(app, documentsDir)

	fmt.Printf("Compiling documents in: %s\n", documentsDir)
	fmt.Printf("Base name: %s\n", baseName)
	fmt.Println()

	var resumePDF, coverPDF string

	// Compile resume.typ if exists
	resumeTyp := filepath.Join(documentsDir, "resume.typ")
	if _, err := os.Stat(resumeTyp); err == nil {
		resumePDF = baseName + "-resume.pdf"
		outPath := filepath.Join(documentsDir, resumePDF)
		fmt.Printf("Compiling resume: %s\n", resumePDF)
		cmd := exec.Command(typstPath, "compile", resumeTyp, outPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error compiling resume: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("  ✓ Resume compiled")
	} else {
		fmt.Println("No resume.typ found")
	}

	// Compile cover-letter.typ if exists
	coverTyp := filepath.Join(documentsDir, "cover-letter.typ")
	if _, err := os.Stat(coverTyp); err == nil {
		coverPDF = baseName + "-cover.pdf"
		outPath := filepath.Join(documentsDir, coverPDF)
		fmt.Printf("Compiling cover letter: %s\n", coverPDF)
		cmd := exec.Command(typstPath, "compile", coverTyp, outPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error compiling cover letter: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("  ✓ Cover letter compiled")
	} else {
		fmt.Println("No cover-letter.typ found")
		if generateMissing {
			fmt.Println("\n--generate-missing: Cover letter generation requires running the agent.")
			fmt.Println("Run: ghosted apply " + filepath.Join(documentsDir, "posting.md"))
		}
	}

	// Update tracker if we have an app
	if app != nil && (resumePDF != "" || coverPDF != "") {
		updates := make(map[string]interface{})
		if resumePDF != "" {
			updates["resume_version"] = resumePDF
		}
		if coverPDF != "" {
			updates["cover_letter"] = coverPDF
		}

		// Apply updates to app
		if resumePDF != "" {
			app.ResumeVersion = resumePDF
		}
		if coverPDF != "" {
			app.CoverLetter = coverPDF
		}

		if err := s.Update(*app); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating tracker: %v\n", err)
			os.Exit(1)
		}

		fmt.Println()
		fmt.Printf("Updated tracker entry: %s @ %s\n", app.Position, app.Company)
		if resumePDF != "" {
			fmt.Printf("  resume_version: %s\n", resumePDF)
		}
		if coverPDF != "" {
			fmt.Printf("  cover_letter: %s\n", coverPDF)
		}
	} else if app == nil {
		fmt.Println()
		fmt.Println("No matching tracker entry found. Add one with:")
		fmt.Printf("  ghosted add --json '{\"company\":\"...\",\"position\":\"...\",\"documents_dir\":\"%s\"}'\n", documentsDir)
	}

	fmt.Println()
	fmt.Println("Done!")
}

// generateBaseName creates a filename base from app or directory
func generateBaseName(app *model.Application, dir string) string {
	if app != nil && app.Company != "" && app.Position != "" {
		// Sanitize company and position for filename
		company := sanitizeForFilename(app.Company)
		position := sanitizeForFilename(app.Position)
		// Truncate if too long
		if len(position) > 20 {
			position = position[:20]
		}
		return company + "-" + position
	}

	// Fall back to directory name
	dir = filepath.Clean(dir)
	return filepath.Base(dir)
}

// sanitizeForFilename makes a string safe for use in filenames
func sanitizeForFilename(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	// Remove characters that are problematic in filenames
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result.WriteRune(r)
		}
	}
	// Remove multiple dashes
	out := result.String()
	for strings.Contains(out, "--") {
		out = strings.ReplaceAll(out, "--", "-")
	}
	return strings.Trim(out, "-")
}
