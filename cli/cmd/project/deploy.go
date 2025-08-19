package project

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/cmd/org"
	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/dotrillcloud"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/local"
	"github.com/rilldata/rill/cli/pkg/printer"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	nonSlugRegex      = regexp.MustCompile(`[^\w-]`)
	ErrInvalidProject = errors.New("invalid project")
)

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

	ArchiveUpload bool
	// Managed indicates if the project should be deployed using Rill Managed Git.
	Managed bool
	// Github indicates if the project should be connected to GitHub for automatic deploys.
	Github bool
}

func (o *DeployOpts) ValidatePathAndSetupGit(ch *cmdutil.Helper) error {
	if o.SubPath != "" && (o.ArchiveUpload || o.Managed) {
		return fmt.Errorf("`subpath` flag cannot be used with `archive` or `managed` deploys")
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

	if o.Managed || o.ArchiveUpload {
		return nil
	}
	if o.SubPath != "" {
		// subpath is already set
		o.Github = true
		return nil
	}

	// detect subpath
	repoRoot, err := gitutil.InferGitRepoRoot(o.GitPath)
	if err != nil {
		// Not a git repository, no need to connect to GitHub
		return nil
	}

	remote, err := gitutil.ExtractGitRemote(repoRoot, o.RemoteName, false)
	if err != nil {
		return err
	}
	if remote.URL == "" {
		// no remote configured
		return nil
	}
	if !strings.HasPrefix(remote.URL, "https://github.com") {
		// not a GitHub repo should not prompt for GitHub connection
		return nil
	}

	subPath, err := filepath.Rel(repoRoot, o.GitPath)
	if err == nil {
		ch.PrintfBold("Detected git repository at: ")
		ch.Printf("%s\n", repoRoot)
		ch.PrintfBold("Connected to Github repository: ")
		ch.Printf("%s\n", remote.URL)
		if subPath != "." {
			ch.PrintfBold("Project location within repo: ")
			ch.Printf("%s\n", subPath)
		}
		confirmed, err := cmdutil.ConfirmPrompt("Enable automatic deploys to Rill Cloud from GitHub?", "", true)
		if err != nil {
			return err
		}
		if confirmed {
			o.SubPath = subPath
			o.GitPath = repoRoot
			o.Github = true
			return nil
		}
		if !o.Github {
			ch.Printf("Skipping GitHub connection. You can connect later using `rill project connect-github`.\n")
			o.Managed = true
		}
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
			err := opts.ValidatePathAndSetupGit(ch)
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
	if !ch.IsDev() {
		if err := deployCmd.Flags().MarkHidden("prod-slots"); err != nil {
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
	_, localProjectPath, err := ValidateLocalProject(ch, opts.GitPath, opts.SubPath)
	if err != nil {
		return err
	}
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
	if ch.Org == "" {
		user, err := adminClient.GetCurrentUser(ctx, &adminv1.GetCurrentUserRequest{})
		if err != nil {
			return err
		}
		// email can have other characters like . and + what to do ?
		username, _, _ := strings.Cut(user.User.Email, "@")
		username = nonSlugRegex.ReplaceAllString(username, "-")
		err = createOrgFlow(ctx, ch, username)
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

	projResp, err := adminClient.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: ch.Org, Name: opts.Name})
	if err != nil {
		if st, ok := status.FromError(err); !ok || st.Code() != codes.NotFound {
			return err
		}
	}

	// check if the project already exists
	if projResp != nil {
		err = redeployUploadedProject(ctx, projResp, ch, adminClient, localProjectPath, opts, repo)
		if err != nil {
			if s, ok := status.FromError(err); ok && s.Code() == codes.PermissionDenied {
				ch.PrintfError("You do not have the permissions needed to update a project in org %q. Please reach out to your Rill admin.\n", ch.Org)
				return nil
			}
			return fmt.Errorf("update project failed with error %w", err)
		}
		return nil
	}

	req := &adminv1.CreateProjectRequest{
		OrganizationName: ch.Org,
		Name:             opts.Name,
		Description:      opts.Description,
		Provisioner:      opts.Provisioner,
		ProdVersion:      opts.ProdVersion,
		ProdSlots:        int64(opts.Slots),
		Public:           opts.Public,
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

	err = dotrillcloud.SetAll(localProjectPath, ch.AdminURL(), &dotrillcloud.Config{
		ProjectID: res.Project.Id,
	})
	if err != nil {
		return err
	}
	if req.GitRemote != "" {
		// also commit dotrillcloud to the repo
		err = ch.GitHelper(ch.Org, opts.Name, localProjectPath).PushToManagedRepo(ctx)
		if err != nil {
			return err
		}
	}

	// Success!
	ch.PrintfSuccess("Created project \"%s/%s\". Use `rill project rename` to change name if required.\n\n", ch.Org, res.Project.Name)

	// Upload .env
	vars, err := local.ParseDotenv(ctx, localProjectPath)
	if err != nil {
		ch.PrintfWarn("Failed to parse .env: %v\n", err)
	} else if len(vars) > 0 {
		_, err = adminClient.UpdateProjectVariables(ctx, &adminv1.UpdateProjectVariablesRequest{
			Organization: ch.Org,
			Project:      opts.Name,
			Variables:    vars,
		})
		if err != nil {
			ch.PrintfWarn("Failed to upload .env: %v\n", err)
		}
	}

	// Open browser
	if res.Project.FrontendUrl != "" {
		ch.PrintfSuccess("Your project can be accessed at: %s\n", res.Project.FrontendUrl)
		if ch.Interactive {
			ch.PrintfSuccess("Opening project in browser...\n")
			time.Sleep(3 * time.Second)
			_ = browser.Open(res.Project.FrontendUrl)
		}
	}
	ch.Telemetry(ctx).RecordBehavioralLegacy(activity.BehavioralEventDeploySuccess)
	return nil
}

func redeployUploadedProject(ctx context.Context, projResp *adminv1.GetProjectResponse, ch *cmdutil.Helper, adminClient *client.Client, localProjectPath string, opts *DeployOpts, repo drivers.RepoStore) error {
	if projResp.Project.GitRemote != "" && projResp.Project.ManagedGitId == "" {
		// connected to user managed github
		ch.PrintfError("Found existing project. But it is already connected to a Github repository.\nPush changes to %q to deploy.\n", projResp.Project.GitRemote)
		return nil
	}
	ch.Printer.Println("Found existing project. Starting re-upload.")
	var updateProjReq *adminv1.UpdateProjectRequest
	if projResp.Project.GitRemote != "" {
		// rill managed git
		err := ch.GitHelper(ch.Org, opts.Name, localProjectPath).PushToManagedRepo(ctx)
		if err != nil {
			return err
		}
	} else {
		// test tarball flow
		if opts.ArchiveUpload {
			assetID, err := cmdutil.UploadRepo(ctx, repo, ch, ch.Org, opts.Name)
			if err != nil {
				return err
			}
			updateProjReq = &adminv1.UpdateProjectRequest{
				OrganizationName: ch.Org,
				Name:             projResp.Project.Name,
				ArchiveAssetId:   &assetID,
			}
		} else {
			// need to migrate to rill managed git
			gitRepo, err := ch.GitHelper(ch.Org, opts.Name, localProjectPath).PushToNewManagedRepo(ctx)
			if err != nil {
				return err
			}
			updateProjReq = &adminv1.UpdateProjectRequest{
				OrganizationName: ch.Org,
				Name:             opts.Name,
				GitRemote:        &gitRepo.Remote,
			}
		}
	}

	if updateProjReq != nil {
		// Update the project
		// Silently ignores other flags like description etc which are handled with project update.
		_, err := adminClient.UpdateProject(ctx, updateProjReq)
		if err != nil {
			if s, ok := status.FromError(err); ok && s.Code() == codes.PermissionDenied {
				ch.PrintfError("You do not have the permissions needed to update a project in org %q. Please reach out to your Rill admin.\n", ch.Org)
				return nil
			}
			return fmt.Errorf("update project failed with error %w", err)
		}
	}

	printer.ColorGreenBold.Printf("All files uploaded successfully.\n\n")
	ch.Telemetry(ctx).RecordBehavioralLegacy(activity.BehavioralEventDeploySuccess)

	// Fetch vars from .env
	vars, err := local.ParseDotenv(ctx, localProjectPath)
	if err != nil {
		ch.PrintfWarn("Failed to parse .env: %v\n", err)
	} else if len(vars) > 0 {
		_, err = adminClient.UpdateProjectVariables(ctx, &adminv1.UpdateProjectVariablesRequest{
			Organization: ch.Org,
			Project:      projResp.Project.Name,
			Variables:    vars,
		})
		if err != nil {
			ch.PrintfWarn("Failed to upload .env: %v\n", err)
		}
	}

	// Success
	ch.PrintfSuccess("Updated project \"%s/%s\".\n\n", ch.Org, projResp.Project.Name)
	return nil
}

func createOrgFlow(ctx context.Context, ch *cmdutil.Helper, defaultName string) error {
	c, err := ch.Client()
	if err != nil {
		return err
	}

	res, err := c.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{
		Name: defaultName,
	})
	if err != nil {
		if !errMsgContains(err, "an org with that name already exists") {
			return err
		}

		ch.PrintfWarn("Rill organizations are derived from the owner of your Github repository.\n")
		ch.PrintfWarn("The %q organization associated with your Github repository already exists.\n", defaultName)
		ch.PrintfWarn("Contact your Rill admin to be added to your org or create a new organization below.\n")

		name, err := orgNamePrompt(ctx, ch)
		if err != nil {
			return err
		}

		res, err = c.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{
			Name: name,
		})
		if err != nil {
			return err
		}
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

	_, err = c.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: name})
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

func projectExists(ctx context.Context, ch *cmdutil.Helper, orgName, projectName string) (bool, error) {
	c, err := ch.Client()
	if err != nil {
		return false, err
	}

	_, err = c.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: orgName, Name: projectName})
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

func errMsgContains(err error, msg string) bool {
	if st, ok := status.FromError(err); ok && st != nil {
		return strings.Contains(st.Message(), msg)
	}
	return false
}
