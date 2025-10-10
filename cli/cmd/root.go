package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rilldata/rill/cli/cmd/admin"
	"github.com/rilldata/rill/cli/cmd/auth"
	"github.com/rilldata/rill/cli/cmd/billing"
	"github.com/rilldata/rill/cli/cmd/chat"
	"github.com/rilldata/rill/cli/cmd/deploy"
	"github.com/rilldata/rill/cli/cmd/devtool"
	"github.com/rilldata/rill/cli/cmd/docs"
	"github.com/rilldata/rill/cli/cmd/env"
	"github.com/rilldata/rill/cli/cmd/org"
	"github.com/rilldata/rill/cli/cmd/project"
	"github.com/rilldata/rill/cli/cmd/publicurl"
	"github.com/rilldata/rill/cli/cmd/query"
	"github.com/rilldata/rill/cli/cmd/runtime"
	"github.com/rilldata/rill/cli/cmd/service"
	"github.com/rilldata/rill/cli/cmd/start"
	"github.com/rilldata/rill/cli/cmd/sudo"
	sudouser "github.com/rilldata/rill/cli/cmd/sudo/user"
	"github.com/rilldata/rill/cli/cmd/token"
	"github.com/rilldata/rill/cli/cmd/uninstall"
	"github.com/rilldata/rill/cli/cmd/upgrade"
	"github.com/rilldata/rill/cli/cmd/user"
	"github.com/rilldata/rill/cli/cmd/usergroup"
	versioncmd "github.com/rilldata/rill/cli/cmd/version"
	"github.com/rilldata/rill/cli/cmd/whoami"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/status"
)

func init() {
	cobra.EnableCommandSorting = false
}

// Run initializes the root command and executes it.
// It also handles errors and prints them in a user-friendly way.
// NOTE: If you change this function, also check if you need to update testcli.Fixture.Run.
func Run(ctx context.Context, ver version.Version) {
	ch, err := cmdutil.NewHelper(ver, "")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Check version.
	// NOTE: Not using PersistentPreRunE due to this issue: https://github.com/spf13/cobra/issues/216.
	err = ch.CheckVersion(ctx)
	if err != nil {
		ch.PrintfWarn("Warning: version check failed: %v\n\n", err)
	}

	// Print warning if currently acting as an assumed user
	representingUser, err := ch.DotRill.GetRepresentingUser()
	if err != nil {
		ch.PrintfWarn("Could not parse representing user email: %v\n\n", err)
	}
	if representingUser != "" {
		expiryTime, err := ch.DotRill.GetRepresentingUserAccessTokenExpiry()
		if err != nil {
			ch.PrintfWarn("Could not parse token expiry %v\n\n", err)
		} else if time.Now().After(expiryTime) {
			// If the assumed user's token has expired, silently unassume and revert to the original user before executing the command.
			err := sudouser.UnassumeUser(ctx, ch)
			if err != nil {
				ch.PrintfWarn("Could not unassume user after the token expired: %v\n\n", err)
			}
		} else {
			ch.PrintfWarn("Warning: Running action as %q\n\n", representingUser)
		}
	}

	// Execute the root command
	err = RootCmd(ch).ExecuteContext(ctx)
	code := HandleExecuteError(ch, err)
	ch.Close()
	os.Exit(code)
}

// RootCmd creates the root command and adds all subcommands.
func RootCmd(ch *cmdutil.Helper) *cobra.Command {
	// Root command
	rootCmd := &cobra.Command{
		Use:   "rill <command> [flags]",
		Short: "A CLI for Rill",
		Long:  `Work with Rill projects directly from the command line.`,
	}
	rootCmd.Version = ch.Version.String()
	// silence usage, usage string will only show up if missing arguments/flags
	rootCmd.SilenceUsage = true
	// we want to override some error messages
	rootCmd.SilenceErrors = true
	rootCmd.PersistentFlags().BoolP("help", "h", false, "Print usage") // Overrides message for help
	rootCmd.PersistentFlags().BoolVar(&ch.Interactive, "interactive", true, "Prompt for missing required parameters")
	rootCmd.PersistentFlags().Var(&ch.Printer.Format, "format", `Output format (options: "human", "json", "csv")`)
	rootCmd.PersistentFlags().StringVar(&ch.AdminURLOverride, "api-url", ch.AdminURLOverride, "Base URL for the cloud API")
	if !ch.IsDev() {
		if err := rootCmd.PersistentFlags().MarkHidden("api-url"); err != nil {
			panic(err)
		}
	}
	rootCmd.PersistentFlags().StringVar(&ch.AdminTokenOverride, "api-token", "", "Token for authenticating with the cloud API")
	rootCmd.Flags().BoolP("version", "v", false, "Show rill version") // Adds option to get version by passing --version or -v

	// Command Groups

	// Project commands
	cmdutil.AddGroup(rootCmd, "Project", false,
		start.StartCmd(ch),
		deploy.DeployCmd(ch),
		project.ProjectCmd(ch),
		chat.ChatCmd(ch),
		query.QueryCmd(ch),
		publicurl.PublicURLCmd(ch),
		env.EnvCmd(ch),
	)

	// Organization commands
	cmdutil.AddGroup(rootCmd, "Organization", false,
		org.OrgCmd(ch),
		user.UserCmd(ch),
		usergroup.UsergroupCmd(ch),
		service.ServiceCmd(ch),
		billing.BillingCmd(ch),
	)

	// Auth commands
	cmdutil.AddGroup(rootCmd, "Auth", false,
		auth.LoginCmd(ch),
		auth.LogoutCmd(ch),
		whoami.WhoamiCmd(ch),
		token.TokenCmd(ch),
	)

	// Internal commands
	cmdutil.AddGroup(rootCmd, "Internal", !ch.IsDev(),
		// These commands are hidden from the help menu
		sudo.SudoCmd(ch),
		devtool.DevtoolCmd(ch),
		verifyInstallCmd(ch),
		admin.AdminCmd(ch),
		runtime.RuntimeCmd(ch),
	)

	// Additional sub-commands
	rootCmd.AddCommand(
		versioncmd.VersionCmd(),
		upgrade.UpgradeCmd(ch),
		uninstall.UninstallCmd(ch),
		docs.DocsCmd(ch, rootCmd),
		completionCmd(ch),
	)

	return rootCmd
}

// HandleExecuteError handles an error returned by RootCmd, returning the desired exit code.
// It contains user-friendly handling for common errors.
// NOTE (2025-03-27): This is a workaround for Cobra not supporting custom logic for printing the error from RunE.
func HandleExecuteError(ch *cmdutil.Helper, err error) int {
	if err == nil {
		return 0
	}

	errMsg := err.Error()
	if strings.Contains(errMsg, "org not found") {
		ch.Println("Org not found. Run `rill org list` to see the orgs. Run `rill org switch` to default org.")
	} else if strings.Contains(errMsg, "project not found") {
		ch.Println("Project not found. Run `rill project list` to check the list of projects.")
	} else if strings.Contains(errMsg, "auth token not found") {
		ch.Println("Auth token is invalid/expired. Login again with `rill login`.")
	} else if strings.Contains(errMsg, "not authenticated as a user") {
		ch.Println("Please log in or sign up for Rill with `rill login`.")
	} else {
		if s, ok := status.FromError(err); ok {
			// rpc error
			ch.Printf("Error: %s (%v)\n", s.Message(), s.Code())
		} else {
			ch.Printf("Error: %s\n", err.Error())
		}
	}

	return 1
}
