package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ResumeGeneratorAgent generates tailored resumes based on job postings and CV data
type ResumeGeneratorAgent struct {
	Config  *AgentConfig
	BaseDir string
}

// CVData represents the candidate's CV in JSON Resume format
type CVData struct {
	Basics    CVBasics     `json:"basics"`
	Work      []CVWork     `json:"work"`
	Education []CVEducation `json:"education"`
	Skills    []CVSkill    `json:"skills"`
	Languages []CVLanguage `json:"languages,omitempty"`
	Projects  []CVProject  `json:"projects,omitempty"`
}

// CVBasics holds basic contact information
type CVBasics struct {
	Name     string      `json:"name"`
	Label    string      `json:"label"`
	Email    string      `json:"email"`
	Phone    string      `json:"phone"`
	URL      string      `json:"url"`
	Summary  string      `json:"summary"`
	Location CVLocation  `json:"location"`
	Profiles []CVProfile `json:"profiles"`
}

// CVLocation represents address information
type CVLocation struct {
	City        string `json:"city"`
	Region      string `json:"region"`
	CountryCode string `json:"countryCode"`
}

// CVProfile represents a social/professional profile
type CVProfile struct {
	Network  string `json:"network"`
	Username string `json:"username"`
	URL      string `json:"url"`
}

// CVWork represents a work experience entry
type CVWork struct {
	Name       string   `json:"name"`
	Position   string   `json:"position"`
	URL        string   `json:"url"`
	StartDate  string   `json:"startDate"`
	EndDate    string   `json:"endDate,omitempty"`
	Summary    string   `json:"summary"`
	Highlights []string `json:"highlights"`
	Location   string   `json:"location,omitempty"`
}

// CVEducation represents an education entry
type CVEducation struct {
	Institution string   `json:"institution"`
	URL         string   `json:"url"`
	Area        string   `json:"area"`
	StudyType   string   `json:"studyType"`
	StartDate   string   `json:"startDate"`
	EndDate     string   `json:"endDate"`
	Score       string   `json:"score,omitempty"`
	Courses     []string `json:"courses,omitempty"`
}

// CVSkill represents a skill category
type CVSkill struct {
	Name     string   `json:"name"`
	Level    string   `json:"level,omitempty"`
	Keywords []string `json:"keywords"`
}

// CVLanguage represents a spoken language
type CVLanguage struct {
	Language string `json:"language"`
	Fluency  string `json:"fluency"`
}

// CVProject represents a project entry
type CVProject struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Highlights  []string `json:"highlights"`
	Keywords    []string `json:"keywords"`
	StartDate   string   `json:"startDate"`
	EndDate     string   `json:"endDate,omitempty"`
	URL         string   `json:"url,omitempty"`
}

// ResumeOutput represents the generated resume paths
type ResumeOutput struct {
	TypstPath string `json:"typst_path"`
	PDFPath   string `json:"pdf_path,omitempty"`
	Company   string `json:"company"`
	Position  string `json:"position"`
}

// NewResumeGeneratorAgent creates a new resume generator agent instance
func NewResumeGeneratorAgent(config *AgentConfig, baseDir string) *ResumeGeneratorAgent {
	return &ResumeGeneratorAgent{
		Config:  config,
		BaseDir: baseDir,
	}
}

// LoadCV loads the candidate CV from the specified path
func (r *ResumeGeneratorAgent) LoadCV(cvPath string) (*CVData, error) {
	data, err := os.ReadFile(cvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CV file: %w", err)
	}

	var cv CVData
	if err := json.Unmarshal(data, &cv); err != nil {
		return nil, fmt.Errorf("failed to parse CV JSON: %w", err)
	}

	if err := r.ValidateCV(&cv); err != nil {
		return nil, err
	}

	return &cv, nil
}

// ValidateCV checks that the CV has required fields
func (r *ResumeGeneratorAgent) ValidateCV(cv *CVData) error {
	if cv.Basics.Name == "" {
		return fmt.Errorf("CV missing required field: basics.name")
	}
	if cv.Basics.Email == "" {
		return fmt.Errorf("CV missing required field: basics.email")
	}
	return nil
}

// LoadTemplate reads the base resume Typst template
func (r *ResumeGeneratorAgent) LoadTemplate(templatePath string) (string, error) {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template: %w", err)
	}
	return string(content), nil
}

// GetSystemPrompt returns the system prompt for resume generation
func (r *ResumeGeneratorAgent) GetSystemPrompt() string {
	return `You are a resume tailoring specialist. Given a job posting and candidate CV data, create a targeted resume in Typst format.

## Tailoring Principles

### 1. Mirror Job Posting Language
- Use exact terminology from the job posting
- Match their tech stack naming (e.g., "ReactJS" vs "React")
- Incorporate their keywords naturally into bullet points

### 2. Prioritize Relevant Experience
- Lead with experiences most relevant to the role
- Highlight projects using their required tech stack
- Emphasize transferable skills that match requirements

### 3. Quantify Achievements
- Use numbers: "Improved performance by 40%"
- Use scale: "Served 10,000+ daily users"
- Use outcomes: "Reduced bug reports by 60%"

### 4. Skills Section Strategy
- Use strong() to bold skills matching job requirements
- Group skills by category (Languages, Frameworks, Tools, etc.)
- Lead each category with the most relevant skills

### 5. Keyword Optimization
- Include keywords from job posting in the metadata
- Naturally weave tech stack items into experience bullets
- Match their terminology for ATS optimization

## Output Format

Generate a complete Typst file using @preview/modern-cv:0.9.0. Return ONLY the Typst content, no markdown formatting or explanations.

CRITICAL: Only use experience and skills that exist in the provided CV. Do not invent achievements, metrics, or skills the candidate doesn't have.`
}

// GetUserPrompt creates the user prompt with job posting, CV, and template
func (r *ResumeGeneratorAgent) GetUserPrompt(posting *ParsedPosting, cv *CVData, template string) (string, error) {
	postingJSON, err := json.MarshalIndent(posting, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to serialize posting: %w", err)
	}

	cvJSON, err := json.MarshalIndent(cv, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to serialize CV: %w", err)
	}

	return fmt.Sprintf(`Generate a tailored resume for this job posting.

## Job Posting Data

%s

## Candidate CV

%s

## Base Template Structure (for reference)

%s

## Instructions

1. Use the candidate's real information from the CV
2. Tailor the content to match the job requirements
3. Prioritize relevant experience and skills
4. Include keywords from the job posting
5. Return ONLY the complete Typst file content`, postingJSON, cvJSON, template), nil
}

// GenerateOutputPath creates the output file path for the resume
func (r *ResumeGeneratorAgent) GenerateOutputPath(posting *ParsedPosting, jobType, outputDir string) string {
	company := sanitizeFilename(posting.Company)
	position := sanitizeFilename(posting.Position)

	// Create job-type directory structure
	if jobType == "" {
		jobType = "swe" // default
	}

	folderName := fmt.Sprintf("%s-%s", company, strings.ReplaceAll(position, "-", "_"))
	return filepath.Join(outputDir, jobType, folderName, "resume.typ")
}

// WriteTypst writes the generated Typst content to a file
func (r *ResumeGeneratorAgent) WriteTypst(content, outputPath string) error {
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
func (r *ResumeGeneratorAgent) CompilePDF(typstPath string) (string, error) {
	pdfPath := strings.TrimSuffix(typstPath, ".typ") + ".pdf"

	cmd := exec.Command("typst", "compile", typstPath, pdfPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("typst compile failed: %w\nOutput: %s", err, string(output))
	}

	return pdfPath, nil
}

// IsTypstAvailable checks if the typst CLI is installed
func (r *ResumeGeneratorAgent) IsTypstAvailable() bool {
	_, err := exec.LookPath("typst")
	return err == nil
}

// ExtractMatchingSkills finds skills from CV that match job requirements
func (r *ResumeGeneratorAgent) ExtractMatchingSkills(cv *CVData, posting *ParsedPosting) []string {
	jobSkills := make(map[string]bool)

	// Normalize and collect job posting skills
	for _, skill := range posting.TechStack {
		jobSkills[strings.ToLower(skill)] = true
	}
	for _, req := range posting.Requirements {
		// Extract potential skill keywords from requirements
		words := strings.Fields(req)
		for _, word := range words {
			jobSkills[strings.ToLower(strings.Trim(word, ".,;:()"))] = true
		}
	}

	// Find matching skills from CV
	var matches []string
	for _, skillGroup := range cv.Skills {
		for _, keyword := range skillGroup.Keywords {
			if jobSkills[strings.ToLower(keyword)] {
				matches = append(matches, keyword)
			}
		}
	}

	return matches
}

// ParseTypstOutput validates and cleans the AI-generated Typst content
func (r *ResumeGeneratorAgent) ParseTypstOutput(output string) (string, error) {
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
	if !strings.Contains(output, "resume.with") && !strings.Contains(output, "modern-cv") {
		return "", fmt.Errorf("invalid Typst output: missing resume template reference")
	}

	return output, nil
}

// Generate runs the full resume generation flow
func (r *ResumeGeneratorAgent) Generate(posting *ParsedPosting, cvPath, templatePath, outputDir, jobType string) (*ResumeOutput, error) {
	// Load CV
	cv, err := r.LoadCV(cvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load CV: %w", err)
	}

	// Load template (optional, for reference)
	template := ""
	if templatePath != "" {
		template, _ = r.LoadTemplate(templatePath) // Ignore error, template is optional
	}

	// Generate output path
	outputPath := r.GenerateOutputPath(posting, jobType, outputDir)

	// The actual Typst content generation would be done by an AI model
	// This method prepares all the inputs and returns the output structure
	result := &ResumeOutput{
		TypstPath: outputPath,
		Company:   posting.Company,
		Position:  posting.Position,
	}

	// Store context for AI generation
	_, err = r.GetUserPrompt(posting, cv, template)
	if err != nil {
		return nil, fmt.Errorf("failed to generate prompt: %w", err)
	}

	return result, nil
}
