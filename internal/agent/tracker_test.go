package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/celloopa/ghosted/internal/store"
)

func TestTrackerAgent_CreateApplication(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "applications.json")

	// Initialize empty store
	os.WriteFile(storePath, []byte("[]"), 0644)
	s, err := store.New(storePath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	agent := NewTrackerAgent(&AgentConfig{Type: AgentTracker}, s, tmpDir)

	input := &TrackerInput{
		Posting: &ParsedPosting{
			Company:   "TechCorp",
			Position:  "Software Engineer",
			Location:  "San Francisco, CA",
			Remote:    true,
			SalaryMin: 150000,
			SalaryMax: 200000,
			TechStack: []string{"Go", "React"},
		},
		Documents: &GeneratedDocuments{
			ResumePDF:      "resume.pdf",
			CoverLetterPDF: "cover-letter.pdf",
		},
		ApplicationFolder: "swe/techcorp-software_engineer",
	}

	app, err := agent.CreateApplication(input)
	if err != nil {
		t.Errorf("CreateApplication() error = %v", err)
		return
	}

	if app.Company != "TechCorp" {
		t.Errorf("Company = %q, want %q", app.Company, "TechCorp")
	}
	if app.Position != "Software Engineer" {
		t.Errorf("Position = %q, want %q", app.Position, "Software Engineer")
	}
	if app.ID == "" {
		t.Error("Application ID should not be empty")
	}
}

func TestTrackerAgent_CreateApplication_NilStore(t *testing.T) {
	agent := NewTrackerAgent(&AgentConfig{Type: AgentTracker}, nil, "")

	input := &TrackerInput{
		Posting: &ParsedPosting{Company: "Test", Position: "Dev"},
	}

	_, err := agent.CreateApplication(input)
	if err == nil {
		t.Error("CreateApplication() expected error when store is nil")
	}
}

func TestTrackerAgent_CreateApplication_NilPosting(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "apps.json")
	os.WriteFile(storePath, []byte("[]"), 0644)
	s, _ := store.New(storePath)

	agent := NewTrackerAgent(&AgentConfig{Type: AgentTracker}, s, tmpDir)

	_, err := agent.CreateApplication(&TrackerInput{})
	if err == nil {
		t.Error("CreateApplication() expected error when posting is nil")
	}
}

func TestTrackerAgent_ValidateInput(t *testing.T) {
	agent := NewTrackerAgent(&AgentConfig{Type: AgentTracker}, nil, "")

	tests := []struct {
		name    string
		input   *TrackerInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: &TrackerInput{
				Posting: &ParsedPosting{
					Company:  "Acme",
					Position: "Engineer",
				},
			},
			wantErr: false,
		},
		{
			name: "missing company",
			input: &TrackerInput{
				Posting: &ParsedPosting{
					Position: "Engineer",
				},
			},
			wantErr: true,
		},
		{
			name: "missing position",
			input: &TrackerInput{
				Posting: &ParsedPosting{
					Company: "Acme",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := agent.ValidateInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTrackerAgent_GenerateNotes(t *testing.T) {
	agent := NewTrackerAgent(&AgentConfig{Type: AgentTracker}, nil, "")

	input := &TrackerInput{
		Posting: &ParsedPosting{
			TechStack:    []string{"Go", "React", "PostgreSQL"},
			Requirements: []string{"5+ years experience", "Distributed systems", "AWS", "Docker"},
			Notes:        "Great company culture",
		},
		DetailedReview: &DetailedReviewResult{
			OverallScore: 85,
		},
	}

	notes := agent.GenerateNotes(input)

	if !strings.Contains(notes, "Tech stack: Go, React, PostgreSQL") {
		t.Error("Notes should contain tech stack")
	}
	if !strings.Contains(notes, "Review score: 85/100") {
		t.Error("Notes should contain review score")
	}
	if !strings.Contains(notes, "Key requirements:") {
		t.Error("Notes should contain key requirements")
	}
	if !strings.Contains(notes, "Great company culture") {
		t.Error("Notes should contain original posting notes")
	}
}

func TestTrackerAgent_GenerateNotes_SimpleReview(t *testing.T) {
	agent := NewTrackerAgent(&AgentConfig{Type: AgentTracker}, nil, "")

	input := &TrackerInput{
		Posting: &ParsedPosting{
			Company:  "Test",
			Position: "Dev",
		},
		ReviewResult: &ReviewResult{
			Score: 8,
		},
	}

	notes := agent.GenerateNotes(input)

	if !strings.Contains(notes, "Review score: 8/10") {
		t.Error("Notes should contain simple review score")
	}
}

func TestTrackerAgent_FormatDocumentPath(t *testing.T) {
	agent := NewTrackerAgent(&AgentConfig{Type: AgentTracker}, nil, "")

	tests := []struct {
		name             string
		fullPath         string
		applicationFolder string
		expected         string
	}{
		{
			name:             "with application folder",
			fullPath:         "/full/path/resume.pdf",
			applicationFolder: "swe/techcorp-engineer",
			expected:         "applications/swe/techcorp-engineer/resume.pdf",
		},
		{
			name:             "already relative path",
			fullPath:         "applications/swe/acme/resume.pdf",
			applicationFolder: "",
			expected:         "applications/swe/acme/resume.pdf",
		},
		{
			name:             "filename only",
			fullPath:         "resume.pdf",
			applicationFolder: "",
			expected:         "resume.pdf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := agent.FormatDocumentPath(tt.fullPath, tt.applicationFolder)
			if got != tt.expected {
				t.Errorf("FormatDocumentPath() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestTrackerAgent_DetermineJobType(t *testing.T) {
	agent := NewTrackerAgent(&AgentConfig{Type: AgentTracker}, nil, "")

	tests := []struct {
		position string
		expected string
	}{
		{"Frontend Developer", "fe-dev"},
		{"React Engineer", "fe-dev"},
		{"Senior Front-End Engineer", "fe-dev"},
		{"UX Designer", "ux-design"},
		{"UI/UX Lead", "fe-dev"}, // "ui" matches frontend before "ux"
		{"Product Designer", "product-design"},
		{"Senior Software Engineer", "swe"},
		{"Backend Developer", "swe"},
		{"Full Stack Engineer", "swe"},
	}

	for _, tt := range tests {
		t.Run(tt.position, func(t *testing.T) {
			posting := &ParsedPosting{Position: tt.position}
			got := agent.DetermineJobType(posting)
			if got != tt.expected {
				t.Errorf("DetermineJobType(%q) = %q, want %q", tt.position, got, tt.expected)
			}
		})
	}
}

func TestTrackerAgent_GenerateApplicationFolder(t *testing.T) {
	agent := NewTrackerAgent(&AgentConfig{Type: AgentTracker}, nil, "")

	posting := &ParsedPosting{
		Company:  "Tech Corp",
		Position: "Software Engineer",
	}

	folder := agent.GenerateApplicationFolder(posting, "swe")

	expected := "swe/tech-corp-software_engineer"
	if folder != expected {
		t.Errorf("GenerateApplicationFolder() = %q, want %q", folder, expected)
	}
}

func TestTrackerAgent_ArchivePosting(t *testing.T) {
	agent := NewTrackerAgent(&AgentConfig{Type: AgentTracker}, nil, "")

	tmpDir := t.TempDir()
	postingPath := filepath.Join(tmpDir, "test-posting.md")
	processedDir := filepath.Join(tmpDir, "processed")
	content := "# Job Posting Content"

	// Create test posting
	if err := os.WriteFile(postingPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test posting: %v", err)
	}

	err := agent.ArchivePosting(postingPath, processedDir)
	if err != nil {
		t.Errorf("ArchivePosting() error = %v", err)
	}

	// Verify archived file exists
	archivedPath := filepath.Join(processedDir, "test-posting.md")
	if _, err := os.Stat(archivedPath); os.IsNotExist(err) {
		t.Error("Archived posting file should exist")
	}

	// Verify original is moved
	if _, err := os.Stat(postingPath); !os.IsNotExist(err) {
		t.Error("Original posting should be removed")
	}
}

func TestTrackerAgent_ArchivePosting_MissingFile(t *testing.T) {
	agent := NewTrackerAgent(&AgentConfig{Type: AgentTracker}, nil, "")

	tmpDir := t.TempDir()
	err := agent.ArchivePosting("/nonexistent/posting.md", tmpDir)
	if err == nil {
		t.Error("ArchivePosting() expected error for missing file")
	}
}

func TestTrackerAgent_ArchivePosting_EmptyPath(t *testing.T) {
	agent := NewTrackerAgent(&AgentConfig{Type: AgentTracker}, nil, "")

	err := agent.ArchivePosting("", "/some/dir")
	if err == nil {
		t.Error("ArchivePosting() expected error for empty path")
	}
}

func TestTrackerAgent_SaveRejectionFeedback(t *testing.T) {
	agent := NewTrackerAgent(&AgentConfig{Type: AgentTracker}, nil, "")

	tmpDir := t.TempDir()

	input := &TrackerInput{
		Posting: &ParsedPosting{
			Company:  "TechCorp",
			Position: "Engineer",
		},
		DetailedReview: &DetailedReviewResult{
			OverallScore:   55,
			Recommendation: "Revise and resubmit",
			ResumeReview: DocumentReview{
				Weaknesses:  []string{"Missing required skills"},
				Suggestions: []string{"Add AWS experience"},
			},
			CoverReview: DocumentReview{
				Weaknesses:  []string{"Too generic"},
				Suggestions: []string{"Be more specific"},
			},
		},
	}

	feedbackPath, err := agent.SaveRejectionFeedback(input, tmpDir)
	if err != nil {
		t.Errorf("SaveRejectionFeedback() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(feedbackPath); os.IsNotExist(err) {
		t.Error("Feedback file should exist")
	}

	// Verify content
	data, _ := os.ReadFile(feedbackPath)
	var feedback RejectionFeedback
	json.Unmarshal(data, &feedback)

	if feedback.OverallScore != 55 {
		t.Errorf("Feedback score = %d, want 55", feedback.OverallScore)
	}
	if len(feedback.Issues) != 2 {
		t.Errorf("Feedback issues = %d, want 2", len(feedback.Issues))
	}
}

func TestTrackerAgent_SaveRejectionFeedback_NoReview(t *testing.T) {
	agent := NewTrackerAgent(&AgentConfig{Type: AgentTracker}, nil, "")

	input := &TrackerInput{
		Posting: &ParsedPosting{Company: "Test", Position: "Dev"},
	}

	_, err := agent.SaveRejectionFeedback(input, "/tmp")
	if err == nil {
		t.Error("SaveRejectionFeedback() expected error when no review")
	}
}

func TestTrackerAgent_CopyPostingToApplicationFolder(t *testing.T) {
	agent := NewTrackerAgent(&AgentConfig{Type: AgentTracker}, nil, "")

	tmpDir := t.TempDir()
	postingPath := filepath.Join(tmpDir, "source-posting.md")
	appDir := filepath.Join(tmpDir, "applications", "swe", "acme")
	content := "# Original Posting"

	// Create source posting
	if err := os.WriteFile(postingPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create source posting: %v", err)
	}

	err := agent.CopyPostingToApplicationFolder(postingPath, appDir)
	if err != nil {
		t.Errorf("CopyPostingToApplicationFolder() error = %v", err)
	}

	// Verify copied file
	copiedPath := filepath.Join(appDir, "posting.md")
	data, err := os.ReadFile(copiedPath)
	if err != nil {
		t.Errorf("Failed to read copied posting: %v", err)
	}
	if string(data) != content {
		t.Error("Copied posting content mismatch")
	}
}

func TestTrackerAgent_CopyPostingToApplicationFolder_EmptyPath(t *testing.T) {
	agent := NewTrackerAgent(&AgentConfig{Type: AgentTracker}, nil, "")

	// Empty path should be a no-op
	err := agent.CopyPostingToApplicationFolder("", "/some/dir")
	if err != nil {
		t.Errorf("CopyPostingToApplicationFolder() should not error for empty path: %v", err)
	}
}

func TestTrackerAgent_GetSystemPrompt(t *testing.T) {
	agent := NewTrackerAgent(&AgentConfig{Type: AgentTracker}, nil, "")
	prompt := agent.GetSystemPrompt()

	if len(prompt) == 0 {
		t.Error("GetSystemPrompt() returned empty string")
	}

	requiredElements := []string{
		"tracker",
		"company",
		"position",
		"notes",
	}

	for _, elem := range requiredElements {
		if !containsIgnoreCase(prompt, elem) {
			t.Errorf("GetSystemPrompt() missing element: %s", elem)
		}
	}
}

func TestTrackerAgent_GetUserPrompt(t *testing.T) {
	agent := NewTrackerAgent(&AgentConfig{Type: AgentTracker}, nil, "")

	input := &TrackerInput{
		Posting: &ParsedPosting{
			Company:  "TestCorp",
			Position: "Developer",
		},
	}

	prompt, err := agent.GetUserPrompt(input)
	if err != nil {
		t.Errorf("GetUserPrompt() error = %v", err)
	}

	if !containsIgnoreCase(prompt, "TestCorp") {
		t.Error("GetUserPrompt() missing company name")
	}
}

func TestTrackerAgent_Integrate_Approved(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "applications.json")
	os.WriteFile(storePath, []byte("[]"), 0644)
	s, _ := store.New(storePath)

	agent := NewTrackerAgent(&AgentConfig{Type: AgentTracker}, s, tmpDir)

	input := &TrackerInput{
		Posting: &ParsedPosting{
			Company:  "ApprovedCorp",
			Position: "Engineer",
		},
		DetailedReview: &DetailedReviewResult{
			Approved:     true,
			OverallScore: 85,
		},
	}

	output, err := agent.Integrate(input)
	if err != nil {
		t.Errorf("Integrate() error = %v", err)
	}

	if output.Status != "applied" {
		t.Errorf("Status = %q, want %q", output.Status, "applied")
	}
	if output.ApplicationID == "" {
		t.Error("ApplicationID should not be empty for approved application")
	}
}

func TestTrackerAgent_Integrate_Rejected(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "applications.json")
	os.WriteFile(storePath, []byte("[]"), 0644)
	s, _ := store.New(storePath)

	agent := NewTrackerAgent(&AgentConfig{Type: AgentTracker}, s, tmpDir)

	// Ensure local/postings directory exists
	os.MkdirAll(filepath.Join(tmpDir, "local", "postings"), 0755)

	input := &TrackerInput{
		Posting: &ParsedPosting{
			Company:  "RejectedCorp",
			Position: "Engineer",
		},
		DetailedReview: &DetailedReviewResult{
			Approved:       false,
			OverallScore:   45,
			Recommendation: "Not recommended",
		},
	}

	output, err := agent.Integrate(input)
	if err != nil {
		t.Errorf("Integrate() error = %v", err)
	}

	if output.Status != "rejected" {
		t.Errorf("Status = %q, want %q", output.Status, "rejected")
	}
	if output.ApplicationID != "" {
		t.Error("ApplicationID should be empty for rejected application")
	}
	if !strings.Contains(output.Notes, "Feedback saved") {
		t.Error("Notes should mention feedback file")
	}
}

func TestTrackerAgent_Integrate_AutoDeterminesJobType(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "applications.json")
	os.WriteFile(storePath, []byte("[]"), 0644)
	s, _ := store.New(storePath)

	agent := NewTrackerAgent(&AgentConfig{Type: AgentTracker}, s, tmpDir)

	input := &TrackerInput{
		Posting: &ParsedPosting{
			Company:  "FrontendCo",
			Position: "React Developer",
		},
		ReviewResult: &ReviewResult{
			Approved: true,
		},
	}

	_, err := agent.Integrate(input)
	if err != nil {
		t.Errorf("Integrate() error = %v", err)
	}

	// Verify job type was auto-determined
	if input.JobType != "fe-dev" {
		t.Errorf("JobType = %q, want %q", input.JobType, "fe-dev")
	}
}
