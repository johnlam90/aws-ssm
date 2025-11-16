package tui

import (
	"fmt"
	"strings"

	"github.com/johnlam90/aws-ssm/pkg/aws"
)

// renderNetworkInterfaces renders the network interfaces view
func (m Model) renderNetworkInterfaces() string {
	instances := m.getNetworkInterfaces()
	if s := m.renderNetworkState(instances); s != "" {
		return s
	}
	var b strings.Builder
	header := m.renderHeader("Network Interfaces", fmt.Sprintf("%d instances", len(instances)))
	b.WriteString(header)
	b.WriteString("\n\n")
	headerRow := fmt.Sprintf("  %-28s %-20s %-32s %6s", "NAME", "INSTANCE ID", "DNS NAME", "IFACES")
	b.WriteString(TableHeaderStyle().Render(headerRow))
	b.WriteString("\n")
	cursor := clampIndex(m.cursor, len(instances))
	startIdx, endIdx := calculateNetworkVisibleRange(len(instances), cursor, m.height-15)
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
	selected := instances[cursor]
	b.WriteString("\n")
	b.WriteString(SubtitleStyle().Render(fmt.Sprintf("Interfaces for %s (%s)", normalizeValue(selected.InstanceName, "(no name)", 0), selected.InstanceID)))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  DNS Name:   %s\n", normalizeValue(selected.DNSName, "n/a", 0)))
	b.WriteString(fmt.Sprintf("  Interfaces: %d\n\n", len(selected.Interfaces)))
	if len(selected.Interfaces) == 0 {
		b.WriteString(HelpStyle().Render("No interfaces found for this instance"))
		b.WriteString("\n")
	} else {
		widths := calculateInterfaceColumnWidths(selected.Interfaces, m.width)
		b.WriteString(TableHeaderStyle().Render(formatInterfaceHeader(widths)))
		b.WriteString("\n")
		for _, iface := range selected.Interfaces {
			b.WriteString(ListItemStyle().Render(formatInterfaceRow(iface, widths)))
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")
	if searchBar := m.renderSearchBar(ViewNetworkInterfaces); searchBar != "" {
		b.WriteString(searchBar)
		b.WriteString("\n")
	}
	b.WriteString(m.renderNetworkFooter())
	b.WriteString("\n")
	b.WriteString(StatusBarStyle().Width(m.width).Render(m.getStatusBar()))
	return b.String()
}

func (m Model) renderNetworkState(instances []aws.InstanceInterfaces) string {
	var b strings.Builder
	header := m.renderHeader("Network Interfaces", fmt.Sprintf("%d instances", len(instances)))
	b.WriteString(header)
	b.WriteString("\n\n")
	if m.loading {
		b.WriteString(m.renderLoading())
		b.WriteString("\n")
		b.WriteString(StatusBarStyle().Render(m.getStatusBar()))
		return b.String()
	}
	if m.err != nil {
		b.WriteString(m.renderError())
		b.WriteString("\n\n")
		b.WriteString(HelpStyle().Render("esc:back"))
		b.WriteString("\n")
		b.WriteString(StatusBarStyle().Render(m.getStatusBar()))
		return b.String()
	}
	if len(instances) == 0 {
		b.WriteString(SubtitleStyle().Render("No instances with network interfaces"))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle().Render("esc:back"))
		b.WriteString("\n")
		b.WriteString(StatusBarStyle().Render(m.getStatusBar()))
		return b.String()
	}
	return ""
}

func calculateNetworkVisibleRange(total, cursor, visibleHeight int) (int, int) {
	if visibleHeight < 5 {
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

// renderNetworkFooter renders footer controls for the network view
func (m Model) renderNetworkFooter() string {
	keys := []struct {
		key  string
		desc string
	}{
		{"↑/k", "up"},
		{"↓/j", "down"},
		{"g/G", "top/bottom"},
		{"r", "refresh"},
		{"/", "search"},
		{"esc", "back"},
	}

	var parts []string
	for _, k := range keys {
		keyStyle := StatusBarKeyStyle().Render(k.key)
		descStyle := StatusBarValueStyle().Render(k.desc)
		parts = append(parts, fmt.Sprintf("%s %s", keyStyle, descStyle))
	}

	return HelpStyle().Render(strings.Join(parts, " • "))
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
	const columnSpacing = 5 // Six columns, five gaps
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
