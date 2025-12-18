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
			err := opts.ValidateAndApplyDefaults(cmd.Context(), ch)
			if err != nil {
				return err
			}

			if !opts.Managed && !opts.ArchiveUpload && !opts.Github {
				confirmed, err := cmdutil.ConfirmPrompt("Enable automatic deploys to Rill Cloud from GitHub?", "", false)
				if err != nil {
					return err
				}
				if confirmed {
					opts.Github = true
				} else {
					opts.Managed = true
				}
			}

			if opts.ArchiveUpload {
				return project.DeployWithUploadFlow(cmd.Context(), ch, opts)
			}
			if opts.Managed {
				return project.DeployWithUploadFlow(cmd.Context(), ch, opts)
			}
			return project.ConnectGithubFlow(cmd.Context(), ch, opts)
		},
	}

	deployCmd.Flags().SortFlags = false
	deployCmd.Flags().StringVar(&opts.GitPath, "path", ".", "Path to project repository (default: current directory)")
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
	deployCmd.Flags().BoolVar(&opts.PushEnv, "push-env", true, "Push local .env file to Rill Cloud")
	if !ch.IsDev() {
		if err := deployCmd.Flags().MarkHidden("prod-slots"); err != nil {
			panic(err)
		}
	}

	deployCmd.Flags().BoolVar(&opts.Managed, "managed", false, "Create project using rill managed repo")
	deployCmd.Flags().BoolVar(&opts.ArchiveUpload, "archive", false, "Create project using tarballs(for testing only)")
	err := deployCmd.Flags().MarkHidden("archive")
	if err != nil {
		panic(err)
	}
	deployCmd.Flags().BoolVar(&opts.Github, "github", false, "Use github repo to create the project")
	// subpath cannot be used with archive or managed deploys
	deployCmd.MarkFlagsMutuallyExclusive("managed", "archive", "subpath")
	deployCmd.MarkFlagsMutuallyExclusive("managed", "archive", "github")

	deployCmd.Flags().BoolVar(&opts.SkipDeploy, "skip-deploy", false, "Skip the runtime deployment step (for testing only)")
	if !ch.IsDev() {
		err = deployCmd.Flags().MarkHidden("skip-deploy")
		if err != nil {
			panic(err)
		}
	}
	return deployCmd
}
