package deploy

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/joho/godotenv"
	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/cmd/auth"
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/cmd/org"
	"github.com/rilldata/rill/cli/cmd/project"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/variable"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type options struct {
	Name        string
	Description string
	ProdBranch  string
	Public      bool
	Variables   []string

	// projectPath string input taken interactively
	// no input taken from user for below
	region   string
	dbDriver string
	dbDSN    string
	slots    int64
}

// DeployCmd is the guided tour for deploying rill projects to rill cloud.
// TODO :: add non interactive mode
func DeployCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Guided tour for deploying rill projects to rill cloud",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			warn := color.New(color.Bold).Add(color.FgYellow)
			info := color.New(color.Bold).Add(color.FgWhite)
			success := color.New(color.Bold).Add(color.FgGreen)
			// admin adminClient
			var adminClient *client.Client

			// log in if not logged in
			if !cfg.IsAuthenticated() {
				msg := fmt.Sprintf("In order to deploy to Rill Cloud, you must login. Opening your browsers to %s to login or sign up...", cfg.AdminURL)
				warn.Println(msg)
				// NOTE : calling commands within commands has both pros and cons
				// PRO : No duplicated code
				// CON : Need to make sure that UX under sub command verifies with UX on this command
				loginCmd := auth.LoginCmd(cfg)
				loginCmd.SetContext(ctx)
				if err := loginCmd.RunE(loginCmd, nil); err != nil {
					exitWithFailure(err)
				}

				// set token in config
				token, err := dotrill.GetAccessToken()
				if err != nil {
					return fmt.Errorf("could not parse access token from ~/.rill: %w", err)
				}
				cfg.AdminTokenDefault = token

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

			path, err := os.Getwd()
			if err != nil {
				return err
			}

			// verify current directory has rill project
			if !hasRillProject(path) {
				warn.Printf("\nCurrent path %s doesn't have a valid rill project. Please select correct path.\n", path)
				warn.Printf("In case there is no valid rill project present, Please use `rill init` to create an empty rill project.\n\n")
				// project path prompt
				path = projectPathPrompt()
			}

			// verify project dir is a git repo with remote on github
			githubURL, err := extractRemote(path)
			if err != nil {
				if errors.Is(err, gitutil.ErrGitRemoteNotFound) || errors.Is(err, git.ErrRepositoryNotExists) {
					info.Print(githubSetupMsg)
					return nil
				}

				return err
			}

			// Check for access to the Github URL
			ghRes, err := project.VerifyAccess(ctx, adminClient, githubURL)
			if err != nil {
				return fmt.Errorf("failed to verify access to github repo, error = %w", err)
			}

			// We now have access to the Github repo
			opts, err := projectParamPrompt(ctx, adminClient, org, githubURL, ghRes.DefaultBranch)
			if err != nil {
				return err
			}

			parsedVariables, err := variable.Parse(opts.Variables)
			if err != nil {
				return err
			}

			// fixing these for now
			opts.dbDriver = "duckdb"
			opts.slots = 2
			// Create the project (automatically deploys prod branch)
			projRes, err := adminClient.CreateProject(ctx, &adminv1.CreateProjectRequest{
				OrganizationName:     org,
				Name:                 opts.Name,
				Description:          opts.Description,
				Region:               opts.region,
				ProductionOlapDriver: opts.dbDriver,
				ProductionOlapDsn:    opts.dbDSN,
				ProductionSlots:      opts.slots,
				ProductionBranch:     opts.ProdBranch,
				Public:               opts.Public,
				GithubUrl:            githubURL,
				Variables:            parsedVariables,
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

	return cmd
}

func projectParamPrompt(ctx context.Context, c *client.Client, orgName, githubURL, prodBranch string) (*options, error) {
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
			Name: "description",
			Prompt: &survey.Input{
				Message: "What is the project description?",
				Default: "",
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
		{
			Name: "variables",
			Prompt: &survey.Editor{
				Message:       "Add variables for your project in format KEY=VALUE. Enter each variable on a new line",
				FileName:      "*.env",
				Default:       envFileDefault,
				HideDefault:   true,
				AppendDefault: true,
			},
			Validate: func(any interface{}) error {
				val := any.(string)
				envs, err := godotenv.Unmarshal(val)
				for key, value := range envs {
					if key == "" {
						return fmt.Errorf("invalid format found empty key")
					} else if key == "GCS_CREDENTIALS_FILE" {
						if _, err := os.Stat(value); err != nil {
							return err
						}
					}
				}
				return err
			},
			Transform: func(any interface{}) interface{} {
				val := any.(string)
				// ignoring error since already validated
				envs, _ := godotenv.Unmarshal(val)
				for k, v := range envs {
					if k == "GCS_CREDENTIALS_FILE" {
						content, _ := os.ReadFile(v)
						envs[k] = string(content)
					}
				}
				return variable.Serialize(envs)
			},
		},
	}

	opts := &options{}
	if err := survey.Ask(questions, opts); err != nil {
		return nil, err
	}

	return opts, nil
}

func projectPathPrompt() string {
	prompt := &survey.Input{
		Message: "What is your project path on local system?",
		Default: ".",
		Help:    "defaults to current directory",
		Suggest: func(toComplete string) []string {
			files, _ := filepath.Glob(toComplete + "*")
			return files
		},
	}
	q := []*survey.Question{
		{
			Name:   "pathParam",
			Prompt: prompt,
			Validate: func(any interface{}) error {
				path := any.(string)
				if !hasRillProject(path) {
					return fmt.Errorf("no rill project on path %v", path)
				}
				return nil
			},
		},
	}

	result := ""
	if err := survey.Ask(q, &result); err != nil {
		exitWithFailure(err)
	}
	return result
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

	envFileDefault = `## add any project specific variables in format KEY=VALUE
## If using private s3 sources uncomment next three and set credentials
# AWS_ACCESS_KEY_ID=
# AWS_SECRET_ACCESS_KEY=
# AWS_SESSION_TOKEN=

## If using private gcs sources set GCS_CREDENTIALS_FILE to a location where credentials.json for gcs is stored on your local system
## this creates an env GCS_CREDENTIALS with the file contents in env variable
# GCS_CREDENTIALS_FILE=

## add any other project specific credentials below:
`
)
