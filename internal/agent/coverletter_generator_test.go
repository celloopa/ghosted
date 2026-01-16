package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCoverLetterGeneratorAgent_LoadCV(t *testing.T) {
	agent := NewCoverLetterGeneratorAgent(&AgentConfig{Type: AgentCover}, "")

	tmpDir := t.TempDir()
	cvPath := filepath.Join(tmpDir, "cv.json")

	validCV := `{
		"basics": {
			"name": "Jane Smith",
			"email": "jane@example.com",
			"summary": "Experienced engineer"
		},
		"work": [
			{
				"name": "TechCorp",
				"position": "Senior Engineer",
				"highlights": ["Led team of 5", "Built React apps"]
			}
		]
	}`

	if err := os.WriteFile(cvPath, []byte(validCV), 0644); err != nil {
		t.Fatalf("Failed to create test CV: %v", err)
	}

	cv, err := agent.LoadCV(cvPath)
	if err != nil {
		t.Errorf("LoadCV() error = %v", err)
		return
	}

	if cv.Basics.Name != "Jane Smith" {
		t.Errorf("CV name = %q, want %q", cv.Basics.Name, "Jane Smith")
	}
}

func TestCoverLetterGeneratorAgent_LoadResume(t *testing.T) {
	agent := NewCoverLetterGeneratorAgent(&AgentConfig{Type: AgentCover}, "")

	tmpDir := t.TempDir()
	resumePath := filepath.Join(tmpDir, "resume.typ")
	resumeContent := `#import "@preview/modern-cv:0.9.0": *`

	if err := os.WriteFile(resumePath, []byte(resumeContent), 0644); err != nil {
		t.Fatalf("Failed to create test resume: %v", err)
	}

	content, err := agent.LoadResume(resumePath)
	if err != nil {
		t.Errorf("LoadResume() error = %v", err)
	}
	if content != resumeContent {
		t.Errorf("LoadResume() content mismatch")
	}
}

func TestCoverLetterGeneratorAgent_LoadResume_Missing(t *testing.T) {
	agent := NewCoverLetterGeneratorAgent(&AgentConfig{Type: AgentCover}, "")

	// Missing file should return empty string, no error
	content, err := agent.LoadResume("/nonexistent/path/resume.typ")
	if err != nil {
		t.Errorf("LoadResume() unexpected error = %v", err)
	}
	if content != "" {
		t.Errorf("LoadResume() expected empty string for missing file")
	}
}

func TestCoverLetterGeneratorAgent_GenerateOutputPath(t *testing.T) {
	agent := NewCoverLetterGeneratorAgent(&AgentConfig{Type: AgentCover}, "")

	tests := []struct {
		name      string
		posting   *ParsedPosting
		jobType   string
		outputDir string
		wantPath  string
	}{
		{
			name: "basic path",
			posting: &ParsedPosting{
				Company:  "Acme Corp",
				Position: "Frontend Developer",
			},
			jobType:   "fe-dev",
			outputDir: "/output",
			wantPath:  "/output/fe-dev/acme-corp-frontend_developer/cover-letter.typ",
		},
		{
			name: "default job type",
			posting: &ParsedPosting{
				Company:  "StartupCo",
				Position: "Engineer",
			},
			jobType:   "",
			outputDir: "/output",
			wantPath:  "/output/swe/startupco-engineer/cover-letter.typ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := agent.GenerateOutputPath(tt.posting, tt.jobType, tt.outputDir)
			if got != tt.wantPath {
				t.Errorf("GenerateOutputPath() = %q, want %q", got, tt.wantPath)
			}
		})
	}
}

func TestCoverLetterGeneratorAgent_WriteTypst(t *testing.T) {
	agent := NewCoverLetterGeneratorAgent(&AgentConfig{Type: AgentCover}, "")

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "nested", "cover-letter.typ")
	content := `#import "@preview/modern-cv:0.9.0": *`

	err := agent.WriteTypst(content, outputPath)
	if err != nil {
		t.Errorf("WriteTypst() error = %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Errorf("Failed to read written file: %v", err)
	}
	if string(data) != content {
		t.Errorf("Written content mismatch")
	}
}

func TestCoverLetterGeneratorAgent_ExtractRelevantExperiences(t *testing.T) {
	agent := NewCoverLetterGeneratorAgent(&AgentConfig{Type: AgentCover}, "")

	cv := &CVData{
		Work: []CVWork{
			{
				Name:       "Backend Corp",
				Position:   "Backend Engineer",
				Highlights: []string{"Built Go services", "Managed databases"},
			},
			{
				Name:       "Frontend Inc",
				Position:   "React Developer",
				Highlights: []string{"Built React components", "TypeScript expert"},
			},
			{
				Name:       "Data Corp",
				Position:   "Data Analyst",
				Highlights: []string{"SQL queries", "Excel reporting"},
			},
		},
	}

	posting := &ParsedPosting{
		Position:  "Frontend Engineer",
		TechStack: []string{"React", "TypeScript"},
		Requirements: []string{
			"3+ years React experience",
			"Strong TypeScript skills",
		},
	}

	relevant := agent.ExtractRelevantExperiences(cv, posting, 2)

	if len(relevant) != 2 {
		t.Errorf("ExtractRelevantExperiences() returned %d experiences, want 2", len(relevant))
	}

	// First result should be the React Developer position (most relevant)
	if len(relevant) > 0 && relevant[0].Position != "React Developer" {
		t.Errorf("First experience = %q, want %q", relevant[0].Position, "React Developer")
	}
}

func TestCoverLetterGeneratorAgent_ExtractRelevantExperiences_DefaultLimit(t *testing.T) {
	agent := NewCoverLetterGeneratorAgent(&AgentConfig{Type: AgentCover}, "")

	cv := &CVData{
		Work: []CVWork{
			{Position: "Job 1", Highlights: []string{"work"}},
			{Position: "Job 2", Highlights: []string{"work"}},
			{Position: "Job 3", Highlights: []string{"work"}},
			{Position: "Job 4", Highlights: []string{"work"}},
			{Position: "Job 5", Highlights: []string{"work"}},
		},
	}

	posting := &ParsedPosting{
		Position: "Engineer",
	}

	// Test with limit=0 should default to 3
	relevant := agent.ExtractRelevantExperiences(cv, posting, 0)

	if len(relevant) != 3 {
		t.Errorf("ExtractRelevantExperiences(limit=0) returned %d experiences, want 3", len(relevant))
	}
}

func TestCoverLetterGeneratorAgent_ParseTypstOutput(t *testing.T) {
	agent := NewCoverLetterGeneratorAgent(&AgentConfig{Type: AgentCover}, "")

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "valid coverletter typst",
			input: `#import "@preview/modern-cv:0.9.0": *
#show: coverletter.with(author: (firstname: "Test"))`,
			wantErr: false,
		},
		{
			name: "with markdown code block",
			input: "```typst\n#import \"@preview/modern-cv:0.9.0\": *\n#show: coverletter.with()\n```",
			wantErr: false,
		},
		{
			name:    "missing import",
			input:   `#show: coverletter.with(author: ())`,
			wantErr: true,
		},
		{
			name:    "missing coverletter template",
			input:   `#import "@preview/some-package:1.0.0": *`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := agent.ParseTypstOutput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTypstOutput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCoverLetterGeneratorAgent_GetSystemPrompt(t *testing.T) {
	agent := NewCoverLetterGeneratorAgent(&AgentConfig{Type: AgentCover}, "")
	prompt := agent.GetSystemPrompt()

	if len(prompt) == 0 {
		t.Error("GetSystemPrompt() returned empty string")
	}

	requiredElements := []string{
		"cover letter",
		"Hook",
		"Bridge",
		"Typst",
		"experience",
	}

	for _, elem := range requiredElements {
		if !containsIgnoreCase(prompt, elem) {
			t.Errorf("GetSystemPrompt() missing element: %s", elem)
		}
	}
}

func TestCoverLetterGeneratorAgent_GetUserPrompt(t *testing.T) {
	agent := NewCoverLetterGeneratorAgent(&AgentConfig{Type: AgentCover}, "")

	posting := &ParsedPosting{
		Company:  "TechCorp",
		Position: "Software Engineer",
	}

	cv := &CVData{
		Basics: CVBasics{
			Name:  "John Doe",
			Email: "john@example.com",
		},
	}

	// Without resume content
	prompt, err := agent.GetUserPrompt(posting, cv, "")
	if err != nil {
		t.Errorf("GetUserPrompt() error = %v", err)
	}

	if !containsIgnoreCase(prompt, "TechCorp") {
		t.Error("GetUserPrompt() missing company name")
	}
	if !containsIgnoreCase(prompt, "John Doe") {
		t.Error("GetUserPrompt() missing candidate name")
	}
}

func TestCoverLetterGeneratorAgent_GetUserPrompt_WithResume(t *testing.T) {
	agent := NewCoverLetterGeneratorAgent(&AgentConfig{Type: AgentCover}, "")

	posting := &ParsedPosting{
		Company:  "Acme",
		Position: "Engineer",
	}

	cv := &CVData{
		Basics: CVBasics{Name: "Test User", Email: "test@test.com"},
	}

	resumeContent := "#import modern-cv... resume content here"

	prompt, err := agent.GetUserPrompt(posting, cv, resumeContent)
	if err != nil {
		t.Errorf("GetUserPrompt() error = %v", err)
	}

	// Should include resume reference
	if !containsIgnoreCase(prompt, "Generated Resume") {
		t.Error("GetUserPrompt() missing resume reference when resume provided")
	}
}

func TestCoverLetterGeneratorAgent_IsTypstAvailable(t *testing.T) {
	agent := NewCoverLetterGeneratorAgent(&AgentConfig{Type: AgentCover}, "")
	// Just test that it runs without error
	_ = agent.IsTypstAvailable()
}
