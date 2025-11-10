package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
    // Use temp dir as fake home
    tmp := t.TempDir()
    os.Setenv("HOME", tmp)
    cfg, err := LoadConfig("")
    if err != nil {
        t.Fatalf("load config failed: %v", err)
    }
    if cfg.Interactive.MaxInstances == 0 || len(cfg.Default.Columns) == 0 {
        t.Fatalf("expected defaults populated: %+v", cfg)
    }
    if !filepath.IsAbs(cfg.Cache.CacheDir) {
        t.Fatalf("expected absolute cache dir path")
    }
}

func TestSaveAndReloadConfig(t *testing.T) {
    tmp := t.TempDir()
    os.Setenv("HOME", tmp)
    cfg, _ := LoadConfig("")
    cfg.Default.Region = "us-west-2"
    path := filepath.Join(tmp, ".aws-ssm", "config.yaml")
    if err := SaveConfig(cfg, ""); err != nil {
        t.Fatalf("save config failed: %v", err)
    }
    cfg2, err := LoadConfig(path)
    if err != nil {
        t.Fatalf("reload failed: %v", err)
    }
    if cfg2.Default.Region != "us-west-2" {
        t.Fatalf("expect region persisted")
    }
}

func TestInvalidPathOutsideHome(t *testing.T) {
    tmp := t.TempDir()
    os.Setenv("HOME", tmp)
    // Create path under different root (/tmp outside HOME) should fail
    invalidPath := filepath.Join("/tmp", "evil", "config.yaml")
    if _, err := LoadConfig(invalidPath); err == nil {
        t.Fatalf("expected error for invalid path outside HOME or /etc")
    }
}
