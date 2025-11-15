// Package fuzzy renders AWS resources using a fuzzy finder user interface.
package fuzzy

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/charmbracelet/bubbles/list"
	"golang.org/x/sync/errgroup"
)

const maxASGDescribeConcurrency = 6

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

	if len(asgNames) == 0 {
		return []ASGInfo{}, nil
	}

	asgs := make([]ASGInfo, len(asgNames))
	workerLimit := maxASGDescribeConcurrency
	if len(asgNames) < workerLimit {
		workerLimit = len(asgNames)
	}
	sem := make(chan struct{}, workerLimit)

	g, gCtx := errgroup.WithContext(ctx)
	for idx, name := range asgNames {
		idx, name := idx, name
		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
			}

			sem <- struct{}{}
			defer func() { <-sem }()

			asgDetail, err := l.client.DescribeAutoScalingGroup(gCtx, name)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return err
				}
				fmt.Printf("Warning: failed to describe ASG %s: %v\n", name, err)
				asgs[idx] = ASGInfo{Name: name}
				return nil
			}

			asgs[idx] = l.convertToASGInfo(asgDetail)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
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
func (f *ASGFuzzyFinder) Select(ctx context.Context) (int, error) {
	fmt.Println("DEBUG: Using NEW bubbles-based fuzzy finder") // DEBUG

	// Create preview renderer
	renderer := NewASGPreviewRenderer(f.colors)

	// Convert ASGs to bubbles list items
	items := make([]list.Item, len(f.asgs))
	for i, asg := range f.asgs {
		items[i] = bubbleItem{
			title:       f.formatASGRow(asg),
			description: "",
			index:       i,
		}
	}

	fmt.Printf("DEBUG: Created %d list items\n", len(items)) // DEBUG

	// Create bubbles finder
	finder := NewBubblesFinder(
		items,
		func(i, width, height int) string {
			if i < 0 || i >= len(f.asgs) {
				return "Select an Auto Scaling Group to view details"
			}
			return renderer.Render(&f.asgs[i], width, height)
		},
		"Auto Scaling Groups",
		f.colors,
	)

	fmt.Println("DEBUG: About to call finder.Select()") // DEBUG
	selectedIndex, err := finder.Select(ctx)
	fmt.Printf("DEBUG: finder.Select() returned: index=%d, err=%v\n", selectedIndex, err) // DEBUG

	if err != nil {
		if strings.Contains(err.Error(), "cancelled") {
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
