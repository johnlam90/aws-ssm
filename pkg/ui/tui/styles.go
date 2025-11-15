package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Design Philosophy: Jony Ive's Minimalist Approach
//
// "Simplicity is not the absence of clutter, that's a consequence of simplicity.
// Simplicity is somehow essentially describing the purpose and place of an object
// and product. The absence of clutter is just a clutter-free product. That's not simple."
// - Jony Ive
//
// Color Strategy:
// 1. PURPOSEFUL: Color is used only where it adds functional value
// 2. HIERARCHY: Color creates visual hierarchy and guides user attention
// 3. RESTRAINT: Muted, refined palette prevents visual fatigue
// 4. ACCESSIBILITY: High contrast ensures readability
// 5. CONSISTENCY: Each color has a specific semantic meaning
//
// Color Meanings:
// - Blue (#60A5FA): Interactive elements, focus, selection - "you can act here"
// - Green (#34D399): Success, active/running states - "healthy and operational"
// - Amber (#FBBF24): Warnings, pending states - "attention needed"
// - Red (#F87171): Errors, stopped states - "issue or inactive"
// - Indigo (#818CF8): Information, keybindings - "helpful reference"
// - Gray scale: Content hierarchy from white (primary) to muted gray (tertiary)

// Jony Ive-inspired minimal color palette
// Philosophy: Color used sparingly and purposefully, only where it adds functional value
// Palette: Muted, refined tones with high contrast for accessibility
var (
	// Primary colors - refined monochromatic base
	ColorPrimary   = lipgloss.Color("#FFFFFF") // Pure white - primary text
	ColorSecondary = lipgloss.Color("#9CA3AF") // Cool gray - secondary text
	ColorMuted     = lipgloss.Color("#6B7280") // Muted gray - tertiary text

	// Functional accent colors - subtle and purposeful
	// These colors guide attention and communicate state
	ColorAccentBlue   = lipgloss.Color("#60A5FA") // Soft blue - interactive elements, focus
	ColorAccentGreen  = lipgloss.Color("#34D399") // Refined green - success, active states
	ColorAccentAmber  = lipgloss.Color("#FBBF24") // Warm amber - warnings, pending states
	ColorAccentRed    = lipgloss.Color("#F87171") // Soft red - errors, stopped states
	ColorAccentIndigo = lipgloss.Color("#818CF8") // Subtle indigo - information, highlights

	// State colors - communicate status at a glance
	ColorRunning    = lipgloss.Color("#34D399") // Green - active, healthy
	ColorStopped    = lipgloss.Color("#9CA3AF") // Gray - inactive, neutral
	ColorPending    = lipgloss.Color("#FBBF24") // Amber - transitional, attention
	ColorTerminated = lipgloss.Color("#6B7280") // Muted gray - ended, archived

	// UI foundation - clean, minimal structure
	ColorBorder     = lipgloss.Color("#374151") // Subtle border - defines space without distraction
	ColorBackground = lipgloss.Color("#000000") // Pure black - maximum contrast
	ColorForeground = lipgloss.Color("#E5E7EB") // Light gray - readable, not harsh
	ColorSelected   = lipgloss.Color("#1E293B") // Slate gray - subtle but visible selection highlight
	ColorFocused    = lipgloss.Color("#1E3A8A") // Deep blue - focused element background
)

// Styles following Jony Ive's design principles:
// - Purposeful use of color to create hierarchy and guide attention
// - Subtle accents that enhance usability without distraction
// - High contrast for accessibility
var (
	// Title style - pure white for maximum hierarchy
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	// Subtitle style - muted gray for supporting information
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// Menu item styles - subtle blue accent on selection for visual feedback
	MenuItemStyle = lipgloss.NewStyle().
			Foreground(ColorForeground)

	MenuItemSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorAccentBlue). // Blue indicates interactivity
				Background(ColorFocused).    // Deep blue background for focus
				Bold(true)

	// List item styles - blue accent guides the eye to selected item
	ListItemStyle = lipgloss.NewStyle().
			Foreground(ColorForeground)

	ListItemSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#60A5FA")). // Bright blue for visibility
				Bold(true)

	// Header styles for tables - refined gray creates visual separation
	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(ColorSecondary).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(ColorBorder).
				Bold(true)

	// Status bar style - muted to stay out of the way
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// Status bar key style - indigo accent for keybindings (functional color)
	StatusBarKeyStyle = lipgloss.NewStyle().
				Foreground(ColorAccentIndigo).
				Bold(true)

	StatusBarValueStyle = lipgloss.NewStyle().
				Foreground(ColorForeground)

	// Help text style - muted gray for non-critical information
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// Help key style - indigo accent makes keybindings scannable
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorAccentIndigo).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	// Border styles - subtle, defines space without visual noise
	BorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(ColorBorder)

	// Error style - soft red communicates issues clearly
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorAccentRed).
			Bold(true)

	// Success style - refined green for positive feedback
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorAccentGreen).
			Bold(true)

	// Loading style - muted to indicate passive state
	LoadingStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	// Panel style - minimal borders for clean structure
	PanelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(ColorBorder)

	// Modal styles - used for focused interactions like scaling prompts
	ModalStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorAccentBlue).
			Foreground(ColorPrimary).
			Background(ColorSelected).
			Padding(1, 2)

	ModalTitleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	ModalLabelStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true)

	ModalInputStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Background(ColorFocused).
			Padding(0, 1).
			Bold(true)

	ModalHelpStyle = lipgloss.NewStyle().
			Foreground(ColorAccentIndigo)
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
	trimmed := strings.TrimSpace(state)
	color := GetStateColor(strings.ToLower(trimmed))
	return lipgloss.NewStyle().Foreground(color).Render(trimmed)
}

// RenderStateCell renders a padded state string so table columns stay aligned
func RenderStateCell(state string, width int) string {
	normalized := strings.ToLower(strings.TrimSpace(state))
	color := GetStateColor(normalized)
	text := fmt.Sprintf("%-*s", width, normalized)
	return lipgloss.NewStyle().Foreground(color).Render(text)
}

// RenderSelectableRow safely renders a row with/without selection indicator
// Selected rows use bright color to indicate selection
func RenderSelectableRow(row string, selected bool) string {
	if !selected {
		return row
	}

	// Apply bright blue color to the entire row
	return ListItemSelectedStyle.Render(row)
}

// RenderStatusMessage renders a status message with appropriate semantic color
func RenderStatusMessage(message string, messageType string) string {
	var style lipgloss.Style

	switch messageType {
	case "success":
		style = SuccessStyle
	case "error":
		style = ErrorStyle
	case "warning":
		style = lipgloss.NewStyle().Foreground(ColorAccentAmber).Bold(true)
	case "info":
		style = lipgloss.NewStyle().Foreground(ColorAccentIndigo)
	default:
		style = lipgloss.NewStyle().Foreground(ColorForeground)
	}

	return style.Render(message)
}

// RenderMetric renders a key metric with subtle color accent for emphasis
func RenderMetric(label string, value string, highlight bool) string {
	labelStyle := lipgloss.NewStyle().Foreground(ColorSecondary)
	valueStyle := lipgloss.NewStyle().Foreground(ColorForeground)

	if highlight {
		// Use blue accent for important metrics
		valueStyle = lipgloss.NewStyle().Foreground(ColorAccentBlue).Bold(true)
	}

	return labelStyle.Render(label) + " " + valueStyle.Render(value)
}
