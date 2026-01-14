package tui

import (
	"fmt"
	"strings"

	"github.com/celloopa/ghosted/internal/model"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// DetailView displays a single job application's details
type DetailView struct {
	application *model.Application
	width       int
	height      int
	keys        KeyMap
	scrollY     int
}

// NewDetailView creates a new detail view
func NewDetailView(app *model.Application, keys KeyMap) DetailView {
	return DetailView{
		application: app,
		keys:        keys,
	}
}

// SetApplication sets the application to display
func (d *DetailView) SetApplication(app *model.Application) {
	d.application = app
	d.scrollY = 0
}

// SetSize sets the view dimensions
func (d *DetailView) SetSize(width, height int) {
	d.width = width
	d.height = height
}

// HandleKey processes a key press and returns the action
func (d *DetailView) HandleKey(msg tea.KeyMsg) (handled bool, action string) {
	switch {
	case key.Matches(msg, d.keys.Back):
		return true, "back"
	case key.Matches(msg, d.keys.Edit):
		return true, "edit"
	case key.Matches(msg, d.keys.Delete):
		return true, "delete"
	case key.Matches(msg, d.keys.Up):
		if d.scrollY > 0 {
			d.scrollY--
		}
		return true, ""
	case key.Matches(msg, d.keys.Down):
		d.scrollY++
		return true, ""
	case key.Matches(msg, d.keys.Status1):
		return true, "status:saved"
	case key.Matches(msg, d.keys.Status2):
		return true, "status:applied"
	case key.Matches(msg, d.keys.Status3):
		return true, "status:screening"
	case key.Matches(msg, d.keys.Status4):
		return true, "status:interview"
	case key.Matches(msg, d.keys.Status5):
		return true, "status:offer"
	case key.Matches(msg, d.keys.Status6):
		return true, "status:accepted"
	case key.Matches(msg, d.keys.Status7):
		return true, "status:rejected"
	case key.Matches(msg, d.keys.Status8):
		return true, "status:withdrawn"
	case key.Matches(msg, d.keys.Quit):
		return true, "quit"
	}
	return false, ""
}

// View renders the detail view
func (d *DetailView) View() string {
	if d.application == nil {
		return SubtleStyle.Render("No application selected")
	}

	if d.width == 0 {
		d.width = 80
	}

	app := d.application
	var b strings.Builder

	// Title with company and position
	title := fmt.Sprintf("%s @ %s", app.Position, app.Company)
	b.WriteString(TitleStyle.Render(title))
	b.WriteString("\n")

	// Status badge
	statusStyle := StatusBadgeStyle(app.Status)
	b.WriteString(statusStyle.Render(model.StatusLabel(app.Status)))
	b.WriteString("\n\n")

	// Basic Info Section
	b.WriteString(SectionStyle.Render("Basic Information"))
	b.WriteString("\n")
	if app.DateApplied != nil {
		b.WriteString(d.renderField("Applied", app.DateApplied.Format("January 2, 2006")))
	} else {
		b.WriteString(d.renderField("Applied", "Not sent"))
	}
	if app.Location != "" {
		loc := app.Location
		if app.Remote {
			loc += " (Remote)"
		}
		b.WriteString(d.renderField("Location", loc))
	} else if app.Remote {
		b.WriteString(d.renderField("Location", "Remote"))
	}
	if salary := app.SalaryRange(); salary != "" {
		b.WriteString(d.renderField("Salary", salary))
	}
	if app.JobURL != "" {
		b.WriteString(d.renderField("URL", app.JobURL))
	}

	// Contact Info
	if app.ContactName != "" || app.ContactEmail != "" {
		b.WriteString("\n")
		b.WriteString(SectionStyle.Render("Contact"))
		b.WriteString("\n")
		if app.ContactName != "" {
			b.WriteString(d.renderField("Name", app.ContactName))
		}
		if app.ContactEmail != "" {
			b.WriteString(d.renderField("Email", app.ContactEmail))
		}
	}

	// Documents
	if app.ResumeVersion != "" || app.CoverLetter != "" {
		b.WriteString("\n")
		b.WriteString(SectionStyle.Render("Documents"))
		b.WriteString("\n")
		if app.ResumeVersion != "" {
			b.WriteString(d.renderField("Resume", app.ResumeVersion))
		}
		if app.CoverLetter != "" {
			b.WriteString(d.renderField("Cover Letter", app.CoverLetter))
		}
	}

	// Interviews
	if len(app.Interviews) > 0 {
		b.WriteString("\n")
		b.WriteString(SectionStyle.Render("Interviews"))
		b.WriteString("\n")
		for i, interview := range app.Interviews {
			prefix := fmt.Sprintf("%d. ", i+1)
			date := interview.Date.Format("Jan 2, 2006 3:04 PM")
			interviewType := interview.Type
			if interviewType == "" {
				interviewType = "Interview"
			}
			b.WriteString(fmt.Sprintf("%s%s - %s\n",
				prefix,
				HighlightStyle.Render(date),
				ValueStyle.Render(interviewType),
			))
			if interview.WithWhom != "" {
				b.WriteString(fmt.Sprintf("   With: %s\n", interview.WithWhom))
			}
			if interview.Notes != "" {
				b.WriteString(fmt.Sprintf("   Notes: %s\n", interview.Notes))
			}
		}
	}

	// Follow-up
	if app.NextFollowUp != nil {
		b.WriteString("\n")
		b.WriteString(SectionStyle.Render("Follow-up"))
		b.WriteString("\n")
		b.WriteString(d.renderField("Next", app.NextFollowUp.Format("January 2, 2006")))
	}

	// Notes
	if app.Notes != "" {
		b.WriteString("\n")
		b.WriteString(SectionStyle.Render("Notes"))
		b.WriteString("\n")
		b.WriteString(app.Notes)
		b.WriteString("\n")
	}

	// Metadata
	b.WriteString("\n")
	b.WriteString(SubtleStyle.Render(fmt.Sprintf("Created: %s | Updated: %s",
		app.CreatedAt.Format("Jan 2, 2006"),
		app.UpdatedAt.Format("Jan 2, 2006"),
	)))
	b.WriteString("\n")

	// Help
	b.WriteString("\n")
	b.WriteString(d.renderHelp())

	// Apply box style
	content := b.String()
	maxWidth := d.width - 4
	if maxWidth > 80 {
		maxWidth = 80
	}
	return BoxStyle.Width(maxWidth).Render(content)
}

func (d *DetailView) renderField(label, value string) string {
	return fmt.Sprintf("%s %s\n",
		LabelStyle.Render(label+":"),
		ValueStyle.Render(value),
	)
}

func (d *DetailView) renderHelp() string {
	return fmt.Sprintf("%s %s  %s %s  %s %s  %s %s",
		HelpKeyStyle.Render("e"),
		HelpDescStyle.Render("edit"),
		HelpKeyStyle.Render("d"),
		HelpDescStyle.Render("delete"),
		HelpKeyStyle.Render("1-7"),
		HelpDescStyle.Render("change status"),
		HelpKeyStyle.Render("esc"),
		HelpDescStyle.Render("back"),
	)
}
