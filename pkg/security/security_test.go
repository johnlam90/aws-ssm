package security

import (
	"net"
	"testing"
	"time"

	"github.com/johnlam90/aws-ssm/pkg/logging"
)

func TestCommandValidation(t *testing.T) {
	m := InitializeSecurity()
	if err := m.ValidateCommand("ls -la"); err != nil {
		t.Fatalf("expected ls to pass: %v", err)
	}
	// Dangerous pattern should fail
	if err := m.ValidateCommand("rm -rf /"); err == nil {
		t.Fatalf("expected dangerous command to fail")
	}
}

func TestCredentialManager(t *testing.T) {
	cm := NewCredentialManager()
	if err := cm.ValidateCredentials("dev-profile"); err != nil {
		t.Fatalf("validate creds failed: %v", err)
	}
	info, ok := cm.GetCredentialInfo("dev-profile")
	if !ok || info.Profile != "dev-profile" || info.Hash == "" {
		t.Fatalf("unexpected credential info")
	}
}

func TestSecureConnWrapper(t *testing.T) {
	c1, c2 := net.Pipe()
	sc := &SecureConn{Conn: c1, timeout: 50 * time.Millisecond, logger: logging.With(logging.String("component", "test_secure_conn"))}
	go func() { _, _ = c2.Write([]byte("hello")) }()
	buf := make([]byte, 5)
	n, err := sc.Read(buf)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if string(buf[:n]) != "hello" {
		t.Fatalf("expected hello got %s", string(buf[:n]))
	}
}
