package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCacheSetAndGet(t *testing.T) {
	dir := t.TempDir()
	svc, err := NewCacheService(dir, 1)
	if err != nil {
		t.Fatalf("new cache service: %v", err)
	}
	key := "abc"
	if err := svc.Set(key, []int{1, 2, 3}, "us-west-2", "query"); err != nil {
		t.Fatalf("set failed: %v", err)
	}
	v, ok := svc.Get(key)
	if !ok {
		t.Fatalf("expected cache hit")
	}
	arr, ok2 := v.([]interface{})
	if !ok2 || len(arr) != 3 {
		t.Fatalf("unexpected data: %#v", v)
	}
}

func TestCacheExpiration(t *testing.T) {
	dir := t.TempDir()
	svc, err := NewCacheService(dir, 0)
	if err != nil {
		t.Fatalf("new cache service: %v", err)
	}
	if err := svc.Set("k", "value", "us-east-1", "q"); err != nil {
		t.Fatalf("set cache entry: %v", err)
	}
	if _, ok := svc.Get("k"); ok {
		t.Fatalf("expected expired entry")
	}
}

func TestCacheStatsAndCleanup(t *testing.T) {
	dir := t.TempDir()
	svc, err := NewCacheService(dir, 1)
	if err != nil {
		t.Fatalf("new cache service: %v", err)
	}
	if err := svc.Set("k1", "v1", "r", "q1"); err != nil {
		t.Fatalf("set first cache entry: %v", err)
	}
	if err := svc.Set("k2", "v2", "r", "q2"); err != nil {
		t.Fatalf("set second cache entry: %v", err)
	}
	total, expired, size, err := svc.GetCacheStats()
	if err != nil || total != 2 || expired != 0 || size <= 0 {
		t.Fatalf("unexpected stats initial: %d %d %d %v", total, expired, size, err)
	}
	path := filepath.Join(dir, "k1.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read cache file: %v", err)
	}
	var e Entry
	if err := json.Unmarshal(b, &e); err != nil {
		t.Fatalf("unmarshal cache entry: %v", err)
	}
	e.Timestamp = time.Now().Add(-2 * time.Hour)
	nb, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("marshal cache entry: %v", err)
	}
	if err := os.WriteFile(path, nb, 0600); err != nil {
		t.Fatalf("write cache file: %v", err)
	}
	if err := svc.Cleanup(); err != nil {
		t.Fatalf("cleanup cache: %v", err)
	}
	totalAfter, expiredAfter, _, err := svc.GetCacheStats()
	if err != nil {
		t.Fatalf("post-cleanup stats: %v", err)
	}
	if totalAfter == 0 || expiredAfter != 0 {
		t.Fatalf("unexpected post-cleanup stats: total=%d expired=%d", totalAfter, expiredAfter)
	}
	if err := svc.Cleanup(); err != nil {
		t.Fatalf("second cleanup: %v", err)
	}
}

func TestGenerateCacheKey(t *testing.T) {
	k := GenerateCacheKey("us", "abc")
	if k != "us_abc" {
		t.Fatalf("unexpected key %s", k)
	}
}
