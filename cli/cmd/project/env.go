package project

import (
	"context"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/cmd/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

// EnvCmd sets/rm variables for a project
func EnvCmd(cfg *config.Config) *cobra.Command {
	envCmd := &cobra.Command{
		Use:   "env",
		Short: "Manage env variables for a project",
	}
	envCmd.AddCommand(RmCmd(cfg))
	envCmd.AddCommand(SetCmd(cfg))
	return envCmd
}

// SetCmd is sub command for env. Sets the env variable for a project
func SetCmd(cfg *config.Config) *cobra.Command {
	setCmd := &cobra.Command{
		Use:   "set",
		Args:  cobra.ExactArgs(3),
		Short: "set env variable",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]
			key := args[1]
			value := args[2]
			client, err := client.New(cfg.AdminURL, cfg.AdminToken())
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := context.Background()
			resp, err := client.GetProject(ctx, &adminv1.GetProjectRequest{
				OrganizationName: cfg.Org(),
				Name:             projectName,
			})
			if err != nil {
				return err
			}

			proj := resp.Project
			if val, ok := proj.Envs[key]; ok && val == value {
				return nil
			}

			if proj.Envs == nil {
				proj.Envs = make(map[string]string)
			}
			proj.Envs[key] = value
			updatedProject, err := client.UpdateProject(context.Background(), &adminv1.UpdateProjectRequest{
				OrganizationName: cfg.Org(),
				Name:             projectName,
				Description:      proj.Description,
				Public:           proj.Public,
				ProductionBranch: proj.ProductionBranch,
				GithubUrl:        proj.GithubUrl,
				Envs:             proj.Envs,
			})
			if err != nil {
				return err
			}

			cmdutil.TextPrinter("Updated project \n")
			cmdutil.TablePrinter(toRow(updatedProject.Project))
			return nil
		},
	}
	return setCmd
}

// RmCmd is sub command for env. Sets the env variable for a project
func RmCmd(cfg *config.Config) *cobra.Command {
	rmCmd := &cobra.Command{
		Use:   "rm",
		Args:  cobra.ExactArgs(2),
		Short: "remove env variable",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]
			key := args[1]
			client, err := client.New(cfg.AdminURL, cfg.AdminToken())
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := context.Background()
			resp, err := client.GetProject(ctx, &adminv1.GetProjectRequest{
				OrganizationName: cfg.Org(),
				Name:             projectName,
			})
			if err != nil {
				return err
			}

			proj := resp.Project
			if _, ok := proj.Envs[key]; !ok {
				return nil
			}

			delete(proj.Envs, key)
			updatedProject, err := client.UpdateProject(context.Background(), &adminv1.UpdateProjectRequest{
				OrganizationName: cfg.Org(),
				Name:             projectName,
				Description:      proj.Description,
				Public:           proj.Public,
				ProductionBranch: proj.ProductionBranch,
				GithubUrl:        proj.GithubUrl,
				Envs:             proj.Envs,
			})
			if err != nil {
				return err
			}

			cmdutil.TextPrinter("Updated project \n")
			cmdutil.TablePrinter(toRow(updatedProject.Project))
			return nil
		},
	}
	return rmCmd
}
