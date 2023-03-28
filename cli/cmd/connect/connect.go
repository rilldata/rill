package connect

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/cli/go-gh"
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

var errUserCancelledGitFlow = fmt.Errorf("user cancelled git flow")

type connectOptions struct {
	Name        string
	Description string
	ProdBranch  string
	Public      bool
	Variables   []string

	// projectPath string
	region string
	// dbDriver    string should we take input ??
	dbDSN string
}

// ConnectCmd is the guided tour for connecting rill projects to rill cloud.
func ConnectCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Guided tour for connecting rill projects to rill cloud",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			warn := color.New(color.Bold).Add(color.FgYellow)
			info := color.New(color.Bold).Add(color.FgWhite)
			success := color.New(color.Bold).Add(color.FgGreen)

			// log in if not logged in
			if cfg.AdminTokenDefault == "" {
				warn.Println("Looks like you are not authenticated with rill cloud!!")
				if !confirmPrompt("Do you want to authenticate with rill cloud?", true) {
					info.Println("GoodBye!!!")
					return nil
				}
				if err := loginPrompt(ctx, cfg); err != nil {
					exitWithFailure(err)
				}
				success.Println("you are now authenticated")
			}

			// Create admin client
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			// select org
			org := ""
			if confirmPrompt("Do you want to create a new org for this project?", false) {
				if err := createOrgPrompt(ctx, cfg, client); err != nil {
					return err
				}
				org, err = dotrill.GetDefaultOrg()
				if err != nil {
					return err
				}
			} else {
				// select org
				orgs, defaultOrg, err := listOrganisations(client)
				if err != nil {
					return err
				}

				if len(orgs) == 0 {
					org = orgs[0]
				} else {
					org = selectPrompt("Listing your orgs. Please select where do you want to deploy project?", orgs, defaultOrg)
				}
			}
			info.Printf("your current org is %v\n", org)

			// verify current directory has rill project
			path, err := os.Getwd()
			if err != nil {
				return err
			}

			githubURL := ""
			if !hasRillProject(path) {
				warn.Printf("\nLooks like current path %s doesn't have a valid rill project. Please select correct dir.\n", path)
				warn.Printf("In case there is no valid rill project present, Please use `rill init` to create an empty rill project\n\n")
				// project path prompt
				path = projectPathPrompt()
			}

			// verify project dir is a git repo with remote on github
			githubURL, err = extractRemote(path)
			if err != nil {
				if !errors.Is(err, gitutil.ErrGitRemoteNotFound) && !errors.Is(err, git.ErrRepositoryNotExists) {
					return err
				}

				warn.Println("\nYour project is not pushed to github")
				warn.Printf("You can exit cli, push project to github and connect later or continue to push project to github repo\n\n")

				if !confirmPrompt("Confirm to push repo to github", false) {
					return errUserCancelledGitFlow
				}

				if !commandExists("gh") {
					warn.Println("\nYou do not have github cli installed on your system. Please install github cli (instructions : https://cli.github.com/manual/installation) to push repo via rill connect or follow instructions below")
					info.Print(githubSetupMsg)
					return nil
				}

				if err := repoCreatePrompt(path, info); err != nil {
					return err
				}
				githubURL, err = extractRemote(path)
				if err != nil {
					return err
				}
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
				ProductionSlots:      2,
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
			success.Printf("Rill projects deploy continuously when you push changes to Github.\n\n")
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

func createOrgPrompt(ctx context.Context, cfg *config.Config, c *client.Client) error {
	// the questions to ask
	qs := []*survey.Question{
		{
			Name:   "name",
			Prompt: &survey.Input{Message: "Please specify a org name"},
			Validate: func(any interface{}) error {
				org := any.(string)
				resp, err := c.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: org})
				if err != nil {
					if st, ok := status.FromError(err); ok {
						if st.Code() == codes.InvalidArgument { // todo :: change to not found in admin server
							return nil
						}
					}
					return err
				}
				if resp.Organization.Name == org {
					// this should always be true but adding this check from completeness POV
					return fmt.Errorf("project with name %v already exists in the org", org)
				}
				return nil
			},
		},
		{
			Name:     "description",
			Prompt:   &survey.Input{Message: "Please specify a org description", Default: ""},
			Validate: survey.Required,
		},
	}

	req := &adminv1.CreateOrganizationRequest{}
	// perform the questions
	err := survey.Ask(qs, req)
	if err != nil {
		return err
	}

	org, err := c.CreateOrganization(context.Background(), req)
	if err != nil {
		return err
	}

	// Switching to the created org
	return dotrill.SetDefaultOrg(org.Organization.Name)
}

func projectParamPrompt(ctx context.Context, c *client.Client, orgName, githubURL, prodBranch string) (*connectOptions, error) {
	v := validator{
		ctx:     ctx,
		client:  c,
		orgName: orgName,
	}
	projectName := path.Base(githubURL)

	questions := []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "What is the project name?",
				Default: projectName,
			},
			Validate: v.validateProjectName,
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

	opts := &connectOptions{}
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

func repoCreatePrompt(dir string, info *color.Color) error {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		if errors.Is(err, git.ErrRepositoryNotExists) {
			if !confirmPrompt("Do you want to create a git repository?", true) {
				return errUserCancelledGitFlow
			}
			repo, err = git.PlainInit(dir, false)
			if err != nil {
				return err
			}
		}
	}

	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	fileGlob := ""
	fileGlobInput := &survey.Input{
		Message: "What files do you want to commit?",
		Default: ".",
	}
	if err := survey.AskOne(fileGlobInput, &fileGlob); err != nil {
		return err
	}
	if err := w.AddGlob(fileGlob); err != nil {
		return err
	}

	// todo :: show staging area
	if !confirmPrompt("Do you want to commit added files?", true) {
		return errUserCancelledGitFlow
	}

	if _, err := w.Commit("files auto committed by rill cli", &git.CommitOptions{}); err != nil {
		return err
	}

	if !confirmPrompt("Do you want to push committed files?", true) {
		return errUserCancelledGitFlow
	}

	stdout, _, err := gh.Exec("repo", "create", "--source=../test", "--public", "--push")
	if err != nil {
		return err
	}
	info.Printf("created remote repo %s\n", stdout.String())
	return nil
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

func hasRillProject(dir string) bool {
	_, err := os.Open(filepath.Join(dir, "rill.yaml"))
	return err == nil
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
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

const (
	githubSetupMsg = `
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
