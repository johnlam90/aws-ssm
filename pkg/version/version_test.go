package version

import (
	"runtime"
	"strings"
	"testing"
)

func TestGetVersion(t *testing.T) {
	version := GetVersion()
	if version == "" {
		t.Error("GetVersion() returned empty string")
	}
	if version != Version {
		t.Errorf("GetVersion() = %q, want %q", version, Version)
	}
}

func TestGetFullVersion(t *testing.T) {
	fullVersion := GetFullVersion()
	if fullVersion == "" {
		t.Error("GetFullVersion() returned empty string")
	}

	// Check that it contains the expected components
	if !strings.Contains(fullVersion, Version) {
		t.Errorf("GetFullVersion() doesn't contain version %q", Version)
	}
	if !strings.Contains(fullVersion, "commit:") {
		t.Error("GetFullVersion() doesn't contain 'commit:'")
	}
	if !strings.Contains(fullVersion, "built:") {
		t.Error("GetFullVersion() doesn't contain 'built:'")
	}
	if !strings.Contains(fullVersion, "go:") {
		t.Error("GetFullVersion() doesn't contain 'go:'")
	}
}

func TestGetBuildInfo(t *testing.T) {
	buildInfo := GetBuildInfo()

	// Check that all expected keys are present
	expectedKeys := []string{"version", "commit", "build_date", "go_version", "os", "arch"}
	for _, key := range expectedKeys {
		if _, exists := buildInfo[key]; !exists {
			t.Errorf("GetBuildInfo() missing key %q", key)
		}
	}

	// Check values
	if buildInfo["version"] != Version {
		t.Errorf("buildInfo[\"version\"] = %q, want %q", buildInfo["version"], Version)
	}
	if buildInfo["commit"] != GitCommit {
		t.Errorf("buildInfo[\"commit\"] = %q, want %q", buildInfo["commit"], GitCommit)
	}
	if buildInfo["build_date"] != BuildDate {
		t.Errorf("buildInfo[\"build_date\"] = %q, want %q", buildInfo["build_date"], BuildDate)
	}
	if buildInfo["go_version"] != GoVersion {
		t.Errorf("buildInfo[\"go_version\"] = %q, want %q", buildInfo["go_version"], GoVersion)
	}
	if buildInfo["os"] != runtime.GOOS {
		t.Errorf("buildInfo[\"os\"] = %q, want %q", buildInfo["os"], runtime.GOOS)
	}
	if buildInfo["arch"] != runtime.GOARCH {
		t.Errorf("buildInfo[\"arch\"] = %q, want %q", buildInfo["arch"], runtime.GOARCH)
	}
}

func TestVersion(t *testing.T) {
	// Test that the version is set
	if Version == "" {
		t.Error("Version is empty")
	}
}

func TestGoVersion(t *testing.T) {
	// Test that GoVersion is set and matches runtime
	if GoVersion == "" {
		t.Error("GoVersion is empty")
	}
	if GoVersion != runtime.Version() {
		t.Errorf("GoVersion = %q, want %q", GoVersion, runtime.Version())
	}
}

func TestGitCommit(t *testing.T) {
	// Test that GitCommit is set
	if GitCommit == "" {
		t.Error("GitCommit is empty")
	}
}

func TestBuildDate(t *testing.T) {
	// Test that BuildDate is set
	if BuildDate == "" {
		t.Error("BuildDate is empty")
	}
}
