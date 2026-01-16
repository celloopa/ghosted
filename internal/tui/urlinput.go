// Package tui provides the terminal user interface for ghosted.
//
// urlinput.go implements the URL input dialog for fetching job postings
// and launching the AI agent pipeline.
//
// # Usage Flow
//
// 1. Press 'u' in list view to open the URL input dialog
// 2. Paste or type a job posting URL
// 3. Press Enter to fetch and launch Claude Code with agent context
// 4. Press Esc to cancel
//
// # Supported URL Formats
//
// The dialog accepts URLs from common job boards:
//   - Lever (jobs.lever.co)
//   - Greenhouse (boards.greenhouse.io)
//   - Workday (*.workday.com)
//   - LinkedIn (linkedin.com/jobs)
//   - Ashby (jobs.ashbyhq.com)
//   - Generic job posting URLs
//
// # What Happens on Submit
//
// When a URL is submitted:
//  1. ghosted fetch <url> is called to download the posting
//  2. claude is spawned with the agent context from ghosted context
//  3. The agent can then run the full pipeline (parse → resume → cover → review → track)
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// URLInputView handles the URL input dialog for fetching job postings
// and launching the AI agent pipeline.
type URLInputView struct {
	input    textinput.Model
	err      string
	width    int
	height   int
	quitting bool
}

// NewURLInputView creates a new URL input dialog.
func NewURLInputView() URLInputView {
	ti := textinput.New()
	ti.Placeholder = "https://jobs.lever.co/company/position-id"
	ti.CharLimit = 500
	ti.Width = 60
	ti.Focus()

	return URLInputView{
		input: ti,
	}
}

// Reset clears the input and error state.
func (u *URLInputView) Reset() {
	u.input.SetValue("")
	u.err = ""
	u.quitting = false
	u.input.Focus()
}

// SetSize sets the dialog dimensions.
func (u *URLInputView) SetSize(width, height int) {
	u.width = width
	u.height = height
}

// Value returns the current input value.
func (u *URLInputView) Value() string {
	return strings.TrimSpace(u.input.Value())
}

// SetError sets an error message to display.
func (u *URLInputView) SetError(err string) {
	u.err = err
}

// Input returns the textinput model for updates.
func (u *URLInputView) Input() textinput.Model {
	return u.input
}

// UpdateInput updates the textinput model.
func (u *URLInputView) UpdateInput(ti textinput.Model) {
	u.input = ti
}

// Validate checks if the input is a valid URL.
func (u *URLInputView) Validate() bool {
	url := u.Value()
	if url == "" {
		u.err = "URL is required"
		return false
	}

	// Basic URL validation
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		u.err = "URL must start with http:// or https://"
		return false
	}

	u.err = ""
	return true
}

// View renders the URL input dialog as a centered overlay.
func (u *URLInputView) View() string {
	var b strings.Builder

	// Dialog title
	b.WriteString(DialogTitleStyle.Render("Fetch Job Posting"))
	b.WriteString("\n\n")

	// Instructions
	b.WriteString(SubtleStyle.Render("Paste a job posting URL to fetch and process with AI"))
	b.WriteString("\n\n")

	// Input field label
	b.WriteString(LabelStyle.Render("URL"))
	b.WriteString("\n")
	b.WriteString(u.input.View())
	b.WriteString("\n")

	// Error message
	if u.err != "" {
		b.WriteString("\n")
		b.WriteString(ErrorStyle.Render(u.err))
		b.WriteString("\n")
	}

	// Supported sites hint
	b.WriteString("\n")
	b.WriteString(SubtleStyle.Render("Supported: Lever, Greenhouse, Workday, LinkedIn, Ashby"))
	b.WriteString("\n\n")

	// What happens next
	b.WriteString(HelpKeyStyle.Render("What happens:"))
	b.WriteString("\n")
	b.WriteString(SubtleStyle.Render("  1. Fetches posting to local/postings/"))
	b.WriteString("\n")
	b.WriteString(SubtleStyle.Render("  2. Launches Claude Code with agent context"))
	b.WriteString("\n")
	b.WriteString(SubtleStyle.Render("  3. AI generates resume & cover letter"))
	b.WriteString("\n\n")

	// Help
	b.WriteString(fmt.Sprintf("%s %s  %s %s",
		HelpKeyStyle.Render("enter"),
		HelpDescStyle.Render("fetch & launch"),
		HelpKeyStyle.Render("esc"),
		HelpDescStyle.Render("cancel"),
	))

	content := DialogStyle.Render(b.String())

	// Center the dialog
	if u.width > 0 && u.height > 0 {
		return lipgloss.Place(
			u.width,
			u.height,
			lipgloss.Center,
			lipgloss.Center,
			content,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#000000")),
		)
	}

	return content
}
