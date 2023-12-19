package check

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
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
		Args:  cobra.ExactArgs(2),
		Short: "Get project details",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := ch.Config

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			var reqErr error
			var res ProjectHealthResponse

			if !cmd.Flags().Changed("org") && !cmd.Flags().Changed("email") && !cmd.Flags().Changed("domain") {
				res, err = client.SudoListProjectsHealth(ctx, &adminv1.SudoListProjectsHealthRequest{
					PageSize:  pageSize,
					PageToken: pageToken,
				})
			} else {
				if cmd.Flags().Changed("org") {
					res, err = client.SudoListProjectsHealthForOrganization(ctx, &adminv1.SudoListProjectsHealthForOrganizationRequest{
						Organization: cfg.Org,
						PageSize:     pageSize,
						PageToken:    pageToken,
					})
					// reqErr = handleResponse(res, ch, err, cmd, pageToken)
				}

				if cmd.Flags().Changed("email") && reqErr == nil {
					res, err = client.SudoListProjectsHealthForUser(ctx, &adminv1.SudoListProjectsHealthForUserRequest{
						Email:     email,
						PageSize:  pageSize,
						PageToken: pageToken,
					})
					// reqErr = handleResponse(res, ch, err, cmd, pageToken)
				}

				if cmd.Flags().Changed("domain") && reqErr == nil {
					res, err = client.SudoListProjectsHealthForDomain(ctx, &adminv1.SudoListProjectsHealthForDomainRequest{
						Domain:    domain,
						PageSize:  pageSize,
						PageToken: pageToken,
					})
					// reqErr = handleResponse(res, ch, err, cmd, pageToken)
				}
			}

			reqErr = handleResponse(res, ch, err, cmd, pageToken)

			return reqErr
		},
	}

	healthCmd.Flags().StringVar(&cfg.Org, "org", cfg.Org, "Organization")
	healthCmd.Flags().StringVar(&email, "email", "", "Email of the user")
	healthCmd.Flags().StringVar(&domain, "domain", "", "Email domain")
	healthCmd.Flags().Uint32Var(&pageSize, "page-size", 50, "Number of projects to return per page")
	healthCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")

	return healthCmd
}

type ProjectHealthResponse interface {
	GetProjects() []*adminv1.ProjectHealth
	GetNextPageToken() string
}

func handleResponse(res ProjectHealthResponse, ch *cmdutil.Helper, err error, cmd *cobra.Command, pageToken string) error {
	if err != nil {
		return err
	}
	projects := res.GetProjects()
	nextPageToken := res.GetNextPageToken()
	if len(projects) > 0 && pageToken == "" {
		cmd.Println()
	}
	err = ch.Printer.PrintResource(projects)
	if err != nil {
		return err
	}
	if nextPageToken != "" {
		cmd.Println()
		cmd.Printf("Next page token: inv%s\n", nextPageToken)
	}
	return nil
}
