package fuzzy

import (
	"fmt"
	"strconv"
	"strings"
)

// LaunchTemplateVersionPreviewRenderer renders launch template version details
type LaunchTemplateVersionPreviewRenderer struct {
	colors *DefaultColorManager
}

// NewLaunchTemplateVersionPreviewRenderer creates a new preview renderer
func NewLaunchTemplateVersionPreviewRenderer(colors *DefaultColorManager) *LaunchTemplateVersionPreviewRenderer {
	return &LaunchTemplateVersionPreviewRenderer{
		colors: colors,
	}
}

// Render renders the launch template version details
func (r *LaunchTemplateVersionPreviewRenderer) Render(v *LaunchTemplateVersionInfo, width, _ int) string {
	var b strings.Builder

	// Calculate responsive separator width
	separatorWidth := responsiveSeparatorWidth(width)

	// Header
	b.WriteString(r.colors.HeaderColor("Launch Template Version Details\n"))
	b.WriteString(strings.Repeat("─", separatorWidth))
	b.WriteString("\n\n")

	// Version information
	b.WriteString(r.renderVersionInfo(v))
	b.WriteString("\n")

	// Template information
	b.WriteString(r.renderTemplateInfo(v))
	b.WriteString("\n")

	// Metadata
	if v.CreateTime != "" || v.CreatedBy != "" {
		b.WriteString(r.renderMetadata(v))
	}

	return b.String()
}

func (r *LaunchTemplateVersionPreviewRenderer) renderVersionInfo(v *LaunchTemplateVersionInfo) string {
	var b strings.Builder

	b.WriteString(r.colors.BoldColor("Version:\n"))

	versionStr := r.formatVersionNumber(v)
	b.WriteString(fmt.Sprintf("  %s\n", versionStr))

	if v.VersionDescription != "" {
		b.WriteString(r.colors.BoldColor("Description:\n"))
		b.WriteString(fmt.Sprintf("  %s\n", v.VersionDescription))
	}

	return b.String()
}

func (r *LaunchTemplateVersionPreviewRenderer) renderTemplateInfo(v *LaunchTemplateVersionInfo) string {
	var b strings.Builder

	b.WriteString(r.colors.BoldColor("Launch Template:\n"))
	b.WriteString(fmt.Sprintf("  Name: %s\n", v.LaunchTemplateName))
	b.WriteString(fmt.Sprintf("  ID:   %s\n", v.LaunchTemplateID))

	return b.String()
}

func (r *LaunchTemplateVersionPreviewRenderer) renderMetadata(v *LaunchTemplateVersionInfo) string {
	var b strings.Builder

	b.WriteString(r.colors.BoldColor("Metadata:\n"))

	if v.CreateTime != "" {
		b.WriteString(fmt.Sprintf("  Created: %s\n", v.CreateTime))
	}

	if v.CreatedBy != "" {
		b.WriteString(fmt.Sprintf("  Created By: %s\n", v.CreatedBy))
	}

	return b.String()
}

func (r *LaunchTemplateVersionPreviewRenderer) formatVersionNumber(v *LaunchTemplateVersionInfo) string {
	if v.VersionNumber == -1 {
		return "$Latest"
	}
	if v.VersionNumber == -2 {
		return "$Default"
	}

	versionStr := strconv.FormatInt(v.VersionNumber, 10)
	if v.DefaultVersion {
		return fmt.Sprintf("%s (Default)", versionStr)
	}
	return versionStr
}
