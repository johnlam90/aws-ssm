package metrics

import (
    "testing"
    "time"
)

func TestCounter(t *testing.T) {
    c := NewCounter("test_counter", nil)
    c.Inc(2)
    c.Add(3)
    c.Set(5)
    if v := c.GetValue(); v != 5 {
        t.Fatalf("expected 5 got %v", v)
    }
    m := c.ToMetric()
    if m.Name != "test_counter" || m.Type != MetricCounter || m.Value != 5 {
        t.Fatalf("unexpected metric %+v", m)
    }
}

func TestGauge(t *testing.T) {
    g := NewGauge("test_gauge", nil)
    g.Set(10)
    g.Inc(5)
    g.Dec(3)
    if v := g.GetValue(); v != 12 {
        t.Fatalf("expected 12 got %v", v)
    }
    if g.ToMetric().Value != 12 {
        t.Fatalf("metric value mismatch")
    }
}

func TestHistogramBuckets(t *testing.T) {
    h := NewHistogram("test_hist", nil, nil)
    h.Observe(0.01)
    h.Observe(0.5)
    h.Observe(5)
    if h.GetCount() != 3 {
        t.Fatalf("expected count 3 got %d", h.GetCount())
    }
    if h.GetSum() <= 0.0 {
        t.Fatalf("expected positive sum")
    }
    b := h.GetBuckets()
    if len(b) == 0 {
        t.Fatalf("expected buckets recorded")
    }
}

func TestTimer(t *testing.T) {
    tm := NewTimer("test_timer", nil)
    tc := tm.Start()
    time.Sleep(5 * time.Millisecond)
    tc.Stop()
    tc2 := tm.Start().WithLabel("k", "v")
    time.Sleep(1 * time.Millisecond)
    tc2.Stop()
    if tm.ToMetric().Name != "test_timer" {
        t.Fatalf("timer metric incorrect")
    }
}

func TestRegistry(t *testing.T) {
    r := NewRegistry()
    c := NewCounter("c", nil)
    r.Register("c", c)
    if _, ok := r.GetMetric("c"); !ok {
        t.Fatalf("expected metric registered")
    }
    if r.GetCount() != 1 {
        t.Fatalf("expected count 1")
    }
    all := r.CollectAll()
    if len(all) != 1 || all[0].Name != "c" {
        t.Fatalf("collect all mismatch")
    }
    r.Unregister("c")
    if r.GetCount() != 0 {
        t.Fatalf("expected 0 after unregister")
    }
}

func TestPerformanceMonitor(t *testing.T) {
    pm := NewPerformanceMonitor()
    tc := pm.StartOperation("op")
    time.Sleep(1 * time.Millisecond)
    tc.Stop()
    pm.RecordOperation("op", 2*time.Millisecond)
}
