package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ParserAgent extracts structured data from job posting files
type ParserAgent struct {
	Config *AgentConfig
}

// NewParserAgent creates a new parser agent instance
func NewParserAgent(config *AgentConfig) *ParserAgent {
	return &ParserAgent{Config: config}
}

// SupportedExtensions returns file extensions the parser can handle
func (p *ParserAgent) SupportedExtensions() []string {
	return []string{".md", ".txt", ".png", ".jpg", ".jpeg"}
}

// IsSupported checks if the given file extension is supported
func (p *ParserAgent) IsSupported(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, supported := range p.SupportedExtensions() {
		if ext == supported {
			return true
		}
	}
	return false
}

// IsImageFile checks if the given file is an image
func (p *ParserAgent) IsImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg"
}

// ReadPosting reads the content of a job posting file
func (p *ParserAgent) ReadPosting(path string) (string, error) {
	if !p.IsSupported(path) {
		return "", fmt.Errorf("unsupported file type: %s", filepath.Ext(path))
	}

	if p.IsImageFile(path) {
		// For images, return a marker that indicates the file should be read as an image
		// The actual image reading will be handled by the AI agent
		return fmt.Sprintf("[IMAGE_FILE: %s]", path), nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// GetSystemPrompt returns the system prompt for the parser agent
func (p *ParserAgent) GetSystemPrompt() string {
	return `You are a job posting parser agent. Your task is to extract structured data from job postings.

Given a job posting (text, markdown, or image), extract the following information and return it as JSON:

{
  "company": "Company name",
  "position": "Job title",
  "team": "Team or department name if mentioned",
  "location": "City, State or location",
  "remote": true/false,
  "salary_min": null or integer (annual salary in USD, e.g., 150000),
  "salary_max": null or integer (annual salary in USD, e.g., 200000),
  "job_url": "URL if provided",
  "requirements": [
    "List of required qualifications",
    "Each as a separate string"
  ],
  "bonus_skills": [
    "Nice-to-have skills",
    "Preferred qualifications"
  ],
  "keywords": [
    "Key industry terms",
    "Domain-specific words",
    "Important themes from the posting"
  ],
  "tech_stack": [
    "Programming languages",
    "Frameworks",
    "Tools mentioned"
  ],
  "company_values": [
    "Company culture keywords",
    "Values emphasized in posting"
  ]
}

Guidelines:
- Extract the company name and job title accurately
- Identify the team/department if explicitly mentioned
- Set remote to true if remote or hybrid work is mentioned
- Parse salary ranges if provided (convert to integers, e.g., "$150k-200k" -> salary_min: 150000, salary_max: 200000)
- Separate required qualifications from nice-to-have/bonus qualifications
- Extract technology stack mentions (languages, frameworks, cloud services, tools)
- Identify company culture keywords and values from the about/culture sections
- Keywords should capture domain-specific terms that indicate what the role is about
- If information is not available, use null for optional fields or empty arrays for lists

Return ONLY the JSON object, no additional text or markdown formatting.`
}

// GetUserPrompt creates the user prompt with the job posting content
func (p *ParserAgent) GetUserPrompt(content string, filePath string) string {
	if strings.HasPrefix(content, "[IMAGE_FILE:") {
		return fmt.Sprintf(`Please read and parse the job posting image at: %s

Extract all the structured information you can identify from the image and return it as JSON.`, filePath)
	}

	return fmt.Sprintf(`Parse the following job posting and extract structured data:

---
%s
---

Return the extracted information as JSON.`, content)
}

// ValidateParsedPosting validates the parsed posting has required fields
func (p *ParserAgent) ValidateParsedPosting(parsed *ParsedPosting) error {
	if parsed.Company == "" {
		return fmt.Errorf("company name is required")
	}
	if parsed.Position == "" {
		return fmt.Errorf("position/job title is required")
	}
	return nil
}

// ParseJSON parses JSON output from the AI into a ParsedPosting struct
func (p *ParserAgent) ParseJSON(jsonStr string) (*ParsedPosting, error) {
	// Clean up the JSON string (remove markdown code blocks if present)
	jsonStr = strings.TrimSpace(jsonStr)
	jsonStr = strings.TrimPrefix(jsonStr, "```json")
	jsonStr = strings.TrimPrefix(jsonStr, "```")
	jsonStr = strings.TrimSuffix(jsonStr, "```")
	jsonStr = strings.TrimSpace(jsonStr)

	var parsed ParsedPosting
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if err := p.ValidateParsedPosting(&parsed); err != nil {
		return nil, err
	}

	return &parsed, nil
}

// GetOutputSchema returns the JSON schema for the parser output
func (p *ParserAgent) GetOutputSchema() string {
	return `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["company", "position"],
  "properties": {
    "company": {
      "type": "string",
      "description": "Company name"
    },
    "position": {
      "type": "string",
      "description": "Job title"
    },
    "team": {
      "type": "string",
      "description": "Team or department name"
    },
    "location": {
      "type": "string",
      "description": "Job location (City, State)"
    },
    "remote": {
      "type": "boolean",
      "description": "Whether remote work is available"
    },
    "salary_min": {
      "type": ["integer", "null"],
      "description": "Minimum salary in USD"
    },
    "salary_max": {
      "type": ["integer", "null"],
      "description": "Maximum salary in USD"
    },
    "job_url": {
      "type": "string",
      "description": "URL to the job posting"
    },
    "requirements": {
      "type": "array",
      "items": {"type": "string"},
      "description": "Required qualifications"
    },
    "bonus_skills": {
      "type": "array",
      "items": {"type": "string"},
      "description": "Nice-to-have skills"
    },
    "keywords": {
      "type": "array",
      "items": {"type": "string"},
      "description": "Key industry terms and themes"
    },
    "tech_stack": {
      "type": "array",
      "items": {"type": "string"},
      "description": "Technologies mentioned"
    },
    "company_values": {
      "type": "array",
      "items": {"type": "string"},
      "description": "Company culture keywords"
    }
  }
}`
}
