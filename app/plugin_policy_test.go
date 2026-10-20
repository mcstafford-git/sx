package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/sleuth-io/sx/v2/internal/config"
	"github.com/sleuth-io/sx/v2/internal/manifest"
	vaultpkg "github.com/sleuth-io/sx/v2/internal/vault"
)

type policyTestVault struct {
	vaultpkg.Vault
	vaultpkg.AppPluginPolicyStore
	policy *manifest.AppPluginPolicy
	err    error
}

func (v *policyTestVault) AppPluginPolicy(context.Context) (*manifest.AppPluginPolicy, error) {
	return v.policy, v.err
}

func TestPluginPolicyFailureCaches(t *testing.T) {
	for _, resolutionFailure := range []bool{false, true} {
		for _, cache := range []string{"missing", "unreadable", "corrupt", "unsupported", "wrong-profile", "disabled", "allowlist", "open"} {
			t.Run(map[bool]string{false: "retrieval", true: "resolution"}[resolutionFailure]+"/"+cache, func(t *testing.T) {
				a := pluginTestApp(t)
				path, err := a.policyCachePath()
				if err != nil {
					t.Fatal(err)
				}
				want := "disabled"
				switch cache {
				case "unreadable":
					err = os.Mkdir(path, 0755)
				case "corrupt":
					err = os.WriteFile(path, []byte("{"), 0644)
				case "unsupported":
					err = os.WriteFile(path, []byte(`{"mode":"future"}`), 0644)
				case "wrong-profile":
					other := filepath.Join(filepath.Dir(filepath.Dir(path)), "other")
					err = os.MkdirAll(other, 0755)
					if err == nil {
						err = os.WriteFile(filepath.Join(other, "policy-cache.json"), []byte(`{"mode":"open"}`), 0644)
					}
				case "disabled", "allowlist", "open":
					want = cache
					a.cachePluginPolicy(path, PluginPolicy{Mode: cache, Allowed: []string{"hello-team"}})
				}
				if err != nil {
					t.Fatal(err)
				}
				if resolutionFailure {
					// A loadable profile with an unsupported vault type fails resolution.
					seed := `{"defaultProfile":"default","activeProfiles":["default"],"profiles":{"default":{"type":"unsupported","repositoryUrl":"x"}}}`
					if err := os.WriteFile(filepath.Join(os.Getenv("SX_CONFIG_DIR"), "config.json"), []byte(seed), 0644); err != nil {
						t.Fatal(err)
					}
					if _, err := a.currentVault(); err == nil {
						t.Fatal("expected vault resolution failure")
					}
				} else {
					a.vault = &policyTestVault{err: errors.New("offline")}
				}
				for _, p := range []PluginPolicy{a.CachedPluginPolicy(), mustPluginPolicy(t, a)} {
					if p.Mode != want || p.Allowed == nil {
						t.Fatalf("got %+v, want %s and nonnil allowed", p, want)
					}
					if cache == "allowlist" && (len(p.Allowed) != 1 || p.Allowed[0] != "hello-team") {
						t.Fatalf("lost allowlist: %+v", p)
					}
				}
			})
		}
	}
}
func mustPluginPolicy(t *testing.T, a *App) PluginPolicy {
	t.Helper()
	p, err := a.GetPluginPolicy()
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestPluginPolicySuccessfulLookups(t *testing.T) {
	for _, mode := range []string{"unsupported-backend", "no-policy", "open", "allowlist", "disabled"} {
		t.Run(mode, func(t *testing.T) {
			a := pluginTestApp(t)
			v := &policyTestVault{}
			want := mode
			switch mode {
			case "unsupported-backend":
				a.vault = struct{ vaultpkg.Vault }{}
				want = "open"
			case "no-policy":
				a.vault = v
				want = "open"
			default:
				v.policy = &manifest.AppPluginPolicy{Mode: mode}
				a.vault = v
			}
			p := mustPluginPolicy(t, a)
			if p.Mode != want || p.Allowed == nil {
				t.Fatalf("got %+v", p)
			}
			if p := a.CachedPluginPolicy(); p.Mode != want || p.Allowed == nil {
				t.Fatalf("successful lookup not cached: %+v", p)
			}
			a.vault = &policyTestVault{err: errors.New("offline")}
			if p := mustPluginPolicy(t, a); p.Mode != want {
				t.Fatalf("successful lookup not cached: %+v", p)
			}
		})
	}
}

type delayedPolicyVault struct {
	policyTestVault
	started chan struct{}
	release chan struct{}
}

func (v *delayedPolicyVault) AppPluginPolicy(context.Context) (*manifest.AppPluginPolicy, error) {
	close(v.started)
	<-v.release
	return &manifest.AppPluginPolicy{Mode: "open"}, nil
}
func TestPluginPolicySwitchDuringRequest(t *testing.T) {
	a := pluginTestApp(t)
	mpc, err := config.LoadMultiProfile()
	if err != nil {
		t.Fatal(err)
	}
	mpc.Profiles["other"] = mpc.Profiles["default"]
	if err := config.SaveMultiProfile(mpc); err != nil {
		t.Fatal(err)
	}
	v := &delayedPolicyVault{started: make(chan struct{}), release: make(chan struct{})}
	a.vault = v
	done := make(chan PluginPolicy, 1)
	go func() { p, _ := a.GetPluginPolicy(); done <- p }()
	<-v.started
	if err := mpc.SetDefaultProfile("other"); err != nil {
		t.Fatal(err)
	}
	if err := config.SaveMultiProfile(mpc); err != nil {
		t.Fatal(err)
	}
	a.resetVault()
	close(v.release)
	if p := <-done; p.Mode != "disabled" {
		t.Errorf("old profile response permitted execution: %+v", p)
	}
	if p := a.CachedPluginPolicy(); p.Mode != "disabled" {
		t.Errorf("old profile response polluted new cache: %+v", p)
	}
}
