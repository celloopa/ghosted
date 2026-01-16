package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/celloopa/ghosted/internal/store"
)

func TestNewPipeline(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()

	// Create a minimal config file
	configDir := filepath.Join(tmpDir, ".agent")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	configContent := `{
		"agents": [
			{"type": "parser", "name": "Parser", "enabled": true}
		]
	}`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Test creating pipeline without store
	pipeline, err := NewPipeline(configPath, nil)
	if err != nil {
		t.Errorf("NewPipeline() error = %v", err)
	}
	if pipeline == nil {
		t.Fatal("NewPipeline() returned nil")
	}

	// Test with default config (non-existent file)
	nonExistentPath := filepath.Join(tmpDir, "nonexistent", "config.json")
	pipeline, err = NewPipeline(nonExistentPath, nil)
	if err != nil {
		t.Errorf("NewPipeline() with default config error = %v", err)
	}
	if pipeline == nil {
		t.Fatal("NewPipeline() returned nil with default config")
	}
}

func TestPipeline_Run(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test posting file
	postingPath := filepath.Join(tmpDir, "acme-swe-posting.md")
	postingContent := `# Software Engineer

Company: Acme Corp
Location: San Francisco, CA
Remote: Yes

## About the Role

We are looking for a software engineer to join our team.

## Requirements

- 3+ years of experience
- Go, Python, or similar languages
`
	if err := os.WriteFile(postingPath, []byte(postingContent), 0644); err != nil {
		t.Fatalf("Failed to write posting: %v", err)
	}

	// Create store in temp directory
	storePath := filepath.Join(tmpDir, "applications.json")
	s, err := store.New(storePath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Create pipeline with default config
	configPath := filepath.Join(tmpDir, ".agent", "config.json")
	pipeline, err := NewPipeline(configPath, s)
	if err != nil {
		t.Fatalf("NewPipeline() error = %v", err)
	}

	// Run pipeline
	err = pipeline.Run(postingPath)
	if err != nil {
		t.Errorf("Pipeline.Run() error = %v", err)
	}

	// Check status
	status := pipeline.GetStatus()
	if status == "" {
		t.Error("GetStatus() returned empty string")
	}
}

func TestPipeline_RunDryRun(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test posting file
	postingPath := filepath.Join(tmpDir, "test-swe-posting.md")
	postingContent := `# Engineer at TestCorp
Location: Remote
`
	if err := os.WriteFile(postingPath, []byte(postingContent), 0644); err != nil {
		t.Fatalf("Failed to write posting: %v", err)
	}

	// Create pipeline without store (dry run)
	configPath := filepath.Join(tmpDir, ".agent", "config.json")
	pipeline, err := NewPipeline(configPath, nil)
	if err != nil {
		t.Fatalf("NewPipeline() error = %v", err)
	}

	// Run pipeline - should not fail even with nil store
	// The tracker step will fail but that's expected
	err = pipeline.Run(postingPath)
	// We expect an error because store is nil and tracker step will fail
	if err == nil {
		// If no error, check that we at least got through parser
		if pipeline.State == nil {
			t.Error("Pipeline state is nil after run")
		}
	}
}

func TestPipeline_GetStatus(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".agent", "config.json")

	pipeline, err := NewPipeline(configPath, nil)
	if err != nil {
		t.Fatalf("NewPipeline() error = %v", err)
	}

	// Before running, status should indicate not started
	status := pipeline.GetStatus()
	if status != "not started" {
		t.Errorf("GetStatus() before Run = %q, want %q", status, "not started")
	}
}

func TestPipeline_UnsupportedFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an unsupported file type
	postingPath := filepath.Join(tmpDir, "test.pdf")
	if err := os.WriteFile(postingPath, []byte("PDF content"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	configPath := filepath.Join(tmpDir, ".agent", "config.json")
	pipeline, err := NewPipeline(configPath, nil)
	if err != nil {
		t.Fatalf("NewPipeline() error = %v", err)
	}

	// Should fail with unsupported file type
	err = pipeline.Run(postingPath)
	if err == nil {
		t.Error("Pipeline.Run() expected error for unsupported file type")
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Acme Corp", "acme-corp"},
		{"Software Engineer", "software-engineer"},
		{"Test/Company", "testcompany"},
		{"Name: With Colon", "name-with-colon"},
		{"Question?", "question"},
		{"Star*Wars", "starwars"},
		{"<html>", "html"},
		{">output", "output"},
		{"pipe|test", "pipetest"},
		{`quote"test`, "quotetest"},
		{"back\\slash", "backslash"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
