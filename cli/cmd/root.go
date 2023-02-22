package cmd

import (
	"context"
	"os"

	"github.com/rilldata/rill/cli/cmd/admin"
	"github.com/rilldata/rill/cli/cmd/build"
	"github.com/rilldata/rill/cli/cmd/docs"
	"github.com/rilldata/rill/cli/cmd/initialize"
	"github.com/rilldata/rill/cli/cmd/runtime"
	"github.com/rilldata/rill/cli/cmd/source"
	"github.com/rilldata/rill/cli/cmd/start"
	versioncmd "github.com/rilldata/rill/cli/cmd/version"
	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
)

func init() {
	cobra.EnableCommandSorting = false
}

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "rill <command>",
	Short: "Rill CLI",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ctx context.Context, ver version.Version) {
	err := runCmd(ctx, ver)
	if err != nil {
		os.Exit(1)
	}
}

func runCmd(ctx context.Context, ver version.Version) error {
	rootCmd.Version = ver.String()

	rootCmd.AddCommand(initialize.InitCmd(ver))
	rootCmd.AddCommand(start.StartCmd(ver))
	rootCmd.AddCommand(build.BuildCmd(ver))
	rootCmd.AddCommand(source.SourceCmd(ver))
	rootCmd.AddCommand(admin.AdminCmd(ver))
	rootCmd.AddCommand(runtime.RuntimeCmd(ver))
	rootCmd.AddCommand(docs.DocsCmd())
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(versioncmd.VersionCmd())

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cli.yaml)")
	rootCmd.PersistentFlags().BoolP("help", "h", false, "Print usage") // Overrides message for help

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().BoolP("version", "v", false, "Show rill version")

	return rootCmd.ExecuteContext(ctx)
}
