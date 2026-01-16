package fetch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// CVResult contains the result of a CV fetch operation
type CVResult struct {
	URL        string `json:"url"`
	OutputPath string `json:"output_path"`
	Name       string `json:"name,omitempty"`
	Label      string `json:"label,omitempty"`
	Size       int    `json:"content_size"`
}

// FetchCV downloads a CV (JSON Resume) from a domain and saves it to local/cv.json
// The input can be:
// - A bare domain: "cello.design" → fetches https://cello.design/cv.json
// - A full URL: "https://example.com/cv.json" → fetches as-is
func (f *Fetcher) FetchCV(input string) (*CVResult, error) {
	// Build the URL
	cvURL := buildCVURL(input)

	// Fetch the CV
	req, err := http.NewRequest("GET", cvURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, */*")

	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CV: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Validate it's valid JSON
	var cvData map[string]interface{}
	if err := json.Unmarshal(body, &cvData); err != nil {
		return nil, fmt.Errorf("response is not valid JSON: %w", err)
	}

	// Extract name and label from JSON Resume format
	name := extractJSONField(cvData, "basics", "name")
	label := extractJSONField(cvData, "basics", "label")

	// Determine output path
	outputPath := filepath.Join("local", "cv.json")

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Pretty-print the JSON for readability
	prettyJSON, err := json.MarshalIndent(cvData, "", "  ")
	if err != nil {
		// Fall back to original body if pretty-print fails
		prettyJSON = body
	}

	// Write to file
	if err := os.WriteFile(outputPath, prettyJSON, 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &CVResult{
		URL:        cvURL,
		OutputPath: outputPath,
		Name:       name,
		Label:      label,
		Size:       len(prettyJSON),
	}, nil
}

// buildCVURL constructs the URL to fetch the CV from
func buildCVURL(input string) string {
	// If it already looks like a full URL
	if IsURL(input) {
		// If it ends with /cv.json, use as-is
		if strings.HasSuffix(input, "/cv.json") {
			return input
		}
		// Otherwise, append /cv.json
		input = strings.TrimRight(input, "/")
		return input + "/cv.json"
	}

	// Bare domain - add https:// and /cv.json
	return "https://" + strings.TrimRight(input, "/") + "/cv.json"
}

// extractJSONField extracts a nested field from a JSON object
func extractJSONField(data map[string]interface{}, keys ...string) string {
	current := data
	for i, key := range keys {
		if val, ok := current[key]; ok {
			if i == len(keys)-1 {
				// Last key - return as string
				if str, ok := val.(string); ok {
					return str
				}
				return ""
			}
			// Not last key - navigate deeper
			if nested, ok := val.(map[string]interface{}); ok {
				current = nested
			} else {
				return ""
			}
		} else {
			return ""
		}
	}
	return ""
}
