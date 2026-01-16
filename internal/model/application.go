package model

import (
	"time"
)

// Status constants for the application pipeline
const (
	StatusSaved     = "saved"
	StatusApplied   = "applied"
	StatusScreening = "screening"
	StatusInterview = "interview"
	StatusOffer     = "offer"
	StatusAccepted  = "accepted"
	StatusRejected  = "rejected"
	StatusWithdrawn = "withdrawn"
)

// AllStatuses returns all valid status values in pipeline order
func AllStatuses() []string {
	return []string{
		StatusSaved,
		StatusApplied,
		StatusScreening,
		StatusInterview,
		StatusOffer,
		StatusAccepted,
		StatusRejected,
		StatusWithdrawn,
	}
}

// StatusLabel returns a human-readable label for a status
func StatusLabel(status string) string {
	labels := map[string]string{
		StatusSaved:     "Saved",
		StatusApplied:   "Applied",
		StatusScreening: "Screening",
		StatusInterview: "Interview",
		StatusOffer:     "Offer",
		StatusAccepted:  "Accepted",
		StatusRejected:  "Rejected",
		StatusWithdrawn: "Withdrawn",
	}
	if label, ok := labels[status]; ok {
		return label
	}
	return status
}

// StatusPriority returns sort priority for a status (higher = more important)
func StatusPriority(status string) int {
	priorities := map[string]int{
		StatusAccepted:  8,
		StatusOffer:     7,
		StatusInterview: 6,
		StatusScreening: 5,
		StatusApplied:   4,
		StatusSaved:     3,
		StatusRejected:  2,
		StatusWithdrawn: 1,
	}
	return priorities[status]
}

// Interview represents a scheduled or completed interview
type Interview struct {
	Date     time.Time `json:"date"`
	Type     string    `json:"type"` // phone, video, onsite, technical
	Notes    string    `json:"notes,omitempty"`
	WithWhom string    `json:"with_whom,omitempty"`
}

// Application represents a job application
type Application struct {
	ID          string    `json:"id"`
	Company     string    `json:"company"`
	Position    string    `json:"position"`
	Status      string    `json:"status"`
	DateApplied *time.Time `json:"date_applied,omitempty"`
	Notes       string    `json:"notes,omitempty"`

	// Extended info
	SalaryMin int    `json:"salary_min,omitempty"`
	SalaryMax int    `json:"salary_max,omitempty"`
	JobURL    string `json:"job_url,omitempty"`
	Location  string `json:"location,omitempty"`
	Remote    bool   `json:"remote,omitempty"`

	// Contact & Interviews
	ContactName  string      `json:"contact_name,omitempty"`
	ContactEmail string      `json:"contact_email,omitempty"`
	Interviews   []Interview `json:"interviews,omitempty"`

	// Documents
	ResumeVersion string `json:"resume_version,omitempty"`
	CoverLetter   string `json:"cover_letter,omitempty"`
	DocumentsDir  string `json:"documents_dir,omitempty"` // Path to application documents folder

	// Follow-up
	NextFollowUp *time.Time `json:"next_follow_up,omitempty"`

	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SalaryRange returns formatted salary range or empty string
func (a *Application) SalaryRange() string {
	if a.SalaryMin == 0 && a.SalaryMax == 0 {
		return ""
	}
	if a.SalaryMin > 0 && a.SalaryMax > 0 {
		return formatSalary(a.SalaryMin) + " - " + formatSalary(a.SalaryMax)
	}
	if a.SalaryMin > 0 {
		return formatSalary(a.SalaryMin) + "+"
	}
	return "Up to " + formatSalary(a.SalaryMax)
}

func formatSalary(amount int) string {
	if amount >= 1000 {
		return "$" + formatNumber(amount/1000) + "k"
	}
	return "$" + formatNumber(amount)
}

func formatNumber(n int) string {
	s := ""
	for n > 0 {
		if s != "" {
			s = "," + s
		}
		if n < 1000 {
			s = string(rune('0'+n%10)) + s
			n /= 10
			for n > 0 {
				s = string(rune('0'+n%10)) + s
				n /= 10
			}
		} else {
			// Add last 3 digits
			for i := 0; i < 3; i++ {
				s = string(rune('0'+n%10)) + s
				n /= 10
			}
		}
	}
	if s == "" {
		return "0"
	}
	return s
}
