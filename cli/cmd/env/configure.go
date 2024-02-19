package env

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/telemetry"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/compilers/rillv1beta"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/spf13/cobra"
)

func ConfigureCmd(ch *cmdutil.Helper) *cobra.Command {
	var projectPath, projectName, subPath string
	var redeploy bool

	configureCommand := &cobra.Command{
		Use:   "configure",
		Short: "Configures connector variables for all sources",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ch.Config
			if projectPath != "" {
				var err error
				projectPath, err = fileutil.ExpandHome(projectPath)
				if err != nil {
					return err
				}
			}

			fullProjectPath := projectPath
			if subPath != "" {
				fullProjectPath = filepath.Join(projectPath, subPath)
			}

			// Verify that the projectPath contains a Rill project
			if !rillv1beta.HasRillProject(fullProjectPath) {
				fullpath, err := filepath.Abs(fullProjectPath)
				if err != nil {
					return err
				}

				ch.Printer.PrintlnWarn(fmt.Sprintf("Directory at %q doesn't contain a valid Rill project.\n", fullpath))
				ch.Printer.PrintlnWarn("Run `rill env configure` from a Rill project directory or use `--path` to pass a project path.")
				return nil
			}

			ctx := cmd.Context()
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if projectName == "" {
				// no project name provided infer name from githubURL
				// Verify projectPath is a Git repo with remote on Github
				_, githubURL, err := gitutil.ExtractGitRemote(projectPath, "")
				if err != nil {
					return err
				}

				// fetch project names for github url
				names, err := cmdutil.ProjectNamesByGithubURL(ctx, client, cfg.Org, githubURL)
				if err != nil {
					return err
				}

				if len(names) == 1 {
					projectName = names[0]
				} else {
					// prompt for name from user
					projectName = cmdutil.SelectPrompt("Select project", names, "")
				}
			}

			variables, err := VariablesFlow(ctx, fullProjectPath, nil)
			if err != nil {
				return fmt.Errorf("failed to get variables: %w", err)
			}

			// get existing variables
			varResp, err := client.GetProjectVariables(ctx, &adminv1.GetProjectVariablesRequest{
				OrganizationName: cfg.Org,
				Name:             projectName,
			})
			if err != nil {
				return fmt.Errorf("failed to list existing variables %w", err)
			}

			if varResp.Variables == nil {
				varResp.Variables = make(map[string]string)
			}

			// update with new variables
			for key, value := range variables {
				varResp.Variables[key] = value
			}

			_, err = client.UpdateProjectVariables(ctx, &adminv1.UpdateProjectVariablesRequest{
				OrganizationName: cfg.Org,
				Name:             projectName,
				Variables:        varResp.Variables,
			})
			if err != nil {
				return fmt.Errorf("failed to update variables %w", err)
			}
			ch.Printer.PrintlnSuccess("Updated project variables")

			if !cmd.Flags().Changed("redeploy") {
				redeploy = cmdutil.ConfirmPrompt("Do you want to redeploy project", "", redeploy)
			}

			if redeploy {
				_, err = client.TriggerRedeploy(ctx, &adminv1.TriggerRedeployRequest{Organization: cfg.Org, Project: projectName})
				if err != nil {
					ch.Printer.PrintlnWarn("Redeploy trigger failed. Trigger redeploy again with `rill project reconcile --reset=true` if required.")
					return err
				}
				ch.Printer.PrintlnSuccess("Redeploy triggered successfully.")
			}
			return nil
		},
	}

	configureCommand.Flags().SortFlags = false
	configureCommand.Flags().StringVar(&projectPath, "path", ".", "Project directory")
	configureCommand.Flags().StringVar(&subPath, "subpath", "", "Project path to sub directory of a larger repository")
	configureCommand.Flags().StringVar(&projectName, "project", "", "")
	configureCommand.Flags().BoolVar(&redeploy, "redeploy", false, "Redeploy project")

	return configureCommand
}

func VariablesFlow(ctx context.Context, projectPath string, tel *telemetry.Telemetry) (map[string]string, error) {
	// Parse the project's connectors
	repo, instanceID, err := cmdutil.RepoForProjectPath(projectPath)
	if err != nil {
		return nil, err
	}
	parser, err := rillv1.Parse(ctx, repo, instanceID, "prod", "duckdb")
	if err != nil {
		return nil, fmt.Errorf("failed to parse project: %w", err)
	}
	connectors, err := parser.AnalyzeConnectors(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to extract connectors: %w", err)
	}

	// Exit early if all connectors can be used anonymously
	foundNotAnonymous := false
	for _, c := range connectors {
		if !c.AnonymousAccess {
			foundNotAnonymous = true
		}
	}
	if !foundNotAnonymous {
		return nil, nil
	}

	// Start the flow
	tel.Emit(telemetry.ActionDataAccessStart)
	fmt.Printf("Finish deploying your project by providing access to the connectors. Rill does not have access to the following data sources:\n\n")
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

	// Prompt for credentials
	variables := make(map[string]string)
	for _, c := range connectors {
		if c.AnonymousAccess {
			continue
		}
		if len(c.Spec.ConfigProperties) == 0 {
			continue
		}

		fmt.Printf("\nConnector %q requires credentials.\n", c.Name)
		if c.Spec.ServiceAccountDocs != "" {
			fmt.Printf("For instructions on how to create a service account, see: %s\n", c.Spec.ServiceAccountDocs)
		}
		fmt.Printf("\n")
		if c.Spec.Help != "" {
			fmt.Println(c.Spec.Help)
		}

		for i := range c.Spec.ConfigProperties {
			prop := c.Spec.ConfigProperties[i] // TODO: Move into range and turn into pointer

			key := fmt.Sprintf("connector.%s.%s", c.Name, prop.Key)
			msg := key
			if prop.Hint != "" {
				msg = fmt.Sprintf(msg+" (%s)", prop.Hint)
			}

			question := &survey.Question{}
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
				return nil, fmt.Errorf("variables prompt failed with error: %w", err)
			}

			if answer != "" {
				variables[key] = answer
			}
		}
	}

	// Continue with the flow
	tel.Emit(telemetry.ActionDataAccessSuccess)
	fmt.Println("")

	return variables, nil
}
