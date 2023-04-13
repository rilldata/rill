package deploy

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/cmd/auth"
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/cmd/org"
	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
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
func DeployCmd(cfg *config.Config) *cobra.Command {
	var description, projectPath, region, dbDriver, dbDSN, prodBranch, name string
	var slots int
	var public bool

	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Guided tour for deploying rill projects to rill cloud",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			warn := color.New(color.Bold).Add(color.FgYellow)
			info := color.New(color.Bold).Add(color.FgWhite)
			success := color.New(color.Bold).Add(color.FgGreen)
			if projectPath != "" {
				var err error
				projectPath, err = fileutil.ExpandHome(projectPath)
				if err != nil {
					return err
				}
			}

			// log in if not logged in
			if !cfg.IsAuthenticated() {
				warn.Println("In order to deploy to Rill Cloud, you must login.")
				time.Sleep(2 * time.Second)
				if err := auth.Login(ctx, cfg); err != nil {
					return fmt.Errorf("failed to login %w", err)
				}
			}
			adminClient, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}

			if cfg.Org == "" {
				// no default org set by user
				res, err := adminClient.ListOrganizations(context.Background(), &adminv1.ListOrganizationsRequest{})
				if err != nil {
					return fmt.Errorf("listing orgs failed with error: %w", err)
				}

				orgs := res.Organizations

				if len(orgs) == 1 {
					if err := dotrill.SetDefaultOrg(orgs[0].Name); err != nil {
						return err
					}
				}

				if len(orgs) > 1 {
					switchCmd := org.SwitchCmd(cfg)
					switchCmd.SetContext(ctx)
					if err := switchCmd.RunE(switchCmd, nil); err != nil {
						return fmt.Errorf("org selection failed %w", err)
					}
				}
			}

			// verify current directory has rill project
			if !hasRillProject(projectPath) {
				fullpath, err := filepath.Abs(projectPath)
				if err != nil {
					return err
				}

				warn.Printf("\nCurrent path `%s` doesn't have a valid rill project.\n\nPlease run `rill deploy` from correct path or pass correct path via `--project` flag.\n\n", fullpath)
				warn.Printf("In case there is no valid rill project present, Please use `rill init` to create an empty rill project.\n\n")
				return nil
			}

			// verify project dir is a git repo with remote on github
			githubURL, err := extractRemote(projectPath)
			if err != nil {
				if errors.Is(err, gitutil.ErrGitRemoteNotFound) || errors.Is(err, git.ErrRepositoryNotExists) {
					info.Print(githubSetupMsg)
					return nil
				}

				return err
			}

			defaultOrg, err := dotrill.GetDefaultOrg()
			if err != nil {
				return err
			}

			if defaultOrg == "" {
				// create an org for the user
				resp, err := createOrg(ctx, adminClient, githubURL)
				if err != nil {
					return fmt.Errorf("org creation failed with error: %w", err)
				}

				defaultOrg = resp.Name
				success.Printf("Created organization %q. Use `rill org edit` to change name if required.\n", defaultOrg)
			} else {
				info.Printf("Using org: %q\n", defaultOrg)
			}

			// Check for access to the Github URL
			ghRes, err := verifyAccess(ctx, adminClient, githubURL)
			if err != nil {
				return fmt.Errorf("failed to verify access to github repo, error = %w", err)
			}

			if prodBranch == "" {
				prodBranch = ghRes.DefaultBranch
			}

			// We now have access to the Github repo
			if name == "" {
				name = path.Base(githubURL)
			}

			variables, err := variablesPrompt(projectPath)
			if err != nil {
				return err
			}

			req := &adminv1.CreateProjectRequest{
				OrganizationName:     defaultOrg,
				Name:                 name,
				Description:          description,
				Region:               region,
				ProductionOlapDriver: dbDriver,
				ProductionOlapDsn:    dbDSN,
				ProductionSlots:      int64(slots),
				ProductionBranch:     prodBranch,
				Public:               public,
				GithubUrl:            githubURL,
				Variables:            variables,
			}
			// Create the project (automatically deploys prod branch)
			projRes, err := createProject(ctx, adminClient, req)
			if err != nil {
				return fmt.Errorf("create project failed with error %w", err)
			}

			// Success!
			success.Printf("Created project %s/%s. Use `rill project edit` to edit name if required.\n", defaultOrg, projRes.Project.Name)
			success.Printf("Rill projects deploy continuously when you push changes to Github.\n")
			if projRes.ProjectUrl != "" {
				success.Printf("Opening project dashboard. Your project can be accessed at %s\n", projRes.ProjectUrl)
				time.Sleep(3 * time.Second)
				_ = browser.Open(projRes.ProjectUrl)
			}
			// TODO :: add rill docs here
			return nil
		},
	}

	deployCmd.Flags().SortFlags = false
	deployCmd.Flags().StringVar(&projectPath, "project", ".", "Project directory")
	deployCmd.Flags().IntVar(&slots, "prod-slots", 2, "Slots to allocate for production deployments")
	deployCmd.Flags().StringVar(&description, "description", "", "Project description")
	deployCmd.Flags().StringVar(&region, "region", "", "Deployment region")
	deployCmd.Flags().StringVar(&dbDriver, "prod-db-driver", "duckdb", "Database driver")
	deployCmd.Flags().StringVar(&dbDSN, "prod-db-dsn", "", "Database driver configuration")
	deployCmd.Flags().BoolVar(&public, "public", false, "Make dashboards publicly accessible")
	deployCmd.Flags().StringVar(&prodBranch, "prod-branch", "", "Git branch to deploy from (default: the default Git branch)")
	deployCmd.Flags().StringVar(&name, "name", "", "Project name (default: taken from rill.yaml)")
	return deployCmd
}

func createOrg(ctx context.Context, adminClient *client.Client, githubURL string) (*adminv1.Organization, error) {
	resp, err := org.Create(adminClient, repoAccount(githubURL), "")
	if err != nil && violatesUniqueConstraint(err) {
		// org name already exists, prompt for the org name and create org with new name again
		name, err := orgNamePrompt(ctx, adminClient)
		if err != nil {
			return nil, err
		}

		return org.Create(adminClient, name, "")
	}
	return resp, err
}

func createProject(ctx context.Context, adminClient *client.Client, req *adminv1.CreateProjectRequest) (*adminv1.CreateProjectResponse, error) {
	// Create the project (automatically deploys prod branch)
	res, err := adminClient.CreateProject(ctx, req)
	if err != nil && violatesUniqueConstraint(err) {
		// project name already exists, prompt for project name and create project with new name again
		name, err := projectNamePrompt(ctx, adminClient, req.OrganizationName)
		if err != nil {
			return nil, err
		}

		req.Name = name
		return adminClient.CreateProject(ctx, req)
	}
	return res, err
}

func projectNamePrompt(ctx context.Context, c *client.Client, orgName string) (string, error) {
	questions := []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "What is the project name?",
			},
			Validate: func(any interface{}) error {
				name := any.(string)
				if name == "" {
					return fmt.Errorf("empty name")
				}
				exists, err := projectExists(ctx, c, orgName, name)
				if err != nil {
					return err
				}
				if exists {
					// this should always be true but adding this check from completeness POV
					return fmt.Errorf("project with name %v already exists in the org", name)
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

func projectExists(ctx context.Context, c *client.Client, orgName, projectName string) (bool, error) {
	resp, err := c.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: orgName, Name: projectName})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				return false, nil
			}
		}
		return false, err
	}
	return resp.Project.Name == projectName, nil
}

func orgNamePrompt(ctx context.Context, adminClient *client.Client) (string, error) {
	qs := []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "Please enter org name",
			},
			Validate: func(any interface{}) error {
				// Validate org name doesn't exist already
				name := any.(string)
				if name == "" {
					return fmt.Errorf("empty name")
				}

				exist, err := orgNameExists(ctx, adminClient, name)
				if err != nil {
					return err
				}

				if exist {
					// this should always be true but adding this check from completeness POV
					return fmt.Errorf("orgnaization with name %v already exists", name)
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

func orgNameExists(ctx context.Context, adminClient *client.Client, name string) (bool, error) {
	resp, err := adminClient.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: name})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound || st.Code() == codes.InvalidArgument { // todo :: remove invalid argument after deployment
				return false, nil
			}
		}
		return false, err
	}
	return resp.Organization.Name == name, nil
}

func variablesPrompt(projectPath string) (map[string]string, error) {
	connectors, err := rillv1beta.ExtractConnectors(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract connectors %w", err)
	}

	vars := make(map[string]string)
	for _, c := range connectors {
		connectorVariables := c.Spec.ConnectorVariables
		if len(connectorVariables) != 0 {
			fmt.Printf("\nConnector %s requires credentials\n\n", c.Type)
		}
		if c.Spec.Help != "" {
			fmt.Println(c.Spec.Help)
		}
		for _, prop := range connectorVariables {
			question := &survey.Question{}
			msg := fmt.Sprintf("connector.%s.%s", c.Name, prop.Key)
			if prop.Help != "" {
				msg = fmt.Sprintf(msg+" (%s)", prop.Help)
			}

			if prop.Secret {
				question.Prompt = &survey.Password{Message: msg}
			} else {
				question.Prompt = &survey.Input{Message: msg, Default: prop.Default}
			}

			if prop.TransformFunc != nil {
				question.Transform = prop.TransformFunc
			}

			if prop.ValidateFunc != nil {
				question.Validate = prop.ValidateFunc
			}

			answer := ""
			if err := survey.Ask([]*survey.Question{question}, &answer); err != nil {
				return nil, fmt.Errorf("variables prompt failed with error %w", err)
			}

			if answer != "" {
				vars[prop.Key] = answer
			}
		}
	}
	return vars, nil
}

func extractRemote(remotePath string) (string, error) {
	remotes, err := gitutil.ExtractRemotes(remotePath)
	if err != nil {
		return "", err
	}
	// Parse into a https://github.com/account/repo (no .git) format
	return gitutil.RemotesToGithubURL(remotes)
}

func hasRillProject(dir string) bool {
	_, err := os.Open(filepath.Join(dir, "rill.yaml"))
	return err == nil
}

func verifyAccess(ctx context.Context, c *client.Client, githubURL string) (*adminv1.GetGithubRepoStatusResponse, error) {
	// Check for access to the Github URL
	ghRes, err := c.GetGithubRepoStatus(ctx, &adminv1.GetGithubRepoStatusRequest{
		GithubUrl: githubURL,
	})
	if err != nil {
		return nil, err
	}

	// If the user has not already granted access, open browser and poll for access
	if !ghRes.HasAccess {
		// Print instructions to grant access
		fmt.Printf("Rill projects deploy continuously when you push changes to Github.\n")
		fmt.Printf("You need to install rill github app to grant read only access to your project.\n\n")
		time.Sleep(3 * time.Second)
		fmt.Printf("Open this URL in your browser to grant Rill access to your Github repository:\n\n")
		fmt.Printf("\t%s\n\n", ghRes.GrantAccessUrl)

		// Open browser if possible
		_ = browser.Open(ghRes.GrantAccessUrl)

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
				return pollRes, nil
			}

			// Sleep and poll again
		}
	}
	return ghRes, nil
}

func repoAccount(githubURL string) string {
	ep, err := transport.NewEndpoint(githubURL)
	if err != nil {
		return ""
	}

	if ep.Host != "github.com" {
		return ""
	}

	account, repo := path.Split(ep.Path)
	account = strings.Trim(account, "/")
	if account == "" || repo == "" || strings.Contains(account, "/") {
		return ""
	}

	return account
}

func violatesUniqueConstraint(err error) bool {
	if st, ok := status.FromError(err); ok && st != nil {
		return strings.Contains(st.Message(), "violates unique constraint")
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
	
5. Push your repository
	
	git push -u origin main
	
6. Deploy Rill to your repository
	
	rill deploy
	
`
)
