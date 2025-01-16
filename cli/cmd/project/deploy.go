package project

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/rilldata/rill/cli/cmd/org"
	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/rilldata/rill/cli/pkg/dotrillcloud"
	"github.com/rilldata/rill/cli/pkg/local"
	"github.com/rilldata/rill/cli/pkg/printer"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	nonSlugRegex      = regexp.MustCompile(`[^\w-]`)
	ErrInvalidProject = errors.New("invalid project")
)

type DeployOpts struct {
	GitPath     string
	SubPath     string
	RemoteName  string
	Name        string
	Description string
	Public      bool
	Provisioner string
	ProdVersion string
	ProdBranch  string
	Slots       int
}

func DeployCmd(ch *cmdutil.Helper) *cobra.Command {
	opts := &DeployOpts{}

	deployCmd := &cobra.Command{
		Use:   "deploy [<path>]",
		Short: "Deploy project to Rill Cloud by uploading the project files",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.GitPath = args[0]
			}
			return DeployWithUploadFlow(cmd.Context(), ch, opts)
		},
	}

	deployCmd.Flags().SortFlags = false
	deployCmd.Flags().StringVar(&opts.GitPath, "path", ".", "Path to project repository (default: current directory)") // This can also be a remote .git URL (undocumented feature)
	deployCmd.Flags().StringVar(&opts.SubPath, "subpath", "", "Relative path to project in the repository (for monorepos)")
	deployCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Org to deploy project in")
	deployCmd.Flags().StringVar(&opts.Name, "project", "", "Project name (default: Git repo name)")
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

func ValidateLocalProject(ch *cmdutil.Helper, gitPath, subPath string) (string, string, error) {
	var localGitPath string
	var err error
	if gitPath != "" {
		localGitPath, err = fileutil.ExpandHome(gitPath)
		if err != nil {
			return "", "", err
		}
	}
	localGitPath, err = filepath.Abs(localGitPath)
	if err != nil {
		return "", "", err
	}

	var localProjectPath string
	if subPath == "" {
		localProjectPath = localGitPath
	} else {
		localProjectPath = filepath.Join(localGitPath, subPath)
	}

	// Verify that localProjectPath contains a Rill project.
	if cmdutil.HasRillProject(localProjectPath) {
		return localGitPath, localProjectPath, nil
	}

	ch.PrintfWarn("Directory %q doesn't contain a valid Rill project.\n", localProjectPath)
	ch.PrintfWarn("Run `rill project deploy` from a Rill project directory or use `--path` to pass a project path.\n")
	ch.PrintfWarn("Run `rill start` to initialize a new Rill project.\n")
	return "", "", ErrInvalidProject
}

func DeployWithUploadFlow(ctx context.Context, ch *cmdutil.Helper, opts *DeployOpts) error {
	_, localProjectPath, err := ValidateLocalProject(ch, opts.GitPath, opts.SubPath)
	if err != nil {
		return err
	}
	// If no project name was provided, default to dir name
	if opts.Name == "" {
		opts.Name = filepath.Base(localProjectPath)
	}

	// Set a default org for the user if necessary
	// (If user is not in an org, we'll create one based on their user name later in the flow.)
	adminClient, err := ch.Client()
	if err != nil {
		return err
	}
	if ch.Org == "" {
		if err := org.SetDefaultOrg(ctx, ch); err != nil {
			return err
		}
	}

	// If no default org is set, it means the user is not in an org yet.
	// We create a default org based on the user name.
	if ch.Org == "" {
		user, err := adminClient.GetCurrentUser(ctx, &adminv1.GetCurrentUserRequest{})
		if err != nil {
			return err
		}
		// email can have other characters like . and + what to do ?
		username, _, _ := strings.Cut(user.User.Email, "@")
		username = nonSlugRegex.ReplaceAllString(username, "-")
		err = createOrgFlow(ctx, ch, username)
		if err != nil {
			return fmt.Errorf("org creation failed with error: %w", err)
		}
		ch.PrintfSuccess("Created org %q. Run `rill org edit` to change name if required.\n\n", ch.Org)
	} else {
		ch.PrintfBold("Using org %q.\n\n", ch.Org)
	}

	// get repo for current project
	repo, _, err := cmdutil.RepoForProjectPath(localProjectPath)
	if err != nil {
		return err
	}

	projResp, err := adminClient.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: ch.Org, Name: opts.Name})
	if err != nil {
		if st, ok := status.FromError(err); !ok || st.Code() != codes.NotFound {
			return err
		}
	}

	// check if the project with name already exists
	if projResp != nil {
		if projResp.Project.GithubUrl != "" {
			ch.PrintfError("Found existing project. But it is connected to a github repo.\nPush any changes to %q to deploy.\n", projResp.Project.GithubUrl)
			return nil
		}

		ch.Printer.Println("Found existing project. Starting re-upload.")
		assetID, err := cmdutil.UploadRepo(ctx, repo, ch, ch.Org, opts.Name)
		if err != nil {
			return err
		}
		printer.ColorGreenBold.Printf("All files uploaded successfully.\n\n")

		// Update the project
		// Silently ignores other flags like description etc which are handled with project update.
		res, err := adminClient.UpdateProject(ctx, &adminv1.UpdateProjectRequest{
			OrganizationName: ch.Org,
			Name:             opts.Name,
			ArchiveAssetId:   &assetID,
		})
		if err != nil {
			if s, ok := status.FromError(err); ok && s.Code() == codes.PermissionDenied {
				ch.PrintfError("You do not have the permissions needed to update a project in org %q. Please reach out to your Rill admin.\n", ch.Org)
				return nil
			}
			return fmt.Errorf("update project failed with error %w", err)
		}
		ch.Telemetry(ctx).RecordBehavioralLegacy(activity.BehavioralEventDeploySuccess)

		// Fetch vars from .env
		vars, err := local.ParseDotenv(ctx, localProjectPath)
		if err != nil {
			ch.PrintfWarn("Failed to parse .env: %v\n", err)
		} else {
			_, err = adminClient.UpdateProjectVariables(ctx, &adminv1.UpdateProjectVariablesRequest{
				Organization: ch.Org,
				Project:      opts.Name,
				Variables:    vars,
			})
			if err != nil {
				ch.PrintfWarn("Failed to upload .env: %v\n", err)
			}
		}

		// Success
		ch.PrintfSuccess("Updated project \"%s/%s\".\n\n", ch.Org, res.Project.Name)
		return nil
	}

	// create a tar archive of the project and upload it
	ch.Printer.Println("Starting upload.")
	assetID, err := cmdutil.UploadRepo(ctx, repo, ch, ch.Org, opts.Name)
	if err != nil {
		return err
	}
	printer.ColorGreenBold.Printf("All files uploaded successfully.\n\n")

	// Create the project
	res, err := adminClient.CreateProject(ctx, &adminv1.CreateProjectRequest{
		OrganizationName: ch.Org,
		Name:             opts.Name,
		Description:      opts.Description,
		Provisioner:      opts.Provisioner,
		ProdVersion:      opts.ProdVersion,
		ProdOlapDriver:   local.DefaultOLAPDriver,
		ProdOlapDsn:      local.DefaultOLAPDSN,
		ProdSlots:        int64(opts.Slots),
		Public:           opts.Public,
		ArchiveAssetId:   assetID,
	})
	if err != nil {
		if s, ok := status.FromError(err); ok && s.Code() == codes.PermissionDenied {
			ch.PrintfError("You do not have the permissions needed to create a project in org %q. Please reach out to your Rill admin.\n", ch.Org)
			return nil
		}
		return fmt.Errorf("create project failed with error %w", err)
	}

	err = dotrillcloud.SetAll(localProjectPath, ch.AdminURL(), &dotrillcloud.Config{
		ProjectID: res.Project.Id,
	})
	if err != nil {
		return err
	}

	// Success!
	ch.PrintfSuccess("Created project \"%s/%s\". Use `rill project rename` to change name if required.\n\n", ch.Org, res.Project.Name)

	// Upload .env
	vars, err := local.ParseDotenv(ctx, localProjectPath)
	if err != nil {
		ch.PrintfWarn("Failed to parse .env: %v\n", err)
	} else {
		_, err = adminClient.UpdateProjectVariables(ctx, &adminv1.UpdateProjectVariablesRequest{
			Organization: ch.Org,
			Project:      opts.Name,
			Variables:    vars,
		})
		if err != nil {
			ch.PrintfWarn("Failed to upload .env: %v\n", err)
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

func createOrgFlow(ctx context.Context, ch *cmdutil.Helper, defaultName string) error {
	c, err := ch.Client()
	if err != nil {
		return err
	}

	res, err := c.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{
		Name: defaultName,
	})
	if err != nil {
		if !errMsgContains(err, "an org with that name already exists") {
			return err
		}

		ch.PrintfWarn("Rill organizations are derived from the owner of your Github repository.\n")
		ch.PrintfWarn("The %q organization associated with your Github repository already exists.\n", defaultName)
		ch.PrintfWarn("Contact your Rill admin to be added to your org or create a new organization below.\n")

		name, err := orgNamePrompt(ctx, ch)
		if err != nil {
			return err
		}

		res, err = c.CreateOrganization(ctx, &adminv1.CreateOrganizationRequest{
			Name: name,
		})
		if err != nil {
			return err
		}
	}

	// Switching to the created org
	ch.Org = res.Organization.Name
	err = dotrill.SetDefaultOrg(ch.Org)
	if err != nil {
		return err
	}

	return nil
}

func orgNamePrompt(ctx context.Context, ch *cmdutil.Helper) (string, error) {
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

				exists, err := orgExists(ctx, ch, name)
				if err != nil {
					return fmt.Errorf("org name %q is already taken", name)
				}

				if exists {
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

func orgExists(ctx context.Context, ch *cmdutil.Helper, name string) (bool, error) {
	c, err := ch.Client()
	if err != nil {
		return false, err
	}

	_, err = c.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: name})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func projectExists(ctx context.Context, ch *cmdutil.Helper, orgName, projectName string) (bool, error) {
	c, err := ch.Client()
	if err != nil {
		return false, err
	}

	_, err = c.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: orgName, Name: projectName})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func errMsgContains(err error, msg string) bool {
	if st, ok := status.FromError(err); ok && st != nil {
		return strings.Contains(st.Message(), msg)
	}
	return false
}
