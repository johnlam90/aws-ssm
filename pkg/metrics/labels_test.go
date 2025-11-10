package metrics

import "testing"

func TestLabelsHelper(t *testing.T) {
    l := Labels("key1", "val1", "key2", "val2")
    if l["key1"] != "val1" || l["key2"] != "val2" {
        t.Fatalf("labels mismatch: %+v", l)
    }
}

func TestLabelsHelperPanic(t *testing.T) {
    defer func() {
        if r := recover(); r == nil {
            t.Fatalf("expected panic on odd arguments")
        }
    }()
    _ = Labels("only_one")
}
