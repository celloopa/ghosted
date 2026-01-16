package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ReviewerAgent reviews generated documents from a hiring manager perspective
type ReviewerAgent struct {
	Config  *AgentConfig
	BaseDir string
}

// DetailedReviewResult holds comprehensive review feedback
type DetailedReviewResult struct {
	Approved      bool          `json:"approved"`
	OverallScore  int           `json:"overall_score"`
	ResumeReview  DocumentReview `json:"resume_review"`
	CoverReview   DocumentReview `json:"cover_letter_review"`
	MatchAnalysis MatchAnalysis  `json:"match_analysis"`
	Recommendation string        `json:"recommendation"`
}

// DocumentReview holds the review for a single document
type DocumentReview struct {
	Score       int      `json:"score"`
	Strengths   []string `json:"strengths"`
	Weaknesses  []string `json:"weaknesses"`
	Suggestions []string `json:"suggestions"`
}

// MatchAnalysis analyzes how well the candidate matches the job requirements
type MatchAnalysis struct {
	RequirementsMet     []string `json:"requirements_met"`
	RequirementsMissing []string `json:"requirements_missing"`
	BonusPointsHit      []string `json:"bonus_points_hit"`
}

// ApprovalThreshold is the minimum score required for approval
const ApprovalThreshold = 70

// NewReviewerAgent creates a new reviewer agent instance
func NewReviewerAgent(config *AgentConfig, baseDir string) *ReviewerAgent {
	return &ReviewerAgent{
		Config:  config,
		BaseDir: baseDir,
	}
}

// LoadDocument reads a document file (resume or cover letter)
func (r *ReviewerAgent) LoadDocument(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read document: %w", err)
	}
	return string(content), nil
}

// GetSystemPrompt returns the system prompt for the reviewer agent
func (r *ReviewerAgent) GetSystemPrompt() string {
	return `You are a hiring manager reviewing job applications. Evaluate the resume and cover letter honestly and constructively, providing actionable feedback.

## Scoring Criteria

### Requirements Match (40% of score)
- How many required qualifications does the candidate meet?
- Are must-have skills clearly demonstrated?
- Is experience level appropriate for the role?

### Experience Relevance (30% of score)
- How closely does past experience match the role?
- Are achievements in relevant domains?
- Is there evidence of growth and impact?

### Communication Quality (20% of score)
- Is writing clear and professional?
- Are bullet points concise and impactful?
- Is the cover letter engaging and specific?

### Cultural Fit Signals (10% of score)
- Does tone match company culture?
- Are company values addressed?
- Is genuine interest demonstrated?

## Resume Review Checklist

- Skills match job requirements
- Experience relevant to role
- Achievements are quantified
- Keywords from posting are present
- Format is clean and readable
- No typos or errors

Red Flags:
- Missing critical required skills
- Experience gap unexplained
- Generic bullet points without impact
- Mismatched seniority level

## Cover Letter Review Checklist

- Mentions company and role specifically
- Connects experience to requirements
- Shows genuine interest
- Professional but personable tone
- Concise (not exceeding one page worth)
- No typos or errors

Red Flags:
- Generic template language
- Doesn't mention the company by name
- Repeats resume verbatim
- Overly long or rambling

## Approval Thresholds

| Score | Decision |
|-------|----------|
| 80-100 | Approve - Ready to submit |
| 60-79 | Conditional - Needs minor edits |
| 40-59 | Revise - Significant improvements needed |
| 0-39 | Reject - Not suitable for this role |

## Output Format

Return a JSON object with your evaluation. Be specific in feedback:
- Strengths: "Strong React experience with 3+ years matches requirement"
- Weaknesses: "Missing Kubernetes experience listed as required"
- Suggestions: "Add the e-commerce project from 2023 to demonstrate payment integration experience"

Return ONLY the JSON object, no additional text or markdown formatting.`
}

// GetUserPrompt creates the user prompt with documents and posting data
func (r *ReviewerAgent) GetUserPrompt(posting *ParsedPosting, resumeContent, coverLetterContent string, cv *CVData) (string, error) {
	postingJSON, err := json.MarshalIndent(posting, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to serialize posting: %w", err)
	}

	prompt := fmt.Sprintf(`Review the following job application materials.

## Job Posting

%s

## Generated Resume

%s

## Generated Cover Letter

%s`, postingJSON, resumeContent, coverLetterContent)

	// Include CV for verification if available
	if cv != nil {
		cvJSON, _ := json.MarshalIndent(cv, "", "  ")
		prompt += fmt.Sprintf(`

## Candidate CV (for verification)

%s`, cvJSON)
	}

	prompt += `

## Instructions

1. Score the resume and cover letter separately (0-100)
2. Identify specific strengths and weaknesses
3. Check that claims match the CV data
4. Provide actionable suggestions for improvement
5. Calculate overall score and make approval recommendation
6. Return ONLY the JSON object`

	return prompt, nil
}

// CalculateOverallScore computes weighted average of component scores
func (r *ReviewerAgent) CalculateOverallScore(resumeScore, coverLetterScore int) int {
	// Resume weighted slightly higher (60%) than cover letter (40%)
	return (resumeScore*60 + coverLetterScore*40) / 100
}

// DetermineRecommendation returns recommendation based on score
func (r *ReviewerAgent) DetermineRecommendation(score int) string {
	switch {
	case score >= 80:
		return "Approve"
	case score >= 70:
		return "Approve with minor edits"
	case score >= 60:
		return "Conditional - Needs minor edits"
	case score >= 40:
		return "Revise and resubmit"
	default:
		return "Not recommended"
	}
}

// IsApproved checks if the score meets the approval threshold
func (r *ReviewerAgent) IsApproved(score int) bool {
	return score >= ApprovalThreshold
}

// ParseReviewOutput parses and validates AI-generated review JSON
func (r *ReviewerAgent) ParseReviewOutput(output string) (*DetailedReviewResult, error) {
	// Clean up the JSON string
	output = strings.TrimSpace(output)
	output = strings.TrimPrefix(output, "```json")
	output = strings.TrimPrefix(output, "```")
	output = strings.TrimSuffix(output, "```")
	output = strings.TrimSpace(output)

	var result DetailedReviewResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, fmt.Errorf("failed to parse review JSON: %w", err)
	}

	// Validate score ranges
	if result.OverallScore < 0 || result.OverallScore > 100 {
		return nil, fmt.Errorf("overall score out of range: %d", result.OverallScore)
	}
	if result.ResumeReview.Score < 0 || result.ResumeReview.Score > 100 {
		return nil, fmt.Errorf("resume score out of range: %d", result.ResumeReview.Score)
	}
	if result.CoverReview.Score < 0 || result.CoverReview.Score > 100 {
		return nil, fmt.Errorf("cover letter score out of range: %d", result.CoverReview.Score)
	}

	// Set approval based on score if not set
	result.Approved = r.IsApproved(result.OverallScore)

	// Set recommendation if empty
	if result.Recommendation == "" {
		result.Recommendation = r.DetermineRecommendation(result.OverallScore)
	}

	return &result, nil
}

// AnalyzeRequirementMatch checks which requirements from posting are met by CV
func (r *ReviewerAgent) AnalyzeRequirementMatch(cv *CVData, posting *ParsedPosting) *MatchAnalysis {
	analysis := &MatchAnalysis{
		RequirementsMet:     []string{},
		RequirementsMissing: []string{},
		BonusPointsHit:      []string{},
	}

	// Build set of candidate skills from CV
	candidateSkills := make(map[string]bool)
	for _, skillGroup := range cv.Skills {
		for _, keyword := range skillGroup.Keywords {
			candidateSkills[strings.ToLower(keyword)] = true
		}
	}

	// Also extract skills mentioned in work highlights
	for _, work := range cv.Work {
		for _, highlight := range work.Highlights {
			// Extract potential technology keywords (simplified)
			words := strings.Fields(strings.ToLower(highlight))
			for _, word := range words {
				word = strings.Trim(word, ".,;:()")
				if len(word) > 2 {
					candidateSkills[word] = true
				}
			}
		}
	}

	// Check tech stack requirements
	for _, tech := range posting.TechStack {
		techLower := strings.ToLower(tech)
		if candidateSkills[techLower] {
			analysis.RequirementsMet = append(analysis.RequirementsMet, tech)
		} else {
			analysis.RequirementsMissing = append(analysis.RequirementsMissing, tech)
		}
	}

	// Check bonus skills
	for _, skill := range posting.BonusSkills {
		skillLower := strings.ToLower(skill)
		// Check if any candidate skill contains the bonus skill
		for cs := range candidateSkills {
			if strings.Contains(cs, skillLower) || strings.Contains(skillLower, cs) {
				analysis.BonusPointsHit = append(analysis.BonusPointsHit, skill)
				break
			}
		}
	}

	return analysis
}

// ConvertToSimpleReview converts DetailedReviewResult to the simpler ReviewResult
func (r *ReviewerAgent) ConvertToSimpleReview(detailed *DetailedReviewResult) *ReviewResult {
	// Combine feedback from both documents
	feedback := append(detailed.ResumeReview.Strengths, detailed.CoverReview.Strengths...)
	feedback = append(feedback, detailed.ResumeReview.Suggestions...)
	feedback = append(feedback, detailed.CoverReview.Suggestions...)

	// Combine issues
	issues := append(detailed.ResumeReview.Weaknesses, detailed.CoverReview.Weaknesses...)

	return &ReviewResult{
		Approved: detailed.Approved,
		Score:    detailed.OverallScore / 10, // Convert 0-100 to 1-10 scale
		Feedback: feedback,
		Issues:   issues,
	}
}

// Review runs the full document review flow
func (r *ReviewerAgent) Review(posting *ParsedPosting, resumePath, coverLetterPath, cvPath string) (*DetailedReviewResult, error) {
	// Load documents
	_, err := r.LoadDocument(resumePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load resume: %w", err)
	}

	_, err = r.LoadDocument(coverLetterPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load cover letter: %w", err)
	}

	// Load CV for verification (optional)
	var cv *CVData
	if cvPath != "" {
		cvData, err := os.ReadFile(cvPath)
		if err == nil {
			json.Unmarshal(cvData, &cv)
		}
	}

	// Pre-analyze requirements match
	analysis := &MatchAnalysis{}
	if cv != nil {
		analysis = r.AnalyzeRequirementMatch(cv, posting)
	}

	// The actual review would be done by an AI model
	// This method prepares inputs and returns a placeholder result
	result := &DetailedReviewResult{
		Approved:       true, // Placeholder
		OverallScore:   80,   // Placeholder
		ResumeReview:   DocumentReview{Score: 80},
		CoverReview:    DocumentReview{Score: 80},
		MatchAnalysis:  *analysis,
		Recommendation: "Approve with minor edits",
	}

	return result, nil
}
