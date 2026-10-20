package config

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestTelemetryConsent(t *testing.T) {
	t.Setenv("SX_CONFIG_DIR", t.TempDir())
	for _, raw := range []string{`{}`, `{"version":0,"enabled":true}`, `{"version":2,"enabled":true}`, `{"version":1,"enabled":false}`, `{"version":1,"enabled":true}`} {
		var consent TelemetryConsent
		if err := json.Unmarshal([]byte(raw), &consent); err != nil {
			t.Fatal(err)
		}
		yes := true
		cfg := &MultiProfileConfig{Profiles: map[string]*Profile{"a": {}, "b": {}}, DefaultProfile: "a", Telemetry: &consent, BootstrapOptions: map[string]*bool{"analytics_hook": &yes}}
		want := consent.Version == 1 && consent.Enabled
		if cfg.GetBootstrapOption("analytics_hook") != want || !cfg.GetBootstrapOption("session_hook") {
			t.Fatal(raw)
		}
		data, err := json.Marshal(cfg)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(os.Getenv("SX_CONFIG_DIR"), "config.json"), data, 0600); err != nil {
			t.Fatal(err)
		}
		loaded, err := LoadMultiProfile()
		if err != nil {
			t.Fatal(err)
		}
		if loaded.TelemetryEnabled() != want {
			t.Fatal(raw)
		}
		if err := loaded.SetDefaultProfile("b"); err != nil {
			t.Fatal(err)
		}
		if err := SaveMultiProfile(loaded); err != nil {
			t.Fatal(err)
		}
		loaded, err = LoadMultiProfile()
		if err != nil || loaded.TelemetryEnabled() != want {
			t.Fatalf("profile round trip: %v", err)
		}
	}
	for _, legacy := range []bool{false, true} {
		cfg := &MultiProfileConfig{BootstrapOptions: map[string]*bool{"analytics_hook": &legacy}}
		if cfg.TelemetryDecided() || cfg.GetBootstrapOption("analytics_hook") {
			t.Fatal("legacy consent accepted")
		}
	}
	var cfg MultiProfileConfig
	if json.Unmarshal([]byte(`{"telemetry":{"version":"bad","enabled":true}}`), &cfg) == nil {
		t.Fatal("invalid version accepted")
	}
}

func TestSaveMultiProfilePreservesNewerConsent(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		t.Run(strconv.FormatBool(enabled), func(t *testing.T) {
			t.Setenv("SX_CONFIG_DIR", t.TempDir())
			ctx := context.Background()
			clear := func() error { return nil }
			cfg := &MultiProfileConfig{Profiles: map[string]*Profile{"a": {}, "b": {}}, DefaultProfile: "a"}
			if err := SaveMultiProfile(cfg); err != nil {
				t.Fatal(err)
			}
			if err := SetTelemetry(ctx, !enabled, clear); err != nil {
				t.Fatal(err)
			}
			stale, err := LoadMultiProfile()
			if err != nil {
				t.Fatal(err)
			}
			if err := SetTelemetry(ctx, enabled, clear); err != nil {
				t.Fatal(err)
			}
			current, err := LoadMultiProfile()
			if err != nil {
				t.Fatal(err)
			}
			if err := stale.SetDefaultProfile("b"); err != nil {
				t.Fatal(err)
			}
			if err := SaveMultiProfile(stale); err != nil {
				t.Fatal(err)
			}
			actual, err := LoadMultiProfile()
			if err != nil {
				t.Fatal(err)
			}
			if actual.TelemetryEnabled() != enabled || *actual.Telemetry != *current.Telemetry {
				t.Fatal("stale save overwrote newer consent")
			}
			if actual.DefaultProfile != "b" {
				t.Fatal("profile edit was lost")
			}
		})
	}
}

func TestSaveMultiProfileCannotGrantConsent(t *testing.T) {
	t.Setenv("SX_CONFIG_DIR", t.TempDir())
	cfg := &MultiProfileConfig{Profiles: map[string]*Profile{}, Telemetry: &TelemetryConsent{Version: TelemetryConsentVersion, Enabled: true}}
	if err := SaveMultiProfile(cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadMultiProfile()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Telemetry != nil {
		t.Fatal("ordinary save granted consent")
	}
}

func TestSaveMultiProfileSharesTelemetryLock(t *testing.T) {
	t.Setenv("SX_CONFIG_DIR", t.TempDir())
	cfg := &MultiProfileConfig{Profiles: map[string]*Profile{"a": {}}}
	if err := SaveMultiProfile(cfg); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	lock, err := LockTelemetry(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lock.Unlock() }()
	done := make(chan error, 1)
	go func() { done <- SaveMultiProfile(cfg) }()
	select {
	case err := <-done:
		t.Fatalf("save bypassed telemetry lock: %v", err)
	case <-time.After(30 * time.Millisecond):
	}
	if err := lock.Unlock(); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
