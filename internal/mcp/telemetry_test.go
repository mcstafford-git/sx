package mcpserver

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/sleuth-io/sx/v2/internal/stats"
)

func TestMCPUsageConsentChangesWithoutRestart(t *testing.T) {
	t.Setenv("SX_CONFIG_DIR", t.TempDir())
	t.Setenv("SX_CACHE_DIR", t.TempDir())
	s := &Server{}
	s.ReportSkillUsage("skill", "1", time.Now())
	if _, err := os.Stat(stats.GetQueuePath()); !os.IsNotExist(err) {
		t.Fatal("MCP queued without consent")
	}
	if err := stats.SetTelemetry(context.Background(), true); err != nil {
		t.Fatal(err)
	}
	// With no configured vault, an enabled report remains in the local queue.
	s.ReportSkillUsage("skill", "1", time.Now())
	events, _, err := stats.DequeueEvents(100)
	if err != nil || len(events) != 1 {
		t.Fatalf("enabled usage: %v %v", events, err)
	}
	if err := stats.SetTelemetry(context.Background(), false); err != nil {
		t.Fatal(err)
	}
	s.ReportSkillUsage("skill", "2", time.Now())
	events, _, err = stats.DequeueEvents(100)
	if err != nil || len(events) != 0 {
		t.Fatalf("revoked usage: %v %v", events, err)
	}
}
