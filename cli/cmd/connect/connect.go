package connect

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/joho/godotenv"
	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/cmd/project"
	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/deviceauth"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/variable"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type connectOptions struct {
	Name        string
	Description string
	ProdBranch  string
	Slots       int
	Public      bool
	Variables   []string

	// projectPath string
	region string
	// dbDriver    string should we take input ??
	dbDSN string
}

func Questions(projectName, prodBranch string, v *validator) []*survey.Question {
	questions := make([]*survey.Question, 0)
	questions = append(questions,
		&survey.Question{
			Name: "name",
			Prompt: &survey.Input{
				Message: "What is the project name?",
				Default: projectName,
			},
			Validate: v.validateProjectName,
		},
		&survey.Question{
			Name: "description",
			Prompt: &survey.Input{
				Message: "What is the project description?",
				Default: "",
			},
		},
		&survey.Question{
			Name: "prodBranch",
			Prompt: &survey.Input{
				Message: "What branch will you deploy on rill cloud?",
				Default: prodBranch,
			},
		},
		&survey.Question{
			Name: "slots",
			Prompt: &survey.Input{
				Message: "How many slots do you want for your project?",
				Default: "2",
			},
		},
		&survey.Question{
			Name: "public",
			Prompt: &survey.Confirm{
				Message: "Do you want to deploy a public dashboard?",
				Default: false,
			},
		},
		&survey.Question{
			Name: "variables",
			Prompt: &survey.Editor{
				Message:  "Add variables for your project in format KEY=VALUE. Enter each variable on a new line",
				FileName: "*.env",
			},
			Validate: func(any interface{}) error {
				val := any.(string)
				envs, err := godotenv.Unmarshal(val)
				for key := range envs {
					if key == "" {
						return fmt.Errorf("invalid format found empty key")
					}
				}
				return err
			},
			Transform: func(any interface{}) interface{} {
				val := any.(string)
				// ignoring error since already validated
				envs, _ := godotenv.Unmarshal(val)
				return variable.Serialize(envs)
			},
		},
	)

	return questions
}

type validator struct {
	ctx     context.Context
	client  *client.Client
	orgName string
}

func (v *validator) validateProjectName(val interface{}) error {
	projectName := val.(string)
	resp, err := v.client.GetProject(v.ctx, &adminv1.GetProjectRequest{OrganizationName: v.orgName, Name: projectName})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.InvalidArgument { // todo :: change to not found
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
}

// ConnectCmd is the guided tour for connecting rill projects to rill cloud.
func ConnectCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Guided tour for connecting rill projects to rill cloud",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			warn := color.New(color.Bold).Add(color.FgYellow)
			info := color.New(color.Bold)
			success := color.New(color.Bold).Add(color.FgGreen)

			// log in if not logged in yet
			if cfg.AdminTokenDefault == "" {
				warn.Println("Looks like you are not authenticated with rill cloud!!")
				if !confirmPrompt("Do you want to authenticate with rill cloud?", true) {
					info.Println("GoodBye!!!")
					return nil
				}
				if err := loginPrompt(ctx, cfg); err != nil {
					exitWithFailure(err)
				}
			}

			// Create admin client
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			// select org
			orgs, defaultOrg, err := listOrganisations(client)
			if err != nil {
				return err
			}

			// todo :: should we add a create org prompt as well ??
			org := selectPrompt("Listing your orgs. Please select where do you want to deploy project?", orgs, defaultOrg)
			info.Printf("your current org is %v\n", org)
			// project path prompt
			projectPath := projectPathPrompt()
			githubURL, err := extractRemote(projectPath)
			if err != nil {
				if errors.Is(err, gitutil.ErrGitRemoteNotFound) || errors.Is(err, git.ErrRepositoryNotExists) {
					warn.Print(githubSetupMsg)
					return nil
				}
				return err
			}

			// Check for access to the Github URL
			ghRes, err := project.VerifyAccess(ctx, client, githubURL)
			if err != nil {
				return err
			}

			// We now have access to the Github repo
			opts, err := projectParamPrompt(ctx, client, org, githubURL, ghRes.DefaultBranch)
			if err != nil {
				return err
			}

			parsedVariables, err := variable.Parse(opts.Variables)
			if err != nil {
				return err
			}

			// Create the project (automatically deploys prod branch)
			projRes, err := client.CreateProject(ctx, &adminv1.CreateProjectRequest{
				OrganizationName:     org,
				Name:                 opts.Name,
				Description:          opts.Description,
				Region:               opts.region,
				ProductionOlapDriver: "duckdb",
				ProductionOlapDsn:    opts.dbDSN,
				ProductionSlots:      int64(opts.Slots),
				ProductionBranch:     opts.ProdBranch,
				Public:               opts.Public,
				GithubUrl:            githubURL,
				Variables:            parsedVariables,
			})
			if err != nil {
				return err
			}

			// Success!
			success.Printf("Created project %s/%s\n", cfg.Org, projRes.Project.Name)
			return nil
		},
	}

	return cmd
}

func loginPrompt(ctx context.Context, cfg *config.Config) error {
	// In production, the REST and gRPC endpoints are the same, but in development, they're served on different ports.
	// We plan to move to connect.build for gRPC, which will allow us to serve both on the same port in development as well.
	// Until we make that change, this is a convenient hack for local development (assumes gRPC on port 9090 and REST on port 8080).
	authURL := cfg.AdminURL
	if strings.Contains(authURL, "http://localhost:9090") {
		authURL = "http://localhost:8080"
	}

	authenticator, err := deviceauth.New(authURL)
	if err != nil {
		return err
	}

	deviceVerification, err := authenticator.VerifyDevice(ctx)
	if err != nil {
		return err
	}

	bold := color.New(color.Bold)
	bold.Printf("\nConfirmation Code: ")
	boldGreen := color.New(color.FgGreen).Add(color.Bold)
	boldGreen.Fprintln(color.Output, deviceVerification.UserCode)

	bold.Printf("\nOpen this URL in your browser to confirm the login: %s\n\n", deviceVerification.VerificationCompleteURL)

	// TODO :: we are asking to open browser and opening as well ??
	_ = browser.Open(deviceVerification.VerificationCompleteURL)

	res1, err := authenticator.GetAccessTokenForDevice(ctx, deviceVerification)
	if err != nil {
		return err
	}

	if err := dotrill.SetAccessToken(res1.AccessToken); err != nil {
		return err
	}
	boldGreen.Printf("\nLogged in successfully")
	return nil
}

func listOrganisations(c *client.Client) ([]string, string, error) {
	res, err := c.ListOrganizations(context.Background(), &adminv1.ListOrganizationsRequest{})
	if err != nil {
		return nil, "", err
	}

	defaultOrg, err := dotrill.GetDefaultOrg()
	if err != nil {
		return nil, "", err
	}

	defaultFound := false
	names := make([]string, len(res.Organizations))
	for i, org := range res.Organizations {
		if org.Name == defaultOrg {
			defaultFound = true
		}
		names[i] = org.Name
	}
	if !defaultFound {
		defaultOrg = ""
	}
	return names, defaultOrg, nil
}

func projectParamPrompt(ctx context.Context, c *client.Client, orgName, githubURL, prodBranch string) (*connectOptions, error) {
	opts := &connectOptions{}
	name := path.Base(githubURL)
	if err := survey.Ask(Questions(name, prodBranch, &validator{ctx: ctx, client: c, orgName: orgName}), opts); err != nil {
		return nil, err
	}

	return opts, nil
}

func projectPathPrompt() string {
	prompt := &survey.Input{
		Message: "What is your project path on local system?",
		Default: ".",
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
				_, err := os.Stat(path)
				return err
			},
		},
	}

	result := ""
	if err := survey.Ask(q, &result); err != nil {
		exitWithFailure(err)
	}
	return result
}

func selectPrompt(msg string, options []string, def string) string {
	prompt := &survey.Select{
		Message: msg,
		Options: options,
		Default: def,
		Description: func(value string, index int) string {
			if value == def {
				return "current default"
			}
			return ""
		},
	}
	result := def
	if err := survey.AskOne(prompt, &result); err != nil {
		exitWithFailure(err)
	}
	return result
}

func confirmPrompt(msg string, def bool) bool {
	prompt := &survey.Confirm{
		Message: msg,
		Default: def,
	}
	result := def
	if err := survey.AskOne(prompt, &result); err != nil {
		exitWithFailure(err)
	}
	return result
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

const githubSetupMsg = `No git remote was found.

Rill projects deploy continuously when you push changes to Github.
Therefore, your project must be on Github before you connect it to Rill.

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

6. Connect Rill to your repository

	rill connect

`
