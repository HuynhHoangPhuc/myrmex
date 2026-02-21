package config

import "testing"

func TestLoad_UsesEnvOverrides(t *testing.T) {
	t.Setenv("APP_FOO", "bar")

	cfg, err := Load("app", t.TempDir())
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if got := cfg.GetString("foo"); got != "bar" {
		t.Fatalf("expected foo=bar, got %q", got)
	}
}
