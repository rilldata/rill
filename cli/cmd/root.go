package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/rilldata/rill/cli/cmd/admin"
	"github.com/rilldata/rill/cli/cmd/auth"
	"github.com/rilldata/rill/cli/cmd/deploy"
	"github.com/rilldata/rill/cli/cmd/devtool"
	"github.com/rilldata/rill/cli/cmd/docs"
	"github.com/rilldata/rill/cli/cmd/env"
	"github.com/rilldata/rill/cli/cmd/org"
	"github.com/rilldata/rill/cli/cmd/project"
	"github.com/rilldata/rill/cli/cmd/runtime"
	"github.com/rilldata/rill/cli/cmd/service"
	"github.com/rilldata/rill/cli/cmd/start"
	"github.com/rilldata/rill/cli/cmd/sudo"
	"github.com/rilldata/rill/cli/cmd/upgrade"
	"github.com/rilldata/rill/cli/cmd/user"
	versioncmd "github.com/rilldata/rill/cli/cmd/version"
	"github.com/rilldata/rill/cli/cmd/whoami"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/rilldata/rill/cli/pkg/printer"
	"github.com/rilldata/rill/cli/pkg/update"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/status"
)

func init() {
	cobra.EnableCommandSorting = false
}

// defaultAdminURL is the default admin server URL.
// Users can override it with the "--api-url" flag or by setting "api-url" in ~/.rill/config.yaml.
const defaultAdminURL = "https://admin.rilldata.com"

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "rill <command>",
	Short: "Rill CLI",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ctx context.Context, ver cmdutil.Version) {
	err := runCmd(ctx, ver)
	if err != nil {
		errMsg := err.Error()
		// check for known messages
		if strings.Contains(errMsg, "org not found") {
			fmt.Println("Org not found. Run `rill org list` to see the orgs. Run `rill org switch` to default org.")
		} else if strings.Contains(errMsg, "project not found") {
			fmt.Println("Project not found. Run `rill project list` to check the list of projects.")
		} else if strings.Contains(errMsg, "auth token not found") {
			fmt.Println("Auth token is invalid/expired. Login again with `rill login`.")
		} else if strings.Contains(errMsg, "not authenticated as a user") {
			fmt.Println("Please log in or sign up for Rill with `rill login`.")
		} else {
			if s, ok := status.FromError(err); ok {
				// rpc error
				fmt.Printf("Error: %s (%v)\n", s.Message(), s.Code())
			} else {
				fmt.Printf("Error: %s\n", err.Error())
			}
		}
		os.Exit(1)
	}
}

func runCmd(ctx context.Context, ver cmdutil.Version) error {
	// Load admin token from .rill (may later be overridden by flag --api-token)
	adminTokenDefault, err := dotrill.GetAccessToken()
	if err != nil {
		return fmt.Errorf("could not parse access token from ~/.rill: %w", err)
	}

	// Load admin URL from .rill (override with --api-url)
	adminURL, err := dotrill.GetDefaultAdminURL()
	if err != nil {
		return fmt.Errorf("could not parse default api URL from ~/.rill: %w", err)
	}
	if adminURL == "" {
		adminURL = defaultAdminURL
	}

	// Load default org from .rill
	defaultOrg, err := dotrill.GetDefaultOrg()
	if err != nil {
		return fmt.Errorf("could not parse default org from ~/.rill: %w", err)
	}

	// Create cmdutil Helper
	ch := &cmdutil.Helper{
		Printer:           printer.NewPrinter(printer.FormatHuman),
		Version:           ver,
		AdminURL:          adminURL,
		AdminTokenDefault: adminTokenDefault,
		Org:               defaultOrg,
		Interactive:       true,
	}
	defer ch.Close()

	// Check version
	err = update.CheckVersion(ctx, ver.Number)
	if err != nil {
		ch.PrintfWarn("Warning: version check failed: %v\n\n", err)
	}

	// Print warning if currently acting as an assumed user
	representingUser, err := dotrill.GetRepresentingUser()
	if err != nil {
		ch.PrintfWarn("Could not parse representing user email\n\n")
	}
	if representingUser != "" {
		ch.PrintfWarn("Warning: Running action as %q\n\n", representingUser)
	}

	// Cobra config
	rootCmd.Version = ver.String()
	// silence usage, usage string will only show up if missing arguments/flags
	rootCmd.SilenceUsage = true
	// we want to override some error messages
	rootCmd.SilenceErrors = true
	rootCmd.PersistentFlags().BoolP("help", "h", false, "Print usage") // Overrides message for help
	rootCmd.PersistentFlags().BoolVar(&ch.Interactive, "interactive", true, "Prompt for missing required parameters")
	rootCmd.PersistentFlags().Var(&ch.Printer.Format, "format", `Output format (options: "human", "json", "csv")`)
	rootCmd.PersistentFlags().StringVar(&ch.AdminURL, "api-url", ch.AdminURL, "Base URL for the cloud API")
	if !ch.IsDev() {
		if err := rootCmd.PersistentFlags().MarkHidden("api-url"); err != nil {
			panic(err)
		}
	}
	rootCmd.PersistentFlags().StringVar(&ch.AdminTokenOverride, "api-token", "", "Token for authenticating with the cloud API")
	rootCmd.Flags().BoolP("version", "v", false, "Show rill version") // Adds option to get version by passing --version or -v

	// Add sub-commands
	rootCmd.AddCommand(
		start.StartCmd(ch),
		deploy.DeployCmd(ch),
		env.EnvCmd(ch),
		user.UserCmd(ch),
		org.OrgCmd(ch),
		project.ProjectCmd(ch),
		service.ServiceCmd(ch),
		auth.LoginCmd(ch),
		auth.LogoutCmd(ch),
		whoami.WhoamiCmd(ch),
		docs.DocsCmd(ch, rootCmd),
		completionCmd,
		versioncmd.VersionCmd(),
		upgrade.UpgradeCmd(ch),
		sudo.SudoCmd(ch),
		devtool.DevtoolCmd(ch),
		admin.AdminCmd(ch),
		runtime.RuntimeCmd(ch),
		verifyInstallCmd(ch),
	)

	return rootCmd.ExecuteContext(ctx)
}
