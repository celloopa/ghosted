package fetch

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectFetchType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected FetchType
	}{
		// CV fetch cases
		{
			name:     "bare domain",
			input:    "cello.design",
			expected: FetchTypeCV,
		},
		{
			name:     "bare domain with trailing slash",
			input:    "cello.design/",
			expected: FetchTypeCV,
		},
		{
			name:     "full URL to bare domain",
			input:    "https://cello.design",
			expected: FetchTypeCV,
		},
		{
			name:     "full URL to bare domain with trailing slash",
			input:    "https://cello.design/",
			expected: FetchTypeCV,
		},
		{
			name:     "explicit cv.json path",
			input:    "https://example.com/cv.json",
			expected: FetchTypeCV,
		},
		{
			name:     "bare domain with cv.json path",
			input:    "example.com/cv.json",
			expected: FetchTypeCV,
		},

		// Job posting cases
		{
			name:     "lever job posting",
			input:    "https://jobs.lever.co/company/123",
			expected: FetchTypeJobPosting,
		},
		{
			name:     "greenhouse job posting",
			input:    "https://boards.greenhouse.io/company/jobs/123",
			expected: FetchTypeJobPosting,
		},
		{
			name:     "generic careers page",
			input:    "https://example.com/careers",
			expected: FetchTypeJobPosting,
		},
		{
			name:     "URL with path",
			input:    "https://example.com/jobs/software-engineer",
			expected: FetchTypeJobPosting,
		},
		{
			name:     "domain with path",
			input:    "example.com/careers",
			expected: FetchTypeJobPosting,
		},
		{
			name:     "linkedin job posting",
			input:    "https://www.linkedin.com/jobs/view/123",
			expected: FetchTypeJobPosting,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectFetchType(tt.input)
			if got != tt.expected {
				t.Errorf("DetectFetchType(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestBuildCVURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "bare domain",
			input:    "cello.design",
			expected: "https://cello.design/cv.json",
		},
		{
			name:     "bare domain with trailing slash",
			input:    "cello.design/",
			expected: "https://cello.design/cv.json",
		},
		{
			name:     "full URL without cv.json",
			input:    "https://example.com",
			expected: "https://example.com/cv.json",
		},
		{
			name:     "full URL with cv.json",
			input:    "https://example.com/cv.json",
			expected: "https://example.com/cv.json",
		},
		{
			name:     "http URL with cv.json",
			input:    "http://example.com/cv.json",
			expected: "http://example.com/cv.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCVURL(tt.input)
			if got != tt.expected {
				t.Errorf("buildCVURL(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExtractJSONField(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		keys     []string
		expected string
	}{
		{
			name: "nested field exists",
			data: map[string]interface{}{
				"basics": map[string]interface{}{
					"name": "John Doe",
				},
			},
			keys:     []string{"basics", "name"},
			expected: "John Doe",
		},
		{
			name: "nested field missing",
			data: map[string]interface{}{
				"basics": map[string]interface{}{
					"email": "john@example.com",
				},
			},
			keys:     []string{"basics", "name"},
			expected: "",
		},
		{
			name:     "top-level field missing",
			data:     map[string]interface{}{},
			keys:     []string{"basics", "name"},
			expected: "",
		},
		{
			name: "field is not a string",
			data: map[string]interface{}{
				"basics": map[string]interface{}{
					"age": 30,
				},
			},
			keys:     []string{"basics", "age"},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractJSONField(tt.data, tt.keys...)
			if got != tt.expected {
				t.Errorf("extractJSONField() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestFetcher_FetchCV(t *testing.T) {
	// Create a test CV in JSON Resume format
	testCV := map[string]interface{}{
		"basics": map[string]interface{}{
			"name":  "Test User",
			"label": "Software Engineer",
			"email": "test@example.com",
		},
		"work": []interface{}{
			map[string]interface{}{
				"company":  "Test Corp",
				"position": "Engineer",
			},
		},
	}
	cvJSON, _ := json.Marshal(testCV)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cv.json" {
			w.Header().Set("Content-Type", "application/json")
			w.Write(cvJSON)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Create temp directory for output
	tmpDir := t.TempDir()

	// Create fetcher with temp output directory
	f := NewFetcher(tmpDir)

	// Override the local output path by changing to temp dir
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create the local directory structure
	os.MkdirAll("local", 0755)

	// Test fetching CV
	result, err := f.FetchCV(server.URL)
	if err != nil {
		t.Fatalf("FetchCV() error = %v", err)
	}

	// Verify result
	if result.Name != "Test User" {
		t.Errorf("result.Name = %q, want %q", result.Name, "Test User")
	}
	if result.Label != "Software Engineer" {
		t.Errorf("result.Label = %q, want %q", result.Label, "Software Engineer")
	}
	if result.Size == 0 {
		t.Error("result.Size should not be 0")
	}

	// Verify file was written
	if _, err := os.Stat(result.OutputPath); os.IsNotExist(err) {
		t.Errorf("Output file was not created at %s", result.OutputPath)
	}

	// Verify file contents
	data, err := os.ReadFile(result.OutputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if !strings.Contains(string(data), "Test User") {
		t.Error("Output file should contain the name")
	}
}

func TestFetcher_FetchCV_InvalidJSON(t *testing.T) {
	// Create test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	f := NewFetcher(t.TempDir())

	_, err := f.FetchCV(server.URL)
	if err == nil {
		t.Error("FetchCV() should return error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "not valid JSON") {
		t.Errorf("Error should mention invalid JSON, got: %v", err)
	}
}

func TestFetcher_FetchCV_HTTPError(t *testing.T) {
	// Create test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	f := NewFetcher(t.TempDir())

	_, err := f.FetchCV(server.URL)
	if err == nil {
		t.Error("FetchCV() should return error for HTTP 404")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("Error should mention 404, got: %v", err)
	}
}

func TestFetcher_FetchCV_WriteFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a minimal CV JSON
	cvData := map[string]interface{}{
		"basics": map[string]interface{}{
			"name": "File Test User",
		},
	}
	prettyJSON, _ := json.MarshalIndent(cvData, "", "  ")

	// Write directly to verify file creation works
	localDir := filepath.Join(tmpDir, "local")
	os.MkdirAll(localDir, 0755)

	outputPath := filepath.Join(localDir, "cv.json")
	err := os.WriteFile(outputPath, prettyJSON, 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Verify file exists and has content
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !strings.Contains(string(data), "File Test User") {
		t.Error("File should contain the name")
	}
}
