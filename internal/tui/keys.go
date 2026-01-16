package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines all keybindings for the application
type KeyMap struct {
	// Navigation
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Top    key.Binding
	Bottom key.Binding

	// Actions
	Add    key.Binding
	Edit   key.Binding
	Delete key.Binding
	Enter  key.Binding
	Back   key.Binding

	// Status shortcuts
	Status1 key.Binding
	Status2 key.Binding
	Status3 key.Binding
	Status4 key.Binding
	Status5 key.Binding
	Status6 key.Binding
	Status7 key.Binding
	Status8 key.Binding

	// Search and filter
	Search key.Binding
	Filter key.Binding
	Clear  key.Binding

	// URL input for job posting fetch
	URLInput key.Binding

	// Open documents folder
	OpenFolder key.Binding

	// General
	Help key.Binding
	Quit key.Binding
	Tab  key.Binding

	// Form specific
	Submit key.Binding
	Cancel key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Navigation
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("k/up", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("j/down", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("h/left", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("l/right", "right"),
		),
		Top: key.NewBinding(
			key.WithKeys("g", "home"),
			key.WithHelp("g", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G", "end"),
			key.WithHelp("G", "bottom"),
		),

		// Actions
		Add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "view"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "backspace"),
			key.WithHelp("esc", "back"),
		),

		// Status shortcuts (1-8 for quick status change)
		Status1: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "saved"),
		),
		Status2: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "applied"),
		),
		Status3: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "screening"),
		),
		Status4: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "interview"),
		),
		Status5: key.NewBinding(
			key.WithKeys("5"),
			key.WithHelp("5", "offer"),
		),
		Status6: key.NewBinding(
			key.WithKeys("6"),
			key.WithHelp("6", "accepted"),
		),
		Status7: key.NewBinding(
			key.WithKeys("7"),
			key.WithHelp("7", "rejected"),
		),
		Status8: key.NewBinding(
			key.WithKeys("8"),
			key.WithHelp("8", "withdrawn"),
		),

		// Search and filter
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter"),
		),
		Clear: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "clear filter"),
		),

		// URL input for job posting fetch
		URLInput: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "fetch URL"),
		),

		// Open documents folder
		OpenFolder: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open folder"),
		),

		// General
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next field"),
		),

		// Form specific
		Submit: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

// ShortHelp returns keybindings to show in the mini help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Add, k.URLInput, k.Enter, k.Filter, k.Search, k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view (excludes status keys shown separately)
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Top, k.Bottom},
		{k.Add, k.Edit, k.Delete, k.Enter},
		{k.Search, k.Filter, k.Clear},
		{k.URLInput, k.Help, k.Quit},
	}
}
