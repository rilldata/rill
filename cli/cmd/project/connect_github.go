package project

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/v71/github"
	"github.com/rilldata/rill/cli/cmd/org"
	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/dotrillcloud"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/local"
	"github.com/rilldata/rill/cli/pkg/printer"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	pollTimeout  = 10 * time.Minute
	pollInterval = 5 * time.Second
)

func GitPushCmd(ch *cmdutil.Helper) *cobra.Command {
	opts := &DeployOpts{}

	deployCmd := &cobra.Command{
		Use:   "connect-github [<path>]",
		Short: "Deploy project to Rill Cloud by pulling project files from a git repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.GitPath = args[0]
			}
			return ConnectGithubFlow(cmd.Context(), ch, opts)
		},
	}

	deployCmd.Flags().SortFlags = false
	deployCmd.Flags().StringVar(&opts.GitPath, "path", ".", "Path to project repository (default: current directory)") // This can also be a remote .git URL (undocumented feature)
	deployCmd.Flags().StringVar(&opts.SubPath, "subpath", "", "Relative path to project in the repository (for monorepos)")
	deployCmd.Flags().StringVar(&opts.RemoteName, "remote", "origin", "Remote name (default: origin)")
	deployCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Org to deploy project in")
	deployCmd.Flags().StringVar(&opts.Name, "name", "", "Project name (default: Git repo name)")
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

func ConnectGithubFlow(ctx context.Context, ch *cmdutil.Helper, opts *DeployOpts) error {
	// Set a default org for the user if necessary
	// (If user is not in an org, we'll create one based on their Github account later in the flow.)
	if ch.Org == "" {
		if err := org.SetDefaultOrg(ctx, ch); err != nil {
			return err
		}
	}

	// The gitPath can be either a local path or a remote .git URL.
	// Determine which it is.
	var isLocalGitPath bool
	var gitRemote string
	if opts.GitPath != "" {
		u, err := url.Parse(opts.GitPath)
		if err != nil || u.Scheme == "" {
			isLocalGitPath = true
		} else {
			gitRemote, err = gitutil.NormalizeGithubRemote(opts.GitPath)
			if err != nil {
				return fmt.Errorf("failed to parse path as a Github remote: %w", err)
			}
		}
	}

	var localGitPath string
	var localProjectPath string
	var err error
	if isLocalGitPath {
		// If it's a local path, we need to do some extra validation and rewrites.
		localGitPath, localProjectPath, err = ValidateLocalProject(ch, opts.GitPath, opts.SubPath)
		if err != nil {
			if errors.Is(err, ErrInvalidProject) {
				return nil
			}
			return err
		}
	}

	if ch.Org != "" {
		adminClient, err := ch.Client()
		if err != nil {
			return err
		}

		var proj *adminv1.Project

		if opts.Name == "" {
			// Try loading the project from the .rillcloud directory
			proj, err = ch.LoadProject(ctx, localProjectPath)
			if err != nil {
				return err
			}
		} else {
			projResp, err := adminClient.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: ch.Org, Name: opts.Name})
			if err != nil {
				if st, ok := status.FromError(err); !ok || st.Code() != codes.NotFound {
					return err
				}
			}
			if projResp != nil {
				proj = projResp.Project
			}
		}

		if proj != nil && proj.GitRemote != "" {
			ch.PrintfError("Found existing project. But it is already connected to a Github repository.\nPlease visit %s to update the Github repository.\n", proj.FrontendUrl)
			return nil
		}
	}

	if isLocalGitPath {
		// Extract and infer the gitRemote.
		remote, err := gitutil.ExtractGitRemote(localGitPath, opts.RemoteName, false)
		if err != nil {
			if !errors.Is(err, gitutil.ErrGitRemoteNotFound) && !errors.Is(err, git.ErrRepositoryNotExists) {
				return err
			}

			// first check if user wants to create a github repo
			ch.Print("No git remote was found.\n")
			ok, confirmErr := cmdutil.ConfirmPrompt("Do you want to create a Github repository?", "", true)
			if confirmErr != nil {
				return confirmErr
			}
			if !ok {
				return nil
			}

			if err := createGithubRepoFlow(ctx, ch, localGitPath); err != nil {
				return err
			}

			// In the rest of the flow we still check for the github access.
			// It just adds some delay and no user action should be required and handles any improbable edge case where we don't have access to newly created repository.
			// Also keeps the code clean.
			remote, err = gitutil.ExtractGitRemote(localGitPath, opts.RemoteName, false)
			if err != nil {
				return err
			}
		}

		// Error if the repository is not in sync with the remote
		ok, err := repoInSyncFlow(ch, localGitPath, opts.ProdBranch, remote.Name)
		if err != nil {
			return err
		}
		if !ok {
			ch.PrintfBold("You can run `rill project connect-github` again when you have pushed your local changes to the remote.\n")
			return nil
		}

		// Set the gitRemote to the normalized Github URL.
		gitRemote, err = remote.Github()
		if err != nil {
			return err
		}
	}

	// We now have a gitRemote.

	// Extract Github account and repo name from the gitRemote
	ghAccount, ghRepo, ok := gitutil.SplitGithubRemote(gitRemote)
	if !ok {
		return fmt.Errorf("remote %q is not a valid github.com remote", gitRemote)
	}

	// Run flow for access to the Github remote (if necessary)
	ghRes, err := githubFlow(ctx, ch, gitRemote)
	if err != nil {
		return fmt.Errorf("failed Github flow: %w", err)
	}

	if opts.ProdBranch == "" {
		opts.ProdBranch = ghRes.DefaultBranch
	}

	// If no project name was provided, default to Git repo name
	if opts.Name == "" {
		opts.Name = ghRepo
	}

	// If no default org is set by now, it means the user is not in an org yet.
	// We create a default org based on their Github account name.
	if ch.Org == "" {
		err := createOrgFlow(ctx, ch, ghAccount)
		if err != nil {
			return fmt.Errorf("org creation failed with error: %w", err)
		}
		ch.PrintfSuccess("Created org %q. Run `rill org edit` to change name if required.\n\n", ch.Org)
	} else {
		ch.PrintfBold("Using org %q.\n\n", ch.Org)
	}

	// Check if a project matching gitRemote already exists in this org
	projects, err := ch.ProjectNamesByGitRemote(ctx, ch.Org, gitRemote, opts.SubPath)
	if err == nil && len(projects) != 0 { // ignoring error since this is just for a confirmation prompt
		for _, p := range projects {
			if strings.EqualFold(opts.Name, p) {
				ch.PrintfWarn("Can't deploy project %q.\n", opts.Name)
				ch.PrintfWarn("It is connected to Github and continuously deploys when you commit to %q\n", gitRemote)
				ch.PrintfWarn("If you want to deploy to a new project, use `rill project connect-github --name new-name`\n")
				return nil
			}
		}
	}

	// Ensure gitignore is set up so that we don't upload files that should not be tracked by Git.
	repo, _, err := cmdutil.RepoForProjectPath(localProjectPath)
	if err != nil {
		return err
	}
	err = cmdutil.SetupGitIgnore(ctx, repo)
	if err != nil {
		return fmt.Errorf("failed to set up .gitignore: %w", err)
	}

	// Create the project (automatically deploys prod branch)
	res, err := createProjectFlow(ctx, ch, &adminv1.CreateProjectRequest{
		OrganizationName: ch.Org,
		Name:             opts.Name,
		Description:      opts.Description,
		Provisioner:      opts.Provisioner,
		ProdVersion:      opts.ProdVersion,
		ProdSlots:        int64(opts.Slots),
		Subpath:          opts.SubPath,
		ProdBranch:       opts.ProdBranch,
		Public:           opts.Public,
		GitRemote:        gitRemote,
	})
	if err != nil {
		if s, ok := status.FromError(err); ok && s.Code() == codes.PermissionDenied {
			ch.PrintfError("You do not have the permissions needed to create a project in org %q. Please reach out to your Rill admin.\n", ch.Org)
			return nil
		}
		return fmt.Errorf("create project failed with error %w", err)
	}

	if localProjectPath != "" {
		err = dotrillcloud.SetAll(localProjectPath, ch.AdminURL(), &dotrillcloud.Config{
			ProjectID: res.Project.Id,
		})
		if err != nil {
			return err
		}
		author, err := ch.GitSignature(ctx, localGitPath)
		if err != nil {
			return err
		}
		err = gitutil.CommitAndForcePush(ctx, localGitPath, &gitutil.Config{Remote: gitRemote, DefaultBranch: opts.ProdBranch}, "Autocommit .rillcloud directory", author)
		if err != nil {
			return fmt.Errorf("failed to push .rillcloud directory to remote: %w", err)
		}
	}

	// Success!
	ch.PrintfSuccess("Created project \"%s/%s\". Use `rill project rename` to change name if required.\n\n", ch.Org, res.Project.Name)
	ch.PrintfSuccess("Rill projects deploy continuously when you push changes to Github.\n")

	// Upload .env
	if isLocalGitPath {
		vars, err := local.ParseDotenv(ctx, localProjectPath)
		if err != nil {
			ch.PrintfWarn("Failed to parse .env: %v\n", err)
		} else if len(vars) > 0 {
			c, err := ch.Client()
			if err != nil {
				return err
			}
			_, err = c.UpdateProjectVariables(ctx, &adminv1.UpdateProjectVariablesRequest{
				Organization: ch.Org,
				Project:      opts.Name,
				Variables:    vars,
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
			time.Sleep(3 * time.Second)
			_ = browser.Open(res.Project.FrontendUrl)
		}
	}

	ch.Telemetry(ctx).RecordBehavioralLegacy(activity.BehavioralEventDeploySuccess)

	return nil
}

func createGithubRepoFlow(ctx context.Context, ch *cmdutil.Helper, localGitPath string) error {
	// Get the admin client
	c, err := ch.Client()
	if err != nil {
		return err
	}

	res, err := c.GetGithubUserStatus(ctx, &adminv1.GetGithubUserStatusRequest{})
	if err != nil {
		return err
	}

	if !res.HasAccess {
		ch.Telemetry(ctx).RecordBehavioralLegacy(activity.BehavioralEventGithubConnectedStart)

		if res.GrantAccessUrl != "" {
			// Print instructions to grant access
			time.Sleep(3 * time.Second)
			ch.Print("Open this URL in your browser to grant Rill access to Github:\n\n")
			ch.Print("\t" + res.GrantAccessUrl + "\n\n")

			// Open browser if possible
			if ch.Interactive {
				_ = browser.Open(res.GrantAccessUrl)
			}
		}
	}

	// Poll for permission granted
	pollCtx, cancel := context.WithTimeout(ctx, pollTimeout)
	defer cancel()
	var pollRes *adminv1.GetGithubUserStatusResponse
	for {
		select {
		case <-pollCtx.Done():
			return pollCtx.Err()
		case <-time.After(pollInterval):
			// Ready to check again.
		}

		// Poll for access to the Github's user account
		pollRes, err = c.GetGithubUserStatus(ctx, &adminv1.GetGithubUserStatusRequest{})
		if err != nil {
			return err
		}
		if pollRes.HasAccess {
			break
		}
		// Sleep and poll again
	}

	// Emit success telemetry
	ch.Telemetry(ctx).RecordBehavioralLegacy(activity.BehavioralEventGithubConnectedSuccess)

	// get orgs on which rill github app is installed with write permission
	var candidateOrgs []string
	if pollRes.UserInstallationPermission == adminv1.GithubPermission_GITHUB_PERMISSION_WRITE {
		candidateOrgs = append(candidateOrgs, pollRes.Account)
	}
	for o, p := range pollRes.OrganizationInstallationPermissions {
		if p == adminv1.GithubPermission_GITHUB_PERMISSION_WRITE {
			candidateOrgs = append(candidateOrgs, o)
		}
	}

	repoOwner := ""
	if len(candidateOrgs) == 0 {
		ch.PrintfWarn("\nRill does not have permissions to create a repository on your Github account. Visit this URL to grant access: %s\n", pollRes.GrantAccessUrl)
		return nil
	} else if len(candidateOrgs) == 1 {
		repoOwner = candidateOrgs[0]
		ok, err := cmdutil.ConfirmPrompt(fmt.Sprintf("Rill will create a new repository in the Github account %q. Do you want to continue?", repoOwner), "", true)
		if err != nil {
			return err
		}
		if !ok {
			ch.PrintfWarn("\nIf you want to deploy to another Github account, visit this URL to grant access: %s\n", pollRes.GrantAccessUrl)
			return nil
		}
	} else {
		repoOwner, err = cmdutil.SelectPrompt("Select a Github account for the new repository", candidateOrgs, candidateOrgs[0])
		if err != nil {
			ch.PrintfWarn("\nIf you want to deploy to another Github account, visit this URL to grant access: %s\n", pollRes.GrantAccessUrl)
			return err
		}
	}

	// create and verify
	githubRepository, err := createGithubRepository(ctx, ch, pollRes, localGitPath, repoOwner)
	if err != nil {
		return err
	}

	printer.ColorGreenBold.Printf("\nSuccessfully created repository on %q\n\n", *githubRepository.HTMLURL)
	ch.Print("Pushing local project to Github\n\n")
	author, err := ch.GitSignature(ctx, localGitPath)
	if err != nil {
		return fmt.Errorf("failed to generate git commit signature: %w", err)
	}
	var branch string
	if githubRepository.DefaultBranch != nil {
		branch = *githubRepository.DefaultBranch
	} else {
		branch = "main"
	}
	config := &gitutil.Config{
		Remote:        *githubRepository.CloneURL,
		Username:      "x-access-token",
		Password:      pollRes.AccessToken,
		DefaultBranch: branch,
	}
	err = gitutil.CommitAndForcePush(ctx, localGitPath, config, "", author)
	if err != nil {
		return fmt.Errorf("failed to push local project to Github: %w", err)
	}

	ch.Print("Successfully pushed your local project to Github\n\n")
	return nil
}

func createGithubRepository(ctx context.Context, ch *cmdutil.Helper, pollRes *adminv1.GetGithubUserStatusResponse, localGitPath, repoOwner string) (*github.Repository, error) {
	githubClient := github.NewTokenClient(ctx, pollRes.AccessToken)

	defaultBranch := "main"
	if repoOwner == pollRes.Account {
		repoOwner = ""
	}
	repoName := filepath.Base(localGitPath)
	private := true

	var githubRepo *github.Repository
	var err error
	for i := 1; i <= 10; i++ {
		githubRepo, _, err = githubClient.Repositories.Create(ctx, repoOwner, &github.Repository{Name: &repoName, DefaultBranch: &defaultBranch, Private: &private})
		if err == nil {
			break
		}
		if strings.Contains(err.Error(), "authentication") || strings.Contains(err.Error(), "credentials") {
			// The users who installed app before we started including repo:write permissions need to accept permissions
			// and then only we can create repositories.
			return nil, fmt.Errorf("rill app does not have permissions to create github repository. Visit `https://github.com/settings/installations` to accept new permissions or reinstall app and try again")
		}

		if !strings.Contains(err.Error(), "name already exists") {
			return nil, fmt.Errorf("failed to create repository: %w", err)
		}

		ch.Printf("Repository name %q is already taken\n", repoName)
		repoName, err = cmdutil.InputPrompt("Please provide alternate name", "")
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	// the create repo API does not wait for repo creation to be fully processed on server. Need to verify by making a get call in a loop
	if repoOwner == "" {
		repoOwner = pollRes.Account
	}

	ch.Print("\nRequest submitted for creating repository. Checking completion status\n")
	pollCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	for {
		select {
		case <-pollCtx.Done():
			return nil, pollCtx.Err()
		case <-time.After(2 * time.Second):
			// Ready to check again.
		}
		_, _, err := githubClient.Repositories.Get(ctx, repoOwner, repoName)
		if err == nil {
			break
		}
	}

	return githubRepo, nil
}

func githubFlow(ctx context.Context, ch *cmdutil.Helper, gitRemote string) (*adminv1.GetGithubRepoStatusResponse, error) {
	// Get the admin client
	c, err := ch.Client()
	if err != nil {
		return nil, err
	}

	// Check for access to the Github repo
	res, err := c.GetGithubRepoStatus(ctx, &adminv1.GetGithubRepoStatusRequest{
		Remote: gitRemote,
	})
	if err != nil {
		return nil, err
	}

	// If the user has not already granted access, open browser and poll for access
	if !res.HasAccess {
		// Emit start telemetry
		ch.Telemetry(ctx).RecordBehavioralLegacy(activity.BehavioralEventGithubConnectedStart)

		// Print instructions to grant access
		ch.Print("Rill projects deploy continuously when you push changes to Github.\n")
		ch.Print("You need to grant Rill read only access to your repository on Github.\n\n")
		time.Sleep(3 * time.Second)
		ch.Print("Open this URL in your browser to grant Rill access to Github:\n\n")
		ch.Print("\t" + res.GrantAccessUrl + "\n\n")

		// Open browser if possible
		if ch.Interactive {
			_ = browser.Open(res.GrantAccessUrl)
		}

		// Poll for permission granted
		pollCtx, cancel := context.WithTimeout(ctx, pollTimeout)
		defer cancel()
		for {
			select {
			case <-pollCtx.Done():
				return nil, pollCtx.Err()
			case <-time.After(pollInterval):
				// Ready to check again.
			}

			// Poll for access to the Github URL
			pollRes, err := c.GetGithubRepoStatus(ctx, &adminv1.GetGithubRepoStatusRequest{
				Remote: gitRemote,
			})
			if err != nil {
				return nil, err
			}

			if pollRes.HasAccess {
				// Emit success telemetry
				ch.Telemetry(ctx).RecordBehavioralLegacy(activity.BehavioralEventGithubConnectedSuccess)

				_, ghRepo, _ := gitutil.SplitGithubRemote(gitRemote)
				ch.PrintfSuccess("You have connected to the %q repository in Github.\n", ghRepo)
				return pollRes, nil
			}

			// Sleep and poll again
		}
	}

	return res, nil
}

func createProjectFlow(ctx context.Context, ch *cmdutil.Helper, req *adminv1.CreateProjectRequest) (*adminv1.CreateProjectResponse, error) {
	// Get the admin client
	c, err := ch.Client()
	if err != nil {
		return nil, err
	}

	// Create the project (automatically deploys prod branch)
	res, err := c.CreateProject(ctx, req)
	if err != nil {
		if !errMsgContains(err, "a project with that name already exists in the org") {
			return nil, err
		}

		ch.PrintfWarn("Rill project names are derived from your Github repository name.\n")
		ch.PrintfWarn("The %q project already exists under org %q. Please enter a different name.\n", req.Name, req.OrganizationName)

		// project name already exists, prompt for project name and create project with new name again
		name, err := projectNamePrompt(ctx, ch, req.OrganizationName)
		if err != nil {
			return nil, err
		}

		req.Name = name
		return c.CreateProject(ctx, req)
	}
	return res, err
}

func repoInSyncFlow(ch *cmdutil.Helper, gitPath, branch, remoteName string) (bool, error) {
	syncStatus, err := gitutil.GetSyncStatus(gitPath, branch, remoteName)
	if err != nil {
		// ignore errors since check is best effort and can fail in multiple cases
		return true, nil
	}

	switch syncStatus {
	case gitutil.SyncStatusUnspecified:
		return true, nil
	case gitutil.SyncStatusSynced:
		return true, nil
	case gitutil.SyncStatusModified:
		ch.PrintfWarn("Some files have been locally modified. These changes will not be present in the deployed project.\n")
	case gitutil.SyncStatusAhead:
		ch.PrintfWarn("Local commits are not pushed to remote yet. These changes will not be present in the deployed project.\n")
	}

	return cmdutil.ConfirmPrompt("Do you want to continue", "", true)
}

func projectNamePrompt(ctx context.Context, ch *cmdutil.Helper, orgName string) (string, error) {
	questions := []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "Enter a project name",
			},
			Validate: func(v any) error {
				name := v.(string)
				if name == "" {
					return fmt.Errorf("empty name")
				}
				exists, err := projectExists(ctx, ch, orgName, name)
				if err != nil {
					return fmt.Errorf("project already exists at %s/%s", orgName, name)
				}
				if exists {
					// this should always be true but adding this check from completeness POV
					return fmt.Errorf("project with name %q already exists in the org", name)
				}
				return nil
			},
		},
	}

	name := ""
	if err := survey.Ask(questions, &name); err != nil {
		return "", err
	}

	return name, nil
}
