package fuzzy

import (
	"context"
	"testing"
	"time"

	"github.com/johnlam90/aws-ssm/pkg/cache"
)

type stubLoader struct {
	instances []Instance
	single    *Instance
}

func (s *stubLoader) LoadInstances(_ context.Context, _ *SearchQuery) ([]Instance, error) {
	return s.instances, nil
}
func (s *stubLoader) LoadInstance(_ context.Context, _ string) (*Instance, error) {
	return s.single, nil
}
func (s *stubLoader) GetRegions() []string     { return []string{"us-west-2"} }
func (s *stubLoader) GetCurrentRegion() string { return "us-west-2" }

func TestCachedInstanceLoader(t *testing.T) {
	svc, err := cache.NewCacheService(t.TempDir(), 1)
	if err != nil {
		t.Fatalf("cache service err: %v", err)
	}
	base := &stubLoader{instances: []Instance{{InstanceID: "i-123", Name: "web", State: "running"}}, single: &Instance{InstanceID: "i-123"}}
	cl := NewCachedInstanceLoader(base, svc, "us-west-2", true)
	ctx := context.Background()
	res, err := cl.LoadInstances(ctx, &SearchQuery{})
	if err != nil || len(res) != 1 {
		t.Fatalf("first load failed: %v %v", err, res)
	}
	// Second load should hit cache (cannot assert directly; rely on speed)
	res2, _ := cl.LoadInstances(ctx, &SearchQuery{})
	if len(res2) != 1 {
		t.Fatalf("second load mismatch")
	}
	// Single instance load
	inst, err := cl.LoadInstance(ctx, "i-123")
	if err != nil || inst.InstanceID != "i-123" {
		t.Fatalf("single load failed")
	}
	// Expire cache artificially by short TTL
	time.Sleep(10 * time.Millisecond)
}
