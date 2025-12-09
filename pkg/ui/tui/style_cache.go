package tui

import (
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// CachedStyles holds pre-computed styles to avoid repeated allocations during rendering.
// Styles are computed once when the theme is set and reused for all subsequent renders.
type CachedStyles struct {
	// Basic text styles
	Title              lipgloss.Style
	Subtitle           lipgloss.Style
	MenuItem           lipgloss.Style
	MenuItemSelected   lipgloss.Style
	ListItem           lipgloss.Style
	ListItemSelected   lipgloss.Style

	// Table styles
	TableHeader lipgloss.Style

	// Status bar styles
	StatusBar      lipgloss.Style
	StatusBarKey   lipgloss.Style
	StatusBarValue lipgloss.Style

	// Help styles
	Help     lipgloss.Style
	HelpKey  lipgloss.Style
	HelpDesc lipgloss.Style

	// Search styles
	SearchBar       lipgloss.Style
	SearchBarActive lipgloss.Style

	// Border and panel styles
	Border lipgloss.Style
	Panel  lipgloss.Style

	// Message styles
	Error   lipgloss.Style
	Success lipgloss.Style
	Warning lipgloss.Style
	Info    lipgloss.Style
	Loading lipgloss.Style

	// Modal styles
	Modal            lipgloss.Style
	ModalTitle       lipgloss.Style
	ModalLabel       lipgloss.Style
	ModalInput       lipgloss.Style
	ModalPlaceholder lipgloss.Style
	ModalHelp        lipgloss.Style

	// Dashboard styles
	DashboardTitle        lipgloss.Style
	DashboardSubtitle     lipgloss.Style
	DashboardHeaderBar    lipgloss.Style
	DashboardSeparator    lipgloss.Style
	DashboardSectionTitle lipgloss.Style
	DashboardServiceName  lipgloss.Style
	DashboardServiceDesc  lipgloss.Style
	DashboardSelectionBar lipgloss.Style
	DashboardSelectedName lipgloss.Style
	DashboardSelectedDesc lipgloss.Style
	DashboardFooter       lipgloss.Style
	DashboardFooterKey    lipgloss.Style
	DashboardFooterAction lipgloss.Style
}

var (
	cachedStyles     *CachedStyles
	cachedStylesMu   sync.RWMutex
	styleCacheInited bool
)

// initStyleCache initializes all cached styles based on the current theme.
// This should be called whenever the theme changes.
func initStyleCache(theme Theme) {
	cachedStylesMu.Lock()
	defer cachedStylesMu.Unlock()

	cachedStyles = &CachedStyles{
		// Basic text styles
		Title: lipgloss.NewStyle().
			Foreground(theme.Primary()).
			Bold(true),
		Subtitle: lipgloss.NewStyle().
			Foreground(theme.Muted()),
		MenuItem: lipgloss.NewStyle().
			Foreground(theme.Foreground()),
		MenuItemSelected: lipgloss.NewStyle().
			Foreground(theme.AccentBlue()).
			Background(theme.Focused()).
			Bold(true),
		ListItem: lipgloss.NewStyle().
			Foreground(theme.Foreground()),
		ListItemSelected: lipgloss.NewStyle().
			Foreground(theme.AccentBlue()).
			Bold(true),

		// Table styles
		TableHeader: lipgloss.NewStyle().
			Foreground(theme.Secondary()).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(theme.Border()).
			Bold(true),

		// Status bar styles
		StatusBar: lipgloss.NewStyle().
			Foreground(theme.Muted()),
		StatusBarKey: lipgloss.NewStyle().
			Foreground(theme.AccentIndigo()).
			Bold(true),
		StatusBarValue: lipgloss.NewStyle().
			Foreground(theme.Foreground()),

		// Help styles
		Help: lipgloss.NewStyle().
			Foreground(theme.Muted()),
		HelpKey: lipgloss.NewStyle().
			Foreground(theme.AccentIndigo()).
			Bold(true),
		HelpDesc: lipgloss.NewStyle().
			Foreground(theme.Secondary()),

		// Search styles
		SearchBar: lipgloss.NewStyle().
			Foreground(theme.Primary()).
			Background(theme.Selected()).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(theme.Border()).
			Padding(0, 1),
		SearchBarActive: lipgloss.NewStyle().
			Foreground(theme.Primary()).
			Background(theme.Highlight()).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(theme.AccentBlue()).
			Padding(0, 1).
			Bold(true),

		// Border and panel styles
		Border: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(theme.Border()),
		Panel: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(theme.Border()),

		// Message styles
		Error: lipgloss.NewStyle().
			Foreground(theme.AccentRed()).
			Bold(true),
		Success: lipgloss.NewStyle().
			Foreground(theme.AccentGreen()).
			Bold(true),
		Warning: lipgloss.NewStyle().
			Foreground(theme.AccentAmber()).
			Bold(true),
		Info: lipgloss.NewStyle().
			Foreground(theme.AccentIndigo()),
		Loading: lipgloss.NewStyle().
			Foreground(theme.Secondary()),

		// Modal styles
		Modal: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.AccentBlue()).
			Foreground(theme.Primary()).
			Background(theme.Selected()).
			Padding(1, 2),
		ModalTitle: lipgloss.NewStyle().
			Foreground(theme.Primary()).
			Bold(true),
		ModalLabel: lipgloss.NewStyle().
			Foreground(theme.Secondary()).
			Bold(true),
		ModalInput: lipgloss.NewStyle().
			Foreground(theme.Primary()).
			Background(theme.Focused()).
			Padding(0, 1).
			Bold(true),
		ModalPlaceholder: lipgloss.NewStyle().
			Foreground(theme.Secondary()).
			Background(theme.Focused()).
			Padding(0, 1).
			Faint(true).
			Italic(true),
		ModalHelp: lipgloss.NewStyle().
			Foreground(theme.AccentIndigo()),

		// Dashboard styles
		DashboardTitle: lipgloss.NewStyle().
			Foreground(theme.AccentBlue()).
			Bold(true).
			MarginBottom(0),
		DashboardSubtitle: lipgloss.NewStyle().
			Foreground(theme.Secondary()).
			MarginBottom(1),
		DashboardHeaderBar: lipgloss.NewStyle().
			Foreground(theme.Muted()).
			MarginBottom(1),
		DashboardSeparator: lipgloss.NewStyle().
			Foreground(theme.Border()).
			MarginTop(1).
			MarginBottom(1),
		DashboardSectionTitle: lipgloss.NewStyle().
			Foreground(theme.Primary()).
			Bold(true).
			MarginBottom(1),
		DashboardServiceName: lipgloss.NewStyle().
			Foreground(theme.Primary()).
			Width(22).
			Bold(true),
		DashboardServiceDesc: lipgloss.NewStyle().
			Foreground(theme.Secondary()).
			Width(40),
		DashboardSelectionBar: lipgloss.NewStyle().
			Foreground(theme.AccentBlue()).
			Bold(true),
		DashboardSelectedName: lipgloss.NewStyle().
			Foreground(theme.AccentBlue()).
			Background(theme.Highlight()).
			Bold(true).
			Width(22),
		DashboardSelectedDesc: lipgloss.NewStyle().
			Foreground(theme.AccentBlue()).
			Background(theme.Highlight()).
			Width(40),
		DashboardFooter: lipgloss.NewStyle().
			Foreground(theme.Muted()).
			MarginTop(1),
		DashboardFooterKey: lipgloss.NewStyle().
			Foreground(theme.AccentIndigo()).
			Bold(true),
		DashboardFooterAction: lipgloss.NewStyle().
			Foreground(theme.Secondary()),
	}

	styleCacheInited = true
}

// GetCachedStyles returns the cached styles, initializing them if necessary.
// This is the primary access point for cached styles.
func GetCachedStyles() *CachedStyles {
	cachedStylesMu.RLock()
	if styleCacheInited && cachedStyles != nil {
		defer cachedStylesMu.RUnlock()
		return cachedStyles
	}
	cachedStylesMu.RUnlock()

	// Initialize with current theme if not already done
	initStyleCache(GetTheme())

	cachedStylesMu.RLock()
	defer cachedStylesMu.RUnlock()
	return cachedStyles
}

// RefreshStyleCache forces a refresh of the cached styles.
// Call this when the theme changes.
func RefreshStyleCache() {
	initStyleCache(GetTheme())
}
