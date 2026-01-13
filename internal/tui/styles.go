package tui

import (
	"jobtrack/internal/model"

	"github.com/charmbracelet/lipgloss"
)

// Colors
var (
	colorPrimary   = lipgloss.Color("#7D56F4")
	colorSecondary = lipgloss.Color("#5B4B8A")
	colorSuccess   = lipgloss.Color("#73D673")
	colorWarning   = lipgloss.Color("#FFD93D")
	colorDanger    = lipgloss.Color("#FF6B6B")
	colorMuted     = lipgloss.Color("#6C6C6C")
	colorHighlight = lipgloss.Color("#FFF")
	colorBg        = lipgloss.Color("#1A1A2E")
	colorBgAlt     = lipgloss.Color("#252541")
)

// Status colors
var statusColors = map[string]lipgloss.Color{
	model.StatusApplied:   lipgloss.Color("#4ECDC4"),
	model.StatusScreening: lipgloss.Color("#FFE66D"),
	model.StatusInterview: lipgloss.Color("#A8E6CF"),
	model.StatusOffer:     lipgloss.Color("#95E1D3"),
	model.StatusAccepted:  lipgloss.Color("#73D673"),
	model.StatusRejected:  lipgloss.Color("#FF6B6B"),
	model.StatusWithdrawn: lipgloss.Color("#6C6C6C"),
}

// GetStatusColor returns the color for a given status
func GetStatusColor(status string) lipgloss.Color {
	if color, ok := statusColors[status]; ok {
		return color
	}
	return colorMuted
}

// Common styles
var (
	// Title style for headers
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1)

	// Subtle text style
	SubtleStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	// Highlight style for selected items
	HighlightStyle = lipgloss.NewStyle().
			Foreground(colorHighlight).
			Bold(true)

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(colorDanger)

	// Success style
	SuccessStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)

	// Warning style
	WarningStyle = lipgloss.NewStyle().
			Foreground(colorWarning)
)

// List view styles
var (
	// Selected row style
	SelectedRowStyle = lipgloss.NewStyle().
				Background(colorBgAlt).
				Foreground(colorHighlight).
				Bold(true)

	// Normal row style
	NormalRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC"))

	// Header style for tables
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(colorSecondary)
)

// Status badge style
func StatusBadgeStyle(status string) lipgloss.Style {
	color := GetStatusColor(status)
	return lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		Padding(0, 1)
}

// Detail view styles
var (
	// Label style for field names
	LabelStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Width(15)

	// Value style for field values
	ValueStyle = lipgloss.NewStyle().
			Foreground(colorHighlight)

	// Section header style
	SectionStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			MarginTop(1).
			MarginBottom(1).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(colorSecondary)

	// Box style for detail panels
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorSecondary).
			Padding(1, 2)
)

// Form styles
var (
	// Focused input style
	FocusedInputStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(colorPrimary)

	// Blurred input style
	BlurredInputStyle = lipgloss.NewStyle().
				Foreground(colorMuted).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(colorMuted)

	// Button style
	ButtonStyle = lipgloss.NewStyle().
			Foreground(colorHighlight).
			Background(colorPrimary).
			Padding(0, 2).
			MarginRight(1)

	// Disabled button style
	DisabledButtonStyle = lipgloss.NewStyle().
				Foreground(colorMuted).
				Background(colorBgAlt).
				Padding(0, 2).
				MarginRight(1)
)

// Help styles
var (
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	HelpSepStyle = lipgloss.NewStyle().
			Foreground(colorSecondary)
)

// Status bar style
var StatusBarStyle = lipgloss.NewStyle().
	Foreground(colorMuted).
	Background(colorBgAlt).
	Padding(0, 1)
