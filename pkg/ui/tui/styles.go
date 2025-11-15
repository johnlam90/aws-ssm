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
	// Primary colors - modern monochromatic with depth
	ColorPrimary   = lipgloss.Color("#E6EDF3") // Soft white - primary text (improved from harsh #FFFFFF)
	ColorSecondary = lipgloss.Color("#8B949E") // Modern gray - secondary text
	ColorMuted     = lipgloss.Color("#484F58") // Subtle gray - tertiary text

	// Functional accent colors - sophisticated and purposeful
	ColorAccentBlue   = lipgloss.Color("#58A6FF") // Modern blue - interactive elements, focus (GitHub's modern blue)
	ColorAccentGreen  = lipgloss.Color("#3FB950") // Vibrant green - success, active states
	ColorAccentAmber  = lipgloss.Color("#D29922") // Sophisticated amber - warnings, pending states
	ColorAccentRed    = lipgloss.Color("#F85149") // Modern red - errors, stopped states
	ColorAccentIndigo = lipgloss.Color("#A371F7") // Rich indigo - information, highlights

	// State colors - clear semantic meaning
	ColorRunning    = lipgloss.Color("#3FB950") // Green - active, healthy
	ColorStopped    = lipgloss.Color("#8B949E") // Gray - inactive, neutral
	ColorPending    = lipgloss.Color("#D29922") // Amber - transitional, attention
	ColorTerminated = lipgloss.Color("#484F58") // Dark gray - ended, archived

	// UI foundation - modern dark theme with depth
	ColorBorder     = lipgloss.Color("#30363D") // Subtle border - modern dark border
	ColorBackground = lipgloss.Color("#0D1117") // Deep dark - modern dark background
	ColorForeground = lipgloss.Color("#C9D1D9") // Soft white - readable foreground
	ColorSelected   = lipgloss.Color("#1F2937") // Modern slate - selection highlight
	ColorFocused    = lipgloss.Color("#1E3A8A") // Deep blue - focused element background

	// Modern CLI enhancement colors
	ColorGradientStart = lipgloss.Color("#58A6FF") // Blue gradient start
	ColorGradientEnd   = lipgloss.Color("#A371F7") // Purple gradient end
	ColorShadow        = lipgloss.Color("#161B22") // Subtle shadow color
	ColorHighlight     = lipgloss.Color("#21262D") // Modern highlight background
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

func (t *ModernTheme) Primary() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorPrimary
}

func (t *ModernTheme) Secondary() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorSecondary
}

func (t *ModernTheme) Muted() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorMuted
}

func (t *ModernTheme) AccentBlue() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorAccentBlue
}

func (t *ModernTheme) AccentGreen() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorAccentGreen
}

func (t *ModernTheme) AccentAmber() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorAccentAmber
}

func (t *ModernTheme) AccentRed() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorAccentRed
}

func (t *ModernTheme) AccentIndigo() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorAccentIndigo
}

func (t *ModernTheme) Running() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorRunning
}

func (t *ModernTheme) Stopped() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorStopped
}

func (t *ModernTheme) Pending() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorPending
}

func (t *ModernTheme) Terminated() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorTerminated
}

func (t *ModernTheme) Border() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorBorder
}

func (t *ModernTheme) Background() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorBackground
}

func (t *ModernTheme) Foreground() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorForeground
}

func (t *ModernTheme) Selected() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorSelected
}

func (t *ModernTheme) Focused() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorFocused
}

func (t *ModernTheme) GradientStart() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorGradientStart
}

func (t *ModernTheme) GradientEnd() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorGradientEnd
}

func (t *ModernTheme) Shadow() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorShadow
}

func (t *ModernTheme) Highlight() lipgloss.Color {
	if !t.colorsEnabled {
		return lipgloss.Color("")
	}
	return ColorHighlight
}

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

// Modern style functions that respect theme settings
func TitleStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Primary()).
		Bold(true)
}

func SubtitleStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Muted())
}

func MenuItemStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Foreground())
}

func MenuItemSelectedStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.AccentBlue()).
		Background(theme.Focused()).
		Bold(true)
}

func ListItemStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Foreground())
}

func ListItemSelectedStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.AccentBlue()).
		Bold(true)
}

func TableHeaderStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Secondary()).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(theme.Border()).
		Bold(true)
}

func StatusBarStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Muted())
}

func StatusBarKeyStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.AccentIndigo()).
		Bold(true)
}

func StatusBarValueStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Foreground())
}

func HelpStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Muted())
}

func HelpKeyStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.AccentIndigo()).
		Bold(true)
}

func HelpDescStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Secondary())
}

func BorderStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.Border())
}

func ErrorStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.AccentRed()).
		Bold(true)
}

func SuccessStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.AccentGreen()).
		Bold(true)
}

func LoadingStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Secondary())
}

func PanelStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.Border())
}

func ModalStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.AccentBlue()).
		Foreground(theme.Primary()).
		Background(theme.Selected()).
		Padding(1, 2)
}

func ModalTitleStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Primary()).
		Bold(true)
}

func ModalLabelStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Secondary()).
		Bold(true)
}

func ModalInputStyle() lipgloss.Style {
	theme := GetTheme()
	return lipgloss.NewStyle().
		Foreground(theme.Primary()).
		Background(theme.Focused()).
		Padding(0, 1).
		Bold(true)
}

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
