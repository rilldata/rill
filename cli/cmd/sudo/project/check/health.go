package check

import (
	"time"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/printer"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func HealthCmd(ch *cmdutil.Helper) *cobra.Command {
	cfg := ch.Config
	var email, domain string
	var pageSize uint32
	var pageToken string

	healthCmd := &cobra.Command{
		Use:   "health",
		Args:  cobra.NoArgs,
		Short: "Get project health",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := ch.Config

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if !cmd.Flags().Changed("org") && !cmd.Flags().Changed("email") && !cmd.Flags().Changed("domain") {
				res, err := client.SudoListProjectsHealth(ctx, &adminv1.SudoListProjectsHealthRequest{
					PageSize:  pageSize,
					PageToken: pageToken,
				})
				if err != nil {
					return err
				}

				err = listProjectsHealth(cmd, ch.Printer, res.GetProjects(), pageToken, res.GetNextPageToken(), pageSize)
				if err != nil {
					return err
				}
			} else {
				if cmd.Flags().Changed("org") {
					res, err := client.SudoListProjectsHealthForOrganization(ctx, &adminv1.SudoListProjectsHealthForOrganizationRequest{
						Organization: cfg.Org,
						PageSize:     pageSize,
						PageToken:    pageToken,
					})
					if err != nil {
						return err
					}

					err = listProjectsHealth(cmd, ch.Printer, res.GetProjects(), pageToken, res.GetNextPageToken(), pageSize)
					if err != nil {
						return err
					}
				}

				if cmd.Flags().Changed("email") {
					res, err := client.SudoListProjectsHealthForUser(ctx, &adminv1.SudoListProjectsHealthForUserRequest{
						Email:     email,
						PageSize:  pageSize,
						PageToken: pageToken,
					})
					if err != nil {
						return err
					}

					err = listProjectsHealth(cmd, ch.Printer, res.GetProjects(), pageToken, res.GetNextPageToken(), pageSize)
					if err != nil {
						return err
					}
				}

				if cmd.Flags().Changed("domain") {
					res, err := client.SudoListProjectsHealthForDomain(ctx, &adminv1.SudoListProjectsHealthForDomainRequest{
						Domain:    domain,
						PageSize:  pageSize,
						PageToken: pageToken,
					})
					if err != nil {
						return err
					}

					err = listProjectsHealth(cmd, ch.Printer, res.GetProjects(), pageToken, res.GetNextPageToken(), pageSize)
					if err != nil {
						return err
					}
				}
			}

			return nil
		},
	}

	healthCmd.Flags().StringVar(&cfg.Org, "org", cfg.Org, "Organization")
	healthCmd.Flags().StringVar(&email, "email", "", "Email of the user")
	healthCmd.Flags().StringVar(&domain, "domain", "", "Email domain")
	healthCmd.Flags().Uint32Var(&pageSize, "page-size", 50, "Number of projects to return per page")
	healthCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")

	return healthCmd
}

func listProjectsHealth(cmd *cobra.Command, p *printer.Printer, projects []*adminv1.ProjectHealth, pageToken, nextPageToken string, pageSize uint32) error {
	if len(projects) == 0 {
		p.PrintlnWarn("No projects found")
		return nil
	}

	// If page token is empty, user is running the command first time and we print separator
	if len(projects) > 0 && pageToken == "" {
		cmd.Println()
	}

	err := p.PrintResource(toTable(projects))
	if err != nil {
		return err
	}

	if nextPageToken != "" {
		cmd.Println()
		cmd.Printf("Next page token: %s\n", nextPageToken)
	}

	return nil
}

func toTable(projects []*adminv1.ProjectHealth) []*projectHealth {
	projs := make([]*projectHealth, 0, len(projects))

	for _, proj := range projects {
		projs = append(projs, toRow(proj))
	}

	return projs
}

func toRow(o *adminv1.ProjectHealth) *projectHealth {
	return &projectHealth{
		ID:            o.ProjectId,
		Name:          o.ProjectName,
		OrgID:         o.OrgId,
		DeploymentID:  o.DeploymentId,
		Status:        o.Status.String(),
		StatusMessage: o.StatusMessage,
		UpdatedOn:     o.DeploymentStatusTimestamp.AsTime().Local().Format(time.RFC3339),
	}
}

type projectHealth struct {
	ID            string `header:"id" json:"id"`
	Name          string `header:"name" json:"name"`
	OrgID         string `header:"orgId" json:"orgId"`
	DeploymentID  string `header:"deploymentId" json:"deploymentId"`
	Status        string `header:"status" json:"status"`
	StatusMessage string `header:"statusMessage" json:"statusMessage"`
	UpdatedOn     string `header:"updatedOn" json:"updatedOn"`
}
