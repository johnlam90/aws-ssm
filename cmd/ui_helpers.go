package cmd

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/johnlam90/aws-ssm/pkg/ui/fuzzy"
)

// createLoadingSpinner creates and returns a configured spinner
func createLoadingSpinner(message string) *spinner.Spinner {
	if noColor {
		// Simple spinner for no-color mode
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " " + message
		return s
	}

	// Stylish spinner with color (dots style - CharSet 14)
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + message
	_ = s.Color("cyan", "bold") // Ignore error - color is optional
	return s
}

// printInteractivePrompt prints a styled prompt for interactive selection
func printInteractivePrompt(selectorName string) {
	if noColor {
		fmt.Printf("Opening interactive %s...\n", selectorName)
		fmt.Println("(Use arrow keys to navigate, type to filter, Enter to select, Esc to cancel)")
	} else {
		// Use cyan color for the chevron and bold for the selector name
		chevron := fuzzy.ColorCyan + "❯" + fuzzy.ColorReset
		boldName := fuzzy.ColorBold + selectorName + fuzzy.ColorReset
		fmt.Printf("%s Opening interactive %s...\n", chevron, boldName)

		// Dim the instruction text
		instruction := fuzzy.ColorDim + "(Use arrow keys to navigate, type to filter, Enter to select, Esc to cancel)" + fuzzy.ColorReset
		fmt.Println(instruction)
	}
}

// printSelectionCancelled prints a styled cancellation message
func printSelectionCancelled() {
	if noColor {
		fmt.Println("\nSelection cancelled.")
	} else {
		msg := fuzzy.ColorYellow + "Selection cancelled." + fuzzy.ColorReset
		fmt.Printf("\n%s\n", msg)
	}
}

// printNoSelection prints a styled message when no item is selected
func printNoSelection(itemType string) {
	if noColor {
		fmt.Printf("\nNo %s selected.\n", itemType)
	} else {
		msg := fuzzy.ColorYellow + fmt.Sprintf("No %s selected.", itemType) + fuzzy.ColorReset
		fmt.Printf("\n%s\n", msg)
	}
}
