package fetch

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsURL(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"https://jobs.lever.co/company/123", true},
		{"http://example.com/job", true},
		{"local/postings/job.md", false},
		{"/absolute/path/job.md", false},
		{"job.md", false},
		{"ftp://example.com/job", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := IsURL(tt.input)
			if got != tt.expected {
				t.Errorf("IsURL(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFetcher_GenerateFilename(t *testing.T) {
	f := NewFetcher("")

	tests := []struct {
		name     string
		company  string
		position string
		urlStr   string
		expected string
	}{
		{
			name:     "with company and position",
			company:  "Acme Corp",
			position: "Software Engineer",
			urlStr:   "https://jobs.lever.co/acme/123",
			expected: "acme-corp-software-engineer-posting",
		},
		{
			name:     "only URL path",
			company:  "",
			position: "",
			urlStr:   "https://jobs.lever.co/acme/senior-engineer",
			expected: "senior-engineer-posting",
		},
		{
			name:     "special characters",
			company:  "Company & Co.",
			position: "Sr. Engineer (Remote)",
			urlStr:   "https://example.com/job",
			expected: "company-co-sr-engineer-remote-posting",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedURL, _ := url.Parse(tt.urlStr)
			got := f.GenerateFilename(tt.company, tt.position, parsedURL)
			if got != tt.expected {
				t.Errorf("GenerateFilename() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestFetcher_FormatOutput(t *testing.T) {
	f := NewFetcher("")

	content := "This is the job description."
	sourceURL := "https://example.com/job"
	company := "Test Corp"
	position := "Engineer"

	output := f.FormatOutput(content, sourceURL, company, position)

	// Check for metadata
	if !strings.Contains(output, "source: https://example.com/job") {
		t.Error("Output missing source URL")
	}
	if !strings.Contains(output, "company: Test Corp") {
		t.Error("Output missing company")
	}
	if !strings.Contains(output, "position: Engineer") {
		t.Error("Output missing position")
	}

	// Check for content structure
	if !strings.Contains(output, "# Engineer") {
		t.Error("Output missing position header")
	}
	if !strings.Contains(output, "**Company:** Test Corp") {
		t.Error("Output missing company line")
	}
	if !strings.Contains(output, "## Job Description") {
		t.Error("Output missing job description header")
	}
	if !strings.Contains(output, content) {
		t.Error("Output missing content")
	}
}

func TestCleanHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
		excludes []string
	}{
		{
			name:     "removes script tags",
			input:    "<p>Hello</p><script>alert('xss')</script><p>World</p>",
			contains: []string{"Hello", "World"},
			excludes: []string{"<script>", "alert"},
		},
		{
			name:     "converts headers",
			input:    "<h1>Title</h1><h2>Subtitle</h2>",
			contains: []string{"# Title", "## Subtitle"},
			excludes: []string{"<h1>", "</h1>"},
		},
		{
			name:     "converts lists",
			input:    "<ul><li>Item 1</li><li>Item 2</li></ul>",
			contains: []string{"- Item 1", "- Item 2"},
			excludes: []string{"<li>", "</li>"},
		},
		{
			name:     "converts bold and italic",
			input:    "<strong>Bold</strong> and <em>italic</em>",
			contains: []string{"**Bold**", "*italic*"},
			excludes: []string{"<strong>", "<em>"},
		},
		{
			name:     "decodes entities",
			input:    "Tom &amp; Jerry &lt;3",
			contains: []string{"Tom & Jerry <3"},
			excludes: []string{"&amp;", "&lt;"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanHTML(tt.input)
			for _, s := range tt.contains {
				if !strings.Contains(got, s) {
					t.Errorf("cleanHTML() should contain %q, got %q", s, got)
				}
			}
			for _, s := range tt.excludes {
				if strings.Contains(got, s) {
					t.Errorf("cleanHTML() should not contain %q, got %q", s, got)
				}
			}
		})
	}
}

func TestCleanText(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<span>Hello</span>", "Hello"},
		{"Tom &amp; Jerry", "Tom & Jerry"},
		{"  Trimmed  ", "Trimmed"},
		{"<a href='#'>Link</a>", "Link"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := cleanText(tt.input)
			if got != tt.expected {
				t.Errorf("cleanText(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Acme Corp", "acme-corp"},
		{"Software Engineer", "software-engineer"},
		{"Sr. Engineer (Remote)", "sr-engineer-remote"},
		{"Company/Division", "companydivision"},
		{"Test--Name", "test-name"},
		{"-Leading-Dash-", "leading-dash"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeFilename(tt.input)
			if got != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExtractBetween(t *testing.T) {
	html := `<div class="content">Hello World</div><div>Other</div>`

	got := extractBetween(html, `<div class="content">`, `</div>`)
	if got != "Hello World" {
		t.Errorf("extractBetween() = %q, want %q", got, "Hello World")
	}

	// Test not found
	got = extractBetween(html, `<span>`, `</span>`)
	if got != "" {
		t.Errorf("extractBetween() should return empty for not found, got %q", got)
	}
}

func TestExtractMetaContent(t *testing.T) {
	html := `
		<html>
		<head>
		<meta property="og:title" content="Software Engineer at Acme">
		<meta property="og:site_name" content="Acme Corp">
		<meta name="description" content="Job description here">
		</head>
		</html>
	`

	tests := []struct {
		property string
		expected string
	}{
		{"og:title", "Software Engineer at Acme"},
		{"og:site_name", "Acme Corp"},
		{"description", "Job description here"},
		{"nonexistent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.property, func(t *testing.T) {
			got := extractMetaContent(html, tt.property)
			if got != tt.expected {
				t.Errorf("extractMetaContent(%q) = %q, want %q", tt.property, got, tt.expected)
			}
		})
	}
}

func TestFetcher_ExtractJobPosting_Generic(t *testing.T) {
	f := NewFetcher("")

	html := `
		<html>
		<head>
		<meta property="og:title" content="Software Engineer">
		<meta property="og:site_name" content="Test Company">
		<meta property="og:description" content="We are looking for a software engineer.">
		</head>
		<body>
		<div class="job-description">
			<h2>About the Role</h2>
			<p>Great opportunity!</p>
		</div>
		</body>
		</html>
	`

	parsedURL, _ := url.Parse("https://example.com/jobs/123")
	content, company, position := f.ExtractJobPosting(html, parsedURL)

	if position != "Software Engineer" {
		t.Errorf("position = %q, want %q", position, "Software Engineer")
	}
	if company != "Test Company" {
		t.Errorf("company = %q, want %q", company, "Test Company")
	}
	if content == "" {
		t.Error("content should not be empty")
	}
}

func TestFetcher_Fetch_WriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	f := NewFetcher(tmpDir)

	// We can't test actual HTTP fetching without a mock server,
	// but we can test the file writing by using the FormatOutput directly
	content := f.FormatOutput("Test content", "https://example.com", "Test Co", "Engineer")

	outputPath := filepath.Join(tmpDir, "test-posting.md")
	err := os.WriteFile(outputPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Verify file exists and has content
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !strings.Contains(string(data), "Test Co") {
		t.Error("File should contain company name")
	}
}

func TestNewFetcher(t *testing.T) {
	f := NewFetcher("/tmp/postings")

	if f.OutputDir != "/tmp/postings" {
		t.Errorf("OutputDir = %q, want %q", f.OutputDir, "/tmp/postings")
	}
	if f.Client == nil {
		t.Error("Client should not be nil")
	}
}
