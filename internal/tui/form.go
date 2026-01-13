package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"jobtrack/internal/model"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// FormField represents the different fields in the form
type FormField int

const (
	FieldCompany FormField = iota
	FieldPosition
	FieldStatus
	FieldDateApplied
	FieldLocation
	FieldRemote
	FieldSalaryMin
	FieldSalaryMax
	FieldJobURL
	FieldContactName
	FieldContactEmail
	FieldResumeVersion
	FieldCoverLetter
	FieldNotes
	FieldCount // Used to track total number of fields
)

// FormView handles adding and editing applications
type FormView struct {
	inputs       []textinput.Model
	focusIndex   int
	keys         KeyMap
	width        int
	height       int
	isEdit       bool
	application  *model.Application
	statusIndex  int
	remoteToggle bool
	err          string
}

// NewFormView creates a new form for adding/editing applications
func NewFormView(keys KeyMap) FormView {
	inputs := make([]textinput.Model, FieldCount)

	// Company
	inputs[FieldCompany] = textinput.New()
	inputs[FieldCompany].Placeholder = "Company name"
	inputs[FieldCompany].CharLimit = 100
	inputs[FieldCompany].Focus()

	// Position
	inputs[FieldPosition] = textinput.New()
	inputs[FieldPosition].Placeholder = "Position title"
	inputs[FieldPosition].CharLimit = 100

	// Status (display only, changed with left/right)
	inputs[FieldStatus] = textinput.New()
	inputs[FieldStatus].Placeholder = "Status"

	// Date Applied
	inputs[FieldDateApplied] = textinput.New()
	inputs[FieldDateApplied].Placeholder = "YYYY-MM-DD (leave empty for today)"
	inputs[FieldDateApplied].CharLimit = 10

	// Location
	inputs[FieldLocation] = textinput.New()
	inputs[FieldLocation].Placeholder = "Location (e.g., San Francisco, CA)"
	inputs[FieldLocation].CharLimit = 100

	// Remote (display only, toggle with space)
	inputs[FieldRemote] = textinput.New()
	inputs[FieldRemote].Placeholder = "Remote"

	// Salary Min
	inputs[FieldSalaryMin] = textinput.New()
	inputs[FieldSalaryMin].Placeholder = "Min salary (e.g., 100000)"
	inputs[FieldSalaryMin].CharLimit = 10

	// Salary Max
	inputs[FieldSalaryMax] = textinput.New()
	inputs[FieldSalaryMax].Placeholder = "Max salary (e.g., 150000)"
	inputs[FieldSalaryMax].CharLimit = 10

	// Job URL
	inputs[FieldJobURL] = textinput.New()
	inputs[FieldJobURL].Placeholder = "Job posting URL"
	inputs[FieldJobURL].CharLimit = 500

	// Contact Name
	inputs[FieldContactName] = textinput.New()
	inputs[FieldContactName].Placeholder = "Contact name"
	inputs[FieldContactName].CharLimit = 100

	// Contact Email
	inputs[FieldContactEmail] = textinput.New()
	inputs[FieldContactEmail].Placeholder = "Contact email"
	inputs[FieldContactEmail].CharLimit = 100

	// Resume Version
	inputs[FieldResumeVersion] = textinput.New()
	inputs[FieldResumeVersion].Placeholder = "Resume version used"
	inputs[FieldResumeVersion].CharLimit = 100

	// Cover Letter
	inputs[FieldCoverLetter] = textinput.New()
	inputs[FieldCoverLetter].Placeholder = "Cover letter path/name"
	inputs[FieldCoverLetter].CharLimit = 100

	// Notes
	inputs[FieldNotes] = textinput.New()
	inputs[FieldNotes].Placeholder = "Notes"
	inputs[FieldNotes].CharLimit = 1000

	return FormView{
		inputs:      inputs,
		focusIndex:  0,
		keys:        keys,
		statusIndex: 0, // "applied" by default
	}
}

// SetApplication populates the form for editing
func (f *FormView) SetApplication(app *model.Application) {
	f.isEdit = true
	f.application = app

	f.inputs[FieldCompany].SetValue(app.Company)
	f.inputs[FieldPosition].SetValue(app.Position)
	f.inputs[FieldDateApplied].SetValue(app.DateApplied.Format("2006-01-02"))
	f.inputs[FieldLocation].SetValue(app.Location)
	f.remoteToggle = app.Remote
	if app.SalaryMin > 0 {
		f.inputs[FieldSalaryMin].SetValue(strconv.Itoa(app.SalaryMin))
	}
	if app.SalaryMax > 0 {
		f.inputs[FieldSalaryMax].SetValue(strconv.Itoa(app.SalaryMax))
	}
	f.inputs[FieldJobURL].SetValue(app.JobURL)
	f.inputs[FieldContactName].SetValue(app.ContactName)
	f.inputs[FieldContactEmail].SetValue(app.ContactEmail)
	f.inputs[FieldResumeVersion].SetValue(app.ResumeVersion)
	f.inputs[FieldCoverLetter].SetValue(app.CoverLetter)
	f.inputs[FieldNotes].SetValue(app.Notes)

	// Find status index
	statuses := model.AllStatuses()
	for i, s := range statuses {
		if s == app.Status {
			f.statusIndex = i
			break
		}
	}
}

// Reset clears the form for a new entry
func (f *FormView) Reset() {
	f.isEdit = false
	f.application = nil
	f.focusIndex = 0
	f.statusIndex = 0
	f.remoteToggle = false
	f.err = ""

	for i := range f.inputs {
		f.inputs[i].SetValue("")
	}

	f.inputs[FieldCompany].Focus()
}

// SetSize sets the view dimensions
func (f *FormView) SetSize(width, height int) {
	f.width = width
	f.height = height
}

// HandleKey processes key input
func (f *FormView) HandleKey(msg tea.KeyMsg) (handled bool, action string) {
	switch {
	case key.Matches(msg, f.keys.Cancel):
		return true, "cancel"
	case key.Matches(msg, f.keys.Submit):
		return true, "submit"
	case key.Matches(msg, f.keys.Tab), key.Matches(msg, f.keys.Down):
		f.nextField()
		return true, ""
	case key.Matches(msg, f.keys.Up):
		f.prevField()
		return true, ""
	}

	// Handle special fields
	field := FormField(f.focusIndex)
	switch field {
	case FieldStatus:
		if key.Matches(msg, f.keys.Left) {
			f.prevStatus()
			return true, ""
		} else if key.Matches(msg, f.keys.Right) {
			f.nextStatus()
			return true, ""
		}
	case FieldRemote:
		if msg.String() == " " || msg.String() == "enter" {
			f.remoteToggle = !f.remoteToggle
			return true, ""
		}
	}

	return false, ""
}

// UpdateInput updates the focused input
func (f *FormView) UpdateInput(msg textinput.Model) {
	field := FormField(f.focusIndex)
	// Don't update status or remote fields
	if field != FieldStatus && field != FieldRemote {
		f.inputs[f.focusIndex] = msg
	}
}

// FocusedInput returns the currently focused input
func (f *FormView) FocusedInput() textinput.Model {
	return f.inputs[f.focusIndex]
}

// FocusedField returns the currently focused field type
func (f *FormView) FocusedField() FormField {
	return FormField(f.focusIndex)
}

func (f *FormView) nextField() {
	f.inputs[f.focusIndex].Blur()
	f.focusIndex = (f.focusIndex + 1) % int(FieldCount)
	f.inputs[f.focusIndex].Focus()
}

func (f *FormView) prevField() {
	f.inputs[f.focusIndex].Blur()
	f.focusIndex--
	if f.focusIndex < 0 {
		f.focusIndex = int(FieldCount) - 1
	}
	f.inputs[f.focusIndex].Focus()
}

func (f *FormView) nextStatus() {
	statuses := model.AllStatuses()
	f.statusIndex = (f.statusIndex + 1) % len(statuses)
}

func (f *FormView) prevStatus() {
	statuses := model.AllStatuses()
	f.statusIndex--
	if f.statusIndex < 0 {
		f.statusIndex = len(statuses) - 1
	}
}

// Validate checks if the form is valid
func (f *FormView) Validate() bool {
	company := strings.TrimSpace(f.inputs[FieldCompany].Value())
	position := strings.TrimSpace(f.inputs[FieldPosition].Value())

	if company == "" {
		f.err = "Company is required"
		return false
	}
	if position == "" {
		f.err = "Position is required"
		return false
	}

	// Validate date if provided
	dateStr := strings.TrimSpace(f.inputs[FieldDateApplied].Value())
	if dateStr != "" {
		_, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			f.err = "Invalid date format (use YYYY-MM-DD)"
			return false
		}
	}

	// Validate salary if provided
	salaryMin := strings.TrimSpace(f.inputs[FieldSalaryMin].Value())
	if salaryMin != "" {
		if _, err := strconv.Atoi(salaryMin); err != nil {
			f.err = "Invalid salary min (use numbers only)"
			return false
		}
	}

	salaryMax := strings.TrimSpace(f.inputs[FieldSalaryMax].Value())
	if salaryMax != "" {
		if _, err := strconv.Atoi(salaryMax); err != nil {
			f.err = "Invalid salary max (use numbers only)"
			return false
		}
	}

	f.err = ""
	return true
}

// GetApplication returns the application from form data
func (f *FormView) GetApplication() model.Application {
	statuses := model.AllStatuses()

	app := model.Application{
		Company:       strings.TrimSpace(f.inputs[FieldCompany].Value()),
		Position:      strings.TrimSpace(f.inputs[FieldPosition].Value()),
		Status:        statuses[f.statusIndex],
		Location:      strings.TrimSpace(f.inputs[FieldLocation].Value()),
		Remote:        f.remoteToggle,
		JobURL:        strings.TrimSpace(f.inputs[FieldJobURL].Value()),
		ContactName:   strings.TrimSpace(f.inputs[FieldContactName].Value()),
		ContactEmail:  strings.TrimSpace(f.inputs[FieldContactEmail].Value()),
		ResumeVersion: strings.TrimSpace(f.inputs[FieldResumeVersion].Value()),
		CoverLetter:   strings.TrimSpace(f.inputs[FieldCoverLetter].Value()),
		Notes:         strings.TrimSpace(f.inputs[FieldNotes].Value()),
	}

	// Parse date
	dateStr := strings.TrimSpace(f.inputs[FieldDateApplied].Value())
	if dateStr != "" {
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			app.DateApplied = t
		}
	}
	if app.DateApplied.IsZero() {
		app.DateApplied = time.Now()
	}

	// Parse salary
	if s := strings.TrimSpace(f.inputs[FieldSalaryMin].Value()); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			app.SalaryMin = v
		}
	}
	if s := strings.TrimSpace(f.inputs[FieldSalaryMax].Value()); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			app.SalaryMax = v
		}
	}

	// If editing, preserve ID and interviews
	if f.isEdit && f.application != nil {
		app.ID = f.application.ID
		app.Interviews = f.application.Interviews
		app.NextFollowUp = f.application.NextFollowUp
		app.CreatedAt = f.application.CreatedAt
	}

	return app
}

// View renders the form
func (f *FormView) View() string {
	var b strings.Builder

	// Title
	title := "Add New Application"
	if f.isEdit {
		title = "Edit Application"
	}
	b.WriteString(TitleStyle.Render(title))
	b.WriteString("\n\n")

	// Error message
	if f.err != "" {
		b.WriteString(ErrorStyle.Render("Error: " + f.err))
		b.WriteString("\n\n")
	}

	// Fields
	b.WriteString(f.renderField(FieldCompany, "Company*"))
	b.WriteString(f.renderField(FieldPosition, "Position*"))
	b.WriteString(f.renderStatusField())
	b.WriteString(f.renderField(FieldDateApplied, "Date Applied"))
	b.WriteString(f.renderField(FieldLocation, "Location"))
	b.WriteString(f.renderRemoteField())
	b.WriteString(f.renderField(FieldSalaryMin, "Salary Min"))
	b.WriteString(f.renderField(FieldSalaryMax, "Salary Max"))
	b.WriteString(f.renderField(FieldJobURL, "Job URL"))
	b.WriteString(f.renderField(FieldContactName, "Contact Name"))
	b.WriteString(f.renderField(FieldContactEmail, "Contact Email"))
	b.WriteString(f.renderField(FieldResumeVersion, "Resume Version"))
	b.WriteString(f.renderField(FieldCoverLetter, "Cover Letter"))
	b.WriteString(f.renderField(FieldNotes, "Notes"))

	// Help
	b.WriteString("\n")
	b.WriteString(f.renderHelp())

	return b.String()
}

func (f *FormView) renderField(field FormField, label string) string {
	focused := FormField(f.focusIndex) == field

	labelStyle := LabelStyle
	if focused {
		labelStyle = labelStyle.Foreground(GetStatusColor(model.StatusApplied))
	}

	input := f.inputs[field].View()

	return fmt.Sprintf("%s\n%s\n\n", labelStyle.Render(label), input)
}

func (f *FormView) renderStatusField() string {
	focused := FormField(f.focusIndex) == FieldStatus
	statuses := model.AllStatuses()
	currentStatus := statuses[f.statusIndex]

	labelStyle := LabelStyle
	if focused {
		labelStyle = labelStyle.Foreground(GetStatusColor(model.StatusApplied))
	}

	statusStyle := StatusBadgeStyle(currentStatus)
	statusText := statusStyle.Render(model.StatusLabel(currentStatus))

	arrows := ""
	if focused {
		arrows = SubtleStyle.Render(" <- ->")
	}

	return fmt.Sprintf("%s\n%s%s\n\n", labelStyle.Render("Status"), statusText, arrows)
}

func (f *FormView) renderRemoteField() string {
	focused := FormField(f.focusIndex) == FieldRemote

	labelStyle := LabelStyle
	if focused {
		labelStyle = labelStyle.Foreground(GetStatusColor(model.StatusApplied))
	}

	toggle := "[ ] No"
	if f.remoteToggle {
		toggle = "[x] Yes"
	}

	toggleStyle := ValueStyle
	if focused {
		toggleStyle = HighlightStyle
	}

	hint := ""
	if focused {
		hint = SubtleStyle.Render(" (space to toggle)")
	}

	return fmt.Sprintf("%s\n%s%s\n\n", labelStyle.Render("Remote"), toggleStyle.Render(toggle), hint)
}

func (f *FormView) renderHelp() string {
	return fmt.Sprintf("%s %s  %s %s  %s %s",
		HelpKeyStyle.Render("tab/down"),
		HelpDescStyle.Render("next field"),
		HelpKeyStyle.Render("ctrl+s"),
		HelpDescStyle.Render("save"),
		HelpKeyStyle.Render("esc"),
		HelpDescStyle.Render("cancel"),
	)
}
