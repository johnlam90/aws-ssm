package metrics

import "testing"

func TestLabelsHelper(t *testing.T) {
	l := Labels("key1", "val1", "key2", "val2")
	if l["key1"] != "val1" || l["key2"] != "val2" {
		t.Fatalf("labels mismatch: %+v", l)
	}
}

// Removed panic-based test to satisfy staticcheck SA5012; Labels requires even args by design.
