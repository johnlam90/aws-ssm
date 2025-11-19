package version

import (
	"strings"
	"testing"
)

func TestGetVersion(t *testing.T) {
	v := GetVersion()
	if v == "" {
		t.Error("GetVersion returned empty string")
	}
}

func TestGetFullVersion(t *testing.T) {
	v := GetFullVersion()
	if !strings.Contains(v, Version) {
		t.Error("GetFullVersion does not contain Version")
	}
	if !strings.Contains(v, GitCommit) {
		t.Error("GetFullVersion does not contain GitCommit")
	}
	if !strings.Contains(v, BuildDate) {
		t.Error("GetFullVersion does not contain BuildDate")
	}
	if !strings.Contains(v, GoVersion) {
		t.Error("GetFullVersion does not contain GoVersion")
	}
}

func TestGetBuildInfo(t *testing.T) {
	info := GetBuildInfo()
	if info["version"] != Version {
		t.Error("GetBuildInfo version mismatch")
	}
	if info["commit"] != GitCommit {
		t.Error("GetBuildInfo commit mismatch")
	}
	if info["build_date"] != BuildDate {
		t.Error("GetBuildInfo build_date mismatch")
	}
	if info["go_version"] != GoVersion {
		t.Error("GetBuildInfo go_version mismatch")
	}
	if _, ok := info["os"]; !ok {
		t.Error("GetBuildInfo missing os")
	}
	if _, ok := info["arch"]; !ok {
		t.Error("GetBuildInfo missing arch")
	}
}
