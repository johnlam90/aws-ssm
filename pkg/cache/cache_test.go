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
    if err != nil { t.Fatalf("new cache service: %v", err) }
    key := "abc"
    if err := svc.Set(key, []int{1,2,3}, "us-west-2", "query" ); err != nil {
        t.Fatalf("set failed: %v", err)
    }
    v, ok := svc.Get(key)
    if !ok { t.Fatalf("expected cache hit") }
    // Original slice marshaled via interface{} becomes []interface{}
    arr, ok2 := v.([]interface{})
    if !ok2 || len(arr) != 3 { t.Fatalf("unexpected data: %#v", v) }
}

func TestCacheExpiration(t *testing.T) {
    dir := t.TempDir()
    svc, _ := NewCacheService(dir, 0) // ttl 0 minutes -> immediate expiry
    _ = svc.Set("k", "value", "us-east-1", "q")
    if _, ok := svc.Get("k"); ok {
        t.Fatalf("expected expired entry")
    }
}

func TestCacheStatsAndCleanup(t *testing.T) {
    dir := t.TempDir()
    svc, _ := NewCacheService(dir, 1)
    _ = svc.Set("k1", "v1", "r", "q1")
    _ = svc.Set("k2", "v2", "r", "q2")
    total, expired, size, err := svc.GetCacheStats()
    if err != nil || total != 2 || expired != 0 || size <= 0 {
        t.Fatalf("unexpected stats initial: %d %d %d %v", total, expired, size, err)
    }
    // Force one expired by manipulating timestamp
    path := filepath.Join(dir, "k1.json")
    b, _ := os.ReadFile(path)
    var e Entry
    _ = json.Unmarshal(b, &e)
    e.Timestamp = time.Now().Add(-2 * time.Hour)
    nb, _ := json.Marshal(e)
    os.WriteFile(path, nb, 0600)
    _ = svc.Cleanup()
    // After cleanup, the expired file should have been removed, leaving >=1 file
    totalAfter, expiredAfter, _, _ := svc.GetCacheStats()
    if totalAfter == 0 || expiredAfter != 0 { t.Fatalf("unexpected post-cleanup stats: total=%d expired=%d", totalAfter, expiredAfter) }
    _ = svc.Cleanup()
}

func TestGenerateCacheKey(t *testing.T) {
    k := GenerateCacheKey("us", "abc")
    if k != "us_abc" { t.Fatalf("unexpected key %s", k) }
}
