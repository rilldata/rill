package deploy

import (
	"path/filepath"
	"strings"

	"github.com/rilldata/rill/cli/cmd/auth"
	"github.com/rilldata/rill/cli/cmd/project"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/local"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
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

			var err error
			github, err = shouldConnectGithub(opts, !managed && !archive, ch)
			if err != nil {
				return err
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

func shouldConnectGithub(opts *project.DeployOpts, inferGitSubPath bool, ch *cmdutil.Helper) (bool, error) {
	var err error
	opts.GitPath, err = fileutil.ExpandHome(opts.GitPath)
	if err != nil {
		return false, err
	}
	opts.GitPath, err = filepath.Abs(opts.GitPath)
	if err != nil {
		return false, err
	}

	if !inferGitSubPath || opts.SubPath != "" {
		return false, nil
	}

	repoRoot, err := gitutil.InferGitRepoRoot(opts.GitPath)
	if err != nil {
		return false, nil // not a git repo
	}

	remote, err := gitutil.ExtractGitRemote(repoRoot, opts.RemoteName, false)
	if err != nil {
		return false, err
	}
	if remote.URL == "" {
		// no remote configured
		return false, nil
	}
	if !strings.HasPrefix(remote.URL, "https://github.com") {
		// not a GitHub repo should not prompt for GitHub connection
		return false, nil
	}

	subPath, err := filepath.Rel(repoRoot, opts.GitPath)
	if err == nil {
		ch.PrintfBold("Detected git repository at: ")
		ch.Printf("%s\n", repoRoot)
		ch.PrintfBold("Connected to Github repository: ")
		ch.Printf("%s\n", remote.URL)
		if subPath != "." {
			ch.PrintfBold("Project location within repo:")
			ch.Printf(" %s\n", subPath)
		}
		confirmed, err := cmdutil.ConfirmPrompt("Enable automatic deploys to Rill Cloud from GitHub?", "", true)
		if err != nil {
			return false, err
		}
		if confirmed {
			opts.SubPath = subPath
			opts.GitPath = repoRoot
		}
	}
	return true, nil
}
