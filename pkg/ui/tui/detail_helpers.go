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

func countRenderedLines(s string) int {
	s = strings.TrimSuffix(s, "\n")
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

func limitRenderedLines(s string, maxLines int) string {
	s = strings.TrimSuffix(s, "\n")
	if s == "" || maxLines <= 0 {
		return ""
	}

	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}
	if maxLines == 1 {
		return "  ..."
	}

	limited := append([]string{}, lines[:maxLines-1]...)
	limited = append(limited, "  ...")
	return strings.Join(limited, "\n")
}

func calculateTableRows(terminalHeight, fixedLines int, details string) int {
	if terminalHeight <= 0 {
		return 5
	}
	rows := terminalHeight - fixedLines - countRenderedLines(details)
	if rows < 1 {
		return 1
	}
	return rows
}

func calculateBoundedVisibleRange(total, cursor, visibleHeight int) (int, int) {
	if total <= 0 {
		return 0, 0
	}
	if visibleHeight < 1 {
		visibleHeight = 1
	}
	if visibleHeight >= total {
		return 0, total
	}

	start := cursor - visibleHeight/2
	if start < 0 {
		start = 0
	}
	end := start + visibleHeight
	if end > total {
		end = total
		start = end - visibleHeight
		if start < 0 {
			start = 0
		}
	}
	return start, end
}
