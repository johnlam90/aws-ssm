package fuzzy

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BubblesFinder provides a fuzzy finder using bubbles/list with full styling control
type BubblesFinder struct {
	items        []list.Item
	previewFunc  func(int, int, int) string
	promptString string
	colors       ColorManager
	selected     int
}

// bubbleItem implements list.Item interface
type bubbleItem struct {
	title       string
	description string
	index       int
}

func (i bubbleItem) Title() string       { return i.title }
func (i bubbleItem) Description() string { return i.description }
func (i bubbleItem) FilterValue() string { return i.title }

// bubbleModel is the bubbletea model for the fuzzy finder
type bubbleModel struct {
	list        list.Model
	previewFunc func(int, int, int) string
	selected    int
	cancelled   bool
	width       int
	height      int
}

func (m bubbleModel) Init() tea.Cmd {
	return nil
}

func (m bubbleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Calculate responsive split based on terminal width
		listWidth, _ := calculateResponsiveSplit(msg.Width)
		m.list.SetSize(listWidth, msg.Height-4)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			if len(m.list.Items()) > 0 {
				if item, ok := m.list.SelectedItem().(bubbleItem); ok {
					m.selected = item.index
				}
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m bubbleModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Calculate responsive split based on terminal width
	_, previewWidth := calculateResponsiveSplit(m.width)

	listView := m.list.View()

	// Generate preview
	var preview string
	if m.previewFunc != nil && len(m.list.Items()) > 0 {
		if item, ok := m.list.SelectedItem().(bubbleItem); ok {
			preview = m.previewFunc(item.index, previewWidth, m.height)
		}
	}

	// Render side-by-side with responsive preview width
	previewStyle := lipgloss.NewStyle().
		Width(previewWidth).
		Height(m.height-4).
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		Padding(0, 1)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		listView,
		previewStyle.Render(preview),
	)
}

// calculateResponsiveSplit calculates responsive list/preview widths based on terminal width
func calculateResponsiveSplit(totalWidth int) (listWidth, previewWidth int) {
	if totalWidth <= 0 {
		return 40, 30
	}

	// For narrow terminals, give more space to the list
	// For wider terminals, balance the preview
	switch {
	case totalWidth >= 160:
		// Extra wide: 55% list, 45% preview
		listWidth = int(float64(totalWidth) * 0.55)
	case totalWidth >= 120:
		// Wide: 58% list, 42% preview
		listWidth = int(float64(totalWidth) * 0.58)
	case totalWidth >= 100:
		// Medium: 60% list, 40% preview
		listWidth = int(float64(totalWidth) * 0.60)
	case totalWidth >= 80:
		// Narrow: 65% list, 35% preview
		listWidth = int(float64(totalWidth) * 0.65)
	default:
		// Very narrow: 70% list, 30% preview
		listWidth = int(float64(totalWidth) * 0.70)
	}

	previewWidth = totalWidth - listWidth - 2 // 2 for border
	if previewWidth < 20 {
		previewWidth = 20
	}

	return listWidth, previewWidth
}

// NewBubblesFinder creates a new fuzzy finder using bubbles/list
func NewBubblesFinder(items []list.Item, previewFunc func(int, int, int) string, promptString string, colors ColorManager) *BubblesFinder {
	return &BubblesFinder{
		items:        items,
		previewFunc:  previewFunc,
		promptString: promptString,
		colors:       colors,
		selected:     -1,
	}
}

// Select displays the fuzzy finder and returns the selected index
func (f *BubblesFinder) Select(ctx context.Context) (int, error) {
	// Create custom delegate with our styling
	delegate := list.NewDefaultDelegate()

	// Customize selected item style - visible color change
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#5B21B6", Dark: "#60A5FA"}). // Purple for light, blue for dark
		Bold(true)

	delegate.Styles.SelectedTitle = selectedStyle
	delegate.Styles.SelectedDesc = selectedStyle

	// Create list
	l := list.New(f.items, delegate, 0, 0)
	l.Title = f.promptString + " [NEW BUBBLES UI]" // DEBUG: Confirm new UI is loading
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)

	// Customize help
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		}
	}

	// Create model
	m := bubbleModel{
		list:        l,
		previewFunc: f.previewFunc,
		selected:    -1,
		cancelled:   false,
	}

	// Run the program
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithContext(ctx))
	finalModel, err := p.Run()
	if err != nil {
		return -1, fmt.Errorf("error running fuzzy finder: %w", err)
	}

	result := finalModel.(bubbleModel)
	if result.cancelled {
		return -1, fmt.Errorf("selection cancelled")
	}

	return result.selected, nil
}
