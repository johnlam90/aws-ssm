package tui

// layout.go provides responsive layout utilities for dynamic terminal width adaptation

// LayoutConfig holds configuration for responsive layouts
type LayoutConfig struct {
	MinWidth int // Minimum terminal width to support
	MaxWidth int // Maximum width before content stops growing
}

// DefaultLayoutConfig returns sensible defaults for terminal layouts
func DefaultLayoutConfig() LayoutConfig {
	return LayoutConfig{
		MinWidth: 60,
		MaxWidth: 200,
	}
}

// ColumnSpec defines a table column with responsive width settings
type ColumnSpec struct {
	Name       string  // Column header name
	MinWidth   int     // Minimum column width
	MaxWidth   int     // Maximum column width (0 = unlimited)
	Weight     float64 // Proportional weight for extra space distribution
	Truncate   bool    // Whether to truncate content with ellipsis
	RightAlign bool    // Whether to right-align content
}

// TableLayout calculates responsive column widths for tables
type TableLayout struct {
	Columns     []ColumnSpec
	TotalWidth  int
	Indent      int
	ColumnGap   int
	ColumnWidth []int // Calculated widths after layout
}

// NewTableLayout creates a new table layout calculator
func NewTableLayout(totalWidth int, columns []ColumnSpec) *TableLayout {
	layout := &TableLayout{
		Columns:     columns,
		TotalWidth:  totalWidth,
		Indent:      2, // Default left indent
		ColumnGap:   1, // Default gap between columns
		ColumnWidth: make([]int, len(columns)),
	}
	layout.calculate()
	return layout
}

// calculate computes the optimal column widths
func (t *TableLayout) calculate() {
	if len(t.Columns) == 0 || t.TotalWidth <= 0 {
		return
	}

	// Start with minimum widths
	usedWidth := t.Indent + (len(t.Columns)-1)*t.ColumnGap
	totalWeight := 0.0

	for i, col := range t.Columns {
		t.ColumnWidth[i] = col.MinWidth
		usedWidth += col.MinWidth
		totalWeight += col.Weight
	}

	// Distribute remaining space by weight
	remainingWidth := t.TotalWidth - usedWidth
	if remainingWidth > 0 && totalWeight > 0 {
		for i, col := range t.Columns {
			extra := int(float64(remainingWidth) * (col.Weight / totalWeight))
			if col.MaxWidth > 0 && t.ColumnWidth[i]+extra > col.MaxWidth {
				extra = col.MaxWidth - t.ColumnWidth[i]
			}
			t.ColumnWidth[i] += extra
		}
	}

	// If still too wide, shrink from right to left (except for right-aligned columns)
	currentTotal := t.Indent + (len(t.Columns)-1)*t.ColumnGap
	for _, w := range t.ColumnWidth {
		currentTotal += w
	}

	if currentTotal > t.TotalWidth {
		excess := currentTotal - t.TotalWidth
		// Shrink columns from right to left, respecting minimum widths
		for i := len(t.Columns) - 1; i >= 0 && excess > 0; i-- {
			col := t.Columns[i]
			available := t.ColumnWidth[i] - col.MinWidth
			if available > 0 {
				shrink := available
				if shrink > excess {
					shrink = excess
				}
				t.ColumnWidth[i] -= shrink
				excess -= shrink
			}
		}
	}
}

// Width returns the calculated width for a column by index
func (t *TableLayout) Width(index int) int {
	if index < 0 || index >= len(t.ColumnWidth) {
		return 0
	}
	return t.ColumnWidth[index]
}

// ResponsiveWidth returns a width that scales with the terminal,
// clamped between min and max values
func ResponsiveWidth(terminalWidth, minWidth, maxWidth int, ratio float64) int {
	if terminalWidth <= 0 {
		return minWidth
	}

	width := int(float64(terminalWidth) * ratio)
	if width < minWidth {
		return minWidth
	}
	if maxWidth > 0 && width > maxWidth {
		return maxWidth
	}
	return width
}

// ClampWidth ensures a width stays within bounds
func ClampWidth(width, minWidth, maxWidth int) int {
	if width < minWidth {
		return minWidth
	}
	if maxWidth > 0 && width > maxWidth {
		return maxWidth
	}
	return width
}

// AvailableContentWidth calculates the usable width after accounting for chrome
func AvailableContentWidth(terminalWidth, leftMargin, rightMargin int) int {
	available := terminalWidth - leftMargin - rightMargin
	if available < 20 {
		return 20 // Minimum usable width
	}
	return available
}

// SplitWidth divides width into two panels with a ratio
func SplitWidth(totalWidth int, leftRatio float64) (leftWidth, rightWidth int) {
	if totalWidth <= 0 {
		return 0, 0
	}
	leftWidth = int(float64(totalWidth) * leftRatio)
	rightWidth = totalWidth - leftWidth
	return
}

// TruncateWithEllipsis truncates a string to fit within maxLen, adding ellipsis
func TruncateWithEllipsis(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}

// VerticalLayout holds the calculated heights for view components
type VerticalLayout struct {
	TotalHeight     int // Total terminal height
	HeaderHeight    int // Height for header (title, subtitle)
	TableHeight     int // Height for the table/list (number of visible rows)
	DetailHeight    int // Height for the detail/preview section
	FooterHeight    int // Height for footer, help, status bar
	AvailableHeight int // Total usable height after fixed elements
}

// CalculateVerticalLayout computes dynamic heights for a two-panel layout (list + detail)
// - fixedHeaderLines: lines used by headers, empty lines before table
// - fixedFooterLines: lines used by footer, help text, status bar, empty lines
// - minTableRows: minimum number of table rows to show
// - minDetailRows: minimum number of detail section rows
// - tableRatio: what percentage of available space goes to the table (0.0-1.0)
func CalculateVerticalLayout(terminalHeight, fixedHeaderLines, fixedFooterLines, minTableRows, minDetailRows int, tableRatio float64) VerticalLayout {
	layout := VerticalLayout{
		TotalHeight:  terminalHeight,
		HeaderHeight: fixedHeaderLines,
		FooterHeight: fixedFooterLines,
	}

	// Calculate available space for table + detail
	layout.AvailableHeight = terminalHeight - fixedHeaderLines - fixedFooterLines
	if layout.AvailableHeight < minTableRows+minDetailRows {
		// Not enough space - use minimums
		layout.TableHeight = minTableRows
		layout.DetailHeight = minDetailRows
		return layout
	}

	// Distribute available space according to ratio
	layout.TableHeight = int(float64(layout.AvailableHeight) * tableRatio)
	layout.DetailHeight = layout.AvailableHeight - layout.TableHeight

	// Enforce minimums
	if layout.TableHeight < minTableRows {
		layout.TableHeight = minTableRows
		layout.DetailHeight = layout.AvailableHeight - layout.TableHeight
	}
	if layout.DetailHeight < minDetailRows {
		layout.DetailHeight = minDetailRows
		layout.TableHeight = layout.AvailableHeight - layout.DetailHeight
	}

	return layout
}

// NodeGroupLayout calculates heights for the node group view
func NodeGroupLayout(terminalHeight int) VerticalLayout {
	// Header: title line + empty line + header row = 3 lines
	// Footer: empty line + subtitle + help line + status bar = 4 lines
	// Details section: ~12 lines for selected node group info
	return CalculateVerticalLayout(terminalHeight, 3, 4, 3, 10, 0.5)
}

// EC2Layout calculates heights for the EC2 instances view
func EC2Layout(terminalHeight int) VerticalLayout {
	// Header: title line + empty line + header row = 3 lines
	// Footer: pagination + subtitle + details + help + status = ~4 lines fixed
	// EC2 details take more space, so give more to details
	return CalculateVerticalLayout(terminalHeight, 3, 4, 3, 8, 0.45)
}

// ASGLayout calculates heights for the ASG view
func ASGLayout(terminalHeight int) VerticalLayout {
	// Similar to node groups
	return CalculateVerticalLayout(terminalHeight, 3, 4, 3, 10, 0.5)
}

// EKSLayout calculates heights for the EKS clusters view
func EKSLayout(terminalHeight int) VerticalLayout {
	return CalculateVerticalLayout(terminalHeight, 3, 4, 3, 10, 0.5)
}

// NetworkLayout calculates heights for the network interfaces view
func NetworkLayout(terminalHeight int) VerticalLayout {
	// Network view has more detail (interface list)
	return CalculateVerticalLayout(terminalHeight, 3, 4, 3, 12, 0.4)
}

// CalculateVisibleRows returns the number of table rows to show based on terminal height
// This is a unified replacement for the various hardcoded m.height-N calculations
func CalculateVisibleRows(terminalHeight, fixedOverhead int) int {
	rows := terminalHeight - fixedOverhead
	if rows < 3 {
		return 3
	}
	return rows
}

// BreakpointSize represents terminal width categories
type BreakpointSize int

// Breakpoint constants for responsive terminal width handling
const (
	// BreakpointNarrow represents terminals with less than 80 columns
	BreakpointNarrow BreakpointSize = iota
	// BreakpointMedium represents terminals with 80-119 columns
	BreakpointMedium
	// BreakpointWide represents terminals with 120-159 columns
	BreakpointWide
	// BreakpointXWide represents terminals with 160+ columns
	BreakpointXWide
)

// GetBreakpoint returns the current breakpoint for responsive layouts
func GetBreakpoint(width int) BreakpointSize {
	switch {
	case width >= 160:
		return BreakpointXWide
	case width >= 120:
		return BreakpointWide
	case width >= 80:
		return BreakpointMedium
	default:
		return BreakpointNarrow
	}
}

// EC2ColumnWidths returns responsive column widths for EC2 instance tables
// Uses proportional scaling to utilize available terminal width
func EC2ColumnWidths(terminalWidth int) (name, id, ip, state, itype int) {
	// Fixed-width columns
	id = 20    // Instance IDs are fixed length (i-xxxxxxxxxxxxxxxxx)
	ip = 15    // IP addresses have max width
	state = 10 // State names are short
	itype = 12 // Instance types are relatively short

	// Calculate space for the name column
	// Account for: 2 indent + fixed columns + spacing between 5 columns (4 gaps)
	fixedWidth := 2 + id + ip + state + itype + 4
	available := terminalWidth - fixedWidth

	if available < 15 {
		name = 15
	} else {
		name = available
	}

	return
}

// ASGColumnWidths returns responsive column widths for ASG tables
// Uses proportional scaling to utilize available terminal width
func ASGColumnWidths(terminalWidth int) (name, desired, minSize, maxSize, current int) {
	// Fixed-width numeric columns
	desired = 8
	minSize = 8
	maxSize = 8
	current = 8

	// Calculate space for the name column
	// Account for: 2 indent + fixed columns + spacing between 5 columns (4 gaps)
	fixedWidth := 2 + desired + minSize + maxSize + current + 4
	available := terminalWidth - fixedWidth

	if available < 25 {
		name = 25
	} else {
		name = available
	}

	return
}

// EKSColumnWidths returns responsive column widths for EKS cluster tables
// Uses proportional scaling to utilize available terminal width
func EKSColumnWidths(terminalWidth int) (name, status, version int) {
	// Fixed-width columns
	status = 12  // Status names are short (ACTIVE, CREATING, etc.)
	version = 10 // Version numbers are short

	// Calculate space for the name column
	// Account for: 2 indent + fixed columns + spacing between 3 columns (2 gaps)
	fixedWidth := 2 + status + version + 2
	available := terminalWidth - fixedWidth

	if available < 20 {
		name = 20
	} else {
		name = available
	}

	return
}

// NodeGroupColumnWidths returns responsive column widths for node group tables
// Uses proportional scaling to utilize available terminal width
func NodeGroupColumnWidths(terminalWidth int) (cluster, nodeGroup, status, desired, minSize, maxSize, current int) {
	// Fixed-width numeric columns
	desired = 8
	minSize = 4 // Compact to give more room to name columns
	maxSize = 4 // Compact to give more room to name columns
	current = 8
	status = 8 // Compact status

	// Calculate space for variable-width columns
	// Account for: 2 indent + fixed columns + spacing between 7 columns (6 gaps)
	fixedWidth := 2 + desired + minSize + maxSize + current + status + 6
	available := terminalWidth - fixedWidth

	if available < 30 {
		// Very narrow - minimum usable widths
		cluster = 15
		nodeGroup = 18
	} else {
		// Distribute available space: 40% cluster, 60% node group
		cluster = available * 40 / 100
		nodeGroup = available * 60 / 100

		// Ensure minimum widths
		if cluster < 20 {
			cluster = 20
		}
		if nodeGroup < 25 {
			nodeGroup = 25
		}
	}

	return
}

// NetworkColumnWidths returns responsive column widths for network interface tables
// Uses proportional scaling to utilize available terminal width
func NetworkColumnWidths(terminalWidth int) (name, id, dns, ifaces int) {
	// Fixed-width columns
	id = 20    // Instance IDs are fixed length
	ifaces = 6 // Number of interfaces is short

	// Calculate space for variable-width columns
	// Account for: 2 indent + fixed columns + spacing between 4 columns (3 gaps)
	fixedWidth := 2 + id + ifaces + 3
	available := terminalWidth - fixedWidth

	if available < 30 {
		// Very narrow - minimum usable widths
		name = 15
		dns = 15
	} else {
		// Distribute available space: 35% name, 65% DNS
		name = available * 35 / 100
		dns = available * 65 / 100

		// Ensure minimum widths
		if name < 15 {
			name = 15
		}
		if dns < 15 {
			dns = 15
		}
	}

	return
}

// PreviewPanelWidth returns the width for preview panels in fuzzy finders
func PreviewPanelWidth(terminalWidth int) int {
	// Preview panel takes 40% of screen, min 30, max 80
	return ClampWidth(int(float64(terminalWidth)*0.4), 30, 80)
}

// ListPanelWidth returns the width for list panels in fuzzy finders
func ListPanelWidth(terminalWidth int) int {
	// List panel takes remaining space after preview
	return terminalWidth - PreviewPanelWidth(terminalWidth) - 2 // 2 for border
}

// CalculateVisibleRange computes the visible start and end indices for a scrollable list.
// This is a consolidated function that replaces the view-specific implementations.
// Parameters:
//   - total: total number of items in the list
//   - cursor: current cursor position (0-indexed)
//   - visibleHeight: number of rows available for display
//   - minHeight: minimum height threshold below which we show all items (default 3)
//
// Returns:
//   - start: index of the first visible item
//   - end: index after the last visible item (for use in slice [start:end])
func CalculateVisibleRange(total, cursor, visibleHeight int) (start, end int) {
	return CalculateVisibleRangeWithThreshold(total, cursor, visibleHeight, 3)
}

// CalculateVisibleRangeWithThreshold is like CalculateVisibleRange but with a custom minimum height threshold.
func CalculateVisibleRangeWithThreshold(total, cursor, visibleHeight, minHeight int) (start, end int) {
	// If visibleHeight is too small or we can show everything, show all items
	if visibleHeight < minHeight || total <= visibleHeight {
		return 0, total
	}

	// Center the cursor in the visible area
	start = cursor - visibleHeight/2
	if start < 0 {
		start = 0
	}

	end = start + visibleHeight
	if end > total {
		end = total
		// Adjust start to fill the visible area if we're at the end
		start = end - visibleHeight
		if start < 0 {
			start = 0
		}
	}

	return start, end
}
