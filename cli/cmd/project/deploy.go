package project

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/rilldata/rill/cli/cmd/org"
	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/local"
	"github.com/rilldata/rill/cli/pkg/printer"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrInvalidProject = errors.New("invalid project")

type DeployOpts struct {
	GitPath     string
	SubPath     string
	RemoteName  string
	Name        string
	Description string
	Public      bool
	Provisioner string
	ProdVersion string
	ProdBranch  string
	Slots       int
	PushEnv     bool

	ArchiveUpload bool
	// Managed indicates if the project should be deployed using Rill Managed Git.
	Managed bool
	// Github indicates if the project should be connected to GitHub for automatic deploys.
	Github bool

	// SkipDeploy skips the runtime deployment step. Used for testing.
	SkipDeploy bool

	// remoteURL is the git remote url of the repository if detected. Set internally.
	remoteURL string
	// pushToProject is set if the deploy should push current changes to this existing project. Set internally.
	pushToProject *adminv1.Project
}

func (o *DeployOpts) LocalProjectPath() string {
	if o.SubPath != "" {
		return filepath.Join(o.GitPath, o.SubPath)
	}
	return o.GitPath
}

func (o *DeployOpts) ValidateAndApplyDefaults(ctx context.Context, ch *cmdutil.Helper) error {
	if o.remoteURL != "" {
		// already validated
		// just a hack to avoid re-validation when `rill project deploy` internally calls `rill project connect-github`
		return nil
	}
	// expand project directory and get absolute path
	var err error
	o.GitPath, err = fileutil.ExpandHome(o.GitPath)
	if err != nil {
		return err
	}
	o.GitPath, err = filepath.Abs(o.GitPath)
	if err != nil {
		return err
	}

	_, _, err = ValidateLocalProject(ch, o.GitPath, o.SubPath)
	if err != nil {
		return err
	}

	// check if specified project already exists
	if o.Name != "" && ch.Org != "" {
		p, err := getProject(ctx, ch, ch.Org, o.Name)
		if err != nil && !errors.Is(err, cmdutil.ErrNoMatchingProject) {
			return err
		}
		if p != nil {
			if ch.Interactive {
				ok, err := cmdutil.ConfirmPrompt(fmt.Sprintf("Project with name %q already exists. Do you want to push current changes to the existing project?", o.Name), "", true)
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("aborting deploy")
				}
			}
			o.pushToProject = p
			o.Managed = o.pushToProject.ManagedGitId != ""
			o.Github = o.pushToProject.ManagedGitId == "" && o.pushToProject.GitRemote != ""
			o.ArchiveUpload = o.pushToProject.ArchiveAssetId != ""
			return nil
		}
	}
	if o.ArchiveUpload {
		return nil
	}

	// detect repo root and subpath
	var repoRoot, subpath string
	if o.SubPath != "" {
		repoRoot = o.GitPath
		subpath = o.SubPath
	} else {
		// detect subpath
		repoRoot, subpath, err = gitutil.InferRepoRootAndSubpath(o.GitPath)
		if err != nil {
			// Not a git repository
			return nil
		}
	}

	// extract remote and check if project already exists
	err = o.detectGitRemoteAndProject(ctx, ch, repoRoot, subpath)
	if err != nil {
		return err
	}

	// if there is a project already connected to this repo+subpath offer to push changes to it
	if o.pushToProject != nil {
		if o.pushToProject.ManagedGitId == "" && o.Managed {
			ch.PrintfError("Project %s/%s is already connected to this GitHub repository. Cannot use --managed flag.\n", o.pushToProject.OrgName, o.pushToProject.Name)
			return fmt.Errorf("aborting deploy")
		}
		if o.pushToProject.ManagedGitId != "" && o.Github {
			ch.Printf("Found another rill managed project %s/%s connected to this folder\n", o.pushToProject.OrgName, o.pushToProject.Name)
			ch.PrintfBold("Run `rill project edit --remote-url <github_remote>` to tranfer the project to GitHub.\n")
			return fmt.Errorf("aborting deploy")
		}
		if o.pushToProject.OrgName != ch.Org {
			ch.PrintfError("A project in another org deploys from this repository. Please switch to org %q to push changes to the project %q.\n", o.pushToProject.OrgName, o.pushToProject.Name)
			return fmt.Errorf("aborting deploy")
		}
		if subpath != "" && o.pushToProject.Subpath != subpath {
			// just for verification confirm that subpath matches the one stored in project
			return fmt.Errorf("current project subpath %q does not match the one stored in rill %q. Try doing deploy using rill cli from github repo root by passing explicit subpath using `rill deploy --subpath %s`", subpath, o.pushToProject.Subpath, o.pushToProject.Subpath)
		}
		// set flags based on existing project
		o.Managed = o.pushToProject.ManagedGitId != ""
		o.Github = o.pushToProject.ManagedGitId == "" && o.pushToProject.GitRemote != ""
		o.ArchiveUpload = o.pushToProject.ArchiveAssetId != ""

		ch.PrintfBold("\nFound existing project: ")
		ch.Printf("%s/%s\n", o.pushToProject.OrgName, o.pushToProject.Name)
		if !ch.Interactive {
			return nil
		}
		ok, err := cmdutil.ConfirmPrompt("Do you want to push current changes to the existing project?", "", true)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("aborting deploy")
		}
		return nil
	}

	if o.remoteURL == "" {
		// no remote configured
		return nil
	}

	// there is a self hosted git repo but no project connected to it
	connectToGithub := true
	ch.PrintfBold("Detected git repository at: ")
	ch.Printf("%s\n", repoRoot)
	ch.PrintfBold("Connected to Github repository: ")
	ch.Printf("%s\n", o.remoteURL)
	if subpath != "" {
		ch.PrintfBold("Project location within repo: ")
		ch.Printf("%s\n", subpath)
	}
	if o.Managed {
		// if user explicitly wants managed deploys confirm if they want to really skip github connection
		ok, err := cmdutil.ConfirmPrompt("Do you want to skip connecting to GitHub and use Rill managed deploys? (Note: Subsequent deploys/push from Rill will not push changes to your GitHub repo)", "", true)
		if err != nil {
			return err
		}
		connectToGithub = !ok
	} else if !o.Github && ch.Interactive {
		// still confirm if user wants to connect to github
		connectToGithub, err = cmdutil.ConfirmPrompt("Enable automatic deploys to Rill Cloud from GitHub?", "", true)
		if err != nil {
			return err
		}
	}
	if connectToGithub {
		o.SubPath = subpath
		o.GitPath = repoRoot
		o.Github = true
		return nil
	}
	o.Managed = true
	return nil
}

func (o *DeployOpts) detectGitRemoteAndProject(ctx context.Context, ch *cmdutil.Helper, repoRoot, subpath string) error {
	remotes, err := gitutil.ExtractRemotes(repoRoot, false)
	if err != nil && !errors.Is(err, gitutil.ErrGitRemoteNotFound) {
		return err
	}
	c, err := ch.Client()
	if err != nil {
		return err
	}

	// find matching projects
	req := &adminv1.ListProjectsForFingerprintRequest{
		DirectoryName: filepath.Base(o.LocalProjectPath()),
		SubPath:       subpath,
	}
	for _, remote := range remotes {
		switch remote.Name {
		case "__rill_remote":
			req.RillMgdGitRemote = remote.URL
		case o.RemoteName:
			gitremote, err := remote.Github()
			if err == nil {
				req.GitRemote = gitremote
			}
		}
	}
	resp, err := c.ListProjectsForFingerprint(ctx, req)
	if err != nil {
		// TODO: check for not found error
		return err
	}
	if resp.UnauthorizedProject != "" {
		ch.PrintfWarn("You do not have access to the project %q which is connected to this repository. Please reach out to your Rill admin\n", resp.UnauthorizedProject)
		return fmt.Errorf("aborting deploy")
	}
	for _, p := range resp.Projects {
		if p.ManagedGitId != "" {
			o.pushToProject = p
			o.remoteURL = p.GitRemote
			return nil
		}
		o.pushToProject = p
		o.remoteURL = p.GitRemote
		// do not return yet, there might be a managed project
		// this is not possible with new flow but keeping it for consistency
	}

	if len(resp.Projects) == 1 && resp.Projects[0].ManagedGitId == "" && req.RillMgdGitRemote != "" {
		err = ch.HandleRepoTransfer(repoRoot, req.GitRemote)
		if err != nil {
			return err
		}
	}
	if req.GitRemote != "" {
		o.remoteURL = req.GitRemote
	}
	return nil
}

func DeployCmd(ch *cmdutil.Helper) *cobra.Command {
	opts := &DeployOpts{}

	deployCmd := &cobra.Command{
		Use:   "deploy [<path>]",
		Short: "Deploy project to Rill Cloud by using a Rill Managed Git repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.GitPath = args[0]
			}
			opts.Managed = true
			err := opts.ValidateAndApplyDefaults(cmd.Context(), ch)
			if err != nil {
				return err
			}
			return DeployWithUploadFlow(cmd.Context(), ch, opts)
		},
	}

	deployCmd.Flags().SortFlags = false
	deployCmd.Flags().StringVar(&opts.GitPath, "path", ".", "Path to project repository (default: current directory)") // This can also be a remote .git URL (undocumented feature)
	deployCmd.Flags().StringVar(&opts.SubPath, "subpath", "", "Relative path to project in the repository (for monorepos)")
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

	deployCmd.Flags().BoolVar(&opts.SkipDeploy, "skip-deploy", false, "Skip the runtime deployment step (for testing only)")
	if !ch.IsDev() {
		err := deployCmd.Flags().MarkHidden("skip-deploy")
		if err != nil {
			panic(err)
		}
	}

	return deployCmd
}

func ValidateLocalProject(ch *cmdutil.Helper, localGitPath, subPath string) (string, string, error) {
	var localProjectPath string
	if subPath == "" {
		localProjectPath = localGitPath
	} else {
		localProjectPath = filepath.Join(localGitPath, subPath)
	}

	// Verify that localProjectPath contains a Rill project.
	if cmdutil.HasRillProject(localProjectPath) {
		return localGitPath, localProjectPath, nil
	}

	ch.PrintfWarn("Directory %q doesn't contain a valid Rill project.\n", localProjectPath)
	ch.PrintfWarn("Run `rill project deploy` from a Rill project directory or use `--path` to pass a project path.\n")
	ch.PrintfWarn("Run `rill start` to initialize a new Rill project.\n")
	return "", "", ErrInvalidProject
}

func DeployWithUploadFlow(ctx context.Context, ch *cmdutil.Helper, opts *DeployOpts) error {
	localProjectPath := opts.LocalProjectPath()
	// If no project name was provided, default to dir name
	if opts.Name == "" {
		opts.Name = filepath.Base(localProjectPath)
	}

	// Set a default org for the user if necessary
	// (If user is not in an org, we'll create one based on their user name later in the flow.)
	adminClient, err := ch.Client()
	if err != nil {
		return err
	}
	if ch.Org == "" {
		if err := org.SetDefaultOrg(ctx, ch); err != nil {
			return err
		}
	}

	// If no default org is set, it means the user is not in an org yet.
	// We create a default org based on the user name.
	// TODO : Ask user prompt similar to UI instead of silently creating org based on email
	if ch.Org == "" {
		err = createOrgFlow(ctx, ch)
		if err != nil {
			return fmt.Errorf("org creation failed with error: %w", err)
		}
		ch.PrintfSuccess("Created org %q. Run `rill org edit` to change name if required.\n\n", ch.Org)
	} else {
		ch.PrintfBold("Using org %q.\n\n", ch.Org)
	}

	// get repo for current project
	repo, _, err := cmdutil.RepoForProjectPath(localProjectPath)
	if err != nil {
		return err
	}
	// Ensure gitignore is set up so that we don't upload files that should not be tracked by Git.
	err = cmdutil.SetupGitIgnore(ctx, repo)
	if err != nil {
		return fmt.Errorf("failed to set up .gitignore: %w", err)
	}

	if opts.pushToProject != nil {
		return redeployProject(ctx, ch, opts)
	}

	// Create a new project
	req := &adminv1.CreateProjectRequest{
		Org:           ch.Org,
		Project:       opts.Name,
		Description:   opts.Description,
		Provisioner:   opts.Provisioner,
		ProdVersion:   opts.ProdVersion,
		ProdSlots:     int64(opts.Slots),
		Public:        opts.Public,
		DirectoryName: filepath.Base(localProjectPath),
		SkipDeploy:    opts.SkipDeploy,
	}

	ch.Printer.Println("Starting upload.")
	if opts.ArchiveUpload {
		// create a tar archive of the project and upload it
		assetID, err := cmdutil.UploadRepo(ctx, repo, ch, ch.Org, opts.Name)
		if err != nil {
			return err
		}
		req.ArchiveAssetId = assetID
	} else {
		gitRepo, err := ch.GitHelper(ch.Org, opts.Name, localProjectPath).PushToNewManagedRepo(ctx)
		if err != nil {
			return err
		}
		req.GitRemote = gitRepo.Remote
	}
	printer.ColorGreenBold.Printf("All files uploaded successfully.\n\n")

	// Create the project
	res, err := adminClient.CreateProject(ctx, req)
	if err != nil {
		if s, ok := status.FromError(err); ok && s.Code() == codes.PermissionDenied {
			ch.PrintfError("You do not have the permissions needed to create a project in org %q. Please reach out to your Rill admin.\n", ch.Org)
			return nil
		}
		return fmt.Errorf("create project failed with error %w", err)
	}

	// Success!
	ch.PrintfSuccess("Created project \"%s/%s\". Use `rill project rename` to change name if required.\n\n", ch.Org, res.Project.Name)

	// Upload .env
	if opts.PushEnv {
		vars, err := local.ParseDotenv(ctx, localProjectPath)
		if err != nil {
			ch.PrintfWarn("Failed to parse .env: %v\n", err)
		} else if len(vars) > 0 {
			_, err = adminClient.UpdateProjectVariables(ctx, &adminv1.UpdateProjectVariablesRequest{
				Org:       ch.Org,
				Project:   opts.Name,
				Variables: vars,
			})
			if err != nil {
				ch.PrintfWarn("Failed to upload .env: %v\n", err)
			}
		}
	}

	// Open browser
	if res.Project.FrontendUrl != "" {
		ch.PrintfSuccess("Your project can be accessed at: %s\n", res.Project.FrontendUrl)
		if ch.Interactive {
			ch.PrintfSuccess("Opening project in browser...\n")
			select {
			case <-time.After(3 * time.Second):
				_ = browser.Open(res.Project.FrontendUrl)
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	ch.Telemetry(ctx).RecordBehavioralLegacy(activity.BehavioralEventDeploySuccess)
	return nil
}

func redeployProject(ctx context.Context, ch *cmdutil.Helper, opts *DeployOpts) error {
	c, err := ch.Client()
	if err != nil {
		return err
	}
	proj := opts.pushToProject
	if proj.ManagedGitId != "" {
		err := ch.GitHelper(ch.Org, proj.Name, opts.LocalProjectPath()).PushToManagedRepo(ctx)
		if err != nil {
			return err
		}
	} else if proj.GitRemote != "" {
		// Infer repo root and subpath for git operations
		repoRoot, subpath, err := gitutil.InferRepoRootAndSubpath(opts.LocalProjectPath())
		if err != nil {
			return err
		}
		// Verify subpath matches the one stored in the project
		if subpath != proj.Subpath {
			return fmt.Errorf("current project subpath %q does not match the one stored in rill %q. Run rill cli from github repo root and pass explicit subpath using `rill deploy --subpath %s`", subpath, proj.Subpath, proj.Subpath)
		}
		config := &gitutil.Config{
			Remote:        opts.pushToProject.GitRemote,
			DefaultBranch: opts.pushToProject.ProdBranch,
			Subpath:       subpath,
		}
		err = ch.CommitAndSafePush(ctx, repoRoot, config, "", nil, "1")
		if err != nil {
			return err
		}
	} else {
		// tarball flow
		var updateProjReq *adminv1.UpdateProjectRequest
		if opts.ArchiveUpload {
			repo, _, err := cmdutil.RepoForProjectPath(opts.LocalProjectPath())
			if err != nil {
				return err
			}
			assetID, err := cmdutil.UploadRepo(ctx, repo, ch, ch.Org, opts.Name)
			if err != nil {
				return err
			}
			updateProjReq = &adminv1.UpdateProjectRequest{
				Org:            ch.Org,
				Project:        proj.Name,
				ArchiveAssetId: &assetID,
			}
		} else {
			// need to migrate to rill managed git
			gitRepo, err := ch.GitHelper(ch.Org, opts.Name, opts.LocalProjectPath()).PushToNewManagedRepo(ctx)
			if err != nil {
				return err
			}
			updateProjReq = &adminv1.UpdateProjectRequest{
				Org:       ch.Org,
				Project:   opts.Name,
				GitRemote: &gitRepo.Remote,
			}
		}
		// Update the project
		// Silently ignores other flags like description etc which are handled with project update.
		_, err = c.UpdateProject(ctx, updateProjReq)
		if err != nil {
			if s, ok := status.FromError(err); ok && s.Code() == codes.PermissionDenied {
				ch.PrintfError("You do not have the permissions needed to update a project in org %q. Please reach out to your Rill admin.\n", ch.Org)
				return nil
			}
			return fmt.Errorf("update project failed with error %w", err)
		}
	}

	// Upload .env
	if opts.PushEnv {
		vars, err := local.ParseDotenv(ctx, opts.LocalProjectPath())
		if err != nil {
			ch.PrintfWarn("Failed to parse .env: %v\n", err)
		} else if len(vars) > 0 {
			_, err = c.UpdateProjectVariables(ctx, &adminv1.UpdateProjectVariablesRequest{
				Org:       ch.Org,
				Project:   proj.Name,
				Variables: vars,
			})
			if err != nil {
				ch.PrintfWarn("Failed to upload .env: %v\n", err)
			}
		}
	}

	// Success
	ch.PrintfSuccess("Updated project \"%s/%s\".\n\n", ch.Org, proj.Name)
	return nil
}

func createOrgFlow(ctx context.Context, ch *cmdutil.Helper) error {
	c, err := ch.Client()
	if err != nil {
		return err
	}

	ch.PrintfBold("No organization found for your account. Creating a new organization.\n")
	name, err := orgNamePrompt(ctx, ch)
	if err != nil {
		return err
	}

	res, err := c.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{
		Name: name,
	})
	if err != nil {
		return err
	}

	// Switching to the created org
	ch.Org = res.Organization.Name
	err = ch.DotRill.SetDefaultOrg(ch.Org)
	if err != nil {
		return err
	}

	return nil
}

func orgNamePrompt(ctx context.Context, ch *cmdutil.Helper) (string, error) {
	qs := []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "Enter an org name",
			},
			Validate: func(v any) error {
				// Validate org name doesn't exist already
				name := v.(string)
				if name == "" {
					return fmt.Errorf("empty name")
				}

				exists, err := orgExists(ctx, ch, name)
				if err != nil {
					return fmt.Errorf("org name %q is already taken", name)
				}

				if exists {
					// this should always be true but adding this check from completeness POV
					return fmt.Errorf("org with name %q already exists", name)
				}
				return nil
			},
		},
	}

	name := ""
	if err := survey.Ask(qs, &name); err != nil {
		return "", err
	}

	return name, nil
}

func orgExists(ctx context.Context, ch *cmdutil.Helper, name string) (bool, error) {
	c, err := ch.Client()
	if err != nil {
		return false, err
	}

	_, err = c.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Org: name})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func projectExists(ctx context.Context, ch *cmdutil.Helper, org, project string) (bool, error) {
	_, err := getProject(ctx, ch, org, project)
	if err != nil {
		if errors.Is(err, cmdutil.ErrNoMatchingProject) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func getProject(ctx context.Context, ch *cmdutil.Helper, org, project string) (*adminv1.Project, error) {
	c, err := ch.Client()
	if err != nil {
		return nil, err
	}

	p, err := c.GetProject(ctx, &adminv1.GetProjectRequest{Org: org, Project: project})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				return nil, cmdutil.ErrNoMatchingProject
			}
		}
		return nil, err
	}
	return p.Project, nil
}

func errMsgContains(err error, msg string) bool {
	if st, ok := status.FromError(err); ok && st != nil {
		return strings.Contains(st.Message(), msg)
	}
	return false
}
