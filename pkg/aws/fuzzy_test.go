package aws

import (
	"testing"

	"github.com/aws-ssm/pkg/ui/fuzzy"
)

func TestColumnNamesToConfig(t *testing.T) {
	tests := []struct {
		name     string
		columns  []string
		expected fuzzy.ColumnConfig
	}{
		{
			name:    "empty columns",
			columns: []string{},
			expected: fuzzy.ColumnConfig{
				Name:       false,
				InstanceID: false,
				PrivateIP:  false,
				State:      false,
				Type:       false,
				AZ:         false,
			},
		},
		{
			name:    "single column - name",
			columns: []string{"name"},
			expected: fuzzy.ColumnConfig{
				Name:       true,
				InstanceID: false,
				PrivateIP:  false,
				State:      false,
				Type:       false,
				AZ:         false,
			},
		},
		{
			name:    "multiple columns",
			columns: []string{"name", "instance-id", "private-ip", "state"},
			expected: fuzzy.ColumnConfig{
				Name:       true,
				InstanceID: true,
				PrivateIP:  true,
				State:      true,
				Type:       false,
				AZ:         false,
			},
		},
		{
			name:    "all columns",
			columns: []string{"name", "instance-id", "private-ip", "state", "type", "az"},
			expected: fuzzy.ColumnConfig{
				Name:       true,
				InstanceID: true,
				PrivateIP:  true,
				State:      true,
				Type:       true,
				AZ:         true,
			},
		},
		{
			name:    "unknown column ignored",
			columns: []string{"name", "unknown-column", "state"},
			expected: fuzzy.ColumnConfig{
				Name:       true,
				InstanceID: false,
				PrivateIP:  false,
				State:      true,
				Type:       false,
				AZ:         false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := columnNamesToConfig(tt.columns)
			if result != tt.expected {
				t.Errorf("columnNamesToConfig(%v) = %+v, want %+v", tt.columns, result, tt.expected)
			}
		})
	}
}

func TestClientInteractiveFlags(t *testing.T) {
	tests := []struct {
		name              string
		interactiveMode   bool
		interactiveCols   []string
		noColor           bool
		width             int
		favorites         bool
		expectedMode      bool
		expectedCols      []string
		expectedNoColor   bool
		expectedWidth     int
		expectedFavorites bool
	}{
		{
			name:              "default flags",
			interactiveMode:   false,
			interactiveCols:   []string{},
			noColor:           false,
			width:             0,
			favorites:         false,
			expectedMode:      false,
			expectedCols:      []string{},
			expectedNoColor:   false,
			expectedWidth:     0,
			expectedFavorites: false,
		},
		{
			name:              "all flags enabled",
			interactiveMode:   true,
			interactiveCols:   []string{"name", "instance-id"},
			noColor:           true,
			width:             120,
			favorites:         true,
			expectedMode:      true,
			expectedCols:      []string{"name", "instance-id"},
			expectedNoColor:   true,
			expectedWidth:     120,
			expectedFavorites: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				InteractiveMode: tt.interactiveMode,
				InteractiveCols: tt.interactiveCols,
				NoColor:         tt.noColor,
				Width:           tt.width,
				Favorites:       tt.favorites,
			}

			if client.InteractiveMode != tt.expectedMode {
				t.Errorf("InteractiveMode = %v, want %v", client.InteractiveMode, tt.expectedMode)
			}
			if len(client.InteractiveCols) != len(tt.expectedCols) {
				t.Errorf("InteractiveCols length = %d, want %d", len(client.InteractiveCols), len(tt.expectedCols))
			}
			for i, col := range client.InteractiveCols {
				if col != tt.expectedCols[i] {
					t.Errorf("InteractiveCols[%d] = %s, want %s", i, col, tt.expectedCols[i])
				}
			}
			if client.NoColor != tt.expectedNoColor {
				t.Errorf("NoColor = %v, want %v", client.NoColor, tt.expectedNoColor)
			}
			if client.Width != tt.expectedWidth {
				t.Errorf("Width = %d, want %d", client.Width, tt.expectedWidth)
			}
			if client.Favorites != tt.expectedFavorites {
				t.Errorf("Favorites = %v, want %v", client.Favorites, tt.expectedFavorites)
			}
		})
	}
}
