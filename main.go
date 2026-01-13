package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"jobtrack/internal/model"
	"jobtrack/internal/store"
	"jobtrack/internal/tui"

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
	fmt.Println(`jobtrack - Job Application Tracker

Usage:
  jobtrack              Launch interactive TUI
  jobtrack <command>    Run a command

Commands:
  add --json '<json>'   Add a new application from JSON
  list [--json]         List all applications (--json for JSON output)
  get <id> [--json]     Get application by ID
  update <id> --json '<json>'  Update application fields
  delete <id>           Delete an application
  help                  Show this help

Environment:
  JOBTRACK_DATA         Path to data file (default: ~/.local/share/jobtrack/applications.json)

Examples:
  jobtrack add --json '{"company":"Acme Corp","position":"Software Engineer"}'
  jobtrack list --json
  jobtrack update abc123 --json '{"status":"interview"}'
  jobtrack delete abc123`)
}

// cmdAdd adds a new application from JSON input
func cmdAdd(s *store.Store, args []string) {
	if len(args) < 2 || args[0] != "--json" {
		fmt.Fprintln(os.Stderr, "Usage: jobtrack add --json '<json>'")
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
	if app.DateApplied.IsZero() {
		app.DateApplied = time.Now()
	}

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
			fmt.Printf("[%s] %s @ %s - %s (%s)\n",
				app.ID[:8],
				app.Position,
				app.Company,
				model.StatusLabel(app.Status),
				app.DateApplied.Format("2006-01-02"),
			)
		}
	}
}

// cmdGet gets a single application by ID
func cmdGet(s *store.Store, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: jobtrack get <id> [--json]")
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
		fmt.Printf("Applied:  %s\n", app.DateApplied.Format("2006-01-02"))
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
		fmt.Fprintln(os.Stderr, "Usage: jobtrack update <id> --json '<json>'")
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
	if v, ok := updates["date_applied"].(string); ok {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			app.DateApplied = t
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
		fmt.Fprintln(os.Stderr, "Usage: jobtrack delete <id>")
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

func getDataPath() string {
	// Check for environment variable override
	if path := os.Getenv("JOBTRACK_DATA"); path != "" {
		return path
	}

	// Default to ~/.local/share/jobtrack/applications.json
	// or current directory if home not available
	home, err := os.UserHomeDir()
	if err != nil {
		return "data/applications.json"
	}

	return filepath.Join(home, ".local", "share", "jobtrack", "applications.json")
}
