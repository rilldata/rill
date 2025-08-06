package deploy

import (
	"github.com/rilldata/rill/cli/cmd/auth"
	"github.com/rilldata/rill/cli/cmd/project"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/local"
	"github.com/spf13/cobra"
)

// DeployCmd is the guided tour for deploying rill projects to rill cloud.
func DeployCmd(ch *cmdutil.Helper) *cobra.Command {
	var managed, github, archive bool
	opts := &project.DeployOpts{}

	deployCmd := &cobra.Command{
		Use:   "deploy [<path>]",
		Short: "Deploy project to Rill Cloud",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !ch.IsAuthenticated() {
				err := auth.LoginWithTelemetry(cmd.Context(), ch, "")
				if err != nil {
					return err
				}
			}

			if len(args) > 0 {
				opts.GitPath = args[0]
			}

			if !managed && !github && !archive {
				confirmed, err := cmdutil.ConfirmPrompt("Enable automatic deploys to Rill Cloud from GitHub?", "", false)
				if err != nil {
					return err
				}
				if confirmed {
					github = true
				} else {
					managed = true
				}
			}

			if archive {
				opts.ArchiveUpload = true
				return project.DeployWithUploadFlow(cmd.Context(), ch, opts)
			}
			if managed {
				return project.DeployWithUploadFlow(cmd.Context(), ch, opts)
			}
			return project.ConnectGithubFlow(cmd.Context(), ch, opts)
		},
	}

	deployCmd.Flags().SortFlags = false
	deployCmd.Flags().StringVar(&opts.GitPath, "path", ".", "Path to project repository (default: current directory)") // This can also be a remote .git URL (undocumented feature)
	deployCmd.Flags().StringVar(&opts.SubPath, "subpath", "", "Relative path to project in the repository (for monorepos)")
	deployCmd.Flags().StringVar(&opts.RemoteName, "remote", "origin", "Remote name (default: origin)")
	deployCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Org to deploy project in")
	deployCmd.Flags().StringVar(&opts.Name, "project", "", "Project name (default: Git repo name)")
	deployCmd.Flags().StringVar(&opts.Description, "description", "", "Project description")
	deployCmd.Flags().BoolVar(&opts.Public, "public", false, "Make dashboards publicly accessible")
	deployCmd.Flags().StringVar(&opts.Provisioner, "provisioner", "", "Project provisioner")
	deployCmd.Flags().StringVar(&opts.ProdVersion, "prod-version", "latest", "Rill version (default: the latest release version)")
	deployCmd.Flags().StringVar(&opts.ProdBranch, "prod-branch", "", "Git branch to deploy from (default: the default Git branch)")
	deployCmd.Flags().IntVar(&opts.Slots, "prod-slots", local.DefaultProdSlots(ch), "Slots to allocate for production deployments")
	if !ch.IsDev() {
		if err := deployCmd.Flags().MarkHidden("prod-slots"); err != nil {
			panic(err)
		}
	}

	deployCmd.Flags().BoolVar(&managed, "managed", false, "Create project using rill managed repo")

	deployCmd.Flags().BoolVar(&archive, "archive", false, "Create project using tarballs(for testing only)")
	err := deployCmd.Flags().MarkHidden("archive")
	if err != nil {
		panic(err)
	}

	deployCmd.Flags().BoolVar(&github, "github", false, "Use github repo to create the project")

	return deployCmd
}
