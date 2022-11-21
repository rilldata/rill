package cmd

import (
	"context"
	"os"

	"github.com/rilldata/rill/cli/cmd/add"
	"github.com/rilldata/rill/cli/cmd/apply"
	"github.com/rilldata/rill/cli/cmd/docs"
	"github.com/rilldata/rill/cli/cmd/drop"
	"github.com/rilldata/rill/cli/cmd/info"
	"github.com/rilldata/rill/cli/cmd/initialize"
	"github.com/rilldata/rill/cli/cmd/source"
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
func Execute(ctx context.Context, ver string, commit string, buildDate string) {
	err := runCmd(ctx, ver, commit, buildDate)
	if err != nil {
		os.Exit(1)
	}
}

func runCmd(ctx context.Context, ver string, commit string, buildDate string) error {
	v := version.Format(ver, commit, buildDate)
	rootCmd.Version = v

	rootCmd.AddCommand(version.VersionCmd(ver, commit, buildDate))
	rootCmd.AddCommand(docs.DocsCmd())
	rootCmd.AddCommand(initialize.InitCmd())
	rootCmd.AddCommand(start.StartCmd())
	rootCmd.AddCommand(apply.ApplyCmd())
	rootCmd.AddCommand(info.InfoCmd())

	//for adding or dropping source or models we can go with either of any approch as below
	//Option 1. Source command followed by subcommands like create, add, drop or other operations will be same for models as well - ref: https://github.com/planetscale/cli/blob/main/internal/cmd/branch/branch.go
	rootCmd.AddCommand(source.SourceCmd())

	//Option 2. Add and Drop command followed by the source and other like models etc
	rootCmd.AddCommand(add.AddCmd())
	rootCmd.AddCommand(drop.DropCmd())

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
