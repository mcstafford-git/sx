package stats

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/sleuth-io/sx/v2/internal/cache"
	"github.com/sleuth-io/sx/v2/internal/config"
	"github.com/sleuth-io/sx/v2/internal/utils"
)

// TelemetryStatus reads current consent; it never treats legacy hook settings as consent.
func TelemetryStatus() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	lock, err := config.LockTelemetry(ctx)
	if err != nil {
		return "disabled", err
	}
	defer func() { _ = lock.Unlock() }()
	state, err := readTelemetryStatus()
	if state != "enabled" {
		err = errors.Join(err, clearUsageQueue())
	}
	return state, err
}

func readTelemetryStatus() (string, error) {
	cfg, err := config.LoadMultiProfile()
	if err != nil {
		path, pathErr := utils.GetConfigFile()
		if pathErr == nil {
			if _, statErr := os.Stat(path); errors.Is(statErr, os.ErrNotExist) {
				return "undecided", nil
			}
		}
		return "disabled", err
	}
	if !cfg.TelemetryDecided() {
		return "undecided", nil
	}
	if cfg.TelemetryEnabled() {
		return "enabled", nil
	}
	return "disabled", nil
}

func clearUsageQueue() error {
	root, err := cache.GetCacheDir()
	if err != nil {
		return err
	}
	// RemoveAll removes a symlink itself, never its target.
	return os.RemoveAll(filepath.Join(root, "usage-queue"))
}

// WithTelemetry serializes automatic usage writes with changes to consent.
func WithTelemetry(ctx context.Context, record func() error) error {
	return WithUsageTelemetry(ctx, time.Time{}, record)
}

// WithUsageTelemetry also rejects delayed work from before the latest opt-in.
func WithUsageTelemetry(ctx context.Context, occurred time.Time, record func() error) error {
	lock, err := config.LockTelemetry(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = lock.Unlock() }()
	cfg, err := config.LoadMultiProfile()
	if err != nil || !cfg.TelemetryEnabled() {
		return clearUsageQueue()
	}
	if !occurred.IsZero() && occurred.Before(cfg.Telemetry.EnabledAt) {
		return nil
	}
	return record()
}

// TelemetryEnabled is an early, best-effort check for event producers.
func TelemetryEnabled() bool {
	enabled := false
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = WithTelemetry(ctx, func() error { enabled = true; return nil })
	return enabled
}

// SetTelemetry changes consent and discards pending events before a fresh opt-in.
func SetTelemetry(ctx context.Context, enabled bool) error {
	return config.SetTelemetry(ctx, enabled, clearUsageQueue)
}
