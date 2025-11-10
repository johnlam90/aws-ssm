package fuzzy

import (
	"context"
	"fmt"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/aws"
	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
)

// ASGInfo represents an Auto Scaling Group for fuzzy finder display
type ASGInfo struct {
	Name                    string
	MinSize                 int32
	MaxSize                 int32
	DesiredCapacity         int32
	CurrentSize             int32
	HealthCheckType         string
	CreatedTime             time.Time
	AvailabilityZones       []string
	Tags                    map[string]string
	LaunchTemplateName      string
	LaunchConfigurationName string
}

// ASGLoader interface for loading Auto Scaling Groups
type ASGLoader interface {
	LoadASGs(ctx context.Context) ([]ASGInfo, error)
	GetASGDetails(ctx context.Context, asgName string) (*ASGInfo, error)
}

// ASGDetail is an interface to avoid circular imports
type ASGDetail interface {
	GetName() string
	GetMinSize() int32
	GetMaxSize() int32
	GetDesiredCapacity() int32
	GetCurrentSize() int32
	GetCreatedTime() time.Time
	GetTags() map[string]string
	GetLaunchTemplateName() string
	GetLaunchConfigurationName() string
	GetAvailabilityZones() []string
	GetHealthCheckType() string
}

// AWSASGClientInterface defines the interface for AWS ASG operations
type AWSASGClientInterface interface {
	ListAutoScalingGroups(ctx context.Context) ([]string, error)
	DescribeAutoScalingGroup(ctx context.Context, asgName string) (*ASGDetail, error)
	GetConfig() awsconfig.Config
}

// AWSASGLoader implements ASGLoader using AWS SDK
type AWSASGLoader struct {
	client AWSASGClientInterface
}

// NewAWSASGLoader creates a new AWS ASG loader
func NewAWSASGLoader(client AWSASGClientInterface) *AWSASGLoader {
	return &AWSASGLoader{client: client}
}

// LoadASGs loads all Auto Scaling Groups
func (l *AWSASGLoader) LoadASGs(ctx context.Context) ([]ASGInfo, error) {
	// Get list of ASG names
	asgNames, err := l.client.ListAutoScalingGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list Auto Scaling Groups: %w", err)
	}

	// Load details for each ASG
	var asgs []ASGInfo
	for _, name := range asgNames {
		asgDetail, err := l.client.DescribeAutoScalingGroup(ctx, name)
		if err != nil {
			fmt.Printf("Warning: failed to describe ASG %s: %v\n", name, err)
			continue
		}

		asgs = append(asgs, l.convertToASGInfo(asgDetail))
	}

	return asgs, nil
}

// GetASGDetails retrieves details for a specific ASG
func (l *AWSASGLoader) GetASGDetails(ctx context.Context, asgName string) (*ASGInfo, error) {
	asgDetail, err := l.client.DescribeAutoScalingGroup(ctx, asgName)
	if err != nil {
		return nil, fmt.Errorf("failed to get ASG details: %w", err)
	}

	info := l.convertToASGInfo(asgDetail)
	return &info, nil
}

// convertToASGInfo converts ASGDetail interface to ASGInfo struct
func (l *AWSASGLoader) convertToASGInfo(asgDetail *ASGDetail) ASGInfo {
	asgInfo := ASGInfo{}

	// Use the interface methods to extract fields
	asg := *asgDetail
	asgInfo.Name = asg.GetName()
	asgInfo.MinSize = asg.GetMinSize()
	asgInfo.MaxSize = asg.GetMaxSize()
	asgInfo.DesiredCapacity = asg.GetDesiredCapacity()
	asgInfo.CurrentSize = asg.GetCurrentSize()
	asgInfo.CreatedTime = asg.GetCreatedTime()
	asgInfo.Tags = asg.GetTags()
	asgInfo.LaunchTemplateName = asg.GetLaunchTemplateName()
	asgInfo.LaunchConfigurationName = asg.GetLaunchConfigurationName()
	asgInfo.AvailabilityZones = asg.GetAvailabilityZones()
	asgInfo.HealthCheckType = asg.GetHealthCheckType()

	return asgInfo
}

// ASGFinder handles ASG selection
type ASGFinder struct {
	loader ASGLoader
	colors ColorManager
}

// NewASGFinder creates a new ASG finder
func NewASGFinder(loader ASGLoader, colors ColorManager) *ASGFinder {
	return &ASGFinder{
		loader: loader,
		colors: colors,
	}
}

// SelectASGInteractive displays the fuzzy finder for ASG selection
func (f *ASGFinder) SelectASGInteractive(ctx context.Context) (*ASGInfo, error) {
	// Load ASGs
	asgs, err := f.loader.LoadASGs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load Auto Scaling Groups: %w", err)
	}

	if len(asgs) == 0 {
		return nil, fmt.Errorf("no Auto Scaling Groups found")
	}

	// Use fuzzyfinder to select
	fuzzyfinder := NewASGFuzzyFinder(asgs, f.colors)
	selectedIndex, err := fuzzyfinder.Select(ctx)
	if err != nil {
		return nil, err
	}

	if selectedIndex < 0 || selectedIndex >= len(asgs) {
		return nil, fmt.Errorf("invalid ASG selection")
	}

	return &asgs[selectedIndex], nil
}

// ASGFuzzyFinder handles the actual fuzzy finding for ASGs
type ASGFuzzyFinder struct {
	asgs   []ASGInfo
	colors ColorManager
}

// NewASGFuzzyFinder creates a new ASG fuzzy finder
func NewASGFuzzyFinder(asgs []ASGInfo, colors ColorManager) *ASGFuzzyFinder {
	return &ASGFuzzyFinder{
		asgs:   asgs,
		colors: colors,
	}
}

// Select displays the fuzzy finder and returns the selected ASG index
func (f *ASGFuzzyFinder) Select(_ context.Context) (int, error) {
	// Create preview renderer
	renderer := NewASGPreviewRenderer(f.colors)

	// Use fuzzyfinder to select
	selectedIndex, err := fuzzyfinder.Find(
		f.asgs,
		func(i int) string {
			return f.formatASGRow(f.asgs[i])
		},
		fuzzyfinder.WithPreviewWindow(func(i, width, height int) string {
			if i < 0 || i >= len(f.asgs) {
				return "Select an Auto Scaling Group to view details"
			}
			return renderer.Render(&f.asgs[i], width, height)
		}),
		fuzzyfinder.WithPromptString("Auto Scaling Group > "),
	)

	if err != nil {
		if err == fuzzyfinder.ErrAbort {
			return -1, nil // User cancelled
		}
		return -1, err
	}

	return selectedIndex, nil
}

// formatASGRow formats an ASG for display in the fuzzy finder
// Format matches v0.2.0: name | desired/min/max | current | health-check
func (f *ASGFuzzyFinder) formatASGRow(asg ASGInfo) string {
	name := asg.Name
	if name == "" {
		name = "(no name)"
	}

	// Truncate name to fit nicely
	if len(name) > 40 {
		name = name[:37] + "..."
	}

	// Format: Name | Desired/Min/Max | Current | Health Check
	scaling := fmt.Sprintf("%d/%d/%d", asg.DesiredCapacity, asg.MinSize, asg.MaxSize)

	healthCheck := asg.HealthCheckType
	if healthCheck == "" {
		healthCheck = "EC2"
	}

	return fmt.Sprintf("%-40s | %-11s | %-7d | %s",
		name,
		scaling,
		asg.CurrentSize,
		healthCheck,
	)
}
