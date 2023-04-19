package env

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/lensesio/tableprinter"
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/variable"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/compilers/rillv1beta"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/spf13/cobra"
)

func EnvCmd(cfg *config.Config) *cobra.Command {
	envCmd := &cobra.Command{
		Use:               "env",
		Short:             "Manage variables for a project",
		Hidden:            !cfg.IsDev(),
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(cfg), cmdutil.CheckOrganization(cfg)),
	}
	envCmd.AddCommand(ConfigureCmd(cfg))
	envCmd.AddCommand(SetCmd(cfg))
	envCmd.AddCommand(RmCmd(cfg))
	envCmd.AddCommand(ShowEnvCmd(cfg))
	return envCmd
}

func ConfigureCmd(cfg *config.Config) *cobra.Command {
	var projectPath, projectName string

	configureCommand := &cobra.Command{
		Use:   "env configure",
		Args:  cobra.ExactArgs(3),
		Short: "configures connector variables for all sources",
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
				githubURL, err := gitutil.ExtractGitRemote(projectPath)
				if err != nil {
					return err
				}

				// fetch project names for github url
				names, err := cmdutil.ProjectNames(ctx, client, cfg.Org, githubURL)
				if err != nil {
					return err
				}

				if len(names) == 1 {
					projectName = names[0]
				} else {
					// prompt for name from user
					projectName = cmdutil.SelectPrompt("select project to configure env", names, "")
				}
			}

			variables, err := VariablesFlow(ctx, projectPath)
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

			updateResp, err := client.UpdateProjectVariables(ctx, &adminv1.UpdateProjectVariablesRequest{
				OrganizationName: cfg.Org,
				Name:             projectName,
				Variables:        varResp.Variables,
			})
			if err != nil {
				return fmt.Errorf("failed to update variables %w", err)
			}

			cmdutil.SuccessPrinter("Updated project variables\n")
			tableprinter.PrintHeadList(os.Stdout, variable.Serialize(updateResp.Variables), "Project Variables")
			return nil
		},
	}

	configureCommand.Flags().SortFlags = false
	configureCommand.Flags().StringVar(&projectPath, "path", ".", "Project directory")
	configureCommand.Flags().StringVar(&projectName, "project", "", "")

	return configureCommand
}

// SetCmd is sub command for env. Sets the variable for a project
func SetCmd(cfg *config.Config) *cobra.Command {
	var projectName string
	setCmd := &cobra.Command{
		Use:   "set <key> <value> --project <project name>",
		Args:  cobra.ExactArgs(3),
		Short: "set variable",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]
			key := args[1]
			value := args[2]
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := cmd.Context()
			resp, err := client.GetProjectVariables(ctx, &adminv1.GetProjectVariablesRequest{
				OrganizationName: cfg.Org,
				Name:             projectName,
			})
			if err != nil {
				return err
			}

			if val, ok := resp.Variables[key]; ok && val == value {
				return nil
			}

			if resp.Variables == nil {
				resp.Variables = make(map[string]string)
			}
			resp.Variables[key] = value
			updateResp, err := client.UpdateProjectVariables(ctx, &adminv1.UpdateProjectVariablesRequest{
				OrganizationName: cfg.Org,
				Name:             projectName,
				Variables:        resp.Variables,
			})
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter("Updated project variables\n")
			tableprinter.PrintHeadList(os.Stdout, variable.Serialize(updateResp.Variables), "Project Variables")
			return nil
		},
	}

	setCmd.Flags().StringVar(&projectName, "project", "", "")
	return setCmd
}

// RmCmd is sub command for env. Removes the variable for a project
func RmCmd(cfg *config.Config) *cobra.Command {
	var projectName string
	rmCmd := &cobra.Command{
		Use:   "rm <key> --project <project name>",
		Args:  cobra.ExactArgs(2),
		Short: "remove variable",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]
			key := args[1]
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := cmd.Context()
			resp, err := client.GetProjectVariables(ctx, &adminv1.GetProjectVariablesRequest{
				OrganizationName: cfg.Org,
				Name:             projectName,
			})
			if err != nil {
				return err
			}

			if _, ok := resp.Variables[key]; !ok {
				return nil
			}

			delete(resp.Variables, key)
			update, err := client.UpdateProjectVariables(ctx, &adminv1.UpdateProjectVariablesRequest{
				OrganizationName: cfg.Org,
				Name:             projectName,
				Variables:        resp.Variables,
			})
			if err != nil {
				return err
			}

			cmdutil.SuccessPrinter("Updated project \n")
			tableprinter.PrintHeadList(os.Stdout, variable.Serialize(update.Variables), "Project Variables")
			return nil
		},
	}
	rmCmd.Flags().StringVar(&projectName, "project", "", "")
	return rmCmd
}

func ShowEnvCmd(cfg *config.Config) *cobra.Command {
	var projectName string
	showCmd := &cobra.Command{
		Use:   "show --project <project name>",
		Args:  cobra.ExactArgs(1),
		Short: "show variable for project",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			resp, err := client.GetProjectVariables(cmd.Context(), &adminv1.GetProjectVariablesRequest{
				OrganizationName: cfg.Org,
				Name:             projectName,
			})
			if err != nil {
				return err
			}

			tableprinter.PrintHeadList(os.Stdout, variable.Serialize(resp.Variables), "Project Variables")
			return nil
		},
	}
	showCmd.Flags().StringVar(&projectName, "project", "", "")
	return showCmd
}

func VariablesFlow(ctx context.Context, projectPath string) (map[string]string, error) {
	connectors, err := rillv1beta.ExtractConnectors(ctx, projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract connectors %w", err)
	}

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

	return vars, nil
}
