package deploy

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
	adminclient "github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/admin/pkg/urlutil"
	"github.com/rilldata/rill/cli/cmd/auth"
	"github.com/rilldata/rill/cli/cmd/org"
	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/deviceauth"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/telemetry"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/compilers/rillv1beta"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
	deployCmd.Flags().StringVar(&ch.Config.Org, "org", ch.Config.Org, "Org to deploy project in")
	deployCmd.Flags().StringVar(&opts.Name, "project", "", "Project name (default: Git repo name)")
	deployCmd.Flags().StringVar(&opts.Description, "description", "", "Project description")
	deployCmd.Flags().BoolVar(&opts.Public, "public", false, "Make dashboards publicly accessible")
	deployCmd.Flags().StringVar(&opts.Provisioner, "provisioner", "", "Project provisioner")
	deployCmd.Flags().StringVar(&opts.ProdVersion, "prod-version", "latest", "Rill version (default: the latest release version)")
	deployCmd.Flags().StringVar(&opts.ProdBranch, "prod-branch", "", "Git branch to deploy from (default: the default Git branch)")
	deployCmd.Flags().StringVar(&opts.DBDriver, "prod-db-driver", "duckdb", "Database driver")
	deployCmd.Flags().StringVar(&opts.DBDSN, "prod-db-dsn", "", "Database driver configuration")
	deployCmd.Flags().IntVar(&opts.Slots, "prod-slots", 2, "Slots to allocate for production deployments")

	if !ch.Config.IsDev() {
		if err := deployCmd.Flags().MarkHidden("prod-slots"); err != nil {
			panic(err)
		}
	}

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
}

func DeployFlow(ctx context.Context, ch *cmdutil.Helper, opts *Options) error {
	// Setup telemetry
	cfg := ch.Config
	tel := telemetry.New(cfg.Version)
	if cfg.IsAuthenticated() {
		userID, err := cmdutil.FetchUserID(ctx, cfg)
		if err == nil {
			tel.WithUserID(userID)
		}
	}
	tel.Emit(telemetry.ActionDeployStart)
	defer func() {
		// give 5s for emitting events over the parent context.
		// this will make sure if user cancelled the command events are still fired.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// telemetry errors shouldn't fail deploy command
		_ = tel.Flush(ctx)
	}()

	// The gitPath can be either a local path or a remote .git URL.
	// Determine which it is.
	var isLocalGitPath bool
	var githubURL string
	if opts.GitPath != "" {
		u, err := url.Parse(opts.GitPath)
		if err != nil || u.Scheme == "" {
			isLocalGitPath = true
		} else {
			githubURL, err = gitutil.RemoteToGithubURL(opts.GitPath)
			if err != nil {
				return fmt.Errorf("failed to parse path as a Github remote: %w", err)
			}
		}
	}

	// If the Git path is local, we'll do some extra steps to infer the githubURL.
	var localGitPath, localProjectPath string
	if isLocalGitPath {
		var err error
		if opts.GitPath != "" {
			localGitPath, err = fileutil.ExpandHome(opts.GitPath)
			if err != nil {
				return err
			}
		}
		localGitPath, err = filepath.Abs(localGitPath)
		if err != nil {
			return err
		}

		if opts.SubPath == "" {
			localProjectPath = localGitPath
		} else {
			localProjectPath = filepath.Join(localGitPath, opts.SubPath)
		}

		// Verify that localProjectPath contains a Rill project.
		// If not, we still navigate user to login and then fail afterwards.
		if !rillv1beta.HasRillProject(localProjectPath) {
			if !cfg.IsAuthenticated() {
				err := loginWithTelemetry(ctx, ch, "", tel)
				if err != nil {
					ch.Printer.PrintlnWarn(fmt.Sprintf("Login failed with error: %s", err.Error()))
				}
				fmt.Println()
			}

			ch.Printer.PrintlnWarn(fmt.Sprintf("Directory %q doesn't contain a valid Rill project.", localProjectPath))
			ch.Printer.PrintlnWarn("Run `rill deploy` from a Rill project directory or use `--path` to pass a project path.")
			ch.Printer.PrintlnWarn("Run `rill start` to initialize a new Rill project.")
			return nil
		}

		// Extract the Git remote and infer the githubURL.
		var remote *gitutil.Remote
		remote, githubURL, err = gitutil.ExtractGitRemote(localGitPath, opts.RemoteName)
		if err != nil {
			// It's not a valid remote for Github. But we still navigate user to login and then fail afterwards.
			if !cfg.IsAuthenticated() {
				err := loginWithTelemetry(ctx, ch, "", tel)
				if err != nil {
					ch.Printer.PrintlnWarn(fmt.Sprintf("Login failed with error: %s", err.Error()))
				}
				fmt.Println()
			}
			if errors.Is(err, gitutil.ErrGitRemoteNotFound) || errors.Is(err, git.ErrRepositoryNotExists) {
				ch.Printer.PrintlnInfo(githubSetupMsg)
				return nil
			}
			return err
		}

		// Error if the repository is not in sync with the remote
		if !repoInSyncFlow(ch, localGitPath, opts.ProdBranch, remote.Name) {
			ch.Printer.PrintlnInfo("You can run `rill deploy` again when you have pushed your local changes to the remote.")
			return nil
		}
	}

	// We now have a githubURL.

	// Extract Github account and repo name from the githubURL
	ghAccount, ghRepo, ok := gitutil.SplitGithubURL(githubURL)
	if !ok {
		ch.Printer.PrintlnError(fmt.Sprintf("Invalid Github URL %q\n", githubURL))
		return nil
	}

	// If user is not authenticated, run login flow.
	// To prevent opening the browser twice, we make it directly redirect to the Github flow.
	silentGitFlow := false
	if !cfg.IsAuthenticated() {
		silentGitFlow = true
		authURL := cfg.AdminURL
		if strings.Contains(authURL, "http://localhost:9090") {
			authURL = "http://localhost:8080"
		}
		redirectURL, err := urlutil.WithQuery(urlutil.MustJoinURL(authURL, "/github/post-auth-redirect"), map[string]string{"remote": githubURL})
		if err != nil {
			return err
		}
		if err := loginWithTelemetry(ctx, ch, redirectURL, tel); err != nil {
			return err
		}
	}

	client, err := cmdutil.Client(cfg)
	if err != nil {
		return err
	}

	// Run flow for access to the Github remote (if necessary)
	ghRes, err := githubFlow(ctx, ch, client, githubURL, silentGitFlow, tel)
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

	// Set a default org for the user if necessary
	// (If user is not in an org, we'll create one based on their Github account later in the flow.)
	if cfg.Org == "" {
		res, err := client.ListOrganizations(context.Background(), &adminv1.ListOrganizationsRequest{})
		if err != nil {
			return fmt.Errorf("listing orgs failed: %w", err)
		}

		if len(res.Organizations) == 1 {
			cfg.Org = res.Organizations[0].Name
			if err := dotrill.SetDefaultOrg(cfg.Org); err != nil {
				return err
			}
		} else if len(res.Organizations) > 1 {
			orgName, err := org.SwitchSelectFlow(res.Organizations)
			if err != nil {
				return fmt.Errorf("org selection failed %w", err)
			}

			cfg.Org = orgName
			if err := dotrill.SetDefaultOrg(cfg.Org); err != nil {
				return err
			}
		}
	}

	// If no default org is set by now, it means the user is not in an org yet.
	// We create a default org based on their Github account name.
	if cfg.Org == "" {
		err := createOrgFlow(ctx, ch, cfg, client, ghAccount)
		if err != nil {
			return fmt.Errorf("org creation failed with error: %w", err)
		}
		ch.Printer.PrintlnSuccess(fmt.Sprintf("Created org %q. Run `rill org edit` to change name if required.\n", cfg.Org))
	} else {
		ch.Printer.PrintlnInfo(fmt.Sprintf("Using org %q.", cfg.Org))
	}

	// Check if a project matching githubURL already exists in this org
	projectNameExists := false
	projects, err := cmdutil.ProjectNamesByGithubURL(ctx, client, cfg.Org, githubURL)
	if err == nil && len(projects) != 0 { // ignoring error since this is just for a confirmation prompt
		for _, p := range projects {
			if strings.EqualFold(opts.Name, p) {
				projectNameExists = true
				break
			}
		}

		ch.Printer.PrintlnWarn(fmt.Sprintf("Another project %q already deploys from %q.", projects[0], githubURL))
		ch.Printer.PrintlnInfo("- To force the existing project to rebuild, press 'n' and run `rill project reconcile --reset`")
		ch.Printer.PrintlnInfo("- To delete the existing project, press 'n' and run `rill project delete`")
		ch.Printer.PrintlnInfo("- To deploy the repository as a new project under another name, press 'y' or enter")
		if !cmdutil.ConfirmPrompt("Do you want to continue?", "", true) {
			ch.Printer.PrintlnWarn("Aborted")
			return nil
		}
	}

	// If the project name already exists, prompt for another name
	if projectNameExists {
		opts.Name, err = projectNamePrompt(ctx, client, cfg.Org)
		if err != nil {
			return err
		}
	}

	// Create the project (automatically deploys prod branch)
	res, err := createProjectFlow(ctx, ch, client, &adminv1.CreateProjectRequest{
		OrganizationName: cfg.Org,
		Name:             opts.Name,
		Description:      opts.Description,
		Provisioner:      opts.Provisioner,
		ProdVersion:      opts.ProdVersion,
		ProdOlapDriver:   opts.DBDriver,
		ProdOlapDsn:      opts.DBDSN,
		ProdSlots:        int64(opts.Slots),
		Subpath:          opts.SubPath,
		ProdBranch:       opts.ProdBranch,
		Public:           opts.Public,
		GithubUrl:        githubURL,
	})
	if err != nil {
		if s, ok := status.FromError(err); ok && s.Code() == codes.PermissionDenied {
			ch.Printer.PrintlnError(fmt.Sprintf("You do not have the permissions needed to create a project in org %q. Please reach out to your Rill admin.\n", cfg.Org))
			return nil
		}
		return fmt.Errorf("create project failed with error %w", err)
	}

	// Success!
	ch.Printer.PrintlnSuccess(fmt.Sprintf("Created project \"%s/%s\". Use `rill project rename` to change name if required.\n\n", cfg.Org, res.Project.Name))
	ch.Printer.PrintlnSuccess("Rill projects deploy continuously when you push changes to Github.")

	// If the Git path is local, we can parse the project and check if credentials are available for the connectors used by the project.
	if isLocalGitPath {
		variablesFlow(ctx, ch, localProjectPath, opts.Name)
	}

	// Open browser
	if res.Project.FrontendUrl != "" {
		ch.Printer.PrintlnSuccess(fmt.Sprintf("Your project can be accessed at: %s", res.Project.FrontendUrl))
		ch.Printer.PrintlnSuccess("Opening project in browser...")
		time.Sleep(3 * time.Second)
		_ = browser.Open(res.Project.FrontendUrl)
	}

	tel.Emit(telemetry.ActionDeploySuccess)
	return nil
}

func loginWithTelemetry(ctx context.Context, ch *cmdutil.Helper, redirectURL string, tel *telemetry.Telemetry) error {
	cfg := ch.Config
	ch.Printer.PrintlnInfo("Please log in or sign up for Rill. Opening browser...")
	time.Sleep(2 * time.Second)

	tel.Emit(telemetry.ActionLoginStart)
	if err := auth.Login(ctx, ch, redirectURL); err != nil {
		if errors.Is(err, deviceauth.ErrAuthenticationTimedout) {
			ch.Printer.PrintlnWarn("Rill login has timed out as the code was not confirmed in the browser.")
			ch.Printer.PrintlnWarn("Run `rill deploy` again.")
			return nil
		} else if errors.Is(err, deviceauth.ErrCodeRejected) {
			ch.Printer.PrintlnError("Login failed: Confirmation code rejected")
			return nil
		}
		return fmt.Errorf("login failed: %w", err)
	}
	userID, err := cmdutil.FetchUserID(ctx, cfg)
	if err == nil {
		tel.WithUserID(userID)
	}
	// fire this after we potentially get the user id
	tel.Emit(telemetry.ActionLoginSuccess)
	return nil
}

func githubFlow(ctx context.Context, ch *cmdutil.Helper, c *adminclient.Client, githubURL string, silent bool, tel *telemetry.Telemetry) (*adminv1.GetGithubRepoStatusResponse, error) {
	// Check for access to the Github repo
	res, err := c.GetGithubRepoStatus(ctx, &adminv1.GetGithubRepoStatusRequest{
		GithubUrl: githubURL,
	})
	if err != nil {
		return nil, err
	}

	// If the user has not already granted access, open browser and poll for access
	if !res.HasAccess {
		tel.Emit(telemetry.ActionGithubConnectedStart)

		// Print instructions to grant access
		if !silent {
			ch.Printer.Print("Rill projects deploy continuously when you push changes to Github.\n")
			ch.Printer.Print("You need to grant Rill read only access to your repository on Github.\n\n")
			time.Sleep(3 * time.Second)
			ch.Printer.Print("Open this URL in your browser to grant Rill access to Github:\n\n")
			ch.Printer.Print("\t" + res.GrantAccessUrl + "\n\n")

			// Open browser if possible
			_ = browser.Open(res.GrantAccessUrl)
		} else {
			ch.Printer.Printf("Polling for Github access for: %q\n", githubURL)
			ch.Printer.Printf("If the browser did not redirect, visit this URL to grant access: %q\n\n", res.GrantAccessUrl)
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
				GithubUrl: githubURL,
			})
			if err != nil {
				return nil, err
			}

			if pollRes.HasAccess {
				// Success
				tel.Emit(telemetry.ActionGithubConnectedSuccess)
				_, ghRepo, _ := gitutil.SplitGithubURL(githubURL)
				ch.Printer.PrintlnSuccess(fmt.Sprintf("You have connected to the %q project in Github.", ghRepo))
				return pollRes, nil
			}

			// Sleep and poll again
		}
	}

	return res, nil
}

func createOrgFlow(ctx context.Context, ch *cmdutil.Helper, cfg *config.Config, client *adminclient.Client, defaultName string) error {
	res, err := client.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{
		Name: defaultName,
	})
	if err != nil {
		if !errMsgContains(err, "an org with that name already exists") {
			return err
		}

		ch.Printer.PrintlnWarn("Rill organizations are derived from the owner of your Github repository.")
		ch.Printer.PrintlnWarn(fmt.Sprintf("The %q organization associated with your Github repository already exists.", defaultName))
		ch.Printer.PrintlnWarn("Contact your Rill admin to be added to your org or create a new organization below.")

		name, err := orgNamePrompt(ctx, client)
		if err != nil {
			return err
		}

		res, err = client.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{
			Name: name,
		})
		if err != nil {
			return err
		}
	}

	// Switching to the created org
	cfg.Org = res.Organization.Name
	err = dotrill.SetDefaultOrg(cfg.Org)
	if err != nil {
		return err
	}

	return nil
}

func orgNamePrompt(ctx context.Context, client *adminclient.Client) (string, error) {
	qs := []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "Enter an org name",
			},
			Validate: func(any interface{}) error {
				// Validate org name doesn't exist already
				name := any.(string)
				if name == "" {
					return fmt.Errorf("empty name")
				}

				exist, err := cmdutil.OrgExists(ctx, client, name)
				if err != nil {
					return fmt.Errorf("org name %q is already taken", name)
				}

				if exist {
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

func createProjectFlow(ctx context.Context, ch *cmdutil.Helper, client *adminclient.Client, req *adminv1.CreateProjectRequest) (*adminv1.CreateProjectResponse, error) {
	// Create the project (automatically deploys prod branch)
	res, err := client.CreateProject(ctx, req)
	if err != nil {
		if !errMsgContains(err, "a project with that name already exists in the org") {
			return nil, err
		}

		ch.Printer.PrintlnWarn("Rill project names are derived from your Github repository name.")
		ch.Printer.PrintlnWarn(fmt.Sprintf("The %q project already exists under org %q. Please enter a different name.\n", req.Name, req.OrganizationName))

		// project name already exists, prompt for project name and create project with new name again
		name, err := projectNamePrompt(ctx, client, req.OrganizationName)
		if err != nil {
			return nil, err
		}

		req.Name = name
		return client.CreateProject(ctx, req)
	}
	return res, err
}

func variablesFlow(ctx context.Context, ch *cmdutil.Helper, gitPath, projectName string) {
	// Parse the project's connectors
	repo, instanceID, err := cmdutil.RepoForProjectPath(gitPath)
	if err != nil {
		return
	}
	parser, err := rillv1.Parse(ctx, repo, instanceID, "prod", "duckdb", []string{"duckdb"})
	if err != nil {
		return
	}
	connectors, err := parser.AnalyzeConnectors(ctx)
	if err != nil {
		return
	}

	// Exit early if all connectors can be used anonymously
	foundNotAnonymous := false
	for _, c := range connectors {
		if !c.AnonymousAccess {
			foundNotAnonymous = true
		}
	}
	if !foundNotAnonymous {
		return
	}

	ch.Printer.PrintlnWarn("\nCould not ingest all sources. Rill requires credentials for the following sources:\n")
	for _, c := range connectors {
		if c.AnonymousAccess {
			continue
		}
		for _, r := range c.Resources {
			fmt.Printf(" - %s", r.Name.Name)
			if len(r.Paths) > 0 {
				fmt.Printf(" (%s)", r.Paths[0])
			}
			fmt.Print("\n")
		}
	}
	ch.Printer.PrintlnWarn(fmt.Sprintf("\nRun `rill env configure --project %s` to provide credentials.\n", projectName))
	time.Sleep(2 * time.Second)
}

func repoInSyncFlow(ch *cmdutil.Helper, gitPath, branch, remoteName string) bool {
	syncStatus, err := gitutil.GetSyncStatus(gitPath, branch, remoteName)
	if err != nil {
		// ignore errors since check is best effort and can fail in multiple cases
		return true
	}

	switch syncStatus {
	case gitutil.SyncStatusUnspecified:
		return true
	case gitutil.SyncStatusSynced:
		return true
	case gitutil.SyncStatusModified:
		ch.Printer.PrintlnWarn("Some files have been locally modified. These changes will not be present in the deployed project.")
	case gitutil.SyncStatusAhead:
		ch.Printer.PrintlnWarn("Local commits are not pushed to remote yet. These changes will not be present in the deployed project.")
	}

	return cmdutil.ConfirmPrompt("Do you want to continue", "", true)
}

func projectNamePrompt(ctx context.Context, client *adminclient.Client, orgName string) (string, error) {
	questions := []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "Enter a project name",
			},
			Validate: func(any interface{}) error {
				name := any.(string)
				if name == "" {
					return fmt.Errorf("empty name")
				}
				exists, err := cmdutil.ProjectExists(ctx, client, orgName, name)
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

func errMsgContains(err error, msg string) bool {
	if st, ok := status.FromError(err); ok && st != nil {
		return strings.Contains(st.Message(), msg)
	}
	return false
}

const (
	githubSetupMsg = `No git remote was found.

Rill projects deploy continuously when you push changes to Github.
Therefore, your project must be on Github before you deploy it to Rill.

Follow these steps to push your project to Github.

1. Initialize git

	git init

2. Add and commit files

	git add .
	git commit -m 'initial commit'

3. Create a new GitHub repository on https://github.com/new

4. Link git to the remote repository

	git remote add origin https://github.com/your-account/your-repo.git

5. Rename master branch to main

	git branch -M main

6. Push your repository

	git push -u origin main

7. Deploy Rill to your repository

	rill deploy

`
)
