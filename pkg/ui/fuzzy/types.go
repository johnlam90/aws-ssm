package fuzzy

import (
	"context"
	"time"
)

// Instance represents an EC2 instance with additional metadata for fuzzy matching
type Instance struct {
	InstanceID       string
	Name             string
	State            string
	PrivateIP        string
	PublicIP         string
	PrivateDNS       string
	PublicDNS        string
	InstanceType     string
	AvailabilityZone string
	LaunchTime       time.Time
	Tags             map[string]string
	SecurityGroups   []string
	InstanceProfile  string
}

// SearchQuery represents a parsed search query with various filters
type SearchQuery struct {
	Raw             string
	Terms           []string          // Fuzzy search terms
	Filters         map[string]string // Exact filters: name:web, id:i-123, etc.
	TagFilters      map[string]string // Tag filters: tag:Env=prod
	NegativeFilters []string          // Negative filters: !Env=dev, !state=stopped
	IPFilters       []string          // IP address patterns
	DNSFilters      []string          // DNS name patterns
	StateFilter     string            // Instance state filter
	TypeFilter      string            // Instance type filter
	AZFilter        string            // Availability zone filter
	HasTags         []string          // Tag existence: has:Environment
	MissingTags     []string          // Tag absence: missing:Team
}

// SortField represents different sorting options
type SortField int

// Sort field constants
const (
	SortByName       SortField = iota // SortByName sorts instances by name
	SortByAZ                          // SortByAZ sorts instances by availability zone
	SortByType                        // SortByType sorts instances by instance type
	SortByLaunchTime                  // SortByLaunchTime sorts instances by launch time
	SortByState                       // SortByState sorts instances by state
	SortByID                          // SortByID sorts instances by instance ID
)

// String returns string representation of SortField
func (sf SortField) String() string {
	switch sf {
	case SortByName:
		return "Name"
	case SortByAZ:
		return "AZ"
	case SortByType:
		return "Type"
	case SortByLaunchTime:
		return "Launch Time"
	case SortByState:
		return "State"
	case SortByID:
		return "Instance ID"
	default:
		return "Unknown"
	}
}

// SortDirection represents sorting direction
type SortDirection int

// Sort direction constants
const (
	SortAsc  SortDirection = iota // SortAsc sorts in ascending order
	SortDesc                      // SortDesc sorts in descending order
)

// ColumnConfig represents column display configuration
type ColumnConfig struct {
	Name       bool
	InstanceID bool
	PrivateIP  bool
	State      bool
	Type       bool
	AZ         bool
}

// DefaultColumnConfig returns the default column configuration
func DefaultColumnConfig() ColumnConfig {
	return ColumnConfig{
		Name:       true,
		InstanceID: true,
		PrivateIP:  true,
		State:      true,
		Type:       false,
		AZ:         false,
	}
}

// WeightConfig represents search field weights for scoring
type WeightConfig struct {
	Name       int
	InstanceID int
	Tags       int
	IP         int
	DNS        int
}

// DefaultWeightConfig returns the default weight configuration
func DefaultWeightConfig() WeightConfig {
	return WeightConfig{
		Name:       5,
		InstanceID: 4,
		Tags:       3,
		IP:         2,
		DNS:        1,
	}
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	Enabled    bool
	TTLMinutes int
	CacheDir   string
}

// Config represents the overall fuzzy finder configuration
type Config struct {
	Columns      ColumnConfig
	Weights      WeightConfig
	Cache        CacheConfig
	MaxInstances int    // For fallback mode
	NoColor      bool   // Disable colors
	Width        int    // Terminal width override
	Favorites    bool   // Show favorites only
	ConfigPath   string // Path to config file
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Columns:      DefaultColumnConfig(),
		Weights:      DefaultWeightConfig(),
		Cache:        CacheConfig{Enabled: true, TTLMinutes: 5, CacheDir: ""}, // Will default to ~/.aws-ssm/cache
		MaxInstances: 10000,
		NoColor:      false,
		Width:        0, // 0 = auto-detect
		Favorites:    false,
		ConfigPath:   "", // Will default to ~/.aws-ssm/config.yaml
	}
}

// Bookmark represents a bookmarked instance
type Bookmark struct {
	InstanceID string    `json:"instance_id"`
	Name       string    `json:"name"`
	Region     string    `json:"region"`
	AddedAt    time.Time `json:"added_at"`
	Notes      string    `json:"notes,omitempty"`
}

// InstanceLoader interface for loading instances
type InstanceLoader interface {
	LoadInstances(ctx context.Context, query *SearchQuery) ([]Instance, error)
	LoadInstance(ctx context.Context, instanceID string) (*Instance, error)
	GetRegions() []string
	GetCurrentRegion() string
}

// Action represents a custom action that can be performed on instances
type Action interface {
	Name() string
	Description() string
	Key() string // Single key to trigger
	Execute(ctx context.Context, instances []Instance) error
}

// PreviewRenderer handles preview window rendering
type PreviewRenderer interface {
	Render(instance *Instance, width, height int) string
	RenderJSON(instance *Instance) string
}

// ColorManager handles color formatting
type ColorManager interface {
	StateColor(state string) string
	HighlightText(text, query string) string
	HeaderColor(text string) string
	TagColor(key, value string) string
	BoldColor(text string) string
	StatusColor(status string) string
	DimColor(text string) string
}

// StateManager manages the current state of the fuzzy finder
type StateManager struct {
	Query         *SearchQuery
	Instances     []Instance
	Filtered      []Instance
	Selected      []int // Indices of selected instances
	SortField     SortField
	SortDirection SortDirection
	Bookmarks     []Bookmark
	Config        Config
	CurrentRegion string
	QueryHistory  []string
}
