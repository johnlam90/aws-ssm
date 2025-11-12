package fuzzy

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
)

// LaunchTemplateVersionInfo represents a launch template version for display
type LaunchTemplateVersionInfo struct {
	LaunchTemplateID   string
	LaunchTemplateName string
	VersionNumber      int64
	VersionDescription string
	CreateTime         string
	CreatedBy          string
	DefaultVersion     bool
	IsLatest           bool
}

// AWSLaunchTemplateClientInterface defines the interface for AWS EC2 launch template operations
type AWSLaunchTemplateClientInterface interface {
	ListLaunchTemplateVersions(ctx context.Context, launchTemplateID string) ([]LaunchTemplateVersionDetail, error)
	GetConfig() aws.Config
}

// LaunchTemplateVersionDetail interface for launch template version details
type LaunchTemplateVersionDetail interface {
	GetLaunchTemplateID() string
	GetLaunchTemplateName() string
	GetVersionNumber() int64
	GetVersionDescription() string
	GetCreateTime() string
	GetCreatedBy() string
	GetDefaultVersion() bool
}

// AWSLaunchTemplateLoader loads launch template versions from AWS
type AWSLaunchTemplateLoader struct {
	client             AWSLaunchTemplateClientInterface
	launchTemplateID   string
	launchTemplateName string
}

// NewAWSLaunchTemplateLoader creates a new AWS launch template version loader
func NewAWSLaunchTemplateLoader(client AWSLaunchTemplateClientInterface, launchTemplateID, launchTemplateName string) *AWSLaunchTemplateLoader {
	return &AWSLaunchTemplateLoader{
		client:             client,
		launchTemplateID:   launchTemplateID,
		launchTemplateName: launchTemplateName,
	}
}

// LoadVersions loads all versions for a launch template
func (l *AWSLaunchTemplateLoader) LoadVersions(ctx context.Context) ([]LaunchTemplateVersionInfo, error) {
	versions, err := l.client.ListLaunchTemplateVersions(ctx, l.launchTemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to list launch template versions: %w", err)
	}

	var versionInfos []LaunchTemplateVersionInfo
	for _, v := range versions {
		versionInfos = append(versionInfos, l.convertToVersionInfo(v))
	}

	return versionInfos, nil
}

// convertToVersionInfo converts a version detail to LaunchTemplateVersionInfo
func (l *AWSLaunchTemplateLoader) convertToVersionInfo(detail LaunchTemplateVersionDetail) LaunchTemplateVersionInfo {
	return LaunchTemplateVersionInfo{
		LaunchTemplateID:   detail.GetLaunchTemplateID(),
		LaunchTemplateName: detail.GetLaunchTemplateName(),
		VersionNumber:      detail.GetVersionNumber(),
		VersionDescription: detail.GetVersionDescription(),
		CreateTime:         detail.GetCreateTime(),
		CreatedBy:          detail.GetCreatedBy(),
		DefaultVersion:     detail.GetDefaultVersion(),
	}
}

// LaunchTemplateVersionFinder provides fuzzy finding for launch template versions
type LaunchTemplateVersionFinder struct {
	loader *AWSLaunchTemplateLoader
	colors *DefaultColorManager
}

// NewLaunchTemplateVersionFinder creates a new launch template version finder
func NewLaunchTemplateVersionFinder(loader *AWSLaunchTemplateLoader, colors *DefaultColorManager) *LaunchTemplateVersionFinder {
	return &LaunchTemplateVersionFinder{
		loader: loader,
		colors: colors,
	}
}

// SelectVersionInteractive displays the fuzzy finder for version selection
func (f *LaunchTemplateVersionFinder) SelectVersionInteractive(ctx context.Context) (*LaunchTemplateVersionInfo, error) {
	// Load versions
	versions, err := f.loader.LoadVersions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load launch template versions: %w", err)
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no launch template versions found")
	}

	// Add special versions at the beginning
	allVersions := []LaunchTemplateVersionInfo{
		{
			LaunchTemplateID:   f.loader.launchTemplateID,
			LaunchTemplateName: f.loader.launchTemplateName,
			VersionNumber:      -1,
			VersionDescription: "Use the latest version",
			DefaultVersion:     false,
		},
		{
			LaunchTemplateID:   f.loader.launchTemplateID,
			LaunchTemplateName: f.loader.launchTemplateName,
			VersionNumber:      -2,
			VersionDescription: "Use the default version",
			DefaultVersion:     false,
		},
	}

	allVersions = append(allVersions, versions...)

	// Use fuzzyfinder to select
	fuzzyfinder := NewLaunchTemplateVersionFuzzyFinder(allVersions, f.colors)
	selectedIndex, err := fuzzyfinder.Select(ctx)
	if err != nil {
		return nil, err
	}

	if selectedIndex < 0 || selectedIndex >= len(allVersions) {
		return nil, fmt.Errorf("invalid version selection")
	}

	return &allVersions[selectedIndex], nil
}

// LaunchTemplateVersionFuzzyFinder wraps the fuzzyfinder for launch template versions
type LaunchTemplateVersionFuzzyFinder struct {
	versions []LaunchTemplateVersionInfo
	colors   *DefaultColorManager
}

// NewLaunchTemplateVersionFuzzyFinder creates a new fuzzy finder for launch template versions
func NewLaunchTemplateVersionFuzzyFinder(versions []LaunchTemplateVersionInfo, colors *DefaultColorManager) *LaunchTemplateVersionFuzzyFinder {
	return &LaunchTemplateVersionFuzzyFinder{
		versions: versions,
		colors:   colors,
	}
}

// Select displays the fuzzy finder and returns the selected version index
func (f *LaunchTemplateVersionFuzzyFinder) Select(ctx context.Context) (int, error) {
	// Create preview renderer
	renderer := NewLaunchTemplateVersionPreviewRenderer(f.colors)

	// Use fuzzyfinder to select with context support for Ctrl+C handling
	selectedIndex, err := fuzzyfinder.Find(
		f.versions,
		func(i int) string {
			return f.formatVersionRow(f.versions[i])
		},
		fuzzyfinder.WithPreviewWindow(func(i, width, height int) string {
			if i < 0 || i >= len(f.versions) {
				return "Select a version to view details"
			}
			return renderer.Render(&f.versions[i], width, height)
		}),
		fuzzyfinder.WithPromptString("Launch Template Version > "),
		fuzzyfinder.WithContext(ctx),
	)

	if err != nil {
		if err == fuzzyfinder.ErrAbort {
			return -1, nil // User cancelled
		}
		return -1, err
	}

	return selectedIndex, nil
}

// formatVersionRow formats a version for display in the list
func (f *LaunchTemplateVersionFuzzyFinder) formatVersionRow(v LaunchTemplateVersionInfo) string {
	versionStr := f.formatVersionNumber(v)
	description := v.VersionDescription
	if description == "" {
		description = "No description"
	}

	// Truncate description if too long
	maxDescLen := 60
	if len(description) > maxDescLen {
		description = description[:maxDescLen-3] + "..."
	}

	return fmt.Sprintf("%-15s | %s", versionStr, description)
}

// formatVersionNumber formats the version number for display
func (f *LaunchTemplateVersionFuzzyFinder) formatVersionNumber(v LaunchTemplateVersionInfo) string {
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

// GetVersionString returns the version string to use for API calls
func GetVersionString(v *LaunchTemplateVersionInfo) string {
	if v.VersionNumber == -1 {
		return "$Latest"
	}
	if v.VersionNumber == -2 {
		return "$Default"
	}
	return strconv.FormatInt(v.VersionNumber, 10)
}
