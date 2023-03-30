package connect

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
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

var errUserCancelledGitFlow = fmt.Errorf("user cancelled git flow")

type connectOptions struct {
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
			// Create admin client
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			// log in if not logged in
			if cfg.AdminTokenDefault == "" {
				warn.Println("You are not authenticated with rill cloud!!")
				if !cmdutil.ConfirmPrompt("Press enter to signup or login", true) {
					info.Println("GoodBye!!!")
					return nil
				}

				// NOTE : calling commands within commands has both pros and cons
				// PRO : No duplicated code
				// CON : Need to make sure that UX under sub command verifies with UX on this command
				loginCmd := auth.LoginCmd(cfg)
				// command on failure prints usage by default. Need to stop since user didn't run this command
				loginCmd.SilenceUsage = true
				loginCmd.SetContext(ctx)
				if err := loginCmd.RunE(loginCmd, nil); err != nil {
					exitWithFailure(err)
				}
			} else {
				// switch is already part of login cmd so running this only when user is already logged in
				defaultOrg, err := dotrill.GetDefaultOrg()
				if err != nil {
					return err
				}

				multipleOrgs, err := multipleOrgs(client)
				if err != nil {
					return fmt.Errorf("listing orgs failed with error %w", err)
				}

				if multipleOrgs {
					msg := fmt.Sprintf("This project will be deployed under %s. Press Y to confirm and N to select a different org", defaultOrg)
					if !cmdutil.ConfirmPrompt(msg, true) {
						switchCmd := org.SwitchCmd(cfg)
						switchCmd.SilenceUsage = true
						if err := switchCmd.ExecuteContext(ctx); err != nil {
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
				warn.Printf("\nLooks like current path %s doesn't have a valid rill project. Please select correct dir.\n", path)
				warn.Printf("In case there is no valid rill project present, Please use `rill init` to create an empty rill project\n\n")
				// project path prompt
				path = projectPathPrompt()
			}

			// verify project dir is a git repo with remote on github
			githubURL, err := extractRemote(path)
			if err != nil {
				if !errors.Is(err, gitutil.ErrGitRemoteNotFound) && !errors.Is(err, git.ErrRepositoryNotExists) {
					return err
				}

				warn.Println("\nYour project is not pushed to github")
				warn.Printf("You can exit cli, push project to github and connect later or continue to push project to github repo\n\n")

				if !cmdutil.ConfirmPrompt("Confirm to push repo to github", false) {
					info.Print(`Rill projects deploy continuously when you push changes to Github.
Therefore, your project must be on Github before you connect it to Rill.
Follow these steps to push your project to Github.`)
					info.Print(githubSetupMsg)
					return nil
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
				return fmt.Errorf("access to github repo failed with error %w", err)
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

			// fixing these for now
			opts.dbDriver = "duckdb"
			opts.slots = 2
			// Create the project (automatically deploys prod branch)
			projRes, err := client.CreateProject(ctx, &adminv1.CreateProjectRequest{
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
			success.Printf("Rill projects deploy continuously when you push changes to Github.\n\n")
			return nil
		},
	}

	return cmd
}

func projectParamPrompt(ctx context.Context, c *client.Client, orgName, githubURL, prodBranch string) (*connectOptions, error) {
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
						if st.Code() == codes.InvalidArgument { // todo :: change to not found in admin
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
			if !cmdutil.ConfirmPrompt("Do you want to create a git repository?", true) {
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

	repoStatus, err := w.Status()
	if err != nil {
		return err
	}

	if !repoStatus.IsClean() {
		if hasUncommittedFiles(repoStatus) {
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
		}

		msg := fmt.Sprintf("Do you want to commit modified files?\n%s\n", repoStatus.String())
		if !cmdutil.ConfirmPrompt(msg, true) {
			return errUserCancelledGitFlow
		}

		if _, err := w.Commit("files auto committed by rill cli", &git.CommitOptions{}); err != nil {
			return err
		}
	}

	if !cmdutil.ConfirmPrompt("Do you want to push committed files?", true) {
		return errUserCancelledGitFlow
	}

	ghClient, err := gh.GQLClient(nil)
	if err != nil {
		return err
	}
	user, orgs, err := currentLoginNameAndOrgs(ghClient)
	if err != nil {
		return err
	}

	name := cmdutil.InputPrompt("Repository name", filepath.Base(dir))
	orgs = append(orgs, user)
	owner := cmdutil.SelectPrompt("Repository owner", orgs, user)

	stdout, _, err := gh.Exec("repo", "create", fmt.Sprintf("%s/%s", owner, name), fmt.Sprintf("--source=%s", dir), "--private", "--push", "--remote=origin")
	if err != nil {
		return err
	}
	info.Printf("created remote repo %s\n", stdout.String())
	return nil
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

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func currentLoginNameAndOrgs(ghClient api.GQLClient) (string, []string, error) {
	type organization struct {
		Login string
	}

	var query struct {
		Viewer struct {
			Login         string
			Organizations struct {
				Nodes []organization
			} `graphql:"organizations(first: 100)"`
		}
	}
	err := ghClient.Query("UserCurrent", &query, nil)
	if err != nil {
		return "", nil, err
	}
	orgNames := []string{}
	for _, org := range query.Viewer.Organizations.Nodes {
		orgNames = append(orgNames, org.Login)
	}
	return query.Viewer.Login, orgNames, err
}

func hasUncommittedFiles(s git.Status) bool {
	for _, status := range s {
		if status.Worktree != git.Unmodified {
			return true
		}
	}

	return false
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
