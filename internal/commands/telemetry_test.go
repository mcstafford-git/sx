package commands

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/sleuth-io/sx/v2/internal/config"
	"github.com/sleuth-io/sx/v2/internal/stats"
)

func TestTelemetryCommands(t *testing.T) {
	t.Setenv("SX_CONFIG_DIR", t.TempDir())
	t.Setenv("SX_CACHE_DIR", t.TempDir())
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"status"}, "undecided"}, {[]string{"enable"}, "enabled"}, {[]string{"status"}, "enabled"}, {[]string{"disable"}, "disabled"}, {[]string{"status"}, "disabled"},
	} {
		c := NewTelemetryCommand()
		out := &bytes.Buffer{}
		c.SetOut(out)
		c.SetArgs(tc.args)
		if err := c.Execute(); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), tc.want) {
			t.Fatal(out.String())
		}
	}
}

func TestOldUsageHookNoConsent(t *testing.T) {
	t.Setenv("SX_CONFIG_DIR", t.TempDir())
	t.Setenv("SX_CACHE_DIR", t.TempDir())
	called := false
	original := flushUsageQueue
	flushUsageQueue = func(context.Context) error { called = true; return nil }
	defer func() { flushUsageQueue = original }()
	cmd := NewReportUsageCommand()
	cmd.SetArgs([]string{`not json`})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("disabled hook flushed")
	}
	if _, err := os.Stat(stats.GetQueuePath()); !os.IsNotExist(err) {
		t.Fatal("disabled hook queued data")
	}
	// Invalid settings must also fail closed, before reading stdin.
	if err := os.WriteFile(filepath.Join(os.Getenv("SX_CONFIG_DIR"), "config.json"), []byte("bad"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("invalid config flushed")
	}
}

func TestInitTelemetryConsent(t *testing.T) {
	for _, tc := range []struct {
		name, input                            string
		interactive, decided, enabled, wantErr bool
	}{
		{name: "legacy_yes_defaults_no", input: "\n", interactive: true, decided: true},
		{name: "fresh_yes", input: "y\n", interactive: true, decided: true, enabled: true},
		{name: "decline", input: "n\n", interactive: true, decided: true},
		{name: "unattended_yes_flag", input: "y\n"},
		{name: "cancel", interactive: true, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("SX_CONFIG_DIR", t.TempDir())
			t.Setenv("SX_CACHE_DIR", t.TempDir())
			yes := true
			cfg := &config.MultiProfileConfig{Profiles: map[string]*config.Profile{}, BootstrapOptions: map[string]*bool{"analytics_hook": &yes}}
			if err := config.SaveMultiProfile(cfg); err != nil {
				t.Fatal(err)
			}
			c := &cobra.Command{}
			c.Flags().Bool("yes", true, "")
			c.SetIn(strings.NewReader(tc.input))
			out := &bytes.Buffer{}
			c.SetOut(out)
			err := promptTelemetryConsent(c, context.Background(), cfg, tc.interactive)
			if (err != nil) != tc.wantErr {
				t.Fatal(err)
			}
			latest, err := config.LoadMultiProfile()
			if err != nil {
				t.Fatal(err)
			}
			if latest.TelemetryDecided() != tc.decided || latest.TelemetryEnabled() != tc.enabled {
				t.Fatalf("unexpected consent: %+v", latest.Telemetry)
			}
			if tc.interactive && !strings.Contains(out.String(), "(y/N)") {
				t.Fatalf("prompt did not default no: %s", out)
			}
			if tc.decided {
				c.SetIn(strings.NewReader(""))
				if err := promptTelemetryConsent(c, context.Background(), latest, true); err != nil {
					t.Fatal("saved decision prompted again", err)
				}
			}
		})
	}
}
