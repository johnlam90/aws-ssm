package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// renderNetworkInterfaces renders the network interfaces view
func (m Model) renderNetworkInterfaces() string {
	var b strings.Builder

	instances := m.getNetworkInterfaces()

	header := m.renderHeader("Network Interfaces", fmt.Sprintf("%d instances", len(instances)))
	b.WriteString(header)
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(m.renderLoading())
		b.WriteString("\n")
		b.WriteString(StatusBarStyle.Render(m.getStatusBar()))
		return b.String()
	}

	if m.err != nil {
		b.WriteString(m.renderError())
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("esc:back"))
		b.WriteString("\n")
		b.WriteString(StatusBarStyle.Render(m.getStatusBar()))
		return b.String()
	}

	if len(instances) == 0 {
		b.WriteString(SubtitleStyle.Render("No instances with network interfaces"))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("esc:back"))
		b.WriteString("\n")
		b.WriteString(StatusBarStyle.Render(m.getStatusBar()))
		return b.String()
	}

	cursor := m.cursor
	total := len(instances)
	if cursor >= total {
		cursor = total - 1
	}
	if cursor < 0 {
		cursor = 0
	}

	visibleHeight := m.height - 15
	if visibleHeight < 5 {
		visibleHeight = total
	}

	startIdx := 0
	endIdx := total
	if total > visibleHeight && visibleHeight > 0 {
		startIdx = cursor - visibleHeight/2
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + visibleHeight
		if endIdx > total {
			endIdx = total
			startIdx = endIdx - visibleHeight
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	headerRow := fmt.Sprintf("  %-28s %-20s %-32s %6s", "NAME", "INSTANCE ID", "DNS NAME", "IFACES")
	b.WriteString(TableHeaderStyle.Render(headerRow))
	b.WriteString("\n")

	for i := startIdx; i < endIdx; i++ {
		inst := instances[i]
		name := normalizeValue(inst.InstanceName, "(no name)", 28)
		instanceID := inst.InstanceID
		if instanceID == "" {
			instanceID = "unknown"
		}
		dns := normalizeValue(inst.DNSName, "n/a", 32)
		ifaceCount := len(inst.Interfaces)

		row := fmt.Sprintf("  %-28s %-20s %-32s %6d", name, instanceID, dns, ifaceCount)
		b.WriteString(RenderSelectableRow(row, i == cursor))
		b.WriteString("\n")
	}

	selected := instances[cursor]

	b.WriteString("\n")
	detailTitle := fmt.Sprintf("Interfaces for %s (%s)", normalizeValue(selected.InstanceName, "(no name)", 0), selected.InstanceID)
	b.WriteString(SubtitleStyle.Render(detailTitle))
	b.WriteString("\n")

	if len(selected.Interfaces) == 0 {
		b.WriteString(HelpStyle.Render("No interfaces found for this instance"))
		b.WriteString("\n")
	} else {
		detailHeader := fmt.Sprintf("  %-10s %-20s %-22s %-22s", "IFACE", "SUBNET", "CIDR", "SECURITY GROUP")
		b.WriteString(TableHeaderStyle.Render(detailHeader))
		b.WriteString("\n")

		for _, iface := range selected.Interfaces {
			subnet := normalizeValue(iface.SubnetID, "N/A", 20)
			cidr := normalizeValue(iface.CIDR, "N/A", 22)
			sg := normalizeValue(iface.SecurityGroup, "N/A", 22)
			row := fmt.Sprintf("  %-10s %-20s %-22s %-22s", iface.InterfaceName, subnet, cidr, sg)
			b.WriteString(ListItemStyle.Render(row))
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
	b.WriteString(StatusBarStyle.Width(m.width).Render(m.getStatusBar()))

	return b.String()
}

// handleNetworkInterfaceKeys handles input in the network interfaces view
func (m Model) handleNetworkInterfaceKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	instances := m.getNetworkInterfaces()
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(instances)-1 {
			m.cursor++
		}
	case "g":
		if len(instances) > 0 {
			m.cursor = 0
		}
	case "G":
		if len(instances) > 0 {
			m.cursor = len(instances) - 1
		}
	case "r":
		m.loading = true
		m.loadingMsg = "Refreshing network interfaces..."
		m.err = nil
		return m, LoadNetworkInterfacesCmd(m.ctx, m.client)
	case "esc":
		return m.navigateBack(), nil
	}

	return m, nil
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
		keyStyle := StatusBarKeyStyle.Render(k.key)
		descStyle := StatusBarValueStyle.Render(k.desc)
		parts = append(parts, fmt.Sprintf("%s %s", keyStyle, descStyle))
	}

	return HelpStyle.Render(strings.Join(parts, " • "))
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
