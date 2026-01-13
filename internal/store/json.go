package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"jobtrack/internal/model"

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
			s.applications = []model.Application{}
			return s, s.save()
		}
		return nil, err
	}

	return s, nil
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
	if app.DateApplied.IsZero() {
		app.DateApplied = time.Now()
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

// List returns all applications, sorted by date applied (newest first)
func (s *Store) List() []model.Application {
	apps := make([]model.Application, len(s.applications))
	copy(apps, s.applications)

	sort.Slice(apps, func(i, j int) bool {
		return apps[i].DateApplied.After(apps[j].DateApplied)
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

	sort.Slice(result, func(i, j int) bool {
		return result[i].DateApplied.After(result[j].DateApplied)
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

	sort.Slice(result, func(i, j int) bool {
		return result[i].DateApplied.After(result[j].DateApplied)
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
	return s.Update(app)
}
