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

func TestFetcher_ExtractMicrosoft_NextData(t *testing.T) {
	f := NewFetcher("")

	// Simulated Microsoft Careers page with __NEXT_DATA__
	html := `
		<html>
		<head>
		<meta property="og:title" content="UX Designer | Microsoft Careers">
		<meta property="og:description" content="Join Microsoft as a UX Designer">
		</head>
		<body>
		<script id="__NEXT_DATA__" type="application/json">
		{
			"props": {
				"pageProps": {
					"job": {
						"title": "UX Designer II",
						"description": "We are looking for a passionate UX Designer to join our team.",
						"qualifications": "5+ years of experience in UX design",
						"responsibilities": "Design user interfaces for Microsoft products",
						"location": "Redmond, WA",
						"employmentType": "Full-time"
					}
				}
			}
		}
		</script>
		</body>
		</html>
	`

	parsedURL, _ := url.Parse("https://careers.microsoft.com/us/en/job/123456")
	content, company, position := f.ExtractJobPosting(html, parsedURL)

	if company != "Microsoft" {
		t.Errorf("company = %q, want %q", company, "Microsoft")
	}
	if position != "UX Designer II" {
		t.Errorf("position = %q, want %q", position, "UX Designer II")
	}
	if !strings.Contains(content, "passionate UX Designer") {
		t.Errorf("content should contain job description, got %q", content)
	}
	if !strings.Contains(content, "Redmond, WA") {
		t.Errorf("content should contain location, got %q", content)
	}
}

func TestFetcher_ExtractMicrosoft_FallbackToMeta(t *testing.T) {
	f := NewFetcher("")

	// Microsoft page without __NEXT_DATA__ - should fall back to meta tags
	html := `
		<html>
		<head>
		<meta property="og:title" content="Software Engineer | Microsoft Careers">
		<meta property="og:description" content="Build amazing software at Microsoft">
		</head>
		<body>
		<div>Some content</div>
		</body>
		</html>
	`

	parsedURL, _ := url.Parse("https://careers.microsoft.com/us/en/job/789")
	content, company, position := f.ExtractJobPosting(html, parsedURL)

	if company != "Microsoft" {
		t.Errorf("company = %q, want %q", company, "Microsoft")
	}
	// Title should have " | Microsoft Careers" stripped
	if position != "Software Engineer" {
		t.Errorf("position = %q, want %q", position, "Software Engineer")
	}
	if !strings.Contains(content, "Build amazing software") {
		t.Errorf("content should contain og:description, got %q", content)
	}
}

func TestFetcher_ExtractMicrosoft_JobDetail(t *testing.T) {
	f := NewFetcher("")

	// Alternative structure with jobDetail instead of job
	html := `
		<html>
		<script id="__NEXT_DATA__" type="application/json">
		{
			"props": {
				"pageProps": {
					"jobDetail": {
						"jobTitle": "Product Manager",
						"jobDescription": "Lead product development at Microsoft",
						"primaryLocation": "Seattle, WA"
					}
				}
			}
		}
		</script>
		</html>
	`

	parsedURL, _ := url.Parse("https://careers.microsoft.com/job/456")
	content, company, position := f.ExtractJobPosting(html, parsedURL)

	if company != "Microsoft" {
		t.Errorf("company = %q, want %q", company, "Microsoft")
	}
	if position != "Product Manager" {
		t.Errorf("position = %q, want %q", position, "Product Manager")
	}
	if !strings.Contains(content, "Lead product development") {
		t.Errorf("content should contain jobDescription, got %q", content)
	}
}

func TestFetcher_ExtractMicrosoft_QualificationsArray(t *testing.T) {
	f := NewFetcher("")

	// Test with qualifications as array
	html := `
		<html>
		<script id="__NEXT_DATA__" type="application/json">
		{
			"props": {
				"pageProps": {
					"job": {
						"title": "Senior Engineer",
						"description": "Join our engineering team",
						"qualifications": ["Bachelor's degree", "5+ years experience", "Strong coding skills"]
					}
				}
			}
		}
		</script>
		</html>
	`

	parsedURL, _ := url.Parse("https://careers.microsoft.com/job/789")
	content, _, _ := f.ExtractJobPosting(html, parsedURL)

	if !strings.Contains(content, "Bachelor's degree") {
		t.Errorf("content should contain qualifications, got %q", content)
	}
	if !strings.Contains(content, "5+ years experience") {
		t.Errorf("content should contain qualifications, got %q", content)
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"123", true},
		{"0", true},
		{"123456789", true},
		{"abc", false},
		{"12a3", false},
		{"", false},
		{"  ", false},
		{" 123 ", true},
		{"Microsoft", false},
		{"123-456", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isNumeric(tt.input)
			if got != tt.expected {
				t.Errorf("isNumeric(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestValidateMicrosoftExtraction(t *testing.T) {
	f := NewFetcher("")

	tests := []struct {
		name            string
		content         string
		company         string
		position        string
		wantCompany     string
		wantPosition    string
	}{
		{
			name:         "valid data",
			content:      "Job description",
			company:      "Microsoft",
			position:     "Software Engineer",
			wantCompany:  "Microsoft",
			wantPosition: "Software Engineer",
		},
		{
			name:         "numeric company rejected",
			content:      "Job description",
			company:      "12345",
			position:     "Engineer",
			wantCompany:  "Microsoft",
			wantPosition: "Engineer",
		},
		{
			name:         "short position rejected",
			content:      "Job description",
			company:      "Microsoft",
			position:     "AB",
			wantCompany:  "Microsoft",
			wantPosition: "",
		},
		{
			name:         "numeric position rejected",
			content:      "Job description",
			company:      "Microsoft",
			position:     "123456",
			wantCompany:  "Microsoft",
			wantPosition: "",
		},
		{
			name:         "empty position rejected",
			content:      "Job description",
			company:      "Microsoft",
			position:     "",
			wantCompany:  "Microsoft",
			wantPosition: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotCompany, gotPosition := f.validateMicrosoftExtraction(tt.content, tt.company, tt.position)
			if gotCompany != tt.wantCompany {
				t.Errorf("company = %q, want %q", gotCompany, tt.wantCompany)
			}
			if gotPosition != tt.wantPosition {
				t.Errorf("position = %q, want %q", gotPosition, tt.wantPosition)
			}
		})
	}
}

func TestFetcher_ExtractMicrosoft_ApplySubdomain(t *testing.T) {
	f := NewFetcher("")

	html := `
		<html>
		<script id="__NEXT_DATA__" type="application/json">
		{
			"props": {
				"pageProps": {
					"job": {
						"title": "Data Scientist",
						"description": "Work with big data at Microsoft"
					}
				}
			}
		}
		</script>
		</html>
	`

	// Test apply.careers.microsoft.com subdomain
	parsedURL, _ := url.Parse("https://apply.careers.microsoft.com/job/123")
	content, company, position := f.ExtractJobPosting(html, parsedURL)

	if company != "Microsoft" {
		t.Errorf("company = %q, want %q", company, "Microsoft")
	}
	if position != "Data Scientist" {
		t.Errorf("position = %q, want %q", position, "Data Scientist")
	}
	if content == "" {
		t.Error("content should not be empty")
	}
}
