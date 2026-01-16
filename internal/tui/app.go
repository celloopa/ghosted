package tui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/celloopa/ghosted/internal/fetch"
	"github.com/celloopa/ghosted/internal/model"
	"github.com/celloopa/ghosted/internal/store"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewState represents the current view
type ViewState int

const (
	ViewSplash ViewState = iota
	ViewList
	ViewDetail
	ViewForm
	ViewFilter
	ViewConfirmDelete
	ViewURLInput
)

// splashDoneMsg signals the splash screen is done
type splashDoneMsg struct{}

// fetchResultMsg contains the result of fetching a job posting URL
type fetchResultMsg struct {
	outputPath string
	company    string
	position   string
	err        error
}

// claudeLaunchedMsg signals that Claude Code was launched
type claudeLaunchedMsg struct {
	err error
}

// App is the main Bubble Tea model
type App struct {
	store        *store.Store
	keys         KeyMap
	width        int
	height       int
	contentWidth int
	viewState    ViewState
	prevState    ViewState
	err          error
	statusMsg    string

	// Views
	listView     ListView
	detailView   DetailView
	formView     FormView
	urlInputView URLInputView

	// URL fetch state
	fetchedPosting string // Path to fetched posting for Claude launch

	// Filter state
	filterOptions  []string
	filterCursor   int
	selectedFilter string

	// Confirm delete
	deleteTarget *model.Application
}

// New creates a new App
func New(s *store.Store) App {
	keys := DefaultKeyMap()

	apps := s.List()
	listView := NewListView(apps, keys)
	detailView := NewDetailView(nil, keys)
	formView := NewFormView(keys)
	urlInputView := NewURLInputView()

	return App{
		store:        s,
		keys:         keys,
		viewState:    ViewSplash,
		listView:     listView,
		detailView:   detailView,
		formView:     formView,
		urlInputView: urlInputView,
		filterOptions: append([]string{"All"}, func() []string {
			statuses := model.AllStatuses()
			labels := make([]string, len(statuses))
			for i, s := range statuses {
				labels[i] = model.StatusLabel(s)
			}
			return labels
		}()...),
	}
}

// Init initializes the app
func (a App) Init() tea.Cmd {
	// Start splash screen timer
	return tea.Tick(1500*time.Millisecond, func(t time.Time) tea.Msg {
		return splashDoneMsg{}
	})
}

// Update handles messages
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Calculate content width (responsive with max)
		a.contentWidth = msg.Width - 4 // Account for padding
		if a.contentWidth > MaxContentWidth {
			a.contentWidth = MaxContentWidth
		}
		if a.contentWidth < MinContentWidth {
			a.contentWidth = MinContentWidth
		}
		a.listView.SetSize(a.contentWidth, msg.Height)
		a.detailView.SetSize(a.contentWidth, msg.Height)
		a.formView.SetSize(a.contentWidth, msg.Height)
		return a, nil

	case splashDoneMsg:
		a.viewState = ViewList
		return a, nil

	case fetchResultMsg:
		if msg.err != nil {
			a.urlInputView.SetError(fmt.Sprintf("Fetch failed: %v", msg.err))
			return a, nil
		}
		// Store the posting path and show success message
		a.fetchedPosting = msg.outputPath
		a.statusMsg = fmt.Sprintf("Fetched: %s @ %s", msg.position, msg.company)
		// Now launch Claude with the context
		return a, a.launchClaude()

	case claudeLaunchedMsg:
		if msg.err != nil {
			a.err = msg.err
			a.viewState = ViewList
			return a, nil
		}
		// Claude was launched successfully, return to list
		a.viewState = ViewList
		a.statusMsg = "Claude Code launched with agent context"
		return a, nil

	case tea.KeyMsg:
		// Skip splash on any key press
		if a.viewState == ViewSplash {
			a.viewState = ViewList
			return a, nil
		}
		return a.handleKey(msg)
	}

	return a, nil
}

func (a App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Clear status message on any key
	a.statusMsg = ""

	switch a.viewState {
	case ViewList:
		return a.handleListKey(msg)
	case ViewDetail:
		return a.handleDetailKey(msg)
	case ViewForm:
		return a.handleFormKey(msg)
	case ViewFilter:
		return a.handleFilterKey(msg)
	case ViewConfirmDelete:
		return a.handleDeleteConfirmKey(msg)
	case ViewURLInput:
		return a.handleURLInputKey(msg)
	}

	return a, nil
}

func (a App) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle search input first
	if a.listView.IsSearchMode() {
		switch msg.String() {
		case "enter":
			query := a.listView.SearchInput().Value()
			if query != "" {
				apps := a.store.Search(query)
				a.listView.SetApplications(apps)
			}
			handled, _ := a.listView.HandleKey(msg)
			if handled {
				return a, nil
			}
		case "esc":
			handled, _ := a.listView.HandleKey(msg)
			if handled {
				return a, nil
			}
		default:
			// Update search input
			input := a.listView.SearchInput()
			var cmd tea.Cmd
			newInput, cmd := input.Update(msg)
			a.listView.UpdateSearchInput(newInput)
			return a, cmd
		}
		return a, nil
	}

	handled, action := a.listView.HandleKey(msg)
	if handled {
		switch action {
		case "quit":
			return a, tea.Quit
		case "add":
			a.formView.Reset()
			a.prevState = a.viewState
			a.viewState = ViewForm
		case "edit":
			if app := a.listView.SelectedApplication(); app != nil {
				a.formView.SetApplication(app)
				a.prevState = a.viewState
				a.viewState = ViewForm
			}
		case "delete":
			if app := a.listView.SelectedApplication(); app != nil {
				a.deleteTarget = app
				a.viewState = ViewConfirmDelete
			}
		case "view":
			if app := a.listView.SelectedApplication(); app != nil {
				a.detailView.SetApplication(app)
				a.prevState = a.viewState
				a.viewState = ViewDetail
			}
		case "filter":
			a.filterCursor = 0
			a.viewState = ViewFilter
		case "urlinput":
			a.urlInputView.Reset()
			a.urlInputView.SetSize(a.width, a.height)
			a.viewState = ViewURLInput
		case "search":
			// Search was just performed, refresh handled above
		case "clear":
			a.listView.SetApplications(a.store.List())
			a.listView.SetFilterStatus("")
		default:
			if strings.HasPrefix(action, "status:") {
				status := strings.TrimPrefix(action, "status:")
				if app := a.listView.SelectedApplication(); app != nil {
					if err := a.store.UpdateStatus(app.ID, status); err == nil {
						a.refreshList()
						a.statusMsg = fmt.Sprintf("Changed status to %s", model.StatusLabel(status))
					}
				}
			}
		}
	}

	return a, nil
}

func (a App) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	handled, action := a.detailView.HandleKey(msg)
	if handled {
		switch action {
		case "quit":
			return a, tea.Quit
		case "back":
			a.viewState = ViewList
		case "edit":
			if a.detailView.application != nil {
				a.formView.SetApplication(a.detailView.application)
				a.prevState = a.viewState
				a.viewState = ViewForm
			}
		case "delete":
			if a.detailView.application != nil {
				a.deleteTarget = a.detailView.application
				a.viewState = ViewConfirmDelete
			}
		case "openfolder":
			if a.detailView.application != nil && a.detailView.application.DocumentsDir != "" {
				if err := openFolder(a.detailView.application.DocumentsDir); err != nil {
					a.err = err
				} else {
					a.statusMsg = "Opened documents folder"
				}
			} else {
				a.statusMsg = "No documents folder set for this application"
			}
		default:
			if strings.HasPrefix(action, "status:") {
				status := strings.TrimPrefix(action, "status:")
				if a.detailView.application != nil {
					if err := a.store.UpdateStatus(a.detailView.application.ID, status); err == nil {
						app, _ := a.store.GetByID(a.detailView.application.ID)
						a.detailView.SetApplication(&app)
						a.statusMsg = fmt.Sprintf("Changed status to %s", model.StatusLabel(status))
					}
				}
			}
		}
	}
	return a, nil
}

func (a App) handleFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	handled, action := a.formView.HandleKey(msg)
	if handled {
		switch action {
		case "cancel":
			a.viewState = a.prevState
		case "submit":
			if a.formView.Validate() {
				app := a.formView.GetApplication()
				if a.formView.isEdit {
					if err := a.store.Update(app); err == nil {
						a.statusMsg = "Application updated"
						a.refreshList()
						// Update detail view if we came from there
						if a.prevState == ViewDetail {
							updated, _ := a.store.GetByID(app.ID)
							a.detailView.SetApplication(&updated)
						}
						a.viewState = a.prevState
					}
				} else {
					if _, err := a.store.Add(app); err == nil {
						a.statusMsg = "Application added"
						a.refreshList()
						a.viewState = ViewList
					}
				}
			}
		}
		return a, nil
	}

	// Handle text input
	field := a.formView.FocusedField()
	if field != FieldStatus && field != FieldRemote {
		input := a.formView.FocusedInput()
		newInput, cmd := input.Update(msg)
		a.formView.UpdateInput(newInput)
		return a, cmd
	}

	return a, nil
}

func (a App) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Cancel):
		a.viewState = ViewList
	case key.Matches(msg, a.keys.Up):
		if a.filterCursor > 0 {
			a.filterCursor--
		}
	case key.Matches(msg, a.keys.Down):
		if a.filterCursor < len(a.filterOptions)-1 {
			a.filterCursor++
		}
	case key.Matches(msg, a.keys.Enter):
		if a.filterCursor == 0 {
			// "All" - clear filter
			a.listView.SetApplications(a.store.List())
			a.listView.SetFilterStatus("")
		} else {
			// Status filter
			statuses := model.AllStatuses()
			status := statuses[a.filterCursor-1]
			apps := a.store.FilterByStatus(status)
			a.listView.SetApplications(apps)
			a.listView.SetFilterStatus(status)
		}
		a.viewState = ViewList
	}
	return a, nil
}

func (a App) handleDeleteConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if a.deleteTarget != nil {
			if err := a.store.Delete(a.deleteTarget.ID); err == nil {
				a.statusMsg = fmt.Sprintf("Deleted %s @ %s", a.deleteTarget.Position, a.deleteTarget.Company)
				a.refreshList()
			}
		}
		a.deleteTarget = nil
		a.viewState = ViewList
	case "n", "N", "esc":
		a.deleteTarget = nil
		if a.prevState == ViewDetail {
			a.viewState = ViewDetail
		} else {
			a.viewState = ViewList
		}
	}
	return a, nil
}

// handleURLInputKey processes key input for the URL input dialog.
// It handles:
//   - Esc: Cancel and return to list view
//   - Enter: Validate URL and start fetch process
//   - Other keys: Update the text input
func (a App) handleURLInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Cancel):
		a.viewState = ViewList
		return a, nil
	case key.Matches(msg, a.keys.Enter):
		if a.urlInputView.Validate() {
			url := a.urlInputView.Value()
			return a, a.fetchPosting(url)
		}
		return a, nil
	}

	// Update text input
	input := a.urlInputView.Input()
	newInput, cmd := input.Update(msg)
	a.urlInputView.UpdateInput(newInput)
	return a, cmd
}

// fetchPosting fetches a job posting from the given URL.
// Returns a tea.Cmd that runs the fetch in the background and
// sends a fetchResultMsg when complete.
func (a App) fetchPosting(url string) tea.Cmd {
	return func() tea.Msg {
		// Create fetcher
		outputDir := "local/postings"
		f := fetch.NewFetcher(outputDir)

		// Fetch the posting
		result, err := f.Fetch(url, "")
		if err != nil {
			return fetchResultMsg{err: err}
		}

		return fetchResultMsg{
			outputPath: result.OutputPath,
			company:    result.Company,
			position:   result.Position,
		}
	}
}

// launchClaude launches Claude Code with the agent context.
// It uses tea.ExecProcess to properly handle the terminal handoff.
//
// The command spawned is:
//
//	claude --print "$(ghosted context)\n\nProcess the posting at {path} using the agent workflow."
//
// This gives Claude full context about the CV, postings, and prompts.
func (a App) launchClaude() tea.Cmd {
	// Build the prompt with context
	prompt := fmt.Sprintf(`Process the job posting at %s using the agent workflow.

Steps:
1. Read the posting file
2. Parse it to extract structured job data
3. Generate a tailored resume (Typst format)
4. Generate a personalized cover letter (Typst format)
5. Review the documents (must score 70+ to approve)
6. Add to the tracker with ghosted add --json

Use the prompts in internal/agent/prompts/ for each step.
Save generated documents to local/applications/{job-type}/{company}/`, a.fetchedPosting)

	// Get the absolute path to ghosted binary (or use 'ghosted' from PATH)
	ghostedBin := "ghosted"
	if exe, err := os.Executable(); err == nil {
		ghostedBin = exe
	}

	// Build the context command
	contextCmd := exec.Command(ghostedBin, "context")
	contextOutput, err := contextCmd.Output()
	if err != nil {
		// Fall back to simpler prompt without context
		contextOutput = []byte("(Context unavailable - run 'ghosted context' manually)")
	}

	// Build the full prompt
	fullPrompt := fmt.Sprintf("%s\n\n%s", string(contextOutput), prompt)

	// Find claude binary
	claudeBin, err := exec.LookPath("claude")
	if err != nil {
		return func() tea.Msg {
			return claudeLaunchedMsg{err: fmt.Errorf("claude not found in PATH: %w", err)}
		}
	}

	// Use tea.ExecProcess to properly hand off the terminal
	c := exec.Command(claudeBin, "--print", fullPrompt)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return claudeLaunchedMsg{err: err}
	})
}

func (a *App) refreshList() {
	if filter := a.listView.FilterStatus(); filter != "" {
		a.listView.SetApplications(a.store.FilterByStatus(filter))
	} else if query := a.listView.SearchQuery(); query != "" {
		a.listView.SetApplications(a.store.Search(query))
	} else {
		a.listView.SetApplications(a.store.List())
	}
}

// View renders the app
func (a App) View() string {
	var b strings.Builder

	switch a.viewState {
	case ViewSplash:
		b.WriteString(a.renderSplash())
	case ViewList:
		b.WriteString(a.listView.View())
	case ViewDetail:
		b.WriteString(a.detailView.View())
	case ViewForm:
		b.WriteString(a.formView.View())
	case ViewFilter:
		b.WriteString(a.renderFilterView())
	case ViewConfirmDelete:
		b.WriteString(a.renderDeleteConfirm())
	case ViewURLInput:
		a.urlInputView.SetSize(a.width, a.height)
		return a.urlInputView.View()
	}

	// Status message
	if a.statusMsg != "" {
		b.WriteString("\n")
		b.WriteString(SuccessStyle.Render(a.statusMsg))
	}

	// Error
	if a.err != nil {
		b.WriteString("\n")
		b.WriteString(ErrorStyle.Render(a.err.Error()))
	}

	content := b.String()

	// Center content both horizontally and vertically
	return lipgloss.Place(
		a.width,
		a.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func (a App) renderFilterView() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Filter by Status"))
	b.WriteString("\n\n")

	for i, opt := range a.filterOptions {
		cursor := "  "
		style := NormalRowStyle
		if i == a.filterCursor {
			cursor = "> "
			style = SelectedRowStyle
		}
		b.WriteString(style.Render(cursor + opt))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("%s %s  %s %s",
		HelpKeyStyle.Render("enter"),
		HelpDescStyle.Render("select"),
		HelpKeyStyle.Render("esc"),
		HelpDescStyle.Render("cancel"),
	))

	return b.String()
}

func (a App) renderDeleteConfirm() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Confirm Delete"))
	b.WriteString("\n\n")

	if a.deleteTarget != nil {
		b.WriteString(fmt.Sprintf("Are you sure you want to delete:\n\n"))
		b.WriteString(HighlightStyle.Render(a.deleteTarget.Position))
		b.WriteString(" @ ")
		b.WriteString(HighlightStyle.Render(a.deleteTarget.Company))
		b.WriteString("\n\n")
	}

	b.WriteString(fmt.Sprintf("%s %s  %s %s",
		HelpKeyStyle.Render("y"),
		HelpDescStyle.Render("yes, delete"),
		HelpKeyStyle.Render("n"),
		HelpDescStyle.Render("no, cancel"),
	))

	return b.String()
}

func (a App) renderSplash() string {
	ghost := `


         .-.
        (o o)  GHOSTED
        | O |  job application tracker
        |   |  for the perpetually ghosted
        '~~~'


`
	return SubtleStyle.Render(ghost)
}

// openFolder opens a folder in the system file manager.
// Works cross-platform: macOS (open), Windows (explorer), Linux (xdg-open).
func openFolder(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("explorer", path)
	default: // Linux and others
		cmd = exec.Command("xdg-open", path)
	}

	return cmd.Start()
}
