package fuzzy

import (
	"testing"
)

func TestParseSearchQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *SearchQuery
	}{
		{
			name:  "empty query",
			input: "",
			expected: &SearchQuery{
				Raw:             "",
				Terms:           []string{},
				Filters:         make(map[string]string),
				TagFilters:      make(map[string]string),
				NegativeFilters:  []string{},
				IPFilters:       []string{},
				DNSFilters:      []string{},
				StateFilter:     "",
				TypeFilter:      "",
				AZFilter:        "",
				HasTags:        []string{},
				MissingTags:    []string{},
			},
		},
		{
			name:  "simple fuzzy term",
			input: "web-server",
			expected: &SearchQuery{
				Raw:             "web-server",
				Terms:           []string{"web-server"},
				Filters:         make(map[string]string),
				TagFilters:      make(map[string]string),
				NegativeFilters:  []string{},
				IPFilters:       []string{},
				DNSFilters:      []string{},
				StateFilter:     "",
				TypeFilter:      "",
				AZFilter:        "",
				HasTags:        []string{},
				MissingTags:    []string{},
			},
		},
		{
			name:  "name filter",
			input: "name:web",
			expected: &SearchQuery{
				Raw:             "name:web",
				Terms:           []string{},
				Filters:         map[string]string{"name": "web"},
				TagFilters:      make(map[string]string),
				NegativeFilters:  []string{},
				IPFilters:       []string{},
				DNSFilters:      []string{},
				StateFilter:     "",
				TypeFilter:      "",
				AZFilter:        "",
				HasTags:        []string{},
				MissingTags:    []string{},
			},
		},
		{
			name:  "multiple filters",
			input: "name:web state:running tag:Env=prod",
			expected: &SearchQuery{
				Raw:             "name:web state:running tag:Env=prod",
				Terms:           []string{},
				Filters:         map[string]string{"name": "web"},
				TagFilters:      map[string]string{"Env": "prod"},
				NegativeFilters:  []string{},
				IPFilters:       []string{},
				DNSFilters:      []string{},
				StateFilter:     "running",
				TypeFilter:      "",
				AZFilter:        "",
				HasTags:        []string{},
				MissingTags:    []string{},
			},
		},
		{
			name:  "negative filter",
			input: "!Env=dev",
			expected: &SearchQuery{
				Raw:             "!Env=dev",
				Terms:           []string{},
				Filters:         make(map[string]string),
				TagFilters:      make(map[string]string),
				NegativeFilters:  []string{"Env=dev"},
				IPFilters:       []string{},
				DNSFilters:      []string{},
				StateFilter:     "",
				TypeFilter:      "",
				AZFilter:        "",
				HasTags:        []string{},
				MissingTags:    []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseSearchQuery(tt.input)
			
			if result.Raw != tt.expected.Raw {
				t.Errorf("Raw = %v, want %v", result.Raw, tt.expected.Raw)
			}
			
			if len(result.Terms) != len(tt.expected.Terms) {
				t.Errorf("Terms length = %v, want %v", len(result.Terms), len(tt.expected.Terms))
			}
			
			if result.StateFilter != tt.expected.StateFilter {
				t.Errorf("StateFilter = %v, want %v", result.StateFilter, tt.expected.StateFilter)
			}
		})
	}
}

func TestInstanceMatchesQuery(t *testing.T) {
	instance := Instance{
		InstanceID:       "i-1234567890abcdef0",
		Name:             "web-server",
		State:            "running",
		PrivateIP:        "10.0.1.100",
		PublicIP:         "54.210.123.45",
		InstanceType:     "t3.medium",
		AvailabilityZone: "us-east-1a",
		Tags: map[string]string{
			"Name":     "web-server",
			"Env":      "production",
			"Team":     "backend",
			"Version":  "v1.2.3",
		},
	}

	tests := []struct {
		name     string
		query    *SearchQuery
		expected bool
	}{
		{
			name: "matches name filter",
			query: &SearchQuery{
				Filters: map[string]string{"name": "web"},
			},
			expected: true,
		},
		{
			name: "doesn't match wrong name",
			query: &SearchQuery{
				Filters: map[string]string{"name": "database"},
			},
			expected: false,
		},
		{
			name: "matches state filter",
			query: &SearchQuery{
				StateFilter: "running",
			},
			expected: true,
		},
		{
			name: "doesn't match wrong state",
			query: &SearchQuery{
				StateFilter: "stopped",
			},
			expected: false,
		},
		{
			name: "matches tag filter",
			query: &SearchQuery{
				TagFilters: map[string]string{"Env": "production"},
			},
			expected: true,
		},
		{
			name: "matches fuzzy term in name",
			query: &SearchQuery{
				Terms: []string{"web"},
			},
			expected: true,
		},
		{
			name: "matches fuzzy term in tags",
			query: &SearchQuery{
				Terms: []string{"backend"},
			},
			expected: true,
		},
		{
			name: "doesn't match non-existent fuzzy term",
			query: &SearchQuery{
				Terms: []string{"database"},
			},
			expected: false, // No match for "database" in any field
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := instance.MatchesQuery(tt.query)
			if result != tt.expected {
				t.Errorf("MatchesQuery() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateScore(t *testing.T) {
	instance := Instance{
		InstanceID:       "i-1234567890abcdef0",
		Name:             "web-server",
		State:            "running",
		PrivateIP:        "10.0.1.100",
		PublicIP:         "54.210.123.45",
		Tags: map[string]string{
			"Name":     "web-server",
			"Env":      "production",
			"Team":     "backend",
		},
	}

	weights := DefaultWeightConfig()

	tests := []struct {
		name     string
		query    *SearchQuery
		expected float64
	}{
		{
			name: "score for name match",
			query: &SearchQuery{
				Terms: []string{"web"},
			},
			expected: 8.0, // Name weight (5) + Tag weight (3) since "web" appears in both name and Name tag
		},
		{
			name: "score for ID match",
			query: &SearchQuery{
				Terms: []string{"i-123"},
			},
			expected: 4.0, // Instance ID weight - "i-123" matches in "i-1234567890abcdef0"
		},
		{
			name: "score for tag match",
			query: &SearchQuery{
				Terms: []string{"production"},
			},
			expected: 3.0, // Tag weight - "production" matches in tags
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := instance.CalculateScore(tt.query, weights)
			if result != tt.expected {
				t.Errorf("CalculateScore() = %v, want %v", result, tt.expected)
			}
		})
	}
}
