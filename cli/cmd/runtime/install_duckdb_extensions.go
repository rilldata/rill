package runtime

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// InstallDuckDBExtensionsCmd adds a CLI command that forces DuckDB to install all required extensions.
// It's used to pre-hydrate the extensions cache in Docker images.
func InstallDuckDBExtensionsCmd(ch *cmdutil.Helper) *cobra.Command {
	installCmd := &cobra.Command{
		Use: "install-duckdb-extensions",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := map[string]any{"dsn": ":memory:"} // In-memory
			h, err := drivers.Open("duckdb", cfg, false, activity.NewNoopClient(), zap.NewNop())
			if err != nil {
				return fmt.Errorf("failed to open ephemeral duckdb: %w", err)
			}
			err = h.Migrate(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to migrate ephemeral duckdb: %w", err)
			}
			err = h.Close()
			if err != nil {
				return fmt.Errorf("failed to close ephemeral duckdb: %w", err)
			}
			return nil
		},
	}

	return installCmd
}
