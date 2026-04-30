package db

import "testing"

func TestDBHealthOptionsNormalize(t *testing.T) {
	options := (DBHealthOptions{}).normalize()
	if options.Name != "default" {
		t.Fatalf("name = %q, want default", options.Name)
	}
	if options.PingInterval <= 0 {
		t.Fatalf("ping interval = %s, want > 0", options.PingInterval)
	}
	if options.FailuresBeforeDegraded != 3 {
		t.Fatalf("failuresBeforeDegraded = %d, want 3", options.FailuresBeforeDegraded)
	}
}
