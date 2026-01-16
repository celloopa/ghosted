package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReviewerAgent_LoadDocument(t *testing.T) {
	agent := NewReviewerAgent(&AgentConfig{Type: AgentReviewer}, "")

	tmpDir := t.TempDir()
	docPath := filepath.Join(tmpDir, "resume.typ")
	content := "#import modern-cv..."

	if err := os.WriteFile(docPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test document: %v", err)
	}

	result, err := agent.LoadDocument(docPath)
	if err != nil {
		t.Errorf("LoadDocument() error = %v", err)
	}
	if result != content {
		t.Errorf("LoadDocument() content mismatch")
	}
}

func TestReviewerAgent_LoadDocument_Missing(t *testing.T) {
	agent := NewReviewerAgent(&AgentConfig{Type: AgentReviewer}, "")

	_, err := agent.LoadDocument("/nonexistent/path/document.typ")
	if err == nil {
		t.Error("LoadDocument() expected error for missing file")
	}
}

func TestReviewerAgent_CalculateOverallScore(t *testing.T) {
	agent := NewReviewerAgent(&AgentConfig{Type: AgentReviewer}, "")

	tests := []struct {
		name             string
		resumeScore      int
		coverLetterScore int
		expected         int
	}{
		{"both 100", 100, 100, 100},
		{"both 50", 50, 50, 50},
		{"resume 100, cover 0", 100, 0, 60},  // 100*0.6 + 0*0.4
		{"resume 0, cover 100", 0, 100, 40},  // 0*0.6 + 100*0.4
		{"resume 80, cover 70", 80, 70, 76},  // 80*0.6 + 70*0.4
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := agent.CalculateOverallScore(tt.resumeScore, tt.coverLetterScore)
			if got != tt.expected {
				t.Errorf("CalculateOverallScore(%d, %d) = %d, want %d",
					tt.resumeScore, tt.coverLetterScore, got, tt.expected)
			}
		})
	}
}

func TestReviewerAgent_DetermineRecommendation(t *testing.T) {
	agent := NewReviewerAgent(&AgentConfig{Type: AgentReviewer}, "")

	tests := []struct {
		score    int
		expected string
	}{
		{100, "Approve"},
		{85, "Approve"},
		{80, "Approve"},
		{75, "Approve with minor edits"},
		{70, "Approve with minor edits"},
		{65, "Conditional - Needs minor edits"},
		{60, "Conditional - Needs minor edits"},
		{50, "Revise and resubmit"},
		{40, "Revise and resubmit"},
		{30, "Not recommended"},
		{0, "Not recommended"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := agent.DetermineRecommendation(tt.score)
			if got != tt.expected {
				t.Errorf("DetermineRecommendation(%d) = %q, want %q", tt.score, got, tt.expected)
			}
		})
	}
}

func TestReviewerAgent_IsApproved(t *testing.T) {
	agent := NewReviewerAgent(&AgentConfig{Type: AgentReviewer}, "")

	tests := []struct {
		score    int
		expected bool
	}{
		{100, true},
		{70, true},
		{69, false},
		{50, false},
		{0, false},
	}

	for _, tt := range tests {
		got := agent.IsApproved(tt.score)
		if got != tt.expected {
			t.Errorf("IsApproved(%d) = %v, want %v", tt.score, got, tt.expected)
		}
	}
}

func TestReviewerAgent_ParseReviewOutput(t *testing.T) {
	agent := NewReviewerAgent(&AgentConfig{Type: AgentReviewer}, "")

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "valid review JSON",
			input: `{
				"approved": true,
				"overall_score": 85,
				"resume_review": {
					"score": 88,
					"strengths": ["Strong technical match"],
					"weaknesses": [],
					"suggestions": []
				},
				"cover_letter_review": {
					"score": 82,
					"strengths": ["Personalized opening"],
					"weaknesses": [],
					"suggestions": []
				},
				"match_analysis": {
					"requirements_met": ["React"],
					"requirements_missing": [],
					"bonus_points_hit": []
				},
				"recommendation": "Approve"
			}`,
			wantErr: false,
		},
		{
			name: "with markdown code block",
			input: "```json\n{\"approved\":true,\"overall_score\":75,\"resume_review\":{\"score\":75},\"cover_letter_review\":{\"score\":75}}\n```",
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   `{invalid}`,
			wantErr: true,
		},
		{
			name:    "overall score out of range",
			input:   `{"approved":true,"overall_score":150,"resume_review":{"score":80},"cover_letter_review":{"score":80}}`,
			wantErr: true,
		},
		{
			name:    "resume score out of range",
			input:   `{"approved":true,"overall_score":80,"resume_review":{"score":-10},"cover_letter_review":{"score":80}}`,
			wantErr: true,
		},
		{
			name:    "cover letter score out of range",
			input:   `{"approved":true,"overall_score":80,"resume_review":{"score":80},"cover_letter_review":{"score":101}}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := agent.ParseReviewOutput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseReviewOutput() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result == nil {
				t.Error("ParseReviewOutput() returned nil for valid input")
			}
		})
	}
}

func TestReviewerAgent_ParseReviewOutput_SetsApproval(t *testing.T) {
	agent := NewReviewerAgent(&AgentConfig{Type: AgentReviewer}, "")

	// Score of 85 should result in approved=true
	input := `{"approved":false,"overall_score":85,"resume_review":{"score":85},"cover_letter_review":{"score":85}}`
	result, err := agent.ParseReviewOutput(input)
	if err != nil {
		t.Fatalf("ParseReviewOutput() error = %v", err)
	}

	if !result.Approved {
		t.Error("ParseReviewOutput() should set Approved=true for score >= 70")
	}
}

func TestReviewerAgent_ParseReviewOutput_SetsRecommendation(t *testing.T) {
	agent := NewReviewerAgent(&AgentConfig{Type: AgentReviewer}, "")

	// Empty recommendation should be auto-set based on score
	input := `{"approved":true,"overall_score":85,"resume_review":{"score":85},"cover_letter_review":{"score":85}}`
	result, err := agent.ParseReviewOutput(input)
	if err != nil {
		t.Fatalf("ParseReviewOutput() error = %v", err)
	}

	if result.Recommendation != "Approve" {
		t.Errorf("Recommendation = %q, want %q", result.Recommendation, "Approve")
	}
}

func TestReviewerAgent_AnalyzeRequirementMatch(t *testing.T) {
	agent := NewReviewerAgent(&AgentConfig{Type: AgentReviewer}, "")

	cv := &CVData{
		Skills: []CVSkill{
			{
				Name:     "Languages",
				Keywords: []string{"Go", "TypeScript", "Python"},
			},
			{
				Name:     "Frameworks",
				Keywords: []string{"React", "Vue"},
			},
		},
		Work: []CVWork{
			{
				Highlights: []string{"Built microservices with Kubernetes"},
			},
		},
	}

	posting := &ParsedPosting{
		TechStack:   []string{"Go", "React", "AWS"},
		BonusSkills: []string{"Python", "Kubernetes"},
	}

	analysis := agent.AnalyzeRequirementMatch(cv, posting)

	// Go and React should be met
	expectedMet := map[string]bool{"Go": true, "React": true}
	for _, skill := range analysis.RequirementsMet {
		if !expectedMet[skill] {
			t.Errorf("Unexpected requirement met: %s", skill)
		}
	}

	// AWS should be missing
	foundAWS := false
	for _, skill := range analysis.RequirementsMissing {
		if skill == "AWS" {
			foundAWS = true
			break
		}
	}
	if !foundAWS {
		t.Error("AWS should be in RequirementsMissing")
	}

	// Python should be in bonus points hit
	foundPython := false
	for _, skill := range analysis.BonusPointsHit {
		if skill == "Python" {
			foundPython = true
			break
		}
	}
	if !foundPython {
		t.Error("Python should be in BonusPointsHit")
	}
}

func TestReviewerAgent_ConvertToSimpleReview(t *testing.T) {
	agent := NewReviewerAgent(&AgentConfig{Type: AgentReviewer}, "")

	detailed := &DetailedReviewResult{
		Approved:     true,
		OverallScore: 85,
		ResumeReview: DocumentReview{
			Score:       88,
			Strengths:   []string{"Good skills match"},
			Weaknesses:  []string{"Missing AWS"},
			Suggestions: []string{"Add cloud projects"},
		},
		CoverReview: DocumentReview{
			Score:       82,
			Strengths:   []string{"Personal touch"},
			Weaknesses:  []string{"Too long"},
			Suggestions: []string{"Be more concise"},
		},
	}

	simple := agent.ConvertToSimpleReview(detailed)

	if !simple.Approved {
		t.Error("Simple review should be approved")
	}
	if simple.Score != 8 { // 85/10 = 8
		t.Errorf("Score = %d, want 8", simple.Score)
	}
	if len(simple.Feedback) == 0 {
		t.Error("Feedback should not be empty")
	}
	if len(simple.Issues) != 2 { // 1 weakness from each doc
		t.Errorf("Issues count = %d, want 2", len(simple.Issues))
	}
}

func TestReviewerAgent_GetSystemPrompt(t *testing.T) {
	agent := NewReviewerAgent(&AgentConfig{Type: AgentReviewer}, "")
	prompt := agent.GetSystemPrompt()

	if len(prompt) == 0 {
		t.Error("GetSystemPrompt() returned empty string")
	}

	requiredElements := []string{
		"hiring manager",
		"Requirements Match",
		"Experience Relevance",
		"Communication Quality",
		"Approval",
	}

	for _, elem := range requiredElements {
		if !containsIgnoreCase(prompt, elem) {
			t.Errorf("GetSystemPrompt() missing element: %s", elem)
		}
	}
}

func TestReviewerAgent_GetUserPrompt(t *testing.T) {
	agent := NewReviewerAgent(&AgentConfig{Type: AgentReviewer}, "")

	posting := &ParsedPosting{
		Company:  "TechCorp",
		Position: "Engineer",
	}

	resumeContent := "Resume content here..."
	coverContent := "Cover letter here..."

	prompt, err := agent.GetUserPrompt(posting, resumeContent, coverContent, nil)
	if err != nil {
		t.Errorf("GetUserPrompt() error = %v", err)
	}

	if !containsIgnoreCase(prompt, "TechCorp") {
		t.Error("GetUserPrompt() missing company name")
	}
	if !containsIgnoreCase(prompt, "Resume content") {
		t.Error("GetUserPrompt() missing resume content")
	}
	if !containsIgnoreCase(prompt, "Cover letter") {
		t.Error("GetUserPrompt() missing cover letter content")
	}
}

func TestReviewerAgent_GetUserPrompt_WithCV(t *testing.T) {
	agent := NewReviewerAgent(&AgentConfig{Type: AgentReviewer}, "")

	posting := &ParsedPosting{Company: "Test", Position: "Dev"}
	cv := &CVData{Basics: CVBasics{Name: "Jane Doe"}}

	prompt, err := agent.GetUserPrompt(posting, "resume", "cover", cv)
	if err != nil {
		t.Errorf("GetUserPrompt() error = %v", err)
	}

	if !containsIgnoreCase(prompt, "Jane Doe") {
		t.Error("GetUserPrompt() missing CV data when CV provided")
	}
	if !containsIgnoreCase(prompt, "verification") {
		t.Error("GetUserPrompt() should mention verification when CV provided")
	}
}

func TestApprovalThreshold(t *testing.T) {
	if ApprovalThreshold != 70 {
		t.Errorf("ApprovalThreshold = %d, want 70", ApprovalThreshold)
	}
}
