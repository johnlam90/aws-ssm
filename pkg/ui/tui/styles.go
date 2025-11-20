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

// Claude/Codex/Droid CLI-inspired modern color palette
// Philosophy: Sophisticated, professional aesthetic with subtle depth and modern appeal
// Design principles:
// 1. Depth through subtle gradients and layered shadows
// 2. Consistent semantic meaning across all colors
// 3. High accessibility contrast ratios
// 4. Smooth transitions and animations
// 5. Contextual highlighting that guides attention naturally

var (
	// ColorPrimary is the primary text color (soft white)
	ColorPrimary = lipgloss.Color("#E6EDF3") // Soft white - primary text (improved from harsh #FFFFFF)
	// ColorSecondary is the secondary text color (modern gray)
	ColorSecondary = lipgloss.Color("#8B949E") // Modern gray - secondary text
	// ColorMuted is the tertiary text color (subtle gray)
	ColorMuted = lipgloss.Color("#484F58") // Subtle gray - tertiary text

	// ColorAccentBlue is the blue accent color for interactive elements
	ColorAccentBlue = lipgloss.Color("#58A6FF") // Modern blue - interactive elements, focus (GitHub's modern blue)
	// ColorAccentGreen is the green accent color for success states
	ColorAccentGreen = lipgloss.Color("#3FB950") // Vibrant green - success, active states
	// ColorAccentAmber is the amber accent color for warnings
	ColorAccentAmber = lipgloss.Color("#D29922") // Sophisticated amber - warnings, pending states
	// ColorAccentRed is the red accent color for errors
	ColorAccentRed = lipgloss.Color("#F85149") // Modern red - errors, stopped states
	// ColorAccentIndigo is the indigo accent color for information
	ColorAccentIndigo = lipgloss.Color("#A371F7") // Rich indigo - information, highlights

	// ColorRunning indicates a running state (green)
	ColorRunning = lipgloss.Color("#3FB950") // Green - active, healthy
	// ColorStopped indicates a stopped state (gray)
	ColorStopped = lipgloss.Color("#8B949E") // Gray - inactive, neutral
	// ColorPending indicates a pending state (amber)
	ColorPending = lipgloss.Color("#D29922") // Amber - transitional, attention
	// ColorTerminated indicates a terminated state (dark gray)
	ColorTerminated = lipgloss.Color("#484F58") // Dark gray - ended, archived

	// ColorBorder is the border color
	ColorBorder = lipgloss.Color("#30363D") // Subtle border - modern dark border
	// ColorBackground is the background color
	ColorBackground = lipgloss.Color("#0D1117") // Deep dark - modern dark background
	// ColorForeground is the foreground color
	ColorForeground = lipgloss.Color("#C9D1D9") // Soft white - readable foreground
	// ColorSelected is the selection highlight color
	ColorSelected = lipgloss.Color("#1F2937") // Modern slate - selection highlight
	// ColorFocused is the focused element background color
	ColorFocused = lipgloss.Color("#1E3A8A") // Deep blue - focused element background

	// ColorGradientStart is the start color for gradients
	ColorGradientStart = lipgloss.Color("#58A6FF") // Blue gradient start
	// ColorGradientEnd is the end color for gradients
	ColorGradientEnd = lipgloss.Color("#A371F7") // Purple gradient end
	// ColorShadow is the shadow color
	ColorShadow = lipgloss.Color("#161B22") // Subtle shadow color
	// ColorHighlight is the highlight background color
	ColorHighlight = lipgloss.Color("#21262D") // Modern highlight background
)

// Theme interface for managing color schemes and no-color support
type Theme interface {
	// Color accessors
	Primary() lipgloss.Color
	Secondary() lipgloss.Color
	Muted() lipgloss.Color
	AccentBlue() lipgloss.Color
	AccentGreen() lipgloss.Color
	AccentAmber() lipgloss.Color
	AccentRed() lipgloss.Color
	AccentIndigo() lipgloss.Color

	// State colors
	Running() lipgloss.Color
	Stopped() lipgloss.Color
	Pending() lipgloss.Color
	Terminated() lipgloss.Color

	// UI colors
	Border() lipgloss.Color
	Background() lipgloss.Color
	Foreground() lipgloss.Color
	Selected() lipgloss.Color
	Focused() lipgloss.Color

	// Enhanced colors
	GradientStart() lipgloss.Color
	GradientEnd() lipgloss.Color
	Shadow() lipgloss.Color
	Highlight() lipgloss.Color

	// Utility
	IsColorEnabled() bool
}

// ModernTheme implements the sophisticated color scheme
type ModernTheme struct {
	colorsEnabled bool
}

// NewModernTheme creates a new modern theme
func NewModernTheme(colorsEnabled bool) *ModernTheme {
	return &ModernTheme{colorsEnabled: colorsEnabled}
}

// Primary returns the primary color
func (t *ModernTheme) Primary() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorPrimary
}

// Secondary returns the secondary color
func (t *ModernTheme) Secondary() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorSecondary
}

// Muted returns the muted color
func (t *ModernTheme) Muted() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorMuted
}

// AccentBlue returns the blue accent color
func (t *ModernTheme) AccentBlue() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorAccentBlue
}

// AccentGreen returns the green accent color
func (t *ModernTheme) AccentGreen() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorAccentGreen
}

// AccentAmber returns the amber accent color
func (t *ModernTheme) AccentAmber() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorAccentAmber
}

// AccentRed returns the red accent color
func (t *ModernTheme) AccentRed() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorAccentRed
}

// AccentIndigo returns the indigo accent color
func (t *ModernTheme) AccentIndigo() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorAccentIndigo
}

// Running returns the running state color
func (t *ModernTheme) Running() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorRunning
}

// Stopped returns the stopped state color
func (t *ModernTheme) Stopped() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorStopped
}

// Pending returns the pending state color
func (t *ModernTheme) Pending() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorPending
}

// Terminated returns the terminated state color
func (t *ModernTheme) Terminated() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorTerminated
}

// Border returns the border color
func (t *ModernTheme) Border() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorBorder
}

// Background returns the background color
func (t *ModernTheme) Background() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorBackground
}

// Foreground returns the foreground color
func (t *ModernTheme) Foreground() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorForeground
}

// Selected returns the selected color
func (t *ModernTheme) Selected() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorSelected
}

// Focused returns the focused color
func (t *ModernTheme) Focused() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorFocused
}

// GradientStart returns the gradient start color
func (t *ModernTheme) GradientStart() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorGradientStart
}

// GradientEnd returns the gradient end color
func (t *ModernTheme) GradientEnd() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorGradientEnd
}

// Shadow returns the shadow color
func (t *ModernTheme) Shadow() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorShadow
}

// Highlight returns the highlight color
func (t *ModernTheme) Highlight() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorHighlight
}

// IsColorEnabled returns whether colors are enabled
func (t *ModernTheme) IsColorEnabled() bool {
	return t.colorsEnabled
}

// Global theme instance
var globalTheme Theme = NewModernTheme(true)

// SetTheme sets the global theme
func SetTheme(theme Theme) {
	globalTheme = theme
}

// GetTheme returns the global theme
func GetTheme() Theme {
	return globalTheme
}

// TitleStyle returns the style for titles
// Modern style functions that respect theme settings
func TitleStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Primary()).
		Bold(true)
}

// SubtitleStyle returns the style for subtitles
func SubtitleStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Muted())
}

// MenuItemStyle returns the style for menu items
func MenuItemStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Foreground())
}

// MenuItemSelectedStyle returns the style for selected menu items
func MenuItemSelectedStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.AccentBlue()).
		Background(theme.Focused()).
		Bold(true)
}

// ListItemStyle returns the style for list items
func ListItemStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Foreground())
}

// ListItemSelectedStyle returns the style for selected list items
func ListItemSelectedStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.AccentBlue()).
		Bold(true)
}

// TableHeaderStyle returns the style for table headers
func TableHeaderStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Secondary()).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(theme.Border()).
		Bold(true)
}

// StatusBarStyle returns the style for the status bar
func StatusBarStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Muted())
}

// StatusBarKeyStyle returns the style for status bar keys
func StatusBarKeyStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.AccentIndigo()).
		Bold(true)
}

// StatusBarValueStyle returns the style for status bar values
func StatusBarValueStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Foreground())
}

// HelpStyle returns the style for help text
func HelpStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Muted())
}

// SearchBarStyle returns the style for the search bar
func SearchBarStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Primary()).
		Background(theme.Selected()).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.Border()).
		Padding(0, 1)
}

// SearchBarActiveStyle returns the style for the active search bar
func SearchBarActiveStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Primary()).
		Background(theme.Highlight()).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.AccentBlue()).
		Padding(0, 1).
		Bold(true)
}

// HelpKeyStyle returns the style for help keys
func HelpKeyStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.AccentIndigo()).
		Bold(true)
}

// HelpDescStyle returns the style for help descriptions
func HelpDescStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Secondary())
}

// BorderStyle returns the style for borders
func BorderStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.Border())
}

// ErrorStyle returns the style for error messages
func ErrorStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.AccentRed()).
		Bold(true)
}

// SuccessStyle returns the style for success messages
func SuccessStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.AccentGreen()).
		Bold(true)
}

// LoadingStyle returns the style for loading text
func LoadingStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Secondary())
}

// PanelStyle returns the style for panels
func PanelStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.Border())
}

// ModalStyle returns the style for modals
func ModalStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.AccentBlue()).
		Foreground(theme.Primary()).
		Background(theme.Selected()).
		Padding(1, 2)
}

// ModalTitleStyle returns the style for modal titles
func ModalTitleStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Primary()).
		Bold(true)
}

// ModalLabelStyle returns the style for modal labels
func ModalLabelStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Secondary()).
		Bold(true)
}

// ModalInputStyle returns the style for modal inputs
func ModalInputStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Primary()).
		Background(theme.Focused()).
		Padding(0, 1).
		Bold(true)
}

// ModalPlaceholderStyle returns the style for modal placeholders
func ModalPlaceholderStyle() lipgloss.Style {
	theme := GetTheme()
	return ModalInputStyle().
		Foreground(theme.Secondary()).
		Faint(true).
		Italic(true)
}

// ModalHelpStyle returns the style for modal help text
func ModalHelpStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.AccentIndigo())
}

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
	return ListItemSelectedStyle().Render(row)
}

// RenderStatusMessage renders a status message with appropriate semantic color
func RenderStatusMessage(message string, messageType string) string {
	var style lipgloss.Style

	switch messageType {
	case "success":
		style = SuccessStyle()
	case "error":
		style = ErrorStyle()
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

// Beautiful Dashboard Styling Functions

// DashboardBannerStyle returns the ASCII banner styling
func DashboardBannerStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.AccentBlue()).
		Bold(true).
		Align(lipgloss.Center)
}

// DashboardWelcomeStyle returns the welcome message styling
func DashboardWelcomeStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Primary()).
		Bold(true)
}

// DashboardInfoStyle returns the system info styling
func DashboardInfoStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Secondary()).
		Italic(true)
}

// DashboardTitleStyle returns the main title styling for the beautiful dashboard
func DashboardTitleStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.AccentBlue()).
		Bold(true).
		MarginBottom(0)
}

// DashboardSubtitleStyle returns the subtitle/context styling
func DashboardSubtitleStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Secondary()).
		MarginBottom(1)
}

// DashboardHeaderBarStyle returns the header bar with context information
func DashboardHeaderBarStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Muted()).
		MarginBottom(1)
}

// DashboardSeparatorStyle returns the horizontal separator line
func DashboardSeparatorStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Border()).
		MarginTop(1).
		MarginBottom(1)
}

// DashboardSectionTitleStyle returns the section title styling
func DashboardSectionTitleStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Primary()).
		Bold(true).
		MarginBottom(1)
}

// DashboardServiceNameStyle returns the service name column styling
func DashboardServiceNameStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Primary()).
		Width(22). // Fixed width for alignment
		Bold(true)
}

// DashboardServiceDescStyle returns the service description column styling
func DashboardServiceDescStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Secondary()).
		Width(40) // Fixed width for alignment
}

// DashboardSelectionBarStyle returns the vertical selection bar styling
func DashboardSelectionBarStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.AccentBlue()).
		Bold(true)
}

// DashboardSelectedNameStyle returns the selected service name styling
func DashboardSelectedNameStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.AccentBlue()).
		Background(theme.Highlight()).
		Bold(true).
		Width(22) // Same width as normal for alignment
}

// DashboardSelectedDescStyle returns the selected service description styling
func DashboardSelectedDescStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.AccentBlue()).
		Background(theme.Highlight()).
		Width(40) // Same width as normal for alignment
}

// DashboardFooterStyle returns the footer styling
func DashboardFooterStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Primary()).
		Bold(true).
		MarginTop(1)
}

// DashboardFooterDetailStyle returns the footer detail styling
func DashboardFooterDetailStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Secondary())
}

// DashboardFooterKeyStyle returns the footer key styling
func DashboardFooterKeyStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.AccentIndigo()).
		Bold(true)
}

// DashboardFooterActionStyle returns the footer action description styling
func DashboardFooterActionStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Secondary())
}
