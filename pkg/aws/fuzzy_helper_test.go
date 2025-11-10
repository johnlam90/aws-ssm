package aws

import "testing"

func TestColumnNamesToConfigMapping(t *testing.T) {
	cfg := columnNamesToConfig([]string{"name", "instance-id", "private-ip", "state", "type", "az"})
	if !cfg.Name || !cfg.InstanceID || !cfg.PrivateIP || !cfg.State || !cfg.Type || !cfg.AZ {
		t.Fatalf("expected all flags true: %+v", cfg)
	}
	cfg2 := columnNamesToConfig([]string{"name"})
	if !cfg2.Name || cfg2.InstanceID || cfg2.PrivateIP || cfg2.State || cfg2.Type || cfg2.AZ {
		t.Fatalf("unexpected other flags set: %+v", cfg2)
	}
}
