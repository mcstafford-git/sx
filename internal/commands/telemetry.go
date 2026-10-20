package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/sleuth-io/sx/v2/internal/stats"
)

func NewTelemetryCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "telemetry", Short: "Control optional usage telemetry", Long: "Control usage telemetry shared by all profiles and the desktop app. Opting in records hook and MCP skill usage and desktop extension installs in your configured vaults (including hosted vaults). Telemetry is off until you explicitly enable it."}
	cmd.AddCommand(&cobra.Command{Use: "status", Short: "Show the current telemetry decision", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		status, err := stats.TelemetryStatus()
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Usage telemetry: %s\n", status)
		return nil
	}})
	for _, enabled := range []bool{true, false} {
		name := "disable"
		short := "Disable telemetry and discard pending local events"
		if enabled {
			name = "enable"
			short = "Opt in to hook/MCP skill usage and desktop extension install reporting to your vaults"
		}
		cmd.AddCommand(&cobra.Command{Use: name, Short: short, Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
			defer cancel()
			if err := stats.SetTelemetry(ctx, enabled); err != nil {
				return fmt.Errorf("update telemetry: %w", err)
			}
			if enabled {
				fmt.Fprintln(cmd.OutOrStdout(), "Usage telemetry enabled: hook and MCP skill usage and desktop extension installs will be recorded in your configured vaults. Run 'sx install' to configure analytics hooks for detected clients.")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "Usage telemetry disabled. Pending local events discarded; historical stats remain available.")
			}
			return nil
		}})
	}
	return cmd
}
