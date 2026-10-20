package main

import (
	"context"
	"testing"
	"time"

	"github.com/sleuth-io/sx/v2/internal/mgmt"
	"github.com/sleuth-io/sx/v2/internal/stats"
	vaultpkg "github.com/sleuth-io/sx/v2/internal/vault"
)

type telemetryVault struct {
	vaultpkg.Vault
	writes int
}

func (v *telemetryVault) RecordUsageEvents(context.Context, []mgmt.UsageEvent) error {
	v.writes++
	return nil
}

func TestDesktopTelemetrySettingsAndReporting(t *testing.T) {
	t.Setenv("SX_CONFIG_DIR", t.TempDir())
	t.Setenv("SX_CACHE_DIR", t.TempDir())
	v := &telemetryVault{}
	a := newTestAppWithVault(t, v)
	settings, err := a.GetSettings()
	if err != nil || settings.Telemetry != "undecided" {
		t.Fatalf("%+v %v", settings, err)
	}
	event := mgmt.UsageEvent{Timestamp: time.Now(), AssetName: "plugin"}
	a.recordPluginUsage(event)
	if v.writes != 0 {
		t.Fatal("desktop reported without consent")
	}
	if err := a.SetTelemetry(true); err != nil {
		t.Fatal(err)
	}
	event.Timestamp = time.Now()
	a.recordPluginUsage(event)
	if v.writes != 1 {
		t.Fatal("desktop omitted enabled usage")
	}
	if err := stats.SetTelemetry(context.Background(), false); err != nil {
		t.Fatal(err)
	}
	a.recordPluginUsage(event)
	if v.writes != 1 {
		t.Fatal("running desktop ignored CLI disable")
	}
	settings, err = a.GetSettings()
	if err != nil || settings.Telemetry != "disabled" {
		t.Fatalf("%+v %v", settings, err)
	}
}

func TestExtensionAuditUnaffectedByDisabledTelemetry(t *testing.T) {
	a, _, vdir := scopedExtensionApp(t, "alice@example.com")
	t.Setenv("SX_CACHE_DIR", t.TempDir())
	a.emitExtensionInstall("audit-only", "1", "me", "fixture", false)
	waitForEvent(t, vdir, "audit", `"plugin.installed"`)
}
