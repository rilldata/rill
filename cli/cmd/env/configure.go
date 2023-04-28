package env

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/telemetry"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/compilers/rillv1beta"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/spf13/cobra"
)

func ConfigureCmd(cfg *config.Config) *cobra.Command {
	var projectPath, projectName string
	var redploy bool

	configureCommand := &cobra.Command{
		Use:   "configure",
		Short: "Configures connector variables for all sources",
		RunE: func(cmd *cobra.Command, args []string) error {
			warn := color.New(color.Bold).Add(color.FgYellow)
			if projectPath != "" {
				var err error
				projectPath, err = fileutil.ExpandHome(projectPath)
				if err != nil {
					return err
				}
			}

			// Verify that the projectPath contains a Rill project
			if !rillv1beta.HasRillProject(projectPath) {
				fullpath, err := filepath.Abs(projectPath)
				if err != nil {
					return err
				}

				warn.Printf("Directory at %q doesn't contain a valid Rill project.\n\n", fullpath)
				warn.Printf("Run \"rill env configure\" from a Rill project directory or use \"--path\" to pass a project path.\n")
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

			variables, err := VariablesFlow(ctx, projectPath, nil)
			if err != nil {
				return fmt.Errorf("failed to get variables %w", err)
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
			cmdutil.SuccessPrinter("Updated project variables")

			if !cmd.Flags().Changed("redeploy") {
				redploy = cmdutil.ConfirmPrompt("Do you want to redeploy project", "", redploy)
			}

			if redploy {
				project, err := client.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: cfg.Org, Name: projectName})
				if err != nil {
					return err
				}

				_, err = client.TriggerRedeploy(ctx, &adminv1.TriggerRedeployRequest{DeploymentId: project.ProdDeployment.Id})
				if err != nil {
					warn.Printf("Redeploy trigger failed. Trigger redeploy again with `rill project reconcile --reset=true` if required.")
					return err
				}
				cmdutil.SuccessPrinter("Redeploy triggered successfully.")
			}
			return nil
		},
	}

	configureCommand.Flags().SortFlags = false
	configureCommand.Flags().StringVar(&projectPath, "path", ".", "Project directory")
	configureCommand.Flags().StringVar(&projectName, "project", "", "")
	configureCommand.Flags().BoolVar(&redploy, "redeploy", false, "Redeploy project")

	return configureCommand
}

func VariablesFlow(ctx context.Context, projectPath string, tel *telemetry.Telemetry) (map[string]string, error) {
	connectors, err := rillv1beta.ExtractConnectors(ctx, projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract connectors %w", err)
	}

	tel.Emit(telemetry.ActionDataAccessStart)

	vars := make(map[string]string)
	for _, c := range connectors {
		if c.AnonymousAccess {
			// ignore asking for credentials if external source can be access anonymously
			continue
		}
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

	tel.Emit(telemetry.ActionDataAccessSuccess)

	return vars, nil
}
