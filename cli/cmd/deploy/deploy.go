package deploy

import (
	"context"
	"regexp"
	"time"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

var nonSlugRegex = regexp.MustCompile(`[^\w-]`)

const (
	pollTimeout  = 10 * time.Minute
	pollInterval = 5 * time.Second
)

// DeployCmd is the guided tour for deploying rill projects to rill cloud.
func DeployCmd(ch *cmdutil.Helper) *cobra.Command {
	opts := &Options{}

	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy project to Rill Cloud",
		RunE: func(cmd *cobra.Command, args []string) error {
			return DeployFlow(cmd.Context(), ch, opts)
		},
	}

	deployCmd.Flags().SortFlags = false
	deployCmd.Flags().StringVar(&opts.GitPath, "path", ".", "Path to project repository (default: current directory)") // This can also be a remote .git URL (undocumented feature)
	deployCmd.Flags().StringVar(&opts.SubPath, "subpath", "", "Relative path to project in the repository (for monorepos)")
	deployCmd.Flags().StringVar(&opts.RemoteName, "remote", "", "Remote name (default: first Git remote)")
	deployCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Org to deploy project in")
	deployCmd.Flags().StringVar(&opts.Name, "project", "", "Project name (default: Git repo name)")
	deployCmd.Flags().StringVar(&opts.Description, "description", "", "Project description")
	deployCmd.Flags().BoolVar(&opts.Public, "public", false, "Make dashboards publicly accessible")
	deployCmd.Flags().StringVar(&opts.Provisioner, "provisioner", "", "Project provisioner")
	deployCmd.Flags().StringVar(&opts.ProdVersion, "prod-version", "latest", "Rill version (default: the latest release version)")
	deployCmd.Flags().StringVar(&opts.ProdBranch, "prod-branch", "", "Git branch to deploy from (default: the default Git branch)")
	deployCmd.Flags().IntVar(&opts.Slots, "prod-slots", 2, "Slots to allocate for production deployments")
	deployCmd.Flags().BoolVarP(&opts.Upload, "upload", "u", false, "Upload project files to Rill managed storage instead of github")
	if !ch.IsDev() {
		if err := deployCmd.Flags().MarkHidden("prod-slots"); err != nil {
			panic(err)
		}
	}

	// 2024-02-19: We have deprecated configuration of the OLAP DB using flags in favor of using rill.yaml.
	// When the migration is complete, we can remove the flags as well as the admin-server support for them.
	deployCmd.Flags().StringVar(&opts.DBDriver, "prod-db-driver", "duckdb", "Database driver")
	deployCmd.Flags().StringVar(&opts.DBDSN, "prod-db-dsn", "", "Database driver configuration")
	if err := deployCmd.Flags().MarkHidden("prod-db-driver"); err != nil {
		panic(err)
	}
	if err := deployCmd.Flags().MarkHidden("prod-db-dsn"); err != nil {
		panic(err)
	}

	deployCmd.MarkFlagsMutuallyExclusive("upload", "subpath")
	deployCmd.MarkFlagsMutuallyExclusive("upload", "remote")
	deployCmd.MarkFlagsMutuallyExclusive("upload", "prod-branch")
	return deployCmd
}

type Options struct {
	GitPath     string
	SubPath     string
	RemoteName  string
	Name        string
	Description string
	Public      bool
	Provisioner string
	ProdVersion string
	ProdBranch  string
	DBDriver    string
	DBDSN       string
	Slots       int
	// Upload repo to rill managed storage instead of GitHub.
	Upload bool
}

func DeployFlow(ctx context.Context, ch *cmdutil.Helper, opts *Options) error {
	return nil
}
