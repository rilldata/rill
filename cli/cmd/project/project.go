package project

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func ProjectCmd(ch *cmdutil.Helper) *cobra.Command {
	projectCmd := &cobra.Command{
		Use:               "project",
		Short:             "Manage projects",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(ch), cmdutil.CheckOrganization(ch)),
	}

	projectCmd.PersistentFlags().StringVar(&ch.Org, "org", ch.Org, "Organization Name")
	projectCmd.AddCommand(ListCmd(ch))
	projectCmd.AddCommand(ShowCmd(ch))
	projectCmd.AddCommand(EditCmd(ch))
	projectCmd.AddCommand(RenameCmd(ch))
	projectCmd.AddCommand(HibernateCmd(ch))
	projectCmd.AddCommand(DeleteCmd(ch))
	projectCmd.AddCommand(StatusCmd(ch))
	projectCmd.AddCommand(PartitionsCmd(ch))
	projectCmd.AddCommand(LogsCmd(ch))
	projectCmd.AddCommand(DescribeCmd(ch))
	projectCmd.AddCommand(RefreshCmd(ch))
	projectCmd.AddCommand(JwtCmd(ch))
	projectCmd.AddCommand(CloneCmd(ch))
	projectCmd.AddCommand(GitPushCmd(ch))
	projectCmd.AddCommand(DeployCmd(ch))
	projectCmd.AddCommand(TablesCmd(ch))

	return projectCmd
}

func ProjectNames(ctx context.Context, ch *cmdutil.Helper) ([]string, error) {
	c, err := ch.Client()
	if err != nil {
		return nil, err
	}

	org := ch.Org

	resp, err := c.ListProjectsForOrganization(ctx, &adminv1.ListProjectsForOrganizationRequest{Org: org})
	if err != nil {
		return nil, err
	}

	if len(resp.Projects) == 0 {
		return nil, fmt.Errorf("no projects found for org %q", org)
	}

	var projNames []string
	for _, proj := range resp.Projects {
		projNames = append(projNames, proj.Name)
	}

	return projNames, nil
}
