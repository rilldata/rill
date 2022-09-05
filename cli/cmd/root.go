package cmd

import (
	"context"
	"os"

	"github.com/rilldata/rill/cli/cmd/dropsource"
	"github.com/rilldata/rill/cli/cmd/example"
	"github.com/rilldata/rill/cli/cmd/importsource"
	"github.com/rilldata/rill/cli/cmd/info"
	"github.com/rilldata/rill/cli/cmd/initialize"
	"github.com/rilldata/rill/cli/cmd/start"
	"github.com/rilldata/rill/cli/cmd/version"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rill <command>",
	Short: "Radically simple metrics dashboards",
	Long:  `Rill makes it easy to create and consume metrics by combining a SQL-based data modeler, real-time database, and metrics dashboard into a single product—a simple alternative to complex BI stacks`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ctx context.Context) {
	err := runCmd(ctx, Version)
	if err != nil {
		os.Exit(1)
	}
}

func runCmd(ctx context.Context, ver string) error {
	v := version.Format(ver)
	rootCmd.SetVersionTemplate(v)
	rootCmd.Version = v

	rootCmd.AddCommand(initialize.InitCmd())
	rootCmd.AddCommand(example.InitExampleCmd())
	rootCmd.AddCommand(start.StartCmd())
	rootCmd.AddCommand(info.InfoCmd())
	rootCmd.AddCommand(importsource.ImportSourceCmd())
	rootCmd.AddCommand(dropsource.DropSourceCmd())
	rootCmd.AddCommand(version.VersionCmd(ver))

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().BoolP("version", "v", false, "Show rill version")

	return rootCmd.ExecuteContext(ctx)
}
