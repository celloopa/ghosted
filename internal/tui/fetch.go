package tui

import (
	"fmt"
	"strings"

	"github.com/celloopa/ghosted/internal/fetch"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// FetchView handles the fetch URL input view
type FetchView struct {
	keys       KeyMap
	urlInput   textinput.Model
	width      int
	height     int
	result     *FetchResultDisplay
	err        error
	isFetching bool
}

// FetchResultDisplay holds the display info for a fetch result
type FetchResultDisplay struct {
	Type       string // "job" or "cv"
	OutputPath string
	Info1      string // Company/Name
	Info2      string // Position/Label
	Size       int
}

// fetchCompleteMsg is sent when a fetch operation completes
type fetchCompleteMsg struct {
	result *FetchResultDisplay
	err    error
}

// NewFetchView creates a new fetch view
func NewFetchView(keys KeyMap) FetchView {
	urlInput := textinput.New()
	urlInput.Placeholder = "URL or domain (e.g., jobs.lever.co/company/123 or cello.design)"
	urlInput.CharLimit = 500
	urlInput.Width = 60
	urlInput.Focus()

	return FetchView{
		keys:     keys,
		urlInput: urlInput,
	}
}

// SetSize sets the view dimensions
func (v *FetchView) SetSize(width, height int) {
	v.width = width
	v.height = height
	v.urlInput.Width = width - 10
	if v.urlInput.Width > 80 {
		v.urlInput.Width = 80
	}
}

// Reset clears the view state
func (v *FetchView) Reset() {
	v.urlInput.SetValue("")
	v.urlInput.Focus()
	v.result = nil
	v.err = nil
	v.isFetching = false
}

// URLInput returns the URL input model
func (v *FetchView) URLInput() textinput.Model {
	return v.urlInput
}

// UpdateURLInput updates the URL input model
func (v *FetchView) UpdateURLInput(input textinput.Model) {
	v.urlInput = input
}

// IsFetching returns whether a fetch is in progress
func (v *FetchView) IsFetching() bool {
	return v.isFetching
}

// HandleKey processes key events
func (v *FetchView) HandleKey(msg tea.KeyMsg) (handled bool, action string) {
	switch {
	case key.Matches(msg, v.keys.Cancel):
		if v.result != nil || v.err != nil {
			// Clear result and allow another fetch
			v.result = nil
			v.err = nil
			v.urlInput.Focus()
			return true, ""
		}
		return true, "cancel"
	case key.Matches(msg, v.keys.Enter):
		if v.isFetching {
			return true, ""
		}
		if v.result != nil {
			// After viewing result, go back
			return true, "done"
		}
		url := strings.TrimSpace(v.urlInput.Value())
		if url != "" {
			return true, "fetch"
		}
		return true, ""
	}
	return false, ""
}

// StartFetch begins a fetch operation
func (v *FetchView) StartFetch() tea.Cmd {
	v.isFetching = true
	v.err = nil
	v.result = nil
	url := strings.TrimSpace(v.urlInput.Value())

	return func() tea.Msg {
		// Detect fetch type
		fetchType := fetch.DetectFetchType(url)

		// Create fetcher
		var result *FetchResultDisplay
		var err error

		if fetchType == fetch.FetchTypeCV {
			f := fetch.NewFetcher("local")
			cvResult, fetchErr := f.FetchCV(url)
			if fetchErr != nil {
				err = fetchErr
			} else {
				result = &FetchResultDisplay{
					Type:       "cv",
					OutputPath: cvResult.OutputPath,
					Info1:      cvResult.Name,
					Info2:      cvResult.Label,
					Size:       cvResult.Size,
				}
			}
		} else {
			f := fetch.NewFetcher("local/postings")
			// Ensure URL has scheme
			fetchURL := url
			if !fetch.IsURL(fetchURL) {
				fetchURL = "https://" + fetchURL
			}
			jobResult, fetchErr := f.Fetch(fetchURL, "")
			if fetchErr != nil {
				err = fetchErr
			} else {
				result = &FetchResultDisplay{
					Type:       "job",
					OutputPath: jobResult.OutputPath,
					Info1:      jobResult.Company,
					Info2:      jobResult.Position,
					Size:       jobResult.ContentSize,
				}
			}
		}

		return fetchCompleteMsg{result: result, err: err}
	}
}

// HandleFetchComplete processes the fetch completion message
func (v *FetchView) HandleFetchComplete(msg fetchCompleteMsg) {
	v.isFetching = false
	v.result = msg.result
	v.err = msg.err
}

// View renders the fetch view
func (v *FetchView) View() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Fetch Job Posting or CV"))
	b.WriteString("\n\n")

	if v.isFetching {
		b.WriteString(SubtleStyle.Render("Fetching..."))
		b.WriteString("\n")
	} else if v.err != nil {
		b.WriteString(ErrorStyle.Render(fmt.Sprintf("Error: %v", v.err)))
		b.WriteString("\n\n")
		b.WriteString(SubtleStyle.Render("Press esc to try again, or esc twice to cancel"))
	} else if v.result != nil {
		b.WriteString(SuccessStyle.Render("Fetch successful!"))
		b.WriteString("\n\n")

		if v.result.Type == "cv" {
			b.WriteString(LabelStyle.Render("Type: "))
			b.WriteString("CV (JSON Resume)\n")
		} else {
			b.WriteString(LabelStyle.Render("Type: "))
			b.WriteString("Job Posting\n")
		}

		b.WriteString(LabelStyle.Render("Saved to: "))
		b.WriteString(v.result.OutputPath)
		b.WriteString("\n")

		if v.result.Info1 != "" {
			if v.result.Type == "cv" {
				b.WriteString(LabelStyle.Render("Name: "))
			} else {
				b.WriteString(LabelStyle.Render("Company: "))
			}
			b.WriteString(v.result.Info1)
			b.WriteString("\n")
		}

		if v.result.Info2 != "" {
			if v.result.Type == "cv" {
				b.WriteString(LabelStyle.Render("Label: "))
			} else {
				b.WriteString(LabelStyle.Render("Position: "))
			}
			b.WriteString(v.result.Info2)
			b.WriteString("\n")
		}

		b.WriteString(LabelStyle.Render("Size: "))
		b.WriteString(fmt.Sprintf("%d bytes", v.result.Size))
		b.WriteString("\n\n")

		if v.result.Type == "job" {
			b.WriteString(SubtleStyle.Render("Next step: ghosted apply " + v.result.OutputPath))
			b.WriteString("\n\n")
		}

		b.WriteString(SubtleStyle.Render("Press enter to close, or esc to fetch another"))
	} else {
		b.WriteString(SubtleStyle.Render("Enter a URL to fetch a job posting, or a domain to fetch a CV"))
		b.WriteString("\n")
		b.WriteString(SubtleStyle.Render("Examples: jobs.lever.co/company/123 (job) or cello.design (CV)"))
		b.WriteString("\n\n")

		b.WriteString(LabelStyle.Render("URL/Domain: "))
		b.WriteString("\n")
		b.WriteString(v.urlInput.View())
		b.WriteString("\n\n")

		b.WriteString(fmt.Sprintf("%s %s  %s %s",
			HelpKeyStyle.Render("enter"),
			HelpDescStyle.Render("fetch"),
			HelpKeyStyle.Render("esc"),
			HelpDescStyle.Render("cancel"),
		))
	}

	return b.String()
}
