package project

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/google/go-github/v71/github"
	"github.com/rilldata/rill/cli/cmd/org"
	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
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
			opts.Github = true
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
	deployCmd.Flags().BoolVar(&opts.PushEnv, "push-env", true, "Push local .env file to Rill Cloud")
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
	// TODO : similar to UI workflow create a org taking user input
	if ch.Org == "" {
		if err := org.SetDefaultOrg(ctx, ch); err != nil {
			return err
		}
	}

	err := opts.ValidateAndApplyDefaults(ctx, ch)
	if err != nil {
		return err
	}

	localGitPath := opts.GitPath
	localProjectPath := opts.LocalProjectPath()

	if opts.pushToProject != nil {
		return redeployProject(ctx, ch, opts)
	}

	if opts.remoteURL == "" {
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
		remote, err := gitutil.ExtractGitRemote(localGitPath, opts.RemoteName, false)
		if err != nil {
			return err
		}
		opts.remoteURL, err = remote.Github()
		opts.RemoteName = remote.Name
		if err != nil {
			return err
		}
	}

	// Error if the repository is not in sync with the remote
	ok, err := repoInSyncFlow(ch, localGitPath, opts.SubPath, opts.RemoteName)
	if err != nil {
		return err
	}
	if !ok {
		ch.PrintfBold("You can run `rill deploy` again when you have pushed your local changes to the remote.\n")
		return nil
	}

	// Extract Github account and repo name from the gitRemote
	_, ghRepo, ok := gitutil.SplitGithubRemote(opts.remoteURL)
	if !ok {
		return fmt.Errorf("remote %q is not a valid github.com remote", opts.remoteURL)
	}

	// Run flow for access to the Github remote (if necessary)
	ghRes, err := githubFlow(ctx, ch, opts.remoteURL)
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
		err := createOrgFlow(ctx, ch)
		if err != nil {
			return fmt.Errorf("org creation failed with error: %w", err)
		}
		ch.PrintfSuccess("Created org %q. Run `rill org edit` to change name if required.\n\n", ch.Org)
	} else {
		ch.PrintfBold("Using org %q.\n\n", ch.Org)
	}

	// Create the project (automatically deploys prod branch)
	res, err := createProjectFlow(ctx, ch, &adminv1.CreateProjectRequest{
		Org:           ch.Org,
		Project:       opts.Name,
		Description:   opts.Description,
		Provisioner:   opts.Provisioner,
		ProdVersion:   opts.ProdVersion,
		ProdSlots:     int64(opts.Slots),
		Subpath:       opts.SubPath,
		ProdBranch:    opts.ProdBranch,
		Public:        opts.Public,
		DirectoryName: filepath.Base(localProjectPath),
		GitRemote:     opts.remoteURL,
		SkipDeploy:    opts.SkipDeploy,
	})
	if err != nil {
		if s, ok := status.FromError(err); ok && s.Code() == codes.PermissionDenied {
			ch.PrintfError("You do not have the permissions needed to create a project in org %q. Please reach out to your Rill admin.\n", ch.Org)
			return nil
		}
		return fmt.Errorf("create project failed with error %w", err)
	}

	// Success!
	ch.PrintfSuccess("Created project \"%s/%s\". Use `rill project rename` to change name if required.\n\n", ch.Org, res.Project.Name)
	ch.PrintfSuccess("Rill projects deploy continuously when you push changes to Github.\n")

	// Upload .env
	if opts.PushEnv {
		vars, err := local.ParseDotenv(ctx, localProjectPath)
		if err != nil {
			ch.PrintfWarn("Failed to parse .env: %v\n", err)
		} else if len(vars) > 0 {
			c, err := ch.Client()
			if err != nil {
				return err
			}
			_, err = c.UpdateProjectVariables(ctx, &adminv1.UpdateProjectVariablesRequest{
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
	err = gitutil.CommitAndPush(ctx, localGitPath, config, "", author)
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

		// Wait three seconds before opening the browser
		select {
		case <-time.After(3 * time.Second):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
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
		ch.PrintfWarn("The %q project already exists under org %q. Please enter a different name.\n", req.Project, req.Org)

		// project name already exists, prompt for project name and create project with new name again
		name, err := projectNamePrompt(ctx, ch, req.Org)
		if err != nil {
			return nil, err
		}

		req.Project = name
		return c.CreateProject(ctx, req)
	}
	return res, err
}

func repoInSyncFlow(ch *cmdutil.Helper, gitPath, subpath, remoteName string) (bool, error) {
	st, err := gitutil.RunGitStatus(gitPath, subpath, remoteName)
	if err != nil {
		return false, err
	}

	if !st.LocalChanges && st.LocalCommits == 0 {
		return true, nil
	}

	if st.LocalChanges {
		ch.PrintfWarn("Some files have been locally modified. These changes will not be present in the deployed project.\n")
	}
	if st.LocalCommits > 0 {
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
