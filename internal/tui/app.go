package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/celloopa/ghosted/internal/model"
	"github.com/celloopa/ghosted/internal/store"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
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
)

// splashDoneMsg signals the splash screen is done
type splashDoneMsg struct{}

// App is the main Bubble Tea model
type App struct {
	store      *store.Store
	keys       KeyMap
	width      int
	height     int
	viewState  ViewState
	prevState  ViewState
	err        error
	statusMsg  string

	// Views
	listView   ListView
	detailView DetailView
	formView   FormView

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

	return App{
		store:      s,
		keys:       keys,
		viewState:  ViewSplash,
		listView:   listView,
		detailView: detailView,
		formView:   formView,
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
		a.listView.SetSize(msg.Width, msg.Height)
		a.detailView.SetSize(msg.Width, msg.Height)
		a.formView.SetSize(msg.Width, msg.Height)
		return a, nil

	case splashDoneMsg:
		a.viewState = ViewList
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

	return b.String()
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
