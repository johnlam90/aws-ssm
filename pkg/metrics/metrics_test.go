package metrics

import (
	"testing"
)

// TestHistogramObserve tests that histogram observations are correctly bucketed
func TestHistogramObserve(t *testing.T) {
	tests := []struct {
		name            string
		observations    []float64
		expectedBuckets map[float64]uint64
	}{
		{
			name:         "single observation in first bucket",
			observations: []float64{0.001},
			expectedBuckets: map[float64]uint64{
				0.005: 1,
			},
		},
		{
			name:         "single observation in middle bucket",
			observations: []float64{0.05},
			expectedBuckets: map[float64]uint64{
				0.05: 1,
			},
		},
		{
			name:         "single observation in last bucket",
			observations: []float64{5e8},
			expectedBuckets: map[float64]uint64{
				1e9: 1,
			},
		},
		{
			name:         "multiple observations",
			observations: []float64{0.001, 0.01, 0.05, 0.1, 1.0, 10.0},
			expectedBuckets: map[float64]uint64{
				0.005: 1, // 0.001
				0.01:  1, // 0.01
				0.05:  1, // 0.05
				0.1:   1, // 0.1
				1.0:   1, // 1.0
				10.0:  1, // 10.0
			},
		},
		{
			name:         "observations at bucket boundaries",
			observations: []float64{0.005, 0.01, 0.025},
			expectedBuckets: map[float64]uint64{
				0.005: 1, // 0.005
				0.01:  1, // 0.01
				0.025: 1, // 0.025
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHistogram("test_histogram", nil, nil)

			// Add observations
			for _, obs := range tt.observations {
				h.Observe(obs)
			}

			// Verify bucket counts
			buckets := h.GetBuckets()
			for bucket, expectedCount := range tt.expectedBuckets {
				actualCount, exists := buckets[bucket]
				if !exists {
					t.Errorf("bucket %f not found in histogram", bucket)
					continue
				}
				if actualCount != expectedCount {
					t.Errorf("bucket %f: expected count %d, got %d", bucket, expectedCount, actualCount)
				}
			}

			// Verify total count
			expectedTotalCount := uint64(len(tt.observations))
			if h.GetCount() != expectedTotalCount {
				t.Errorf("expected total count %d, got %d", expectedTotalCount, h.GetCount())
			}
		})
	}
}

// TestHistogramSum tests that histogram sum is correctly calculated
func TestHistogramSum(t *testing.T) {
	h := NewHistogram("test_histogram", nil, nil)

	observations := []float64{0.1, 0.2, 0.3, 0.4}
	expectedSum := 1.0

	for _, obs := range observations {
		h.Observe(obs)
	}

	if h.GetSum() != expectedSum {
		t.Errorf("expected sum %f, got %f", expectedSum, h.GetSum())
	}
}

// TestHistogramCount tests that histogram count is correctly tracked
func TestHistogramCount(t *testing.T) {
	h := NewHistogram("test_histogram", nil, nil)

	observations := []float64{0.1, 0.2, 0.3, 0.4, 0.5}

	for i, obs := range observations {
		h.Observe(obs)
		expectedCount := uint64(i + 1)
		if h.GetCount() != expectedCount {
			t.Errorf("after observation %d: expected count %d, got %d", i+1, expectedCount, h.GetCount())
		}
	}
}

// TestHistogramConcurrency tests that histogram is thread-safe
func TestHistogramConcurrency(t *testing.T) {
	h := NewHistogram("test_histogram", nil, nil)

	// Simulate concurrent observations
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(value float64) {
			h.Observe(value)
			done <- true
		}(float64(i) * 0.1)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	if h.GetCount() != 10 {
		t.Errorf("expected count 10, got %d", h.GetCount())
	}
}

// TestHistogramGetBuckets tests that GetBuckets returns a copy
func TestHistogramGetBuckets(t *testing.T) {
	h := NewHistogram("test_histogram", nil, nil)
	h.Observe(0.05)

	buckets1 := h.GetBuckets()
	buckets2 := h.GetBuckets()

	// Modify the first copy
	buckets1[0.05] = 999

	// Verify the second copy is unchanged
	if buckets2[0.05] != 1 {
		t.Errorf("GetBuckets should return a copy, but modification affected the original")
	}

	// Verify the histogram's internal state is unchanged
	if h.GetBuckets()[0.05] != 1 {
		t.Errorf("GetBuckets modification should not affect histogram state")
	}
}

// TestHistogramLargeBucket tests observation in the largest bucket
func TestHistogramLargeBucket(t *testing.T) {
	h := NewHistogram("test_histogram", nil, nil)
	h.Observe(5e8)

	buckets := h.GetBuckets()
	t.Logf("Buckets after observing 5e8: %v", buckets)

	// The value 5e8 should match the 1e9 bucket (since 5e8 <= 1e9)
	if count, exists := buckets[1e9]; !exists {
		t.Errorf("bucket 1e9 not found in histogram")
	} else if count != 1 {
		t.Errorf("bucket 1e9: expected count 1, got %d", count)
	}
}

// TestHistogramBoundaryValues tests values at exact bucket boundaries
func TestHistogramBoundaryValues(t *testing.T) {
	tests := []struct {
		name           string
		value          float64
		expectedBucket float64
		description    string
	}{
		{
			name:           "exact boundary 0.005",
			value:          0.005,
			expectedBucket: 0.005,
			description:    "Value at exact bucket boundary should be placed in that bucket",
		},
		{
			name:           "exact boundary 0.01",
			value:          0.01,
			expectedBucket: 0.01,
			description:    "Value at exact bucket boundary should be placed in that bucket",
		},
		{
			name:           "exact boundary 0.1",
			value:          0.1,
			expectedBucket: 0.1,
			description:    "Value at exact bucket boundary should be placed in that bucket",
		},
		{
			name:           "exact boundary 1.0",
			value:          1.0,
			expectedBucket: 1.0,
			description:    "Value at exact bucket boundary should be placed in that bucket",
		},
		{
			name:           "just below boundary",
			value:          0.0049,
			expectedBucket: 0.005,
			description:    "Value just below boundary should be placed in that bucket",
		},
		{
			name:           "just above boundary",
			value:          0.0051,
			expectedBucket: 0.01,
			description:    "Value just above boundary should be placed in next bucket",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHistogram("test_histogram", nil, nil)
			h.Observe(tt.value)

			buckets := h.GetBuckets()
			if count, exists := buckets[tt.expectedBucket]; !exists {
				t.Errorf("%s: bucket %f not found in histogram", tt.description, tt.expectedBucket)
			} else if count != 1 {
				t.Errorf("%s: bucket %f expected count 1, got %d", tt.description, tt.expectedBucket, count)
			}
		})
	}
}
