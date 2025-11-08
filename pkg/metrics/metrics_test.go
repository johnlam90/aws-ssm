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
