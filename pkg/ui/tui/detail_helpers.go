package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// formatRelativeTimestamp renders a timestamp plus its relative age.
func formatRelativeTimestamp(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	return fmt.Sprintf("%s (%s ago)", t.Format("2006-01-02 15:04:05"), humanDuration(time.Since(t)))
}

// humanDuration converts a duration into a concise human readable string.
func humanDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}

	if len(parts) == 0 {
		return "<1m"
	}

	return strings.Join(parts, " ")
}

// renderTagLines returns sorted tag lines while skipping the provided keys.
func renderTagLines(tags map[string]string, skipKeys ...string) []string {
	if len(tags) == 0 {
		return nil
	}

	skip := make(map[string]struct{}, len(skipKeys))
	for _, key := range skipKeys {
		skip[strings.ToLower(strings.TrimSpace(key))] = struct{}{}
	}

	keys := make([]string, 0, len(tags))
	for key := range tags {
		if _, ignored := skip[strings.ToLower(key)]; ignored {
			continue
		}
		keys = append(keys, key)
	}

	sort.Strings(keys)
	if len(keys) == 0 {
		return nil
	}

	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("  %s=%s", key, tags[key]))
	}
	return lines
}

// KeyBinding represents a key-description pair for footer rendering
type FooterKey struct {
	Key  string
	Desc string
}

// RenderFooter creates a styled footer from a slice of key bindings.
// This consolidates the common footer rendering pattern used across all views.
func RenderFooter(keys []FooterKey) string {
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		keyStyle := StatusBarKeyStyle().Render(k.Key)
		descStyle := StatusBarValueStyle().Render(k.Desc)
		parts = append(parts, fmt.Sprintf("%s %s", keyStyle, descStyle))
	}
	return HelpStyle().Render(strings.Join(parts, " • "))
}

// CommonNavigationKeys returns the standard navigation keybindings
func CommonNavigationKeys() []FooterKey {
	return []FooterKey{
		{"↑/k", "up"},
		{"↓/j", "down"},
		{"g/G", "top/bottom"},
	}
}

// CommonFooterKeys returns the standard footer keys appended with common actions
func CommonFooterKeys(enterAction string) []FooterKey {
	keys := CommonNavigationKeys()
	return append(keys,
		FooterKey{"enter", enterAction},
		FooterKey{"r", "refresh"},
		FooterKey{"/", "search"},
		FooterKey{"esc", "back"},
	)
}
