package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Minimal, professional color palette
var (
	// Primary colors - muted and professional
	ColorPrimary   = lipgloss.Color("#FFFFFF") // White
	ColorSecondary = lipgloss.Color("#A8A8A8") // Light Gray
	ColorAccent    = lipgloss.Color("#D0D0D0") // Subtle highlight
	ColorWarning   = lipgloss.Color("#C0C0C0") // Muted warning
	ColorDanger    = lipgloss.Color("#B0B0B0") // Muted danger
	ColorInfo      = lipgloss.Color("#E0E0E0") // Light info
	ColorMuted     = lipgloss.Color("#808080") // Gray

	// State colors - minimal and subtle
	ColorRunning    = lipgloss.Color("#90EE90") // Soft green
	ColorStopped    = lipgloss.Color("#D3D3D3") // Light gray
	ColorPending    = lipgloss.Color("#F0E68C") // Soft yellow
	ColorTerminated = lipgloss.Color("#A9A9A9") // Dark gray

	// UI colors - clean and minimal
	ColorBorder     = lipgloss.Color("#404040") // Subtle border
	ColorBackground = lipgloss.Color("#000000") // Black background
	ColorForeground = lipgloss.Color("#E0E0E0") // Light gray text
	ColorSelected   = lipgloss.Color("#303030") // Subtle selection
)

// Minimal styles for different components
var (
	// Title style - clean and simple
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary)

	// Subtitle style - subtle
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// Menu item styles - minimal
	MenuItemStyle = lipgloss.NewStyle().
			Foreground(ColorForeground)

	MenuItemSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Background(ColorSelected)

	// List item styles - clean
	ListItemStyle = lipgloss.NewStyle().
			Foreground(ColorForeground)

	ListItemSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Background(ColorSelected)

	// Header styles for tables - minimal
	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(ColorSecondary).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(ColorBorder)

	// Status bar style - clean
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StatusBarKeyStyle = lipgloss.NewStyle().
				Foreground(ColorSecondary)

	StatusBarValueStyle = lipgloss.NewStyle().
				Foreground(ColorForeground)

	// Help text style - minimal
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	// Border styles - simple lines only
	BorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(ColorBorder)

	// Error style - minimal
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorDanger)

	// Success style - minimal
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorAccent)

	// Loading style - subtle
	LoadingStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	// Panel style - no decorative borders
	PanelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(ColorBorder)
)

// GetStateColor returns the appropriate color for an instance state
func GetStateColor(state string) lipgloss.Color {
	switch state {
	case "running":
		return ColorRunning
	case "stopped":
		return ColorStopped
	case "pending", "stopping":
		return ColorPending
	case "terminated", "shutting-down":
		return ColorTerminated
	default:
		return ColorMuted
	}
}

// StateStyle returns a styled state string - minimal styling
func StateStyle(state string) string {
	color := GetStateColor(state)
	return lipgloss.NewStyle().Foreground(color).Render(state)
}
