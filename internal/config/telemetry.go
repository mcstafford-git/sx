package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"

	"github.com/sleuth-io/sx/v2/internal/utils"
)

const TelemetryConsentVersion = 1

// TelemetryConsent is separate from legacy permission to install analytics hooks.
type TelemetryConsent struct {
	Version   int       `json:"version"`
	Enabled   bool      `json:"enabled"`
	EnabledAt time.Time `json:"enabledAt,omitzero"`
}

func (mpc *MultiProfileConfig) TelemetryDecided() bool {
	return mpc.Telemetry != nil && mpc.Telemetry.Version == TelemetryConsentVersion
}

func (mpc *MultiProfileConfig) TelemetryEnabled() bool {
	return mpc.TelemetryDecided() && mpc.Telemetry.Enabled
}

// LockTelemetry coordinates consent changes, reporting, and configuration saves.
func LockTelemetry(ctx context.Context) (*flock.Flock, error) {
	path, err := utils.GetConfigFile()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	lock := flock.New(filepath.Join(filepath.Dir(path), "telemetry.lock"))
	ok, err := lock.TryLockContext(ctx, 10*time.Millisecond)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("could not lock telemetry settings")
	}
	return lock, nil
}

// SetTelemetry changes consent while the caller discards pending usage under the same lock.
func SetTelemetry(ctx context.Context, enabled bool, discardPending func() error) error {
	lock, err := LockTelemetry(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = lock.Unlock() }()
	cfg, err := LoadMultiProfile()
	if err != nil {
		path, pathErr := utils.GetConfigFile()
		if pathErr != nil {
			return pathErr
		}
		if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
			return err
		}
		cfg = &MultiProfileConfig{Profiles: map[string]*Profile{}}
	}
	if enabled && cfg.TelemetryEnabled() {
		return nil
	}
	if !enabled {
		cfg.Telemetry = &TelemetryConsent{Version: TelemetryConsentVersion}
		if err := saveMultiProfile(cfg); err != nil {
			return err
		}
	}
	if err := discardPending(); err != nil {
		return fmt.Errorf("cannot discard pending telemetry: %w", err)
	}
	if enabled {
		cfg.Telemetry = &TelemetryConsent{Version: TelemetryConsentVersion, Enabled: true, EnabledAt: time.Now()}
		return saveMultiProfile(cfg)
	}
	return nil
}
