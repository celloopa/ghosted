package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParserAgent_SupportedExtensions(t *testing.T) {
	agent := NewParserAgent(&AgentConfig{Type: AgentParser})

	tests := []struct {
		path     string
		expected bool
	}{
		{"job-posting.md", true},
		{"job-posting.txt", true},
		{"job-posting.png", true},
		{"job-posting.jpg", true},
		{"job-posting.jpeg", true},
		{"job-posting.pdf", false},
		{"job-posting.docx", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := agent.IsSupported(tt.path); got != tt.expected {
				t.Errorf("IsSupported(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestParserAgent_IsImageFile(t *testing.T) {
	agent := NewParserAgent(&AgentConfig{Type: AgentParser})

	tests := []struct {
		path     string
		expected bool
	}{
		{"job-posting.png", true},
		{"job-posting.jpg", true},
		{"job-posting.jpeg", true},
		{"job-posting.md", false},
		{"job-posting.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := agent.IsImageFile(tt.path); got != tt.expected {
				t.Errorf("IsImageFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestParserAgent_ReadPosting(t *testing.T) {
	agent := NewParserAgent(&AgentConfig{Type: AgentParser})

	// Create a temp file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-posting.md")
	content := "# Software Engineer\nAcme Corp\nSan Francisco, CA"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test reading text file
	result, err := agent.ReadPosting(testFile)
	if err != nil {
		t.Errorf("ReadPosting() error = %v", err)
	}
	if result != content {
		t.Errorf("ReadPosting() = %q, want %q", result, content)
	}

	// Test unsupported file
	unsupported := filepath.Join(tmpDir, "test.pdf")
	os.WriteFile(unsupported, []byte("test"), 0644)
	_, err = agent.ReadPosting(unsupported)
	if err == nil {
		t.Error("ReadPosting() expected error for unsupported file type")
	}
}

func TestParserAgent_ParseJSON(t *testing.T) {
	agent := NewParserAgent(&AgentConfig{Type: AgentParser})

	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name: "valid JSON",
			json: `{
				"company": "Twitch",
				"position": "Software Engineer",
				"location": "Seattle, WA",
				"remote": false,
				"requirements": ["1+ years experience"],
				"tech_stack": ["Go", "TypeScript"]
			}`,
			wantErr: false,
		},
		{
			name: "JSON with markdown code block",
			json: "```json\n{\"company\": \"Acme\", \"position\": \"Engineer\"}\n```",
			wantErr: false,
		},
		{
			name:    "missing company",
			json:    `{"position": "Engineer"}`,
			wantErr: true,
		},
		{
			name:    "missing position",
			json:    `{"company": "Acme"}`,
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			json:    `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := agent.ParseJSON(tt.json)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && parsed == nil {
				t.Error("ParseJSON() returned nil for valid input")
			}
		})
	}
}

func TestParserAgent_GetSystemPrompt(t *testing.T) {
	agent := NewParserAgent(&AgentConfig{Type: AgentParser})
	prompt := agent.GetSystemPrompt()

	// Verify prompt contains key elements
	if len(prompt) == 0 {
		t.Error("GetSystemPrompt() returned empty string")
	}

	requiredElements := []string{
		"company",
		"position",
		"requirements",
		"bonus_skills",
		"tech_stack",
		"company_values",
		"keywords",
	}

	for _, elem := range requiredElements {
		if !contains(prompt, elem) {
			t.Errorf("GetSystemPrompt() missing required element: %s", elem)
		}
	}
}

func TestExtractBasicInfo(t *testing.T) {
	content := `Software Engineer
Seattle, WA
About Us
Twitch is the world's biggest live streaming service.
This is a remote-friendly position.`

	parsed := extractBasicInfo(content, "twitch-swe-posting.md")

	if parsed.Company != "Twitch" {
		t.Errorf("Company = %q, want %q", parsed.Company, "Twitch")
	}
	if parsed.Position != "Swe" {
		t.Errorf("Position = %q, want %q", parsed.Position, "Swe")
	}
	if parsed.Location != "Seattle, WA" {
		t.Errorf("Location = %q, want %q", parsed.Location, "Seattle, WA")
	}
	if !parsed.Remote {
		t.Error("Remote = false, want true")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
