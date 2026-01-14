package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/celloopa/ghosted/internal/model"

	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("application not found")
)

// Store manages job applications in a JSON file
type Store struct {
	filepath     string
	applications []model.Application
}

// New creates a new Store with the given file path
func New(path string) (*Store, error) {
	s := &Store{filepath: path}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Load existing data or create empty file
	if err := s.load(); err != nil {
		if os.IsNotExist(err) {
			s.applications = sampleData()
			return s, s.save()
		}
		return nil, err
	}

	// Seed with sample data if empty
	if len(s.applications) == 0 {
		s.applications = sampleData()
		return s, s.save()
	}

	return s, nil
}

// sampleData returns pre-seeded sample applications for new users
func sampleData() []model.Application {
	now := time.Now()
	weekAgo := now.AddDate(0, 0, -7)
	threeDAgo := now.AddDate(0, 0, -3)

	return []model.Application{
		{
			ID:          "sample-001",
			Company:     "Acme Corp",
			Position:    "Software Engineer",
			Status:      model.StatusInterview,
			DateApplied: &weekAgo,
			Notes:       "Tech stack: Go, React, PostgreSQL. Team of 5 engineers.",
			SalaryMin:   130000,
			SalaryMax:   160000,
			JobURL:      "https://example.com/jobs/acme-swe",
			Location:    "San Francisco, CA",
			Remote:      true,
			ContactName: "Jane Smith",
			ContactEmail: "jane@acme.example.com",
			Interviews: []model.Interview{
				{
					Date:     weekAgo.AddDate(0, 0, 5),
					Type:     "phone",
					Notes:    "Initial recruiter screen - went well",
					WithWhom: "Jane Smith (Recruiter)",
				},
			},
			ResumeVersion: "resume-v2.pdf",
			CreatedAt:     weekAgo,
			UpdatedAt:     now,
		},
		{
			ID:          "sample-002",
			Company:     "TechStart Inc",
			Position:    "Frontend Developer",
			Status:      model.StatusApplied,
			DateApplied: &threeDAgo,
			Notes:       "Early-stage startup, Series A. Building developer tools.",
			SalaryMin:   120000,
			SalaryMax:   150000,
			JobURL:      "https://example.com/jobs/techstart-fe",
			Location:    "Remote",
			Remote:      true,
			CreatedAt:   threeDAgo,
			UpdatedAt:   threeDAgo,
		},
		{
			ID:          "sample-003",
			Company:     "BigCo Industries",
			Position:    "Full Stack Engineer",
			Status:      model.StatusSaved,
			DateApplied: nil, // Saved applications don't need a date
			Notes:       "Fortune 500, great benefits. Need to tailor resume for this one.",
			SalaryMin:   140000,
			SalaryMax:   180000,
			Location:    "Seattle, WA",
			Remote:      false,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
}

// load reads applications from the JSON file
func (s *Store) load() error {
	data, err := os.ReadFile(s.filepath)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		s.applications = []model.Application{}
		return nil
	}

	return json.Unmarshal(data, &s.applications)
}

// save writes applications to the JSON file
func (s *Store) save() error {
	data, err := json.MarshalIndent(s.applications, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filepath, data, 0644)
}

// Add creates a new application and returns it
func (s *Store) Add(app model.Application) (model.Application, error) {
	app.ID = uuid.New().String()
	app.CreatedAt = time.Now()
	app.UpdatedAt = time.Now()

	if app.Status == "" {
		app.Status = model.StatusApplied
	}
	if app.DateApplied == nil && app.Status != model.StatusSaved {
		now := time.Now()
		app.DateApplied = &now
	}

	s.applications = append(s.applications, app)
	return app, s.save()
}

// Update modifies an existing application
func (s *Store) Update(app model.Application) error {
	for i, a := range s.applications {
		if a.ID == app.ID {
			app.UpdatedAt = time.Now()
			app.CreatedAt = a.CreatedAt // Preserve original creation time
			s.applications[i] = app
			return s.save()
		}
	}
	return ErrNotFound
}

// Delete removes an application by ID
func (s *Store) Delete(id string) error {
	for i, a := range s.applications {
		if a.ID == id {
			s.applications = append(s.applications[:i], s.applications[i+1:]...)
			return s.save()
		}
	}
	return ErrNotFound
}

// GetByID returns a single application by ID
func (s *Store) GetByID(id string) (model.Application, error) {
	for _, a := range s.applications {
		if a.ID == id {
			return a, nil
		}
	}
	return model.Application{}, ErrNotFound
}

// List returns all applications, sorted by status priority (later stages first), then by date (newest first)
func (s *Store) List() []model.Application {
	apps := make([]model.Application, len(s.applications))
	copy(apps, s.applications)

	sort.Slice(apps, func(i, j int) bool {
		// Primary sort: status priority (higher first)
		pi, pj := model.StatusPriority(apps[i].Status), model.StatusPriority(apps[j].Status)
		if pi != pj {
			return pi > pj
		}
		// Secondary sort: date (newest first, nil dates last)
		if apps[i].DateApplied == nil && apps[j].DateApplied == nil {
			return false
		}
		if apps[i].DateApplied == nil {
			return false
		}
		if apps[j].DateApplied == nil {
			return true
		}
		return apps[i].DateApplied.After(*apps[j].DateApplied)
	})

	return apps
}

// FilterByStatus returns applications with the given status
func (s *Store) FilterByStatus(status string) []model.Application {
	var result []model.Application
	for _, a := range s.applications {
		if a.Status == status {
			result = append(result, a)
		}
	}

	// Sort by date within same status (newest first, nil dates last)
	sort.Slice(result, func(i, j int) bool {
		if result[i].DateApplied == nil && result[j].DateApplied == nil {
			return false
		}
		if result[i].DateApplied == nil {
			return false
		}
		if result[j].DateApplied == nil {
			return true
		}
		return result[i].DateApplied.After(*result[j].DateApplied)
	})

	return result
}

// Search returns applications matching the query in company or position
func (s *Store) Search(query string) []model.Application {
	query = strings.ToLower(query)
	var result []model.Application

	for _, a := range s.applications {
		if strings.Contains(strings.ToLower(a.Company), query) ||
			strings.Contains(strings.ToLower(a.Position), query) {
			result = append(result, a)
		}
	}

	// Sort by status priority (later stages first), then by date (newest first)
	sort.Slice(result, func(i, j int) bool {
		pi, pj := model.StatusPriority(result[i].Status), model.StatusPriority(result[j].Status)
		if pi != pj {
			return pi > pj
		}
		if result[i].DateApplied == nil && result[j].DateApplied == nil {
			return false
		}
		if result[i].DateApplied == nil {
			return false
		}
		if result[j].DateApplied == nil {
			return true
		}
		return result[i].DateApplied.After(*result[j].DateApplied)
	})

	return result
}

// CountByStatus returns a map of status to count
func (s *Store) CountByStatus() map[string]int {
	counts := make(map[string]int)
	for _, a := range s.applications {
		counts[a.Status]++
	}
	return counts
}

// Total returns the total number of applications
func (s *Store) Total() int {
	return len(s.applications)
}

// UpdateStatus is a convenience method to change just the status
func (s *Store) UpdateStatus(id string, status string) error {
	app, err := s.GetByID(id)
	if err != nil {
		return err
	}
	app.Status = status
	// Auto-set date when transitioning to non-saved status
	if app.DateApplied == nil && status != model.StatusSaved {
		now := time.Now()
		app.DateApplied = &now
	}
	return s.Update(app)
}
