package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/celloopa/ghosted/internal/model"
	"github.com/celloopa/ghosted/internal/store"
)

// Pipeline orchestrates the multi-agent document generation workflow
type Pipeline struct {
	Config    *PipelineConfig
	Store     *store.Store
	State     *PipelineState
	BaseDir   string
	StateFile string
}

// NewPipeline creates a new pipeline instance
func NewPipeline(configPath string, store *store.Store) (*Pipeline, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		// Use default config if file doesn't exist
		if os.IsNotExist(err) {
			config = DefaultConfig()
		} else {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	baseDir := filepath.Dir(configPath)

	return &Pipeline{
		Config:    config,
		Store:     store,
		BaseDir:   baseDir,
		StateFile: filepath.Join(baseDir, "state.json"),
	}, nil
}

// Run executes the full pipeline for a job posting
func (p *Pipeline) Run(postingPath string) error {
	// Initialize state
	p.State = &PipelineState{
		PostingPath: postingPath,
		StartedAt:   time.Now().Format(time.RFC3339),
		Status:      "running",
		Results:     make(map[AgentType]StepResult),
	}

	// Initialize all steps as pending
	for _, agent := range p.Config.Agents {
		p.State.Results[agent.Type] = StepResult{Status: "pending"}
	}

	// Save initial state
	if err := p.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	// Run each enabled agent in sequence
	var lastOutput json.RawMessage
	for _, agent := range p.Config.EnabledAgents() {
		p.State.CurrentStep = agent.Type

		result, err := p.runStep(agent, lastOutput, postingPath)
		if err != nil {
			p.State.Results[agent.Type] = StepResult{
				Status: "failed",
				Error:  err.Error(),
			}
			p.State.Status = "failed"
			p.saveState()
			return fmt.Errorf("step %s failed: %w", agent.Type, err)
		}

		p.State.Results[agent.Type] = result
		lastOutput = result.Output

		if err := p.saveState(); err != nil {
			return fmt.Errorf("failed to save state: %w", err)
		}
	}

	p.State.Status = "completed"
	return p.saveState()
}

// runStep executes a single agent step
func (p *Pipeline) runStep(agent AgentConfig, input json.RawMessage, postingPath string) (StepResult, error) {
	start := time.Now()

	result := StepResult{
		Status: "running",
		Input:  input,
	}

	var output json.RawMessage
	var err error

	switch agent.Type {
	case AgentParser:
		output, err = p.runParserStep(postingPath)
	case AgentResume:
		output, err = p.runResumeStep(input)
	case AgentCover:
		output, err = p.runCoverStep(input)
	case AgentReviewer:
		output, err = p.runReviewerStep(input)
	case AgentTracker:
		output, err = p.runTrackerStep(input)
	default:
		err = fmt.Errorf("unknown agent type: %s", agent.Type)
	}

	result.Duration = time.Since(start).String()

	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		return result, err
	}

	result.Status = "completed"
	result.Output = output
	return result, nil
}

// runParserStep extracts structured data from a job posting
func (p *Pipeline) runParserStep(postingPath string) (json.RawMessage, error) {
	// Get the parser agent config
	agentConfig := p.Config.GetAgentConfig(AgentParser)
	if agentConfig == nil {
		return nil, fmt.Errorf("parser agent not configured")
	}

	// Create parser agent
	parser := NewParserAgent(agentConfig)

	// Check if file is supported
	if !parser.IsSupported(postingPath) {
		return nil, fmt.Errorf("unsupported file type: %s", filepath.Ext(postingPath))
	}

	// Read the posting content
	content, err := parser.ReadPosting(postingPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read posting: %w", err)
	}

	// For image files or when running without an AI backend,
	// store the content for later processing
	// The actual AI parsing happens when invoked via Claude Code
	parsed := ParsedPosting{
		Description: content,
		Notes:       fmt.Sprintf("Parsed from: %s", filepath.Base(postingPath)),
	}

	// If this is a text file, try to extract basic info
	if !parser.IsImageFile(postingPath) {
		parsed = extractBasicInfo(content, postingPath)
	}

	return json.Marshal(parsed)
}

// extractBasicInfo attempts to extract basic information from posting text
// This is a fallback when AI parsing is not available
func extractBasicInfo(content string, postingPath string) ParsedPosting {
	parsed := ParsedPosting{
		Description: content,
		Notes:       fmt.Sprintf("Parsed from: %s", filepath.Base(postingPath)),
	}

	lines := strings.Split(content, "\n")

	// Try to extract company and position from filename
	// Format: company-position-posting.md
	base := filepath.Base(postingPath)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	base = strings.TrimSuffix(base, "-posting")
	parts := strings.SplitN(base, "-", 2)
	if len(parts) >= 1 {
		parsed.Company = strings.ReplaceAll(parts[0], "_", " ")
		parsed.Company = strings.Title(parsed.Company)
	}
	if len(parts) >= 2 {
		parsed.Position = strings.ReplaceAll(parts[1], "_", " ")
		parsed.Position = strings.Title(parsed.Position)
	}

	// Try to find location from content
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for common location patterns
		if strings.Contains(strings.ToLower(line), "location:") ||
			strings.Contains(strings.ToLower(line), "based in") {
			parsed.Location = strings.TrimSpace(strings.Split(line, ":")[len(strings.Split(line, ":"))-1])
			break
		}
		// Check if line looks like a location (City, State format)
		if len(line) < 50 && strings.Contains(line, ",") &&
			!strings.Contains(line, ".") && !strings.Contains(line, "and") {
			// Could be a standalone location line
			if parsed.Location == "" {
				parsed.Location = line
			}
		}
	}

	// Check for remote mentions
	contentLower := strings.ToLower(content)
	if strings.Contains(contentLower, "remote") ||
		strings.Contains(contentLower, "work from home") ||
		strings.Contains(contentLower, "hybrid") {
		parsed.Remote = true
	}

	return parsed
}

// runResumeStep generates a tailored resume
// In production, this would invoke Claude Code with resume templates
func (p *Pipeline) runResumeStep(input json.RawMessage) (json.RawMessage, error) {
	var parsed ParsedPosting
	if err := json.Unmarshal(input, &parsed); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// Placeholder: In production, generates .typ file and compiles to PDF
	docs := GeneratedDocuments{
		ResumePath: filepath.Join(p.Config.Paths.OutputDir, p.formatFilename(parsed, "resume.typ")),
	}

	return json.Marshal(docs)
}

// runCoverStep generates a cover letter
// In production, this would invoke Claude Code with cover letter templates
func (p *Pipeline) runCoverStep(input json.RawMessage) (json.RawMessage, error) {
	var docs GeneratedDocuments
	if err := json.Unmarshal(input, &docs); err != nil {
		// Try parsing as ParsedPosting for backward compatibility
		var parsed ParsedPosting
		if err := json.Unmarshal(input, &parsed); err != nil {
			return nil, fmt.Errorf("invalid input: %w", err)
		}
		docs = GeneratedDocuments{
			CoverLetterPath: filepath.Join(p.Config.Paths.OutputDir, p.formatFilename(parsed, "cover.typ")),
		}
	}

	// Placeholder: In production, generates .typ file and compiles to PDF
	return json.Marshal(docs)
}

// runReviewerStep reviews generated documents
// In production, this would invoke Claude Code to review from hiring manager perspective
func (p *Pipeline) runReviewerStep(_ json.RawMessage) (json.RawMessage, error) {
	// TODO: Use input (GeneratedDocuments) to read and review the documents
	// Placeholder review result
	review := ReviewResult{
		Approved: true,
		Score:    8,
		Feedback: []string{
			"Resume highlights relevant experience",
			"Cover letter is personalized to the role",
		},
	}

	return json.Marshal(review)
}

// runTrackerStep creates an application entry in the store
func (p *Pipeline) runTrackerStep(_ json.RawMessage) (json.RawMessage, error) {
	if p.Store == nil {
		// Skip tracker step gracefully (dry-run mode)
		return json.Marshal(map[string]string{"status": "skipped", "reason": "dry-run mode"})
	}

	// Get the parsed posting from the parser step
	parserResult, ok := p.State.Results[AgentParser]
	if !ok || parserResult.Status != "completed" {
		return nil, fmt.Errorf("parser step not completed")
	}

	var parsed ParsedPosting
	if err := json.Unmarshal(parserResult.Output, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse posting data: %w", err)
	}

	// Get generated documents from resume/cover steps
	var docs GeneratedDocuments
	if coverResult, ok := p.State.Results[AgentCover]; ok && coverResult.Status == "completed" {
		json.Unmarshal(coverResult.Output, &docs)
	}

	// Create the application entry
	app := model.Application{
		Company:       parsed.Company,
		Position:      parsed.Position,
		Status:        model.StatusSaved, // Start as saved, user will mark as applied
		Location:      parsed.Location,
		Remote:        parsed.Remote,
		SalaryMin:     parsed.SalaryMin,
		SalaryMax:     parsed.SalaryMax,
		JobURL:        parsed.JobURL,
		Notes:         parsed.Notes,
		ResumeVersion: filepath.Base(docs.ResumePath),
		CoverLetter:   filepath.Base(docs.CoverLetterPath),
	}

	created, err := p.Store.Add(app)
	if err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	return json.Marshal(created)
}

// formatFilename creates an output filename from posting data
func (p *Pipeline) formatFilename(parsed ParsedPosting, suffix string) string {
	pattern := p.Config.Output.Naming
	if pattern == "" {
		pattern = "{company}-{position}"
	}

	name := pattern
	name = strings.ReplaceAll(name, "{company}", sanitizeFilename(parsed.Company))
	name = strings.ReplaceAll(name, "{position}", sanitizeFilename(parsed.Position))

	return name + "-" + suffix
}

// sanitizeFilename removes invalid characters from a filename
func sanitizeFilename(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	// Remove characters that are invalid in filenames
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalid {
		s = strings.ReplaceAll(s, char, "")
	}
	return s
}

// saveState persists the current pipeline state to disk
func (p *Pipeline) saveState() error {
	data, err := json.MarshalIndent(p.State, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(p.StateFile), 0755); err != nil {
		return err
	}

	return os.WriteFile(p.StateFile, data, 0644)
}

// LoadState loads pipeline state from disk
func (p *Pipeline) LoadState() error {
	data, err := os.ReadFile(p.StateFile)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &p.State)
}

// Resume continues a paused or failed pipeline from the current step
func (p *Pipeline) Resume() error {
	if p.State == nil {
		if err := p.LoadState(); err != nil {
			return fmt.Errorf("no state to resume: %w", err)
		}
	}

	if p.State.Status == "completed" {
		return fmt.Errorf("pipeline already completed")
	}

	// Find where we left off and continue
	p.State.Status = "running"
	foundCurrent := false

	var lastOutput json.RawMessage
	for _, agent := range p.Config.EnabledAgents() {
		if !foundCurrent {
			if agent.Type == p.State.CurrentStep {
				foundCurrent = true
			} else if result, ok := p.State.Results[agent.Type]; ok && result.Status == "completed" {
				lastOutput = result.Output
				continue
			}
		}

		result, err := p.runStep(agent, lastOutput, p.State.PostingPath)
		if err != nil {
			p.State.Results[agent.Type] = StepResult{
				Status: "failed",
				Error:  err.Error(),
			}
			p.State.Status = "failed"
			p.saveState()
			return fmt.Errorf("step %s failed: %w", agent.Type, err)
		}

		p.State.Results[agent.Type] = result
		lastOutput = result.Output
		p.State.CurrentStep = agent.Type

		if err := p.saveState(); err != nil {
			return fmt.Errorf("failed to save state: %w", err)
		}
	}

	p.State.Status = "completed"
	return p.saveState()
}

// GetStatus returns a summary of the pipeline status
func (p *Pipeline) GetStatus() string {
	if p.State == nil {
		return "not started"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Status: %s\n", p.State.Status))
	sb.WriteString(fmt.Sprintf("Started: %s\n", p.State.StartedAt))
	sb.WriteString(fmt.Sprintf("Posting: %s\n\n", p.State.PostingPath))

	sb.WriteString("Steps:\n")
	for _, agent := range p.Config.Agents {
		result := p.State.Results[agent.Type]
		marker := " "
		switch result.Status {
		case "completed":
			marker = "✓"
		case "running":
			marker = "→"
		case "failed":
			marker = "✗"
		}
		sb.WriteString(fmt.Sprintf("  [%s] %s: %s\n", marker, agent.Name, result.Status))
		if result.Error != "" {
			sb.WriteString(fmt.Sprintf("      Error: %s\n", result.Error))
		}
	}

	return sb.String()
}
