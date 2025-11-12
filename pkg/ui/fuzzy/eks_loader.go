package fuzzy

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	fuzzyfinder "github.com/ktr0731/go-fuzzyfinder"
	"golang.org/x/sync/errgroup"
)

const maxEKSDescribeConcurrency = 6

// EKSCluster represents an EKS cluster for fuzzy finder display
type EKSCluster struct {
	Name                string
	ARN                 string
	Status              string
	Version             string
	Endpoint            string
	CreatedAt           string
	Tags                map[string]string
	NodeGroupCount      int
	FargateProfileCount int
	VpcID               string
	SubnetCount         int
	SecurityGroupCount  int
}

// EKSClusterLoader interface for loading EKS clusters
type EKSClusterLoader interface {
	LoadClusters(ctx context.Context) ([]EKSCluster, error)
	LoadCluster(ctx context.Context, clusterName string) (*EKSCluster, error)
	GetRegions() []string
	GetCurrentRegion() string
}

// AWSEKSClientInterface defines the interface for AWS EKS client operations
// Note: Cluster type is defined in pkg/aws/eks.go
type AWSEKSClientInterface interface {
	ListClusters(ctx context.Context) ([]string, error)
	DescribeClusterBasic(ctx context.Context, clusterName string) (any, error)
	GetConfig() aws.Config
}

// AWSEKSLoader implements EKSClusterLoader interface using the AWS client
type AWSEKSLoader struct {
	client        AWSEKSClientInterface
	regions       []string
	currentRegion string
	describer     ClusterDescriber       // Interface for describing clusters
	detailsCache  map[string]*EKSCluster // Cache for full cluster details
}

// ClusterDescriber interface for describing EKS clusters
// Using any instead of interface{} for Go 1.18+
type ClusterDescriber interface {
	DescribeCluster(ctx context.Context, clusterName string) (any, error)
}

// NewAWSEKSLoader creates a new AWS EKS loader
func NewAWSEKSLoader(client AWSEKSClientInterface, describer ClusterDescriber) *AWSEKSLoader {
	return &AWSEKSLoader{
		client:        client,
		describer:     describer,
		regions:       []string{client.GetConfig().Region},
		currentRegion: client.GetConfig().Region,
		detailsCache:  make(map[string]*EKSCluster),
	}
}

// LoadClusters loads basic EKS cluster information for fast initial display
// Full details (node groups, Fargate profiles) are loaded lazily when needed
func (l *AWSEKSLoader) LoadClusters(ctx context.Context) ([]EKSCluster, error) {
	clusterNames, err := l.client.ListClusters(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list EKS clusters: %w", err)
	}

	if len(clusterNames) == 0 {
		return []EKSCluster{}, nil
	}

	clusters := make([]EKSCluster, len(clusterNames))
	workerLimit := maxEKSDescribeConcurrency
	if len(clusterNames) < workerLimit {
		workerLimit = len(clusterNames)
	}
	sem := make(chan struct{}, workerLimit)

	g, gCtx := errgroup.WithContext(ctx)
	for idx, name := range clusterNames {
		idx, name := idx, name
		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
			}

			sem <- struct{}{}
			defer func() { <-sem }()

			clusterDetail, err := l.client.DescribeClusterBasic(gCtx, name)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return err
				}

				fmt.Printf("Warning: failed to describe cluster %s: %v\n", name, err)
				clusters[idx] = EKSCluster{Name: name}
				return nil
			}

			eksCluster := l.convertToEKSCluster(clusterDetail)
			clusters[idx] = eksCluster
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return clusters, nil
}

// GetClusterDetails fetches full cluster details with caching
// This is called lazily when a cluster is previewed or selected
func (l *AWSEKSLoader) GetClusterDetails(ctx context.Context, clusterName string) (*EKSCluster, error) {
	// Check cache first
	if cached, ok := l.detailsCache[clusterName]; ok {
		return cached, nil
	}

	// Fetch full cluster details
	clusterDetail, err := l.describer.DescribeCluster(ctx, clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to describe cluster %s: %w", clusterName, err)
	}

	// Convert to EKSCluster
	eksCluster := l.convertToEKSCluster(clusterDetail)

	// Cache the result
	l.detailsCache[clusterName] = &eksCluster

	return &eksCluster, nil
}

// convertToEKSCluster converts a full cluster detail to EKSCluster for display
func (l *AWSEKSLoader) convertToEKSCluster(clusterDetail any) EKSCluster {
	eksCluster := EKSCluster{}

	if clusterDetail == nil {
		return eksCluster
	}

	val := reflect.ValueOf(clusterDetail)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return eksCluster
	}

	// Extract basic fields
	l.extractBasicFields(val, &eksCluster)

	// Extract VPC information
	l.extractVPCInfo(val, &eksCluster)

	// Extract resource counts
	l.extractResourceCounts(val, &eksCluster)

	return eksCluster
}

// extractBasicFields extracts basic cluster information
func (l *AWSEKSLoader) extractBasicFields(val reflect.Value, cluster *EKSCluster) {
	// Extract string fields
	extractStringField(val, "Name", &cluster.Name)
	extractStringField(val, "ARN", &cluster.ARN)
	extractStringField(val, "Status", &cluster.Status)
	extractStringField(val, "Version", &cluster.Version)
	extractStringField(val, "Endpoint", &cluster.Endpoint)

	// Extract tags
	if tagsField := val.FieldByName("Tags"); tagsField.IsValid() && tagsField.Kind() == reflect.Map {
		cluster.Tags = extractTagMap(tagsField)
	}

	// Extract created time
	if createdField := val.FieldByName("CreatedAt"); createdField.IsValid() {
		if timeValue, ok := createdField.Interface().(time.Time); ok {
			cluster.CreatedAt = timeValue.Format("2006-01-02 15:04:05")
		}
	}
}

// extractVPCInfo extracts VPC-related information
func (l *AWSEKSLoader) extractVPCInfo(val reflect.Value, cluster *EKSCluster) {
	if vpcField := val.FieldByName("VPC"); vpcField.IsValid() {
		vpcVal := vpcField
		if vpcVal.Kind() == reflect.Ptr {
			vpcVal = vpcVal.Elem()
		}
		if vpcIDField := vpcVal.FieldByName("VpcID"); vpcIDField.IsValid() {
			cluster.VpcID = vpcIDField.String()
		}
		if subnetsField := vpcVal.FieldByName("SubnetIDs"); subnetsField.IsValid() && subnetsField.Kind() == reflect.Slice {
			cluster.SubnetCount = subnetsField.Len()
		}
		if sgField := vpcVal.FieldByName("SecurityGroupIDs"); sgField.IsValid() && sgField.Kind() == reflect.Slice {
			cluster.SecurityGroupCount = sgField.Len()
		}
	}
}

// extractResourceCounts extracts node group and Fargate profile counts
func (l *AWSEKSLoader) extractResourceCounts(val reflect.Value, cluster *EKSCluster) {
	if ngField := val.FieldByName("NodeGroups"); ngField.IsValid() && ngField.Kind() == reflect.Slice {
		cluster.NodeGroupCount = ngField.Len()
	}
	if fpField := val.FieldByName("FargateProfiles"); fpField.IsValid() && fpField.Kind() == reflect.Slice {
		cluster.FargateProfileCount = fpField.Len()
	}
}

// extractStringField safely extracts a string field using reflection
func extractStringField(val reflect.Value, fieldName string, target *string) {
	if field := val.FieldByName(fieldName); field.IsValid() {
		*target = field.String()
	}
}

// extractTagMap extracts tag map using reflection
func extractTagMap(tagsField reflect.Value) map[string]string {
	tagMap := make(map[string]string)
	for _, key := range tagsField.MapKeys() {
		tagMap[key.String()] = tagsField.MapIndex(key).String()
	}
	return tagMap
}

// LoadCluster loads a single EKS cluster by name
func (l *AWSEKSLoader) LoadCluster(_ context.Context, clusterName string) (*EKSCluster, error) {
	// Return a basic cluster entry - full details will be loaded on selection
	return &EKSCluster{
		Name: clusterName,
	}, nil
}

// GetRegions returns available regions
func (l *AWSEKSLoader) GetRegions() []string {
	return l.regions
}

// GetCurrentRegion returns the current region
func (l *AWSEKSLoader) GetCurrentRegion() string {
	return l.currentRegion
}

// ProvidedEKSLoader loads EKS clusters from a provided slice
type ProvidedEKSLoader struct {
	clusters      []EKSCluster
	currentRegion string
}

// NewProvidedEKSLoader creates a new provided EKS loader
func NewProvidedEKSLoader(clusters []EKSCluster, region string) *ProvidedEKSLoader {
	return &ProvidedEKSLoader{
		clusters:      clusters,
		currentRegion: region,
	}
}

// LoadClusters returns the provided clusters
func (l *ProvidedEKSLoader) LoadClusters(_ context.Context) ([]EKSCluster, error) {
	return l.clusters, nil
}

// LoadCluster loads a single cluster by name
func (l *ProvidedEKSLoader) LoadCluster(_ context.Context, clusterName string) (*EKSCluster, error) {
	for i, cluster := range l.clusters {
		if cluster.Name == clusterName {
			return &l.clusters[i], nil
		}
	}
	return nil, fmt.Errorf("cluster not found: %s", clusterName)
}

// GetRegions returns available regions
func (l *ProvidedEKSLoader) GetRegions() []string {
	return []string{l.currentRegion}
}

// GetCurrentRegion returns the current region
func (l *ProvidedEKSLoader) GetCurrentRegion() string {
	return l.currentRegion
}

// EKSFinder represents the EKS cluster fuzzy finder
type EKSFinder struct {
	loader EKSClusterLoader
	colors ColorManager
	config Config
}

// NewEKSFinder creates a new EKS cluster fuzzy finder
func NewEKSFinder(loader EKSClusterLoader, config Config) *EKSFinder {
	colors := NewDefaultColorManager(config.NoColor)
	return &EKSFinder{
		loader: loader,
		colors: colors,
		config: config,
	}
}

// SelectClusterInteractive displays the fuzzy finder for EKS cluster selection
func (f *EKSFinder) SelectClusterInteractive(ctx context.Context) (*EKSCluster, error) {
	// Load clusters (basic info only for fast initial display)
	clusters, err := f.loader.LoadClusters(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load EKS clusters: %w", err)
	}

	if len(clusters) == 0 {
		return nil, fmt.Errorf("no EKS clusters found")
	}

	// Use fuzzyfinder to select (pass loader for lazy loading)
	fuzzyfinder := NewEKSFuzzyFinder(clusters, f.colors, f.loader)
	selectedIndex, err := fuzzyfinder.Select(ctx)
	if err != nil {
		return nil, err
	}

	if selectedIndex < 0 || selectedIndex >= len(clusters) {
		return nil, fmt.Errorf("invalid cluster selection")
	}

	return &clusters[selectedIndex], nil
}

// EKSFuzzyFinder handles the actual fuzzy finding for EKS clusters
type EKSFuzzyFinder struct {
	clusters []EKSCluster
	colors   ColorManager
	loader   EKSClusterLoader // For lazy loading cluster details
}

// NewEKSFuzzyFinder creates a new EKS fuzzy finder
func NewEKSFuzzyFinder(clusters []EKSCluster, colors ColorManager, loader EKSClusterLoader) *EKSFuzzyFinder {
	return &EKSFuzzyFinder{
		clusters: clusters,
		colors:   colors,
		loader:   loader,
	}
}

// Select displays the fuzzy finder and returns the selected cluster index
func (f *EKSFuzzyFinder) Select(ctx context.Context) (int, error) {
	// Create preview renderer with loader for lazy loading
	renderer := NewEKSPreviewRenderer(f.colors, f.loader)

	// Use fuzzyfinder to select with context support for Ctrl+C handling
	selectedIndex, err := fuzzyfinder.Find(
		f.clusters,
		func(i int) string {
			return f.formatClusterRow(f.clusters[i])
		},
		fuzzyfinder.WithPreviewWindow(func(i, width, height int) string {
			if i < 0 || i >= len(f.clusters) {
				return "Select an EKS cluster to view details"
			}
			// Render with lazy loading - full details fetched on-demand
			return renderer.RenderWithLazyLoad(ctx, &f.clusters[i], width, height)
		}),
		fuzzyfinder.WithPromptString("EKS Cluster > "),
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

// formatClusterRow formats a cluster for display in the fuzzy finder
// Format matches v0.2.0: name | status | version
func (f *EKSFuzzyFinder) formatClusterRow(cluster EKSCluster) string {
	name := cluster.Name
	if name == "" {
		name = "(no name)"
	}

	// Truncate name to fit nicely
	if len(name) > 30 {
		name = name[:27] + "..."
	}

	// Format: Name | Status | Version
	status := cluster.Status
	if status == "" {
		status = "UNKNOWN"
	}

	version := cluster.Version
	if version == "" {
		version = "N/A"
	}

	return fmt.Sprintf("%-30s | %-10s | %s",
		name,
		status,
		version,
	)
}
