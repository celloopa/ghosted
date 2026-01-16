package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AgentType identifies the type of agent in the pipeline
type AgentType string

const (
	AgentParser   AgentType = "parser"
	AgentResume   AgentType = "resume"
	AgentCover    AgentType = "cover"
	AgentReviewer AgentType = "reviewer"
	AgentTracker  AgentType = "tracker"
)

// AgentConfig holds configuration for a single agent
type AgentConfig struct {
	Type        AgentType `json:"type"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	PromptFile  string    `json:"prompt_file"`  // Path to prompt template file
	Enabled     bool      `json:"enabled"`
	Model       string    `json:"model,omitempty"` // Optional model override (e.g., "sonnet", "opus")
}

// PipelineConfig holds the full pipeline configuration
type PipelineConfig struct {
	Version string `json:"version"`

	// Paths configuration
	Paths PathsConfig `json:"paths"`

	// Agent configurations
	Agents []AgentConfig `json:"agents"`

	// Output settings
	Output OutputConfig `json:"output"`
}

// PathsConfig defines paths used by the pipeline
type PathsConfig struct {
	PostingsDir    string `json:"postings_dir"`
	ResumesDir     string `json:"resumes_dir"`
	CoverLettersDir string `json:"cover_letters_dir"`
	TemplatesDir   string `json:"templates_dir"`
	OutputDir      string `json:"output_dir"`
}

// OutputConfig defines output generation settings
type OutputConfig struct {
	GeneratePDF  bool   `json:"generate_pdf"`
	PDFEngine    string `json:"pdf_engine"`    // "typst" or "pandoc"
	KeepTypst    bool   `json:"keep_typst"`    // Keep .typ source files
	Naming       string `json:"naming"`        // Output file naming pattern
}

// PipelineState tracks the state of a pipeline run
type PipelineState struct {
	PostingPath string            `json:"posting_path"`
	StartedAt   string            `json:"started_at"`
	Status      string            `json:"status"` // "running", "completed", "failed", "paused"
	CurrentStep AgentType         `json:"current_step"`
	Results     map[AgentType]StepResult `json:"results"`
}

// StepResult holds the output of a single pipeline step
type StepResult struct {
	Status   string          `json:"status"` // "pending", "running", "completed", "failed", "skipped"
	Input    json.RawMessage `json:"input,omitempty"`
	Output   json.RawMessage `json:"output,omitempty"`
	Error    string          `json:"error,omitempty"`
	Duration string          `json:"duration,omitempty"`
}

// ParsedPosting represents structured data extracted from a job posting
type ParsedPosting struct {
	Company       string   `json:"company"`
	Position      string   `json:"position"`
	Team          string   `json:"team,omitempty"`
	Location      string   `json:"location,omitempty"`
	Remote        bool     `json:"remote,omitempty"`
	SalaryMin     int      `json:"salary_min,omitempty"`
	SalaryMax     int      `json:"salary_max,omitempty"`
	JobURL        string   `json:"job_url,omitempty"`
	Requirements  []string `json:"requirements,omitempty"`
	BonusSkills   []string `json:"bonus_skills,omitempty"`
	Keywords      []string `json:"keywords,omitempty"`
	TechStack     []string `json:"tech_stack,omitempty"`
	CompanyValues []string `json:"company_values,omitempty"`
	Description   string   `json:"description,omitempty"`
	Notes         string   `json:"notes,omitempty"`
}

// GeneratedDocuments holds paths to generated resume and cover letter
type GeneratedDocuments struct {
	ResumePath      string `json:"resume_path,omitempty"`
	CoverLetterPath string `json:"cover_letter_path,omitempty"`
	ResumePDF       string `json:"resume_pdf,omitempty"`
	CoverLetterPDF  string `json:"cover_letter_pdf,omitempty"`
}

// ReviewResult holds the reviewer agent's feedback
type ReviewResult struct {
	Approved bool     `json:"approved"`
	Score    int      `json:"score,omitempty"`    // 1-10 rating
	Feedback []string `json:"feedback,omitempty"` // List of suggestions
	Issues   []string `json:"issues,omitempty"`   // Critical issues to fix
}

// DefaultConfig returns a default pipeline configuration
func DefaultConfig() *PipelineConfig {
	return &PipelineConfig{
		Version: "1.0",
		Paths: PathsConfig{
			PostingsDir:     "local/postings",
			ResumesDir:      "local/resumes",
			CoverLettersDir: "local/cover-letters",
			TemplatesDir:    "local/document-generation/.agent/templates",
			OutputDir:       "local/document-generation/output",
		},
		Agents: []AgentConfig{
			{
				Type:        AgentParser,
				Name:        "Posting Parser",
				Description: "Extracts structured data from job postings",
				PromptFile:  "prompts/parser.md",
				Enabled:     true,
			},
			{
				Type:        AgentResume,
				Name:        "Resume Generator",
				Description: "Generates tailored resume for the position",
				PromptFile:  "prompts/resume.md",
				Enabled:     true,
			},
			{
				Type:        AgentCover,
				Name:        "Cover Letter Generator",
				Description: "Generates personalized cover letter",
				PromptFile:  "prompts/cover.md",
				Enabled:     true,
			},
			{
				Type:        AgentReviewer,
				Name:        "Hiring Manager Reviewer",
				Description: "Reviews documents from hiring manager perspective",
				PromptFile:  "prompts/reviewer.md",
				Enabled:     true,
			},
			{
				Type:        AgentTracker,
				Name:        "Tracker Integration",
				Description: "Creates application entry in ghosted tracker",
				PromptFile:  "prompts/tracker.md",
				Enabled:     true,
			},
		},
		Output: OutputConfig{
			GeneratePDF: true,
			PDFEngine:   "typst",
			KeepTypst:   true,
			Naming:      "{company}-{position}",
		},
	}
}

// LoadConfig loads pipeline configuration from a JSON file
func LoadConfig(path string) (*PipelineConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config PipelineConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig saves pipeline configuration to a JSON file
func SaveConfig(config *PipelineConfig, path string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetAgentConfig returns the configuration for a specific agent type
func (c *PipelineConfig) GetAgentConfig(agentType AgentType) *AgentConfig {
	for i := range c.Agents {
		if c.Agents[i].Type == agentType {
			return &c.Agents[i]
		}
	}
	return nil
}

// EnabledAgents returns a list of enabled agents in pipeline order
func (c *PipelineConfig) EnabledAgents() []AgentConfig {
	var enabled []AgentConfig
	for _, agent := range c.Agents {
		if agent.Enabled {
			enabled = append(enabled, agent)
		}
	}
	return enabled
}
