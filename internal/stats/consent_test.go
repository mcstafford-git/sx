package stats

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/sleuth-io/sx/v2/internal/config"
	"github.com/sleuth-io/sx/v2/internal/vault"
)

type usageVault struct {
	vault.Vault
	calls int
}

func (v *usageVault) PostUsageStats(context.Context, string) error { v.calls++; return nil }

func setupConsent(t *testing.T) {
	t.Helper()
	t.Setenv("SX_CONFIG_DIR", t.TempDir())
	t.Setenv("SX_CACHE_DIR", t.TempDir())
}
func TestConsentLifecycle(t *testing.T) {
	setupConsent(t)
	ctx := context.Background()
	if s, e := TelemetryStatus(); s != "undecided" || e != nil {
		t.Fatalf("%s %v", s, e)
	}
	seed := func() {
		t.Helper()
		if e := os.MkdirAll(GetQueuePath(), 0700); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(filepath.Join(GetQueuePath(), "old.json"), []byte(`{}`), 0600); e != nil {
			t.Fatal(e)
		}
	}
	seed()
	v := &usageVault{}
	if e := FlushQueue(ctx, v); e != nil {
		t.Fatal(e)
	}
	if v.calls != 0 {
		t.Fatal("sent without consent")
	}
	if _, e := os.Stat(GetQueuePath()); !os.IsNotExist(e) {
		t.Fatal("legacy queue remains")
	}
	seed()
	if e := SetTelemetry(ctx, true); e != nil {
		t.Fatal(e)
	}
	if _, e := os.Stat(GetQueuePath()); !os.IsNotExist(e) {
		t.Fatal("old events survived opt-in")
	}
	if e := EnqueueEvent(UsageEvent{AssetName: "new", Timestamp: time.Now().UTC().Format(time.RFC3339Nano)}); e != nil {
		t.Fatal(e)
	}
	if e := SetTelemetry(ctx, true); e != nil {
		t.Fatal(e)
	}
	if e := FlushQueue(ctx, v); e != nil {
		t.Fatal(e)
	}
	if v.calls != 1 {
		t.Fatalf("enabled sends: %d", v.calls)
	}
	seed()
	if e := SetTelemetry(ctx, false); e != nil {
		t.Fatal(e)
	}
	if TelemetryEnabled() {
		t.Fatal("still enabled")
	}
	if s, e := TelemetryStatus(); s != "disabled" || e != nil {
		t.Fatalf("%s %v", s, e)
	}
	if e := EnqueueEvent(UsageEvent{Timestamp: time.Now().UTC().Format(time.RFC3339Nano)}); e != nil {
		t.Fatal(e)
	}
	if _, e := os.Stat(GetQueuePath()); !os.IsNotExist(e) {
		t.Fatal("off created queue")
	}
	seed()
	if e := SetTelemetry(ctx, true); e != nil {
		t.Fatal(e)
	}
	if e := FlushQueue(ctx, v); e != nil {
		t.Fatal(e)
	}
	if v.calls != 1 {
		t.Fatal("reenable sent old events")
	}
}

func TestConsentInvalidConfigAndLegacy(t *testing.T) {
	setupConsent(t)
	yes := true
	cfg := &config.MultiProfileConfig{Profiles: map[string]*config.Profile{}, BootstrapOptions: map[string]*bool{"analytics_hook": &yes}}
	if e := config.SaveMultiProfile(cfg); e != nil {
		t.Fatal(e)
	}
	if TelemetryEnabled() {
		t.Fatal("legacy true enabled reporting")
	}
	file := filepath.Join(os.Getenv("SX_CONFIG_DIR"), "config.json")
	if e := os.WriteFile(file, []byte(`{"profiles":{},"telemetry":{"version":"bad","enabled":true}}`), 0600); e != nil {
		t.Fatal(e)
	}
	called := false
	if e := WithTelemetry(context.Background(), func() error { called = true; return nil }); e != nil {
		t.Fatal(e)
	}
	if called {
		t.Fatal("invalid config reported")
	}
	if e := SetTelemetry(context.Background(), true); e == nil {
		t.Fatal("overwrote invalid config")
	}
}

func TestQueueSymlinkDoesNotDeleteTarget(t *testing.T) {
	setupConsent(t)
	target := t.TempDir()
	file := filepath.Join(target, "preserve")
	if e := os.WriteFile(file, []byte("keep"), 0600); e != nil {
		t.Fatal(e)
	}
	if e := os.Symlink(target, GetQueuePath()); e != nil {
		t.Skip(e)
	}
	if e := SetTelemetry(context.Background(), true); e != nil {
		t.Fatal(e)
	}
	if _, e := os.Stat(file); e != nil {
		t.Fatal("deleted symlink target", e)
	}
}

func TestCleanupFailureCannotEnable(t *testing.T) {
	setupConsent(t)
	cacheFile := filepath.Join(t.TempDir(), "file")
	if e := os.WriteFile(cacheFile, []byte("file"), 0600); e != nil {
		t.Fatal(e)
	}
	t.Setenv("SX_CACHE_DIR", cacheFile)
	if e := SetTelemetry(context.Background(), true); e == nil {
		t.Fatal("cleanup unexpectedly succeeded")
	}
	if TelemetryEnabled() {
		t.Fatal("enabled after cleanup failure")
	}
}

func TestTelemetryProcess(t *testing.T) {
	mode := os.Getenv("SX_TELEMETRY_TEST_PROCESS")
	if mode == "" {
		return
	}
	err := WithTelemetry(context.Background(), func() error {
		if err := os.WriteFile(os.Getenv("SX_TELEMETRY_TEST_READY"), []byte("ready"), 0600); err != nil {
			return err
		}
		if mode == "hold" {
			deadline := time.Now().Add(5 * time.Second)
			for {
				if _, err := os.Stat(os.Getenv("SX_TELEMETRY_TEST_RELEASE")); err == nil {
					break
				}
				if time.Now().After(deadline) {
					return errors.New("release timed out")
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRevocationAcrossProcesses(t *testing.T) {
	setupConsent(t)
	ctx := context.Background()
	if err := SetTelemetry(ctx, true); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	ready := filepath.Join(dir, "ready")
	release := filepath.Join(dir, "release")
	child := func(mode, marker string) *exec.Cmd {
		c := exec.Command(os.Args[0], "-test.run=^TestTelemetryProcess$")
		c.Env = append(os.Environ(), "SX_TELEMETRY_TEST_PROCESS="+mode, "SX_TELEMETRY_TEST_READY="+marker, "SX_TELEMETRY_TEST_RELEASE="+release)
		return c
	}
	c := child("hold", ready)
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Process.Kill() }()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("reporter never started")
		}
		time.Sleep(10 * time.Millisecond)
	}
	// Disable must wait for the reporter already holding the lifecycle lock.
	done := make(chan error, 1)
	go func() { done <- SetTelemetry(ctx, false) }()
	select {
	case err := <-done:
		t.Fatalf("disable returned before active reporter completed: %v", err)
	case <-time.After(30 * time.Millisecond):
	}
	if err := os.WriteFile(release, []byte("release"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := c.Wait(); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(dir, "second-reporter")
	if output, err := child("report", marker).CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, output)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("a second process reported after disable returned")
	}
}

func TestDelayedEventCannotCrossFreshOptIn(t *testing.T) {
	setupConsent(t)
	ctx := context.Background()
	if err := SetTelemetry(ctx, true); err != nil {
		t.Fatal(err)
	}
	old := UsageEvent{Timestamp: time.Now().UTC().Format(time.RFC3339Nano)}
	if err := SetTelemetry(ctx, false); err != nil {
		t.Fatal(err)
	}
	if err := SetTelemetry(ctx, true); err != nil {
		t.Fatal(err)
	}
	if err := EnqueueEvent(old); err != nil {
		t.Fatal(err)
	}
	called := false
	occurred, err := time.Parse(time.RFC3339Nano, old.Timestamp)
	if err != nil {
		t.Fatal(err)
	}
	if err := WithUsageTelemetry(ctx, occurred, func() error { called = true; return nil }); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("old asynchronous work crossed the fresh opt-in")
	}
	events, _, err := DequeueEvents(100)
	if err != nil || len(events) != 0 {
		t.Fatalf("old events accepted: %+v %v", events, err)
	}
}

func TestDisablePersistsBeforeCleanupFailure(t *testing.T) {
	setupConsent(t)
	ctx := context.Background()
	if err := SetTelemetry(ctx, true); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SX_CACHE_DIR", file)
	if err := SetTelemetry(ctx, false); err == nil {
		t.Fatal("cleanup did not fail")
	}
	cfg, err := config.LoadMultiProfile()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.TelemetryEnabled() || !cfg.TelemetryDecided() {
		t.Fatal("off decision not persisted")
	}
}
