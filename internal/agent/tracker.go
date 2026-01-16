package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/celloopa/ghosted/internal/model"
	"github.com/celloopa/ghosted/internal/store"
)

// TrackerAgent integrates pipeline results with the ghosted tracker
type TrackerAgent struct {
	Config  *AgentConfig
	Store   *store.Store
	BaseDir string
}

// TrackerInput holds the data needed to create a tracker entry
type TrackerInput struct {
	Posting          *ParsedPosting      `json:"posting"`
	Documents        *GeneratedDocuments `json:"documents"`
	ReviewResult     *ReviewResult       `json:"review_result,omitempty"`
	DetailedReview   *DetailedReviewResult `json:"detailed_review,omitempty"`
	ApplicationFolder string             `json:"application_folder"`
	JobType          string              `json:"job_type"`
}

// TrackerOutput holds the result of tracker integration
type TrackerOutput struct {
	ApplicationID   string `json:"application_id"`
	Status          string `json:"status"`
	ResumeVersion   string `json:"resume_version"`
	CoverLetter     string `json:"cover_letter"`
	PostingArchived bool   `json:"posting_archived"`
	Notes           string `json:"notes"`
}

// RejectionFeedback holds feedback when documents are rejected
type RejectionFeedback struct {
	Posting       *ParsedPosting `json:"posting"`
	OverallScore  int            `json:"overall_score"`
	Issues        []string       `json:"issues"`
	Suggestions   []string       `json:"suggestions"`
	Recommendation string        `json:"recommendation"`
}

// NewTrackerAgent creates a new tracker agent instance
func NewTrackerAgent(config *AgentConfig, appStore *store.Store, baseDir string) *TrackerAgent {
	return &TrackerAgent{
		Config:  config,
		Store:   appStore,
		BaseDir: baseDir,
	}
}

// CreateApplication creates a new application entry in the store
func (t *TrackerAgent) CreateApplication(input *TrackerInput) (*model.Application, error) {
	if t.Store == nil {
		return nil, fmt.Errorf("store not initialized")
	}

	if input.Posting == nil {
		return nil, fmt.Errorf("posting data is required")
	}

	if err := t.ValidateInput(input); err != nil {
		return nil, err
	}

	// Build the application model
	app := model.Application{
		Company:   input.Posting.Company,
		Position:  input.Posting.Position,
		Status:    model.StatusApplied,
		Location:  input.Posting.Location,
		Remote:    input.Posting.Remote,
		SalaryMin: input.Posting.SalaryMin,
		SalaryMax: input.Posting.SalaryMax,
		JobURL:    input.Posting.JobURL,
		Notes:     t.GenerateNotes(input),
	}

	// Set document paths if available
	if input.Documents != nil {
		if input.Documents.ResumePDF != "" {
			app.ResumeVersion = t.FormatDocumentPath(input.Documents.ResumePDF, input.ApplicationFolder)
		} else if input.Documents.ResumePath != "" {
			app.ResumeVersion = t.FormatDocumentPath(input.Documents.ResumePath, input.ApplicationFolder)
		}

		if input.Documents.CoverLetterPDF != "" {
			app.CoverLetter = t.FormatDocumentPath(input.Documents.CoverLetterPDF, input.ApplicationFolder)
		} else if input.Documents.CoverLetterPath != "" {
			app.CoverLetter = t.FormatDocumentPath(input.Documents.CoverLetterPath, input.ApplicationFolder)
		}
	}

	// Add to store
	created, err := t.Store.Add(app)
	if err != nil {
		return nil, fmt.Errorf("failed to add application to store: %w", err)
	}

	return &created, nil
}

// ValidateInput checks that required fields are present
func (t *TrackerAgent) ValidateInput(input *TrackerInput) error {
	if input.Posting.Company == "" {
		return fmt.Errorf("company name is required")
	}
	if input.Posting.Position == "" {
		return fmt.Errorf("position is required")
	}
	return nil
}

// GenerateNotes creates notes from the parsed posting and review
func (t *TrackerAgent) GenerateNotes(input *TrackerInput) string {
	var notes strings.Builder

	// Tech stack
	if len(input.Posting.TechStack) > 0 {
		notes.WriteString("Tech stack: ")
		notes.WriteString(strings.Join(input.Posting.TechStack, ", "))
		notes.WriteString("\n")
	}

	// Review score if available
	if input.DetailedReview != nil {
		notes.WriteString(fmt.Sprintf("Review score: %d/100\n", input.DetailedReview.OverallScore))
	} else if input.ReviewResult != nil && input.ReviewResult.Score > 0 {
		notes.WriteString(fmt.Sprintf("Review score: %d/10\n", input.ReviewResult.Score))
	}

	// Key requirements
	if len(input.Posting.Requirements) > 0 {
		notes.WriteString("Key requirements: ")
		// Show first 3 requirements
		count := 3
		if len(input.Posting.Requirements) < count {
			count = len(input.Posting.Requirements)
		}
		notes.WriteString(strings.Join(input.Posting.Requirements[:count], ", "))
		notes.WriteString("\n")
	}

	// Original notes from posting
	if input.Posting.Notes != "" {
		notes.WriteString("\n")
		notes.WriteString(input.Posting.Notes)
	}

	return strings.TrimSpace(notes.String())
}

// FormatDocumentPath creates a relative path for storage in the tracker
func (t *TrackerAgent) FormatDocumentPath(fullPath, applicationFolder string) string {
	// If it's already a relative path starting with applications/, return as-is
	if strings.HasPrefix(fullPath, "applications/") {
		return fullPath
	}

	// Extract just the filename
	filename := filepath.Base(fullPath)

	// If we have an application folder, use that structure
	if applicationFolder != "" {
		return filepath.Join("applications", applicationFolder, filename)
	}

	return filename
}

// ArchivePosting moves the original posting to the processed folder
func (t *TrackerAgent) ArchivePosting(postingPath, processedDir string) error {
	if postingPath == "" {
		return fmt.Errorf("posting path is required")
	}

	// Ensure processed directory exists
	if err := os.MkdirAll(processedDir, 0755); err != nil {
		return fmt.Errorf("failed to create processed directory: %w", err)
	}

	// Generate destination path
	filename := filepath.Base(postingPath)
	destPath := filepath.Join(processedDir, filename)

	// Check if source exists
	if _, err := os.Stat(postingPath); os.IsNotExist(err) {
		return fmt.Errorf("posting file not found: %s", postingPath)
	}

	// Move the file
	if err := os.Rename(postingPath, destPath); err != nil {
		// If rename fails (cross-device), try copy and delete
		data, err := os.ReadFile(postingPath)
		if err != nil {
			return fmt.Errorf("failed to read posting: %w", err)
		}
		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write archived posting: %w", err)
		}
		os.Remove(postingPath) // Best effort delete
	}

	return nil
}

// SaveRejectionFeedback saves feedback when documents are rejected
func (t *TrackerAgent) SaveRejectionFeedback(input *TrackerInput, outputDir string) (string, error) {
	if input.DetailedReview == nil && input.ReviewResult == nil {
		return "", fmt.Errorf("review result is required")
	}

	feedback := RejectionFeedback{
		Posting: input.Posting,
	}

	// Extract feedback from review
	if input.DetailedReview != nil {
		feedback.OverallScore = input.DetailedReview.OverallScore
		feedback.Recommendation = input.DetailedReview.Recommendation
		feedback.Issues = append(
			input.DetailedReview.ResumeReview.Weaknesses,
			input.DetailedReview.CoverReview.Weaknesses...,
		)
		feedback.Suggestions = append(
			input.DetailedReview.ResumeReview.Suggestions,
			input.DetailedReview.CoverReview.Suggestions...,
		)
	} else if input.ReviewResult != nil {
		feedback.OverallScore = input.ReviewResult.Score * 10
		feedback.Issues = input.ReviewResult.Issues
		feedback.Suggestions = input.ReviewResult.Feedback
	}

	// Ensure directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create feedback directory: %w", err)
	}

	// Generate feedback filename
	company := sanitizeFilename(input.Posting.Company)
	filename := fmt.Sprintf("%s-feedback.json", company)
	outputPath := filepath.Join(outputDir, filename)

	// Write feedback JSON
	data, err := json.MarshalIndent(feedback, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to serialize feedback: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write feedback file: %w", err)
	}

	return outputPath, nil
}

// DetermineJobType infers the job type from posting data
func (t *TrackerAgent) DetermineJobType(posting *ParsedPosting) string {
	positionLower := strings.ToLower(posting.Position)

	// Check for frontend keywords
	frontendKeywords := []string{"frontend", "front-end", "front end", "react", "vue", "angular", "ui"}
	for _, keyword := range frontendKeywords {
		if strings.Contains(positionLower, keyword) {
			return "fe-dev"
		}
	}

	// Check for UX/UI keywords
	uxKeywords := []string{"ux", "ui design", "user experience", "user interface"}
	for _, keyword := range uxKeywords {
		if strings.Contains(positionLower, keyword) {
			return "ux-design"
		}
	}

	// Check for product design keywords
	pdKeywords := []string{"product design", "product designer"}
	for _, keyword := range pdKeywords {
		if strings.Contains(positionLower, keyword) {
			return "product-design"
		}
	}

	// Default to general SWE
	return "swe"
}

// GenerateApplicationFolder creates the folder name for an application
func (t *TrackerAgent) GenerateApplicationFolder(posting *ParsedPosting, jobType string) string {
	company := sanitizeFilename(posting.Company)
	position := sanitizeFilename(posting.Position)
	position = strings.ReplaceAll(position, "-", "_")

	return filepath.Join(jobType, fmt.Sprintf("%s-%s", company, position))
}

// CopyPostingToApplicationFolder copies the original posting to the application folder
func (t *TrackerAgent) CopyPostingToApplicationFolder(postingPath, applicationDir string) error {
	if postingPath == "" {
		return nil // No posting to copy
	}

	// Ensure application directory exists
	if err := os.MkdirAll(applicationDir, 0755); err != nil {
		return fmt.Errorf("failed to create application directory: %w", err)
	}

	// Read source posting
	data, err := os.ReadFile(postingPath)
	if err != nil {
		return fmt.Errorf("failed to read posting: %w", err)
	}

	// Write to application folder as posting.md
	destPath := filepath.Join(applicationDir, "posting.md")
	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return fmt.Errorf("failed to copy posting: %w", err)
	}

	return nil
}

// GetSystemPrompt returns the system prompt for tracker integration
func (t *TrackerAgent) GetSystemPrompt() string {
	return `You are a job application tracker agent. Your task is to create entries in the ghosted CLI tracker using data from the pipeline.

## Field Mapping

| Pipeline Data | Tracker Field |
|---------------|---------------|
| parsed.company | company |
| parsed.position | position |
| "applied" | status (default) |
| parsed.location | location |
| parsed.remote | remote |
| parsed.salary_min | salary_min |
| parsed.salary_max | salary_max |
| parsed.job_url | job_url |
| documents.resume_pdf | resume_version |
| documents.cover_pdf | cover_letter |
| (generated) | notes |

## Notes Generation

Compile useful information into the notes field:
- Tech stack
- Review score
- Key requirements (first 3)

## Output

Return the bash command to add the application using ./ghosted add --json`
}

// GetUserPrompt creates the user prompt with pipeline data
func (t *TrackerAgent) GetUserPrompt(input *TrackerInput) (string, error) {
	inputJSON, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to serialize input: %w", err)
	}

	return fmt.Sprintf(`Create a tracker entry for this job application.

## Pipeline Data

%s

## Instructions

1. Verify company and position are present
2. Generate notes from tech stack, review score, and requirements
3. Format document paths correctly
4. Return the ./ghosted add --json command`, inputJSON), nil
}

// Integrate runs the full tracker integration flow
func (t *TrackerAgent) Integrate(input *TrackerInput) (*TrackerOutput, error) {
	// Validate input
	if err := t.ValidateInput(input); err != nil {
		return nil, err
	}

	// Determine job type if not provided
	if input.JobType == "" {
		input.JobType = t.DetermineJobType(input.Posting)
	}

	// Generate application folder path
	if input.ApplicationFolder == "" {
		input.ApplicationFolder = t.GenerateApplicationFolder(input.Posting, input.JobType)
	}

	// Check if approved
	approved := true
	if input.DetailedReview != nil {
		approved = input.DetailedReview.Approved
	} else if input.ReviewResult != nil {
		approved = input.ReviewResult.Approved
	}

	output := &TrackerOutput{
		Status: "applied",
	}

	if !approved {
		// Save rejection feedback
		feedbackDir := filepath.Join(t.BaseDir, "local", "postings")
		feedbackPath, err := t.SaveRejectionFeedback(input, feedbackDir)
		if err != nil {
			return nil, fmt.Errorf("failed to save rejection feedback: %w", err)
		}
		output.Status = "rejected"
		output.Notes = fmt.Sprintf("Feedback saved to: %s", feedbackPath)
		return output, nil
	}

	// Create application
	app, err := t.CreateApplication(input)
	if err != nil {
		return nil, err
	}

	output.ApplicationID = app.ID
	output.ResumeVersion = app.ResumeVersion
	output.CoverLetter = app.CoverLetter

	return output, nil
}
