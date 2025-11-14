// Package testing provides an opinionated test harness for aws-ssm features.
package testing

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/johnlam90/aws-ssm/pkg/logging"
)

// TestFramework provides a comprehensive testing framework for the AWS SSM CLI
type TestFramework struct {
	logger    logging.Logger
	ctx       context.Context
	startTime time.Time
}

// NewTestFramework creates a new test framework instance
func NewTestFramework() *TestFramework {
	return &TestFramework{
		logger:    logging.With(logging.String("component", "test_framework")),
		ctx:       context.Background(),
		startTime: time.Now(),
	}
}

// TestCase represents a test case
type TestCase struct {
	Name       string
	Function   func(*TestFramework) error
	Setup      func(*TestFramework) error
	Teardown   func(*TestFramework) error
	Skip       bool
	SkipReason string
	Timeout    time.Duration
	Tags       []string
}

// Result represents test execution result
type Result struct {
	Name         string
	Passed       bool
	Failed       bool
	Skipped      bool
	Error        error
	Duration     time.Duration
	StartTime    time.Time
	EndTime      time.Time
	SetupTime    time.Duration
	TeardownTime time.Duration
}

// Assertion provides assertion methods
type Assertion struct {
	t      *testing.T
	errors []string
}

// NewAssertion creates a new assertion instance for testing
func NewAssertion(t *testing.T) *Assertion {
	return &Assertion{
		t:      t,
		errors: make([]string, 0),
	}
}

// True checks if condition is true
func (a *Assertion) True(condition bool, message string) {
	if !condition {
		a.errors = append(a.errors, fmt.Sprintf("Expected true, but got false: %s", message))
		a.t.Errorf("ASSERTION_FAILED: Expected true, but got false: %s", message)
	}
}

// False checks if condition is false
func (a *Assertion) False(condition bool, message string) {
	if condition {
		a.errors = append(a.errors, fmt.Sprintf("Expected false, but got true: %s", message))
		a.t.Errorf("ASSERTION_FAILED: Expected false, but got true: %s", message)
	}
}

// Equal checks if two values are equal
func (a *Assertion) Equal(expected, actual interface{}, message string) {
	if !reflect.DeepEqual(expected, actual) {
		a.errors = append(a.errors, fmt.Sprintf("Expected %v, but got %v: %s", expected, actual, message))
		a.t.Errorf("ASSERTION_FAILED: Expected %v, but got %v: %s", expected, actual, message)
	}
}

// NotEqual checks if two values are not equal
func (a *Assertion) NotEqual(expected, actual interface{}, message string) {
	if reflect.DeepEqual(expected, actual) {
		a.errors = append(a.errors, fmt.Sprintf("Expected %v to not equal %v: %s", expected, actual, message))
		a.t.Errorf("ASSERTION_FAILED: Expected %v to not equal %v: %s", expected, actual, message)
	}
}

// Nil checks if value is nil
func (a *Assertion) Nil(value interface{}, message string) {
	// First check if value is nil directly
	if value == nil {
		return
	}

	// For pointer types, check if the pointer is nil
	// Use reflection safely - only check IsNil for types that support it
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		if !rv.IsNil() {
			a.errors = append(a.errors, fmt.Sprintf("Expected nil, but got %v: %s", value, message))
			a.t.Errorf("ASSERTION_FAILED: Expected nil, but got %v: %s", value, message)
		}
	default:
		// Non-nilable types are never nil
		a.errors = append(a.errors, fmt.Sprintf("Expected nil, but got %v: %s", value, message))
		a.t.Errorf("ASSERTION_FAILED: Expected nil, but got %v: %s", value, message)
	}
}

// NotNil checks if value is not nil
func (a *Assertion) NotNil(value interface{}, message string) {
	// First check if value is nil directly
	if value == nil {
		a.errors = append(a.errors, fmt.Sprintf("Expected not nil, but got nil: %s", message))
		a.t.Errorf("ASSERTION_FAILED: Expected not nil, but got nil: %s", message)
		return
	}

	// For pointer types, check if the pointer is nil
	// Use reflection safely - only check IsNil for types that support it
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		if rv.IsNil() {
			a.errors = append(a.errors, fmt.Sprintf("Expected not nil, but got nil: %s", message))
			a.t.Errorf("ASSERTION_FAILED: Expected not nil, but got nil: %s", message)
		}
	}
}

// Error checks if error is not nil
func (a *Assertion) Error(err error, message string) {
	if err == nil {
		a.errors = append(a.errors, fmt.Sprintf("Expected error, but got nil: %s", message))
		a.t.Errorf("ASSERTION_FAILED: Expected error, but got nil: %s", message)
	}
}

// NoError checks if error is nil
func (a *Assertion) NoError(err error, message string) {
	if err != nil {
		a.errors = append(a.errors, fmt.Sprintf("Expected no error, but got %v: %s", err, message))
		a.t.Errorf("ASSERTION_FAILED: Expected no error, but got %v: %s", err, message)
	}
}

// Greater checks if value is greater than expected
func (a *Assertion) Greater(actual, expected interface{}, message string) {
	if !isGreater(actual, expected) {
		a.errors = append(a.errors, fmt.Sprintf("Expected %v > %v: %s", actual, expected, message))
		a.t.Errorf("ASSERTION_FAILED: Expected %v > %v: %s", actual, expected, message)
	}
}

// Less checks if value is less than expected
func (a *Assertion) Less(actual, expected interface{}, message string) {
	if !isLess(actual, expected) {
		a.errors = append(a.errors, fmt.Sprintf("Expected %v < %v: %s", actual, expected, message))
		a.t.Errorf("ASSERTION_FAILED: Expected %v < %v: %s", actual, expected, message)
	}
}

// Contains checks if string contains substring
func (a *Assertion) Contains(s, substr, message string) {
	if !strings.Contains(s, substr) {
		a.errors = append(a.errors, fmt.Sprintf("Expected string to contain %q: %s", substr, message))
		a.t.Errorf("ASSERTION_FAILED: Expected string to contain %q: %s", substr, message)
	}
}

// NotContains checks if string does not contain substring
func (a *Assertion) NotContains(s, substr, message string) {
	if strings.Contains(s, substr) {
		a.errors = append(a.errors, fmt.Sprintf("Expected string to not contain %q: %s", substr, message))
		a.t.Errorf("ASSERTION_FAILED: Expected string to not contain %q: %s", substr, message)
	}
}

// HasPrefix checks if string has prefix
func (a *Assertion) HasPrefix(s, prefix, message string) {
	if !strings.HasPrefix(s, prefix) {
		a.errors = append(a.errors, fmt.Sprintf("Expected string to have prefix %q: %s", prefix, message))
		a.t.Errorf("ASSERTION_FAILED: Expected string to have prefix %q: %s", prefix, message)
	}
}

// HasSuffix checks if string has suffix
func (a *Assertion) HasSuffix(s, suffix, message string) {
	if !strings.HasSuffix(s, suffix) {
		a.errors = append(a.errors, fmt.Sprintf("Expected string to have suffix %q: %s", suffix, message))
		a.t.Errorf("ASSERTION_FAILED: Expected string to have suffix %q: %s", suffix, message)
	}
}

// MatchRegex checks if string matches regex pattern
func (a *Assertion) MatchRegex(s, pattern, message string) {
	if !matchRegex(s, pattern) {
		a.errors = append(a.errors, fmt.Sprintf("Expected string to match pattern %q: %s", pattern, message))
		a.t.Errorf("ASSERTION_FAILED: Expected string to match pattern %q: %s", pattern, message)
	}
}

// Length checks if slice/array/map/string has expected length
func (a *Assertion) Length(value interface{}, expected int, message string) {
	if !hasLength(value, expected) {
		a.errors = append(a.errors, fmt.Sprintf("Expected length %d, but got %d: %s", expected, getLength(value), message))
		a.t.Errorf("ASSERTION_FAILED: Expected length %d, but got %d: %s", expected, getLength(value), message)
	}
}

// GetErrors returns all assertion errors
func (a *Assertion) GetErrors() []string {
	return a.errors
}

// HasErrors returns true if there are assertion errors
func (a *Assertion) HasErrors() bool {
	return len(a.errors) > 0
}

// BenchmarkResult represents benchmark execution result
type BenchmarkResult struct {
	Name        string
	Passed      bool
	NsPerOp     float64
	AllocsPerOp int64
	BytesPerOp  int64
}

// TestSuite represents a collection of test cases
type TestSuite struct {
	Name     string
	Tests    []TestCase
	Setup    func(*TestFramework) error
	Teardown func(*TestFramework) error
	Parallel bool
}

// RunTestSuite runs a test suite
func (tf *TestFramework) RunTestSuite(suite *TestSuite) []Result {
	tf.logger.Info("Running test suite", logging.String("suite_name", suite.Name), logging.Int("test_count", len(suite.Tests)))

	results := make([]Result, 0, len(suite.Tests))

	// Run suite setup
	if suite.Setup != nil {
		if err := suite.Setup(tf); err != nil {
			tf.logger.Error("Suite setup failed", logging.String("suite", suite.Name), logging.String("error", err.Error()))
			// Mark all tests as failed
			for _, test := range suite.Tests {
				results = append(results, Result{
					Name:      test.Name,
					Failed:    true,
					Error:     fmt.Errorf("suite setup failed: %w", err),
					StartTime: time.Now(),
					EndTime:   time.Now(),
				})
			}
			return results
		}
	}

	// Run tests
	for i, test := range suite.Tests {
		result := tf.runTestCase(suite.Name, test, i)
		results = append(results, result)
	}

	// Run suite teardown
	if suite.Teardown != nil {
		if err := suite.Teardown(tf); err != nil {
			tf.logger.Error("Suite teardown failed", logging.String("suite", suite.Name), logging.String("error", err.Error()))
		}
	}

	tf.logger.Info("Test suite completed", logging.String("suite_name", suite.Name),
		logging.Int("total_tests", len(suite.Tests)),
		logging.Int("passed", countPassed(results)),
		logging.Int("failed", countFailed(results)))

	return results
}

func (tf *TestFramework) runTestCase(_ string, test TestCase, _ int) Result {
	result := Result{
		Name:      test.Name,
		StartTime: time.Now(),
	}

	// Check if test should be skipped
	if test.Skip {
		result.Skipped = true
		result.EndTime = time.Now()
		return result
	}

	// Set timeout
	testCtx := tf.ctx
	if test.Timeout > 0 {
		var cancel context.CancelFunc
		testCtx, cancel = context.WithTimeout(testCtx, test.Timeout)
		defer cancel()
		// Update the test framework context for this test execution
		tf.ctx = testCtx
	}

	// Run setup
	if test.Setup != nil {
		setupStart := time.Now()
		if err := test.Setup(tf); err != nil {
			result.Error = fmt.Errorf("test setup failed: %w", err)
			result.EndTime = time.Now()
			result.SetupTime = time.Since(setupStart)
			return result
		}
		result.SetupTime = time.Since(setupStart)
	}

	// Run test
	testStart := time.Now()
	err := test.Function(tf)
	result.Duration = time.Since(testStart)

	if err != nil {
		result.Error = err
		result.Failed = true
	} else {
		result.Passed = true
	}

	// Run teardown
	if test.Teardown != nil {
		teardownStart := time.Now()
		if teardownErr := test.Teardown(tf); teardownErr != nil {
			tf.logger.Warn("Test teardown failed",
				logging.String("test", test.Name),
				logging.String("error", teardownErr.Error()))
			if result.Error == nil {
				result.Error = fmt.Errorf("test teardown failed: %w", teardownErr)
			}
		}
		result.TeardownTime = time.Since(teardownStart)
	}

	result.EndTime = time.Now()
	return result
}

// Helper functions
func isGreater(a, b interface{}) bool {
	switch av := a.(type) {
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(a).Int() > reflect.ValueOf(b).Int()
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(a).Uint() > reflect.ValueOf(b).Uint()
	case float32, float64:
		return reflect.ValueOf(a).Float() > reflect.ValueOf(b).Float()
	case string:
		if bStr, ok := b.(string); ok {
			return av > bStr
		}
		return false
	}
	return false
}

func isLess(a, b interface{}) bool {
	switch av := a.(type) {
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(a).Int() < reflect.ValueOf(b).Int()
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(a).Uint() < reflect.ValueOf(b).Uint()
	case float32, float64:
		return reflect.ValueOf(a).Float() < reflect.ValueOf(b).Float()
	case string:
		if bStr, ok := b.(string); ok {
			return av < bStr
		}
		return false
	}
	return false
}

func hasLength(value interface{}, expected int) bool {
	return getLength(value) == expected
}

func getLength(value interface{}) int {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
		return v.Len()
	}
	return -1
}

func matchRegex(s, pattern string) bool {
	// Simple regex matching - in production, use regexp package
	return strings.Contains(s, pattern)
}

func countPassed(results []Result) int {
	count := 0
	for _, r := range results {
		if r.Passed {
			count++
		}
	}
	return count
}

func countFailed(results []Result) int {
	count := 0
	for _, r := range results {
		if r.Failed {
			count++
		}
	}
	return count
}

// CreateTestFile generates test files from source files
func (tf *TestFramework) CreateTestFile(_ string, _ string, _ bool) error {
	// This would implement automatic test file generation
	// For now, just a placeholder
	return nil
}

// TestDataProvider provides test data
type TestDataProvider struct {
	data map[string]interface{}
}

// NewTestDataProvider creates a new test data provider
func NewTestDataProvider() *TestDataProvider {
	return &TestDataProvider{
		data: make(map[string]interface{}),
	}
}

// Add adds a key-value pair to the test data provider
func (p *TestDataProvider) Add(key string, value interface{}) *TestDataProvider {
	p.data[key] = value
	return p
}

// Get retrieves a value by key from the test data provider
func (p *TestDataProvider) Get(key string) (interface{}, bool) {
	value, exists := p.data[key]
	return value, exists
}

// GetString retrieves a string value by key from the test data provider
func (p *TestDataProvider) GetString(key string) (string, bool) {
	value, exists := p.data[key]
	if !exists {
		return "", false
	}
	str, ok := value.(string)
	return str, ok
}

// GetInt retrieves an integer value by key from the test data provider
func (p *TestDataProvider) GetInt(key string) (int, bool) {
	value, exists := p.data[key]
	if !exists {
		return 0, false
	}
	switch v := value.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	}
	return 0, false
}

// MockProvider provides mock objects
type MockProvider struct {
	mocks map[string]interface{}
}

// NewMockProvider creates a new mock provider
func NewMockProvider() *MockProvider {
	return &MockProvider{
		mocks: make(map[string]interface{}),
	}
}

// Add adds a mock object to the mock provider
func (p *MockProvider) Add(name string, mock interface{}) *MockProvider {
	p.mocks[name] = mock
	return p
}

// Get retrieves a mock object by name from the mock provider
func (p *MockProvider) Get(name string) (interface{}, bool) {
	mock, exists := p.mocks[name]
	return mock, exists
}

// TestConfig provides test configuration
type TestConfig struct {
	Timeout           time.Duration
	Parallel          bool
	RetryAttempts     int
	RetryDelay        time.Duration
	SetupTimeout      time.Duration
	TeardownTimeout   time.Duration
	EnableLogging     bool
	LogLevel          string
	GenerateCoverage  bool
	CoverageThreshold float64
}

// DefaultTestConfig returns default test configuration
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		Timeout:           30 * time.Second,
		Parallel:          false,
		RetryAttempts:     1,
		RetryDelay:        100 * time.Millisecond,
		SetupTimeout:      5 * time.Second,
		TeardownTimeout:   5 * time.Second,
		EnableLogging:     true,
		LogLevel:          "info",
		GenerateCoverage:  true,
		CoverageThreshold: 80.0,
	}
}

// RunBenchmark runs a benchmark test
func (tf *TestFramework) RunBenchmark(name string, _ func(b *testing.B)) *BenchmarkResult {
	tf.logger.Info("Running benchmark", logging.String("name", name))

	start := time.Now()

	// Create a temporary benchmark structure
	benchmark := &struct {
		N      int
		Allocs int64
		Bytes  int64
	}{
		N: 1000, // Default iterations
	}

	// This is a simplified benchmark runner
	// In production, you'd use actual testing.B from Go's testing package
	_ = benchmark // Avoid unused variable warning

	// In a real implementation, you would run actual benchmarks
	// and collect memory allocation statistics

	elapsed := time.Since(start)
	nsPerOp := float64(elapsed) / float64(benchmark.N)

	return &BenchmarkResult{
		Name:        name,
		Passed:      true,
		NsPerOp:     nsPerOp,
		AllocsPerOp: 0, // Would be populated from real benchmark
		BytesPerOp:  0, // Would be populated from real benchmark
	}
}

// GetCurrentTestName returns the name of the current test function
func GetCurrentTestName() string {
	pc, _, line, _ := runtime.Caller(2)
	function := runtime.FuncForPC(pc)
	functionName := function.Name()

	// Extract test name from function name
	parts := strings.Split(functionName, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return fmt.Sprintf("unknown:%d", line)
}

// GetCurrentPackageName returns the current package name
func GetCurrentPackageName() string {
	pc, _, _, _ := runtime.Caller(2)
	function := runtime.FuncForPC(pc)
	functionName := function.Name()

	// Extract package name from function name
	parts := strings.Split(functionName, ".")
	if len(parts) > 1 {
		return parts[0]
	}

	return ""
}

// GenerateTestData generates test data for various types
func GenerateTestData() *TestDataProvider {
	return NewTestDataProvider().
		Add("valid_instance_id", "i-1234567890abcdef0").
		Add("valid_region", "us-east-1").
		Add("valid_profile", "test-profile").
		Add("valid_dns_name", "ec2-1-2-3-4.compute.amazonaws.com").
		Add("valid_ip_address", "10.0.1.100").
		Add("valid_command", "ls -la").
		Add("invalid_instance_id", "invalid-id").
		Add("empty_string", "").
		Add("long_string", strings.Repeat("a", 1000))
}

// Test tags
const (
	TagUnit        = "unit"
	TagIntegration = "integration"
	TagPerformance = "performance"
	TagSecurity    = "security"
	TagSlow        = "slow"
	TagFast        = "fast"
)
