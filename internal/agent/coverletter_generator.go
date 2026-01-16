package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CoverLetterGeneratorAgent generates tailored cover letters based on job postings and CV data
type CoverLetterGeneratorAgent struct {
	Config  *AgentConfig
	BaseDir string
}

// CoverLetterOutput represents the generated cover letter paths
type CoverLetterOutput struct {
	TypstPath string `json:"typst_path"`
	PDFPath   string `json:"pdf_path,omitempty"`
	Company   string `json:"company"`
	Position  string `json:"position"`
}

// NewCoverLetterGeneratorAgent creates a new cover letter generator agent instance
func NewCoverLetterGeneratorAgent(config *AgentConfig, baseDir string) *CoverLetterGeneratorAgent {
	return &CoverLetterGeneratorAgent{
		Config:  config,
		BaseDir: baseDir,
	}
}

// LoadCV loads the candidate CV from the specified path
func (c *CoverLetterGeneratorAgent) LoadCV(cvPath string) (*CVData, error) {
	data, err := os.ReadFile(cvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CV file: %w", err)
	}

	var cv CVData
	if err := json.Unmarshal(data, &cv); err != nil {
		return nil, fmt.Errorf("failed to parse CV JSON: %w", err)
	}

	return &cv, nil
}

// LoadResume reads the generated resume for consistency reference
func (c *CoverLetterGeneratorAgent) LoadResume(resumePath string) (string, error) {
	content, err := os.ReadFile(resumePath)
	if err != nil {
		// Resume is optional for cover letter generation
		return "", nil
	}
	return string(content), nil
}

// LoadTemplate reads the base cover letter Typst template
func (c *CoverLetterGeneratorAgent) LoadTemplate(templatePath string) (string, error) {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template: %w", err)
	}
	return string(content), nil
}

// GetSystemPrompt returns the system prompt for cover letter generation
func (c *CoverLetterGeneratorAgent) GetSystemPrompt() string {
	return `You are a cover letter writing specialist. Create compelling, personalized cover letters that connect candidate experience to job requirements.

## Writing Principles

### 1. The Hook (Opening)
- Reference the specific role and company by name
- Show genuine interest in their mission or product
- Connect your background to their needs in one sentence
- Avoid generic openings ("I am writing to apply for...")

### 2. The Bridge (Experience)
- Pick 2-3 most relevant experiences
- Mirror language from the job posting
- Include specific, quantified achievements
- Show how past work prepares you for this role

### 3. The Value (What You Bring)
- Connect to company values from the posting
- Emphasize unique strengths that match their needs
- Show enthusiasm without being over-the-top
- Include a forward-looking statement

### 4. The Close
- Express interest in discussing further
- Thank them for their consideration
- Keep it brief and professional

## Tone Guidelines

- Professional but personable - Not stiff, not overly casual
- Confident but not arrogant - State achievements factually
- Specific over generic - Use concrete examples
- Concise - Each paragraph should be 3-5 sentences
- Active voice - "I built" not "It was built by me"

## Content Rules

1. Never lie or exaggerate - Only include real experience
2. Be specific - "I improved API response time by 40%" not "I improved performance"
3. Connect to requirements - Reference skills from job posting naturally
4. Show, don't tell - Demonstrate skills through examples
5. Keep it to one page - 3 concise paragraphs maximum

## Output Format

Generate a complete Typst file using @preview/modern-cv:0.9.0. Return ONLY the Typst content, no markdown formatting or explanations.

CRITICAL: Only reference real experience from the provided CV. Do not invent projects, metrics, or achievements.`
}

// GetUserPrompt creates the user prompt with job posting, CV, and resume context
func (c *CoverLetterGeneratorAgent) GetUserPrompt(posting *ParsedPosting, cv *CVData, resumeContent string) (string, error) {
	postingJSON, err := json.MarshalIndent(posting, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to serialize posting: %w", err)
	}

	cvJSON, err := json.MarshalIndent(cv, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to serialize CV: %w", err)
	}

	prompt := fmt.Sprintf(`Generate a tailored cover letter for this job posting.

## Job Posting Data

%s

## Candidate CV

%s`, postingJSON, cvJSON)

	// Include resume content if available for consistency
	if resumeContent != "" {
		prompt += fmt.Sprintf(`

## Generated Resume (for consistency reference)

%s`, resumeContent)
	}

	prompt += `

## Instructions

1. Use the candidate's real information from the CV
2. Reference 2-3 most relevant experiences that match job requirements
3. Mirror language and terminology from the job posting
4. Show genuine interest in the company and role
5. Keep it to 3 paragraphs maximum
6. Return ONLY the complete Typst file content`

	return prompt, nil
}

// GenerateOutputPath creates the output file path for the cover letter
func (c *CoverLetterGeneratorAgent) GenerateOutputPath(posting *ParsedPosting, jobType, outputDir string) string {
	company := sanitizeFilename(posting.Company)
	position := sanitizeFilename(posting.Position)

	// Create job-type directory structure
	if jobType == "" {
		jobType = "swe" // default
	}

	folderName := fmt.Sprintf("%s-%s", company, strings.ReplaceAll(position, "-", "_"))
	return filepath.Join(outputDir, jobType, folderName, "cover-letter.typ")
}

// WriteTypst writes the generated Typst content to a file
func (c *CoverLetterGeneratorAgent) WriteTypst(content, outputPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write Typst file: %w", err)
	}

	return nil
}

// CompilePDF compiles the Typst file to PDF using the typst CLI
func (c *CoverLetterGeneratorAgent) CompilePDF(typstPath string) (string, error) {
	pdfPath := strings.TrimSuffix(typstPath, ".typ") + ".pdf"

	cmd := exec.Command("typst", "compile", typstPath, pdfPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("typst compile failed: %w\nOutput: %s", err, string(output))
	}

	return pdfPath, nil
}

// IsTypstAvailable checks if the typst CLI is installed
func (c *CoverLetterGeneratorAgent) IsTypstAvailable() bool {
	_, err := exec.LookPath("typst")
	return err == nil
}

// ExtractRelevantExperiences finds the most relevant work experiences for the job
func (c *CoverLetterGeneratorAgent) ExtractRelevantExperiences(cv *CVData, posting *ParsedPosting, limit int) []CVWork {
	if limit <= 0 {
		limit = 3
	}

	// Build a relevance map based on job requirements and tech stack
	jobKeywords := make(map[string]bool)
	for _, skill := range posting.TechStack {
		jobKeywords[strings.ToLower(skill)] = true
	}
	for _, req := range posting.Requirements {
		words := strings.Fields(strings.ToLower(req))
		for _, word := range words {
			if len(word) > 3 { // Skip short words
				jobKeywords[strings.Trim(word, ".,;:()")] = true
			}
		}
	}

	// Score each work experience by relevance
	type scoredWork struct {
		work  CVWork
		score int
	}

	scored := make([]scoredWork, 0, len(cv.Work))
	for _, work := range cv.Work {
		score := 0
		// Check highlights for keyword matches
		for _, highlight := range work.Highlights {
			highlightLower := strings.ToLower(highlight)
			for keyword := range jobKeywords {
				if strings.Contains(highlightLower, keyword) {
					score++
				}
			}
		}
		// Check position title
		positionLower := strings.ToLower(work.Position)
		for keyword := range jobKeywords {
			if strings.Contains(positionLower, keyword) {
				score += 2 // Weight title matches higher
			}
		}
		scored = append(scored, scoredWork{work: work, score: score})
	}

	// Sort by score (simple bubble sort for small arrays)
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// Return top N experiences
	result := make([]CVWork, 0, limit)
	for i := 0; i < len(scored) && i < limit; i++ {
		result = append(result, scored[i].work)
	}

	return result
}

// ParseTypstOutput validates and cleans the AI-generated Typst content
func (c *CoverLetterGeneratorAgent) ParseTypstOutput(output string) (string, error) {
	// Remove markdown code blocks if present
	output = strings.TrimSpace(output)
	output = strings.TrimPrefix(output, "```typst")
	output = strings.TrimPrefix(output, "```")
	output = strings.TrimSuffix(output, "```")
	output = strings.TrimSpace(output)

	// Basic validation - check for required Typst elements
	if !strings.Contains(output, "#import") {
		return "", fmt.Errorf("invalid Typst output: missing #import statement")
	}
	if !strings.Contains(output, "coverletter") && !strings.Contains(output, "modern-cv") {
		return "", fmt.Errorf("invalid Typst output: missing coverletter template reference")
	}

	return output, nil
}

// Generate runs the full cover letter generation flow
func (c *CoverLetterGeneratorAgent) Generate(posting *ParsedPosting, cvPath, resumePath, outputDir, jobType string) (*CoverLetterOutput, error) {
	// Load CV
	cv, err := c.LoadCV(cvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load CV: %w", err)
	}

	// Load resume for consistency (optional)
	resumeContent, _ := c.LoadResume(resumePath)

	// Generate output path
	outputPath := c.GenerateOutputPath(posting, jobType, outputDir)

	// The actual Typst content generation would be done by an AI model
	// This method prepares all the inputs and returns the output structure
	result := &CoverLetterOutput{
		TypstPath: outputPath,
		Company:   posting.Company,
		Position:  posting.Position,
	}

	// Store context for AI generation
	_, err = c.GetUserPrompt(posting, cv, resumeContent)
	if err != nil {
		return nil, fmt.Errorf("failed to generate prompt: %w", err)
	}

	return result, nil
}
