package deploy

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
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
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	pollTimeout  = 10 * time.Minute
	pollInterval = 5 * time.Second
)

type promptOptions struct {
	Name       string
	ProdBranch string
	Public     bool
}

// DeployCmd is the guided tour for deploying rill projects to rill cloud.
func DeployCmd(cfg *config.Config) *cobra.Command {
	var description, projectPath, region, dbDriver, dbDSN string
	var slots int

	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Guided tour for deploying rill projects to rill cloud",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			warn := color.New(color.Bold).Add(color.FgYellow)
			info := color.New(color.Bold).Add(color.FgWhite)
			success := color.New(color.Bold).Add(color.FgGreen)
			var adminClient *client.Client

			// log in if not logged in
			if !cfg.IsAuthenticated() {
				warn.Println("In order to deploy to Rill Cloud, you must login.")
				time.Sleep(2 * time.Second)
				// NOTE : calling commands within commands has both pros and cons
				// PRO : No duplicated code
				// CON : Need to make sure that UX under sub command verifies with UX on this command
				loginCmd := auth.LoginCmd(cfg)
				loginCmd.SetContext(ctx)
				if err := loginCmd.RunE(loginCmd, nil); err != nil {
					return err
				}

				var err error
				// init admin client
				adminClient, err = cmdutil.Client(cfg)
				if err != nil {
					return err
				}
			} else {
				// switch is already part of login cmd so running this only when user is already logged in
				defaultOrg, err := dotrill.GetDefaultOrg()
				if err != nil {
					return err
				}

				// init admin client
				adminClient, err = cmdutil.Client(cfg)
				if err != nil {
					return err
				}

				multipleOrgs, err := multipleOrgs(adminClient)
				if err != nil {
					return fmt.Errorf("listing orgs failed with error %w", err)
				}

				if multipleOrgs {
					msg := fmt.Sprintf("This project will be deployed under %s. Press Y to confirm and N to select a different org", defaultOrg)
					if !cmdutil.ConfirmPrompt(msg, true) {
						switchCmd := org.SwitchCmd(cfg)
						switchCmd.SetContext(ctx)
						if err := switchCmd.RunE(switchCmd, nil); err != nil {
							exitWithFailure(err)
						}
					}
				}
			}

			org, err := dotrill.GetDefaultOrg()
			if err != nil {
				return err
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

			// Check for access to the Github URL
			ghRes, err := verifyAccess(ctx, adminClient, githubURL)
			if err != nil {
				return fmt.Errorf("failed to verify access to github repo, error = %w", err)
			}

			// We now have access to the Github repo
			opts, err := projectParamPrompt(ctx, adminClient, org, githubURL, ghRes.DefaultBranch)
			if err != nil {
				return err
			}

			variables, err := variablesPrompt(projectPath)
			if err != nil {
				return err
			}

			// Create the project (automatically deploys prod branch)
			projRes, err := adminClient.CreateProject(ctx, &adminv1.CreateProjectRequest{
				OrganizationName:     org,
				Name:                 opts.Name,
				Description:          description,
				Region:               region,
				ProductionOlapDriver: dbDriver,
				ProductionOlapDsn:    dbDSN,
				ProductionSlots:      int64(slots),
				ProductionBranch:     opts.ProdBranch,
				Public:               opts.Public,
				GithubUrl:            githubURL,
				Variables:            variables,
			})
			if err != nil {
				return fmt.Errorf("create project failed with error %w", err)
			}

			// Success!
			success.Printf("Created project %s/%s\n", cfg.Org, projRes.Project.Name)
			success.Printf("Rill projects deploy continuously when you push changes to Github.\n")
			if projRes.ProjectUrl != "" {
				success.Printf("Your project can be accessed at %s\n", projRes.ProjectUrl)
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
	return deployCmd
}

func projectParamPrompt(ctx context.Context, c *client.Client, orgName, githubURL, prodBranch string) (*promptOptions, error) {
	projectName := path.Base(githubURL)

	questions := []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "What is the project name?",
				Default: projectName,
			},
			Validate: func(any interface{}) error {
				projectName := any.(string)
				resp, err := c.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: orgName, Name: projectName})
				if err != nil {
					if st, ok := status.FromError(err); ok {
						if st.Code() == codes.NotFound || st.Code() == codes.InvalidArgument { // todo :: remove InvalidArgument once admin server is deployed
							return nil
						}
					}
					return err
				}
				if resp.Project.Name == projectName {
					// this should always be true but adding this check from completeness POV
					return fmt.Errorf("project with name %v already exists in the org", projectName)
				}
				return nil
			},
		},
		{
			Name: "prodBranch",
			Prompt: &survey.Input{
				Message: "What branch will you deploy on rill cloud?",
				Default: prodBranch,
			},
		},
		{
			Name: "public",
			Prompt: &survey.Confirm{
				Message: "Do you want to deploy a public dashboard?",
				Default: false,
			},
		},
	}

	opts := &promptOptions{}
	if err := survey.Ask(questions, opts); err != nil {
		return nil, err
	}

	return opts, nil
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
				exitWithFailure(fmt.Errorf("variables prompt failed with error %w", err))
			}

			if answer != "" {
				vars[prop.Key] = answer
			}
		}
	}
	return vars, nil
}

func multipleOrgs(c *client.Client) (bool, error) {
	res, err := c.ListOrganizations(context.Background(), &adminv1.ListOrganizationsRequest{PageSize: 2})
	if err != nil {
		return false, err
	}

	return len(res.Organizations) > 1, nil
}

func extractRemote(remotePath string) (string, error) {
	remotes, err := gitutil.ExtractRemotes(remotePath)
	if err != nil {
		return "", err
	}
	// Parse into a https://github.com/account/repo (no .git) format
	return gitutil.RemotesToGithubURL(remotes)
}

func exitWithFailure(err error) {
	errormsg := color.New(color.Bold).Add(color.FgRed)
	errormsg.Printf("Prompt failed %v\n", err)
	os.Exit(1)
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
