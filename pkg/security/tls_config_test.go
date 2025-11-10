package security

import "testing"

func TestNewTLSConfigInvalidPaths(t *testing.T) {
	// Expect error when CA path provided but missing
	if _, err := NewTLSConfig("", "", "nonexistent-ca.crt"); err == nil {
		t.Fatalf("expected error for missing CA cert")
	}
}

func TestTLSConfigGetTLSConfig(t *testing.T) {
	tc, _ := NewTLSConfig("", "", "")
	if tc.GetTLSConfig() == nil {
		t.Fatalf("expected tls config")
	}
}
