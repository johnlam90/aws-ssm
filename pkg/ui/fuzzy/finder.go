package fuzzy

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/ktr0731/go-fuzzyfinder"
)

// EnhancedFinder represents the enhanced fuzzy finder
type EnhancedFinder struct {
	state     *StateManager
	loader    InstanceLoader
	renderer  PreviewRenderer
	colors    ColorManager
	bookmarks []Bookmark
	config    Config
}

// NewEnhancedFinder creates a new enhanced fuzzy finder
func NewEnhancedFinder(loader InstanceLoader, config Config) *EnhancedFinder {
	colors := NewDefaultColorManager(config.NoColor)
	renderer := NewDefaultPreviewRenderer(colors)

	state := &StateManager{
		Config:        config,
		SortField:     SortByName,
		SortDirection: SortAsc,
		Bookmarks:     []Bookmark{},
		QueryHistory:  []string{},
	}

	return &EnhancedFinder{
		state:    state,
		loader:   loader,
		renderer: renderer,
		colors:   colors,
		config:   config,
	}
}

// SelectInstanceInteractive displays the enhanced interactive fuzzy finder
func (f *EnhancedFinder) SelectInstanceInteractive(ctx context.Context) ([]Instance, error) {
	// Initialize state with empty query if needed
	if f.state.Query == nil {
		f.state.Query = &SearchQuery{Raw: ""}
	}

	// Load initial instances
	allInstances, err := f.loadAllInstances(ctx, f.state.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to load instances: %w", err)
	}

	f.state.Instances = allInstances
	f.state.Filtered = allInstances

	// Apply initial sort
	f.sortInstances()

	selectedIndices, err := fuzzyfinder.FindMulti(
		f.state.Filtered,
		func(i int) string {
			return f.formatInstanceRow(f.state.Filtered[i])
		},
		fuzzyfinder.WithPreviewWindow(func(i, width, height int) string {
			if i < 0 || i >= len(f.state.Filtered) {
				return f.formatHelp()
			}
			return f.renderer.Render(&f.state.Filtered[i], width, height)
		}),
		fuzzyfinder.WithPromptString(f.formatPrompt()),
	)

	if err != nil {
		if err == fuzzyfinder.ErrAbort {
			return nil, nil
		}
		return nil, err
	}

	// Convert selected indices to instances
	var selectedInstances []Instance
	for _, idx := range selectedIndices {
		selectedInstances = append(selectedInstances, f.state.Filtered[idx])
	}

	return selectedInstances, nil
}

// loadAllInstances loads all instances directly (no channel overhead)
func (f *EnhancedFinder) loadAllInstances(ctx context.Context, query *SearchQuery) ([]Instance, error) {
	// Load instances directly - much faster than channel-based approach
	instances, err := f.loader.LoadInstances(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to load instances: %w", err)
	}

	return instances, nil
}

// formatInstanceRow formats an instance for display in the list
func (f *EnhancedFinder) formatInstanceRow(instance Instance) string {
	var parts []string

	if f.config.Columns.Name {
		name := instance.Name
		if name == "" {
			name = "(no name)"
		}
		name = f.truncateString(name, 30)
		parts = append(parts, fmt.Sprintf("%-30s", name))
	}

	if f.config.Columns.InstanceID {
		parts = append(parts, fmt.Sprintf("%-19s", instance.InstanceID))
	}

	if f.config.Columns.PrivateIP {
		parts = append(parts, fmt.Sprintf("%-15s", instance.PrivateIP))
	}

	if f.config.Columns.State {
		state := f.colors.StateColor(instance.State)
		parts = append(parts, fmt.Sprintf("%-10s", state))
	}

	if f.config.Columns.Type {
		parts = append(parts, fmt.Sprintf("%-12s", instance.InstanceType))
	}

	if f.config.Columns.AZ {
		parts = append(parts, fmt.Sprintf("%-12s", instance.AvailabilityZone))
	}

	return strings.Join(parts, " | ")
}

// truncateString truncates a string to the specified length
func (f *EnhancedFinder) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// formatPrompt formats the search prompt
func (f *EnhancedFinder) formatPrompt() string {
	var parts []string

	if f.config.Favorites {
		parts = append(parts, "Favorites")
	}

	if len(f.state.Query.Terms) > 0 || len(f.state.Query.Filters) > 0 {
		parts = append(parts, fmt.Sprintf("Query: %s", f.state.Query.Raw))
	}

	parts = append(parts, fmt.Sprintf("%d instances", len(f.state.Filtered)))

	if f.state.SortField != SortByName {
		parts = append(parts, fmt.Sprintf("Sort: %s", f.state.SortField))
	}

	// Add trailing space so user's typed query doesn't run into the prompt
	return strings.Join(parts, " • ") + " "
}

// formatHelp formats the help text
func (f *EnhancedFinder) formatHelp() string {
	help := []string{
		f.colors.HeaderColor("AWS SSM Instance Selector"),
		"",
		f.colors.BoldColor("Navigation:"),
		"  ↑↓       - Navigate up/down",
		"  Enter    - Select instance and connect",
		"  Esc      - Cancel and exit",
		"",
		f.colors.BoldColor("Search:"),
		"  Type to filter instances by name, ID, IP, or tags",
		"",
		f.colors.BoldColor("Advanced Search (not yet implemented):"),
		"  name:web      - Filter by name",
		"  id:i-123      - Filter by instance ID",
		"  state:running - Filter by state",
		"  tag:Env=prod  - Filter by tags",
		"",
	}

	return strings.Join(help, "\n")
}

// sortInstances sorts the instances based on current sort field and direction
func (f *EnhancedFinder) sortInstances() {
	sort.Slice(f.state.Filtered, func(i, j int) bool {
		var result bool

		switch f.state.SortField {
		case SortByName:
			result = strings.ToLower(f.state.Filtered[i].Name) < strings.ToLower(f.state.Filtered[j].Name)
		case SortByAZ:
			result = f.state.Filtered[i].AvailabilityZone < f.state.Filtered[j].AvailabilityZone
		case SortByType:
			result = f.state.Filtered[i].InstanceType < f.state.Filtered[j].InstanceType
		case SortByLaunchTime:
			result = f.state.Filtered[i].LaunchTime.Before(f.state.Filtered[j].LaunchTime)
		case SortByState:
			result = f.state.Filtered[i].State < f.state.Filtered[j].State
		case SortByID:
			result = f.state.Filtered[i].InstanceID < f.state.Filtered[j].InstanceID
		}

		if f.state.SortDirection == SortDesc {
			result = !result
		}

		return result
	})
}
