package tui

import (
	"fmt"
	"strings"

	"github.com/celloopa/ghosted/internal/model"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ListView displays a list of job applications
type ListView struct {
	applications []model.Application
	cursor       int
	width        int
	height       int
	keys         KeyMap

	// Search/filter state
	searchMode   bool
	searchInput  textinput.Model
	searchQuery  string
	filterStatus string

	// Help
	showHelp bool
}

// NewListView creates a new list view
func NewListView(apps []model.Application, keys KeyMap) ListView {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.CharLimit = 50

	return ListView{
		applications: apps,
		cursor:       0,
		keys:         keys,
		searchInput:  ti,
	}
}

// SetApplications updates the applications list
func (l *ListView) SetApplications(apps []model.Application) {
	l.applications = apps
	if l.cursor >= len(apps) {
		l.cursor = max(0, len(apps)-1)
	}
}

// SetSize sets the view dimensions
func (l *ListView) SetSize(width, height int) {
	l.width = width
	l.height = height
}

// SelectedApplication returns the currently selected application
func (l *ListView) SelectedApplication() *model.Application {
	if len(l.applications) == 0 || l.cursor >= len(l.applications) {
		return nil
	}
	return &l.applications[l.cursor]
}

// HandleKey processes a key press and returns true if handled
func (l *ListView) HandleKey(msg tea.KeyMsg) (handled bool, action string) {
	// If in search mode, handle search input
	if l.searchMode {
		switch {
		case key.Matches(msg, l.keys.Cancel):
			l.searchMode = false
			l.searchInput.Blur()
			return true, ""
		case key.Matches(msg, l.keys.Enter):
			l.searchMode = false
			l.searchQuery = l.searchInput.Value()
			l.searchInput.Blur()
			return true, "search"
		}
		return false, ""
	}

	// Normal mode key handling
	switch {
	case key.Matches(msg, l.keys.Up):
		l.MoveUp()
		return true, ""
	case key.Matches(msg, l.keys.Down):
		l.MoveDown()
		return true, ""
	case key.Matches(msg, l.keys.Top):
		l.cursor = 0
		return true, ""
	case key.Matches(msg, l.keys.Bottom):
		l.cursor = max(0, len(l.applications)-1)
		return true, ""
	case key.Matches(msg, l.keys.Search):
		l.searchMode = true
		l.searchInput.Focus()
		return true, ""
	case key.Matches(msg, l.keys.Clear):
		l.searchQuery = ""
		l.filterStatus = ""
		return true, "clear"
	case key.Matches(msg, l.keys.Help):
		l.showHelp = !l.showHelp
		return true, ""
	case key.Matches(msg, l.keys.Add):
		return true, "add"
	case key.Matches(msg, l.keys.Edit):
		return true, "edit"
	case key.Matches(msg, l.keys.Delete):
		return true, "delete"
	case key.Matches(msg, l.keys.Enter):
		return true, "view"
	case key.Matches(msg, l.keys.Filter):
		return true, "filter"
	case key.Matches(msg, l.keys.Status1):
		return true, "status:saved"
	case key.Matches(msg, l.keys.Status2):
		return true, "status:applied"
	case key.Matches(msg, l.keys.Status3):
		return true, "status:screening"
	case key.Matches(msg, l.keys.Status4):
		return true, "status:interview"
	case key.Matches(msg, l.keys.Status5):
		return true, "status:offer"
	case key.Matches(msg, l.keys.Status6):
		return true, "status:accepted"
	case key.Matches(msg, l.keys.Status7):
		return true, "status:rejected"
	case key.Matches(msg, l.keys.Status8):
		return true, "status:withdrawn"
	case key.Matches(msg, l.keys.Quit):
		return true, "quit"
	}

	return false, ""
}

// UpdateSearchInput updates the search input
func (l *ListView) UpdateSearchInput(msg textinput.Model) {
	l.searchInput = msg
}

// SearchInput returns the search input model
func (l *ListView) SearchInput() textinput.Model {
	return l.searchInput
}

// IsSearchMode returns true if in search mode
func (l *ListView) IsSearchMode() bool {
	return l.searchMode
}

// SearchQuery returns the current search query
func (l *ListView) SearchQuery() string {
	return l.searchQuery
}

// SetFilterStatus sets the filter status
func (l *ListView) SetFilterStatus(status string) {
	l.filterStatus = status
}

// FilterStatus returns the current filter status
func (l *ListView) FilterStatus() string {
	return l.filterStatus
}

// MoveUp moves the cursor up
func (l *ListView) MoveUp() {
	if l.cursor > 0 {
		l.cursor--
	}
}

// MoveDown moves the cursor down
func (l *ListView) MoveDown() {
	if l.cursor < len(l.applications)-1 {
		l.cursor++
	}
}

// View renders the list view
func (l *ListView) View() string {
	if l.width == 0 {
		l.width = 80
	}
	if l.height == 0 {
		l.height = 24
	}

	var b strings.Builder

	// Title with ghost ASCII art
	ghost := `   ██████╗ ██╗  ██╗ ██████╗ ███████╗████████╗███████╗██████╗
  ██╔════╝ ██║  ██║██╔═══██╗██╔════╝╚══██╔══╝██╔════╝██╔══██╗
  ██║  ███╗███████║██║   ██║███████╗   ██║   █████╗  ██║  ██║
  ██║   ██║██╔══██║██║   ██║╚════██║   ██║   ██╔══╝  ██║  ██║
  ╚██████╔╝██║  ██║╚██████╔╝███████║   ██║   ███████╗██████╔╝
   ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚══════╝   ╚═╝   ╚══════╝╚═════╝ `
	b.WriteString(SubtleStyle.Render(ghost))
	b.WriteString("\n\n")

	// Search input if in search mode
	if l.searchMode {
		b.WriteString(l.searchInput.View())
		b.WriteString("\n\n")
	} else if l.searchQuery != "" || l.filterStatus != "" {
		// Show active filters
		var filters []string
		if l.searchQuery != "" {
			filters = append(filters, fmt.Sprintf("Search: %s", l.searchQuery))
		}
		if l.filterStatus != "" {
			filters = append(filters, fmt.Sprintf("Status: %s", model.StatusLabel(l.filterStatus)))
		}
		b.WriteString(SubtleStyle.Render(strings.Join(filters, " | ")))
		b.WriteString(" ")
		b.WriteString(SubtleStyle.Render("(c to clear)"))
		b.WriteString("\n\n")
	}

	// Empty state
	if len(l.applications) == 0 {
		emptyMsg := "No applications yet. Press 'a' to add one."
		if l.searchQuery != "" || l.filterStatus != "" {
			emptyMsg = "No matching applications. Press 'c' to clear filters."
		}
		b.WriteString(SubtleStyle.Render(emptyMsg))
		b.WriteString("\n")
	} else {
		// Header
		header := l.renderHeader()
		b.WriteString(header)
		b.WriteString("\n\n")

		// Calculate visible rows
		listHeight := l.height - 12 // Account for header, footer, spacing, etc.
		if listHeight < 3 {
			listHeight = 3
		}

		// Determine scroll window
		start := 0
		if l.cursor >= listHeight {
			start = l.cursor - listHeight + 1
		}
		end := start + listHeight
		if end > len(l.applications) {
			end = len(l.applications)
		}

		// Render visible rows
		for i := start; i < end; i++ {
			row := l.renderRow(i, i == l.cursor)
			b.WriteString(row)
			b.WriteString("\n")
		}

		// Scroll indicator
		if len(l.applications) > listHeight {
			scrollInfo := fmt.Sprintf(" %d/%d ", l.cursor+1, len(l.applications))
			b.WriteString(SubtleStyle.Render(scrollInfo))
			b.WriteString("\n")
		}
	}

	// Help
	if l.showHelp {
		b.WriteString("\n\n")
		b.WriteString(l.renderFullHelp())
	} else {
		b.WriteString("\n\n")
		b.WriteString(l.renderShortHelp())
	}

	return b.String()
}

func (l *ListView) renderHeader() string {
	// Column widths
	companyW := 20
	positionW := 25
	statusW := 12
	dateW := 12

	company := truncate("COMPANY", companyW)
	position := truncate("POSITION", positionW)
	status := truncate("STATUS", statusW)
	date := truncate("DATE SENT", dateW)

	header := fmt.Sprintf("%-*s %-*s %-*s %-*s",
		companyW, company,
		positionW, position,
		statusW+10, status, // Match row padding for ANSI codes
		dateW, date,
	)

	return HeaderStyle.Render(header)
}

func (l *ListView) renderRow(index int, selected bool) string {
	app := l.applications[index]

	companyW := 20
	positionW := 25
	statusW := 12
	dateW := 12

	company := truncate(app.Company, companyW)
	position := truncate(app.Position, positionW)
	status := truncate(model.StatusLabel(app.Status), statusW)
	date := app.DateApplied.Format("2006-01-02")

	// Status with color
	statusStyle := StatusBadgeStyle(app.Status)
	statusText := statusStyle.Render(status)

	row := fmt.Sprintf("%-*s %-*s %s %-*s",
		companyW, company,
		positionW, position,
		padRight(statusText, statusW+10), // Extra padding for ANSI codes
		dateW, date,
	)

	if selected {
		// Highlight the entire row
		return SelectedRowStyle.Render("> " + row)
	}
	return NormalRowStyle.Render("  " + row)
}

func (l *ListView) renderShortHelp() string {
	bindings := l.keys.ShortHelp()
	var parts []string
	for _, b := range bindings {
		parts = append(parts, fmt.Sprintf("%s %s",
			HelpKeyStyle.Render(b.Help().Key),
			HelpDescStyle.Render(b.Help().Desc),
		))
	}
	return strings.Join(parts, HelpSepStyle.Render(" | "))
}

func (l *ListView) renderFullHelp() string {
	var b strings.Builder
	b.WriteString(SectionStyle.Render("Keyboard Shortcuts"))
	b.WriteString("\n\n")

	groups := l.keys.FullHelp()
	for _, group := range groups {
		var parts []string
		for _, binding := range group {
			parts = append(parts, fmt.Sprintf("%s %s",
				HelpKeyStyle.Render(binding.Help().Key),
				HelpDescStyle.Render(binding.Help().Desc),
			))
		}
		b.WriteString(strings.Join(parts, "  "))
		b.WriteString("\n")
	}

	return b.String()
}

// Helper functions
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func padRight(s string, length int) string {
	// Calculate visible length (excluding ANSI codes)
	visible := lipgloss.Width(s)
	if visible >= length {
		return s
	}
	return s + strings.Repeat(" ", length-visible)
}
