package deploy

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	adminclient "github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/cmd/auth"
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/cmd/org"
	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/telemetry"
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
		Short: "Deploy project to Rill Cloud",
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

			// telemetry errors shouldn't fail deploy command
			tel, _ := telemetry.NewTelemetry(cfg.Version)
			tel.EmitDeployStart(ctx)

			// Verify that the projectPath contains a Rill project
			if !hasRillProject(projectPath) {
				fullpath, err := filepath.Abs(projectPath)
				if err != nil {
					return err
				}

				warn.Printf("Directory at %q doesn't contain a valid Rill project.\n\n", fullpath)
				warn.Printf("Run \"rill deploy\" from a Rill project directory or use \"--project\" to pass a project path.\n")
				warn.Printf("Run \"rill start\" to initialize a new Rill project.\n")
				return nil
			}

			// Verify projectPath is a Git repo with remote on Github
			githubURL, err := extractGitRemote(projectPath)
			if err != nil {
				if errors.Is(err, gitutil.ErrGitRemoteNotFound) || errors.Is(err, git.ErrRepositoryNotExists) {
					info.Print(githubSetupMsg)
					return nil
				}
				return err
			}

			// Extract Github account and repo name from remote URL
			ghAccount, ghRepo, ok := gitutil.SplitGithubURL(githubURL)
			if !ok {
				return fmt.Errorf("invalid remote %q", githubURL)
			}

			// If user is not authenticated, run login flow
			if !cfg.IsAuthenticated() {
				warn.Println("You are not yet authenticated. Opening browser to log in or sign up for Rill Cloud.")
				time.Sleep(2 * time.Second)
				if err := auth.Login(ctx, cfg); err != nil {
					return fmt.Errorf("login failed: %w", err)
				}
				fmt.Println("")
			}
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
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

			// Run flow for access to the Github remote (if necessary)
			ghRes, err := githubFlow(ctx, client, githubURL, tel)
			if err != nil {
				return fmt.Errorf("failed Github flow: %w", err)
			}

			if prodBranch == "" {
				prodBranch = ghRes.DefaultBranch
			}

			// If no project name was provided, default to Git repo name
			if name == "" {
				name = ghRepo
			}

			// If no default org is set by now, it means the user is not in an org yet.
			// We create a default org based on their Github account name.
			if cfg.Org == "" {
				err := createOrgFlow(ctx, cfg, client, ghAccount)
				if err != nil {
					return fmt.Errorf("org creation failed with error: %w", err)
				}
				success.Printf("Created org %q. Run \"rill org edit\" to change name if required.\n", cfg.Org)
			} else {
				info.Printf("Using org %q.\n", cfg.Org)
			}

			// Run flow to get connector credentials and other variables
			variables, err := variablesFlow(projectPath)
			if err != nil {
				return err
			}

			// Create the project (automatically deploys prod branch)
			res, err := createProjectFlow(ctx, client, &adminv1.CreateProjectRequest{
				OrganizationName:     cfg.Org,
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
			})
			if err != nil {
				return fmt.Errorf("create project failed with error %w", err)
			}

			// Success!
			success.Printf("Created project \"%s/%s\". Use \"rill project rename\" to change name if required.\n", cfg.Org, res.Project.Name)
			success.Printf("Rill projects deploy continuously when you push changes to Github.\n")
			if res.ProjectUrl != "" {
				success.Printf("Your project can be accessed at: %s\n", res.ProjectUrl)
				success.Printf("Opening project in browser...\n")
				time.Sleep(3 * time.Second)
				_ = browser.Open(res.ProjectUrl)
			}

			// telemetry errors shouldn't fail deploy command
			tel.EmitDeploySuccess(ctx)
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

func githubFlow(ctx context.Context, c *adminclient.Client, githubURL string, tel *telemetry.Telemetry) (*adminv1.GetGithubRepoStatusResponse, error) {
	// Check for access to the Github repo
	res, err := c.GetGithubRepoStatus(ctx, &adminv1.GetGithubRepoStatusRequest{
		GithubUrl: githubURL,
	})
	if err != nil {
		return nil, err
	}

	// If the user has not already granted access, open browser and poll for access
	if !res.HasAccess {
		// Print instructions to grant access
		fmt.Printf("Rill projects deploy continuously when you push changes to Github.\n")
		fmt.Printf("You need to grant Rill read only access to your repository on Github.\n\n")
		time.Sleep(3 * time.Second)
		fmt.Printf("Open this URL in your browser to grant Rill access to Github:\n\n")
		fmt.Printf("\t%s\n\n", res.GrantAccessUrl)
		tel.EmitGithubConnectedStart(ctx)

		// Open browser if possible
		_ = browser.Open(res.GrantAccessUrl)

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
				tel.EmitGithubConnectedSuccess(ctx)
				return pollRes, nil
			}

			// Sleep and poll again
		}
	}

	return res, nil
}

func variablesFlow(projectPath string) (map[string]string, error) {
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

	if len(connectors) > 0 {
		fmt.Println("")
	}

	return vars, nil
}

func createOrgFlow(ctx context.Context, cfg *config.Config, client *adminclient.Client, defaultName string) error {
	res, err := client.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{
		Name: defaultName,
	})
	if err != nil {
		if !isNameExistsErr(err) {
			return err
		}

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

				exist, err := orgNameExists(ctx, client, name)
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

func orgNameExists(ctx context.Context, client *adminclient.Client, name string) (bool, error) {
	resp, err := client.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: name})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				return false, nil
			}
		}
		return false, err
	}
	return resp.Organization.Name == name, nil
}

func createProjectFlow(ctx context.Context, client *adminclient.Client, req *adminv1.CreateProjectRequest) (*adminv1.CreateProjectResponse, error) {
	// Create the project (automatically deploys prod branch)
	res, err := client.CreateProject(ctx, req)
	if err != nil {
		if !isNameExistsErr(err) {
			return nil, err
		}

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
				exists, err := projectExists(ctx, client, orgName, name)
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

func projectExists(ctx context.Context, client *adminclient.Client, orgName, projectName string) (bool, error) {
	resp, err := client.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: orgName, Name: projectName})
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

func hasRillProject(dir string) bool {
	_, err := os.Open(filepath.Join(dir, "rill.yaml"))
	return err == nil
}

func extractGitRemote(projectPath string) (string, error) {
	remotes, err := gitutil.ExtractRemotes(projectPath)
	if err != nil {
		return "", err
	}

	// Parse into a https://github.com/account/repo (no .git) format
	return gitutil.RemotesToGithubURL(remotes)
}

func isNameExistsErr(err error) bool {
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
