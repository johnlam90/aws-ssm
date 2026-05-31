package table

import (
	"strings"
)

// FormatHeader returns the column-headers row using the allocated
// column widths. Dropped columns are omitted. Adjacent columns are
// separated by a single space.
func FormatHeader(cols []AllocatedColumn) string {
	cells := make([]string, 0, len(cols))
	for _, c := range cols {
		if c.Dropped {
			continue
		}
		cells = append(cells, FitCell(c.Spec.Header, c.Width, c.Spec.Align))
	}
	return strings.Join(cells, " ")
}

// FormatRow joins per-column rendered cells. The values slice must be
// the same length as cols (use empty strings for cells that have no
// value). Cells corresponding to dropped columns are omitted.
//
// Each value is fit to the column's allocated width using the
// column's Align setting; the caller is responsible for any styling
// (color/bold) that should survive truncation — apply it AFTER
// FormatRow if styling spans the whole row.
func FormatRow(values []string, cols []AllocatedColumn) string {
	if len(values) != len(cols) {
		// Defensive fallback: pad/truncate values to match cols.
		out := make([]string, len(cols))
		copy(out, values)
		values = out
	}
	cells := make([]string, 0, len(cols))
	for i, c := range cols {
		if c.Dropped {
			continue
		}
		cells = append(cells, FitCell(values[i], c.Width, c.Spec.Align))
	}
	return strings.Join(cells, " ")
}
