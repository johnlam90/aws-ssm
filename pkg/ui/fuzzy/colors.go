package fuzzy

import (
	"fmt"
	"strings"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
)

// DefaultColorManager implements ColorManager interface
type DefaultColorManager struct {
	noColor bool
}

// NewDefaultColorManager creates a new color manager
func NewDefaultColorManager(noColor bool) *DefaultColorManager {
	return &DefaultColorManager{noColor: noColor}
}

// StateColor returns colorized state text
func (c *DefaultColorManager) StateColor(state string) string {
	if c.noColor {
		return state
	}

	switch strings.ToLower(state) {
	case "running":
		return ColorGreen + state + ColorReset
	case "stopped":
		return ColorRed + state + ColorReset
	case "pending":
		return ColorYellow + state + ColorReset
	case "stopping", "shutting-down":
		return ColorYellow + state + ColorReset
	case "terminated":
		return ColorRed + state + ColorReset
	default:
		return state
	}
}

// HighlightText highlights matching substrings in text
func (c *DefaultColorManager) HighlightText(text, query string) string {
	if c.noColor || query == "" {
		return text
	}

	queryLower := strings.ToLower(query)
	textLower := strings.ToLower(text)

	var result strings.Builder
	textRunes := []rune(text)
	textLowerRunes := []rune(textLower)
	queryRunes := []rune(queryLower)

	i := 0
	for i < len(textRunes) {
		// Check if query matches at current position
		if i+len(queryRunes) <= len(textRunes) {
			match := true
			for j := 0; j < len(queryRunes); j++ {
				if textLowerRunes[i+j] != queryRunes[j] {
					match = false
					break
				}
			}

			if match {
				// Highlight the match
				result.WriteString(ColorCyan)
				result.WriteString(string(textRunes[i : i+len(queryRunes)]))
				result.WriteString(ColorReset)
				i += len(queryRunes)
				continue
			}
		}

		// No match, add current character
		result.WriteRune(textRunes[i])
		i++
	}

	return result.String()
}

// HeaderColor returns colorized header text
func (c *DefaultColorManager) HeaderColor(text string) string {
	if c.noColor {
		return text
	}
	return ColorBold + ColorBlue + text + ColorReset
}

// TagColor returns colorized tag text
func (c *DefaultColorManager) TagColor(key, value string) string {
	if c.noColor {
		return fmt.Sprintf("%s=%s", key, value)
	}
	return ColorPurple + key + ColorReset + "=" + ColorCyan + value + ColorReset
}

// StatusColor returns colorized status text based on status type
func (c *DefaultColorManager) StatusColor(status string) string {
	if c.noColor {
		return status
	}

	switch strings.ToLower(status) {
	case "ok", "passed":
		return ColorGreen + status + ColorReset
	case "warning", "impaired":
		return ColorYellow + status + ColorReset
	case "error", "failed", "critical":
		return ColorRed + status + ColorReset
	default:
		return status
	}
}

// DimColor returns dimmed text
func (c *DefaultColorManager) DimColor(text string) string {
	if c.noColor {
		return text
	}
	return ColorDim + text + ColorReset
}

// BoldColor returns bold text
func (c *DefaultColorManager) BoldColor(text string) string {
	if c.noColor {
		return text
	}
	return ColorBold + text + ColorReset
}
