package tui

import (
	"fmt"
	"strings"

	"github.com/johnlam90/aws-ssm/pkg/aws"
)

// renderNetworkInterfaces renders the network interfaces main-panel content.
func (m Model) renderNetworkInterfaces() string {
	instances := m.getNetworkInterfaces()
	if s := m.renderNetworkState(instances); s != "" {
		return s
	}
	var b strings.Builder
	cursor := clampIndex(m.cursor, len(instances))
	selected := instances[cursor]
	details := limitRenderedLines(renderNetworkDetails(selected, m.width), max(1, m.height-8))
	visibleRows := calculateTableRows(m.height, 7, details)

	headerRow := fmt.Sprintf("  %-28s %-20s %-32s %6s", "NAME", "INSTANCE ID", "DNS NAME", "IFACES")
	b.WriteString(TableHeaderStyle().Render(headerRow))
	b.WriteString("\n")
	startIdx, endIdx := calculateNetworkVisibleRange(len(instances), cursor, visibleRows)
	for i := startIdx; i < endIdx; i++ {
		inst := instances[i]
		name := normalizeValue(inst.InstanceName, "(no name)", 28)
		id := inst.InstanceID
		if id == "" {
			id = "unknown"
		}
		dns := normalizeValue(inst.DNSName, "n/a", 32)
		row := fmt.Sprintf("  %-28s %-20s %-32s %6d", name, id, dns, len(inst.Interfaces))
		b.WriteString(RenderSelectableRow(row, i == cursor))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(details)

	if searchBar := m.renderSearchBar(ViewNetworkInterfaces); searchBar != "" {
		b.WriteString("\n")
		b.WriteString(searchBar)
	}
	return b.String()
}

func renderNetworkDetails(selected aws.InstanceInterfaces, width int) string {
	var b strings.Builder
	b.WriteString(SubtitleStyle().Render(fmt.Sprintf("Interfaces for %s (%s)", normalizeValue(selected.InstanceName, "(no name)", 0), selected.InstanceID)))
	b.WriteString("\n")
	fmt.Fprintf(&b, "  DNS Name:   %s\n", normalizeValue(selected.DNSName, "n/a", 0))
	fmt.Fprintf(&b, "  Interfaces: %d\n\n", len(selected.Interfaces))
	if len(selected.Interfaces) == 0 {
		b.WriteString(HelpStyle().Render("No interfaces found for this instance"))
		b.WriteString("\n")
	} else {
		widths := calculateInterfaceColumnWidths(selected.Interfaces, width)
		b.WriteString(TableHeaderStyle().Render(formatInterfaceHeader(widths)))
		b.WriteString("\n")
		for _, iface := range selected.Interfaces {
			b.WriteString(ListItemStyle().Render(formatInterfaceRow(iface, widths)))
			b.WriteString("\n")
		}
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func (m Model) renderNetworkState(instances []aws.InstanceInterfaces) string {
	if m.loading {
		return m.renderLoading()
	}
	if m.err != nil {
		return m.renderError()
	}
	if len(instances) == 0 {
		return SubtitleStyle().Render("No instances with network interfaces")
	}
	return ""
}

func calculateNetworkVisibleRange(total, cursor, visibleHeight int) (int, int) {
	return calculateBoundedVisibleRange(total, cursor, visibleHeight)
}

type interfaceColumnWidths struct {
	iface  int
	card   int
	device int
	subnet int
	cidr   int
	sg     int
}

func (w interfaceColumnWidths) totalWidth() int {
	const indent = 2
	const columnSpacing = 5
	return indent + columnSpacing + w.iface + w.card + w.device + w.subnet + w.cidr + w.sg
}

func (w *interfaceColumnWidths) clamp(maxWidth int) {
	if maxWidth <= 0 {
		return
	}
	current := w.totalWidth()
	if current <= maxWidth {
		return
	}

	targets := []struct {
		ptr *int
		min int
	}{
		{&w.sg, len("SECURITY GROUP")},
		{&w.subnet, len("SUBNET")},
		{&w.cidr, len("CIDR")},
	}

	for _, target := range targets {
		if current <= maxWidth {
			break
		}
		if *target.ptr <= target.min {
			continue
		}

		diff := current - maxWidth
		reducible := *target.ptr - target.min
		if reducible > diff {
			reducible = diff
		}
		*target.ptr -= reducible
		current -= reducible
	}
}

func calculateInterfaceColumnWidths(ifaces []aws.NetworkInterface, totalWidth int) interfaceColumnWidths {
	widths := interfaceColumnWidths{
		iface:  len("IFACE"),
		card:   len("CARD"),
		device: len("DEVICE"),
		subnet: len("SUBNET"),
		cidr:   len("CIDR"),
		sg:     len("SECURITY GROUP"),
	}

	for _, iface := range ifaces {
		widths.iface = maxInt(widths.iface, len(normalizeValue(iface.InterfaceName, "n/a", 0)))
		widths.card = maxInt(widths.card, len(fmt.Sprintf("%d", iface.NetworkCardIndex)))
		widths.device = maxInt(widths.device, len(fmt.Sprintf("%d", iface.DeviceIndex)))
		widths.subnet = maxInt(widths.subnet, len(normalizeValue(iface.SubnetID, "N/A", 0)))
		widths.cidr = maxInt(widths.cidr, len(normalizeValue(iface.CIDR, "N/A", 0)))
		widths.sg = maxInt(widths.sg, len(normalizeValue(iface.SecurityGroup, "N/A", 0)))
	}

	widths.clamp(totalWidth)
	return widths
}

func formatInterfaceHeader(widths interfaceColumnWidths) string {
	return fmt.Sprintf("  %-*s %-*s %-*s %-*s %-*s %-*s",
		widths.iface, "IFACE",
		widths.card, "CARD",
		widths.device, "DEVICE",
		widths.subnet, "SUBNET",
		widths.cidr, "CIDR",
		widths.sg, "SECURITY GROUP",
	)
}

func formatInterfaceRow(iface aws.NetworkInterface, widths interfaceColumnWidths) string {
	name := normalizeValue(iface.InterfaceName, "n/a", widths.iface)
	subnet := normalizeValue(iface.SubnetID, "N/A", widths.subnet)
	cidr := normalizeValue(iface.CIDR, "N/A", widths.cidr)
	sg := normalizeValue(iface.SecurityGroup, "N/A", widths.sg)

	return fmt.Sprintf("  %-*s %*d %*d %-*s %-*s %-*s",
		widths.iface, name,
		widths.card, iface.NetworkCardIndex,
		widths.device, iface.DeviceIndex,
		widths.subnet, subnet,
		widths.cidr, cidr,
		widths.sg, sg,
	)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// normalizeValue returns a fallback if value is empty and optionally truncates it
func normalizeValue(value, fallback string, maxLen int) string {
	if strings.TrimSpace(value) == "" || value == "N/A" {
		value = fallback
	}

	if maxLen > 0 && len(value) > maxLen {
		if maxLen <= 3 {
			return value[:maxLen]
		}
		return value[:maxLen-3] + "..."
	}

	return value
}
